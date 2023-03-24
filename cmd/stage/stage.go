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
	cmdMap    = make(map[string]int)
	app       = kingpin.New("stage", "stage commands or files for remote systems")
	file      = app.Command("file", "stage a file")
	viewFile  = file.Command("view", "view current file stage")
	serveFile = file.Command("serve", "serve a specified file")
	deleteF   = file.Command("clear", "remove a file from the stage")
	filename  = serveFile.Arg("filename", "name of file to stage").Required().String()
	outfile   = serveFile.Arg("outfile", "name of file to write to remote disk").Default("").String()
	cmd       = app.Command("cmd", "stage a command")
	exec      = cmd.Command("exec", "execute a shell command on target")
	viewCmd   = cmd.Command("view", "view current command stage")
	deleteC   = cmd.Command("clear", "remove a command from the stage")
	cmdstring = exec.Arg("command", "command to execute").Required().String()
	forwho    = app.Flag("for", "name of target to receive command").String()
	unsafe    = serveFile.Flag("unsafe", "allows un-encrypted files to be written to disk remotely").Bool()
)

func main() {
	db, err := datastore.New()
	if err != nil {
		log.Fatal(err)
	}
	cmdMap["exec"] = 1
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case viewFile.FullCommand():
		content, err := datastore.ShowStage(filestage)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(content)
		return
	case serveFile.FullCommand():
		bytes, err := os.ReadFile(*filename)
		if err != nil {
			log.Fatal(err)
		}
		var uuid string
		if len(*forwho) != 0 {
			uuid, err = datastore.GetUUIDByName(*forwho)
			if err != nil {
				log.Fatal(err)
			}
			aeskey, X1, X2 := db.GetKeysByUUID(uuid)
			bytes, err = kes.EncryptObject(bytes, aeskey, X1, X2)
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(*forwho) == 0 && !*unsafe {
			log.Fatal("'for' flag was not set, meaning file will be unencrypted on disk. set --unsafe to allow.")
		}
		if len(*forwho) == 0 && *unsafe {
			log.Warn("'for' flag was not set, meaning file will be unencrypted on disk.")
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
	case exec.FullCommand():
		handleExec(db)
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

func handleExec(db *datastore.Kdb) {
	cmdlet := cmdMap["exec"]
	cmdObj := &models.KarObjectCmd{}
	cmdObj.Cmd = cmdlet
	cmdObj.Args = append(cmdObj.Args, models.MultiType{StrValue: *cmdstring})
	bytes, err := json.Marshal(cmdObj)
	if err != nil {
		log.Fatal(err)
	}
	index := 0
	for i, n := range os.Args {
		if n == "exec" {
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
	uuid, err := datastore.GetUUIDByName(*forwho)
	var encObj []byte
	if err != nil || len(uuid) == 0 {
		encObj = bytes
	} else {
		aeskey, x1, x2 := db.GetKeysByUUID(uuid)
		encObj, err = kes.EncryptObject(bytes, aeskey, x1, x2)
		if err != nil {
			log.Fatal(err)
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

package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/pygrum/karmine/config"
	log "github.com/sirupsen/logrus"
)

var (
	app     = kingpin.New("deploy", "create karma folder (for zipping) and sharing. HIDE KARMA+DATA+DLL FILES BEFORE COMPRESSION & DISTRIBUTION")
	outfile = app.Flag("outfile", "name of resulting folder").Required().String()
	data    = app.Flag("data_path", "path to the dummy data to load after execution (e.g. pdf file)").Short('d').Required().String()
	path    = app.Arg("pe_path", "path to karma instance to deploy").Required().String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	conf, err := config.GetFullConfig()
	if err != nil {
		log.Fatal(err)
	}
	scriptPath := conf.SrcPath + "/deploy/deploy.sh"
	cmd := exec.Command(
		scriptPath,
		*path,
		*outfile,
		filepath.Base(*data),
		*data,
	)
	var buf bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &buf)

	cmd.Stdout = mw
	cmd.Stderr = mw
	if err = cmd.Run(); err != nil {
		log.Fatalf("could not build new instance: %v", err)
	}
}

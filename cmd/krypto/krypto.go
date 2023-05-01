package main

import (
	"os"

	"github.com/pygrum/karmine/krypto/kes"
)

// This is a helper cmdlet to encrypt an embedded payload before packing into another file.
// It should ideally always run without errors and never be executed directly by a user.

func main() {
	target := os.Args[1]
	aeskey := os.Args[2]
	x1 := os.Args[3]
	x2 := os.Args[4]

	bytes, err := os.ReadFile(target)
	if err != nil {
		panic(err)
	}

	encBytes, err := kes.EncryptObject(bytes, aeskey, x1, x2)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(target, encBytes, 0600); err != nil {
		panic(err)
	}
}

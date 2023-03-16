// The karmine/cryptor package implements a custom obfuscation algorithm for hiding strings, and strings only.
//If you come across this package, I'd advise against using it. There are much better encryption alternatives out there.
package kryptor

import (
	"crypto/rand"
	"encoding/base64"
)

// De-obfuscates a provided karmine string, and returns the result.
// Neither the result nor provided string are verified for validity.
func Decrypt(str, encKey1, encKey2 string) ([]byte, error) {
	bytes, err := base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		return bytes, err
	}
	encKey1New, err := base64.RawStdEncoding.DecodeString(encKey1)
	if err != nil {
		return bytes, err
	}
	encKey2New, err := base64.RawStdEncoding.DecodeString(encKey2)
	if err != nil {
		return bytes, err
	}
	for i := 0; i < len(bytes); i++ {
		bytes[i] ^= encKey1New[i%len(encKey1New)] ^ encKey2New[i%len(encKey2New)]
	}
	return bytes, nil
}

type decryptResult struct {
	Value []byte
	Error error
}

func DecryptAll(encKey1, encKey2 []byte, strArray ...string) (d []decryptResult) {
	for _, data := range strArray {
		v, err := Decrypt(data, string(encKey1), string(encKey2))
		d = append(d, decryptResult{
			Value: v,
			Error: err,
		})
	}
	return
}

// Obfuscates a provided string and returns the result.
// This function must not be imported anywhere in production.
func Encrypt(bytes []byte, x1, x2 string) (string, error) {
	encKey1, err := base64.RawStdEncoding.DecodeString(x1)
	if err != nil {
		return "", err
	}
	encKey2, err := base64.RawStdEncoding.DecodeString(x2)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(bytes); i++ {
		bytes[i] ^= encKey1[i%len(encKey1)] ^ encKey2[i%len(encKey2)]
	}
	dst0 := make([]byte, base64.RawStdEncoding.EncodedLen(len(bytes)))
	base64.RawStdEncoding.Encode(dst0, bytes)

	return string(dst0), nil
}

func GetXORKeyPair(length int) (string, string) {
	encKey1 := keyGen(length)
	encKey2 := keyGen(length)
	dst1 := make([]byte, base64.RawStdEncoding.EncodedLen(len(encKey1)))
	base64.RawStdEncoding.Encode(dst1, encKey1)
	dst2 := make([]byte, base64.RawStdEncoding.EncodedLen(len(encKey2)))
	base64.RawStdEncoding.Encode(dst2, encKey2)
	return string(dst1), string(dst2)
}

func keyGen(length int) []byte {
	b := make([]byte, length)
	rand.Read(b)
	return b
}

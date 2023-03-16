package kes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/pygrum/karmine/krypto/kryptor"
)

func EncryptObject(obj []byte, aKey, kX1, kX2 string) ([]byte, error) {
	aesKey, err := kryptor.Decrypt(aKey, kX1, kX2)
	if err != nil {
		return nil, err
	}
	encObj, err := encrypt(obj, aesKey)
	if err != nil {
		return encObj, err
	}
	return encObj, nil
	// encrypt object with KES (AES) engine, provided the aesKey, and return
}

func encrypt(obj, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create new cipher: %v", err)
	}

	cipherText := make([]byte, aes.BlockSize+len(obj))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("could not encrypt: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], obj)

	return cipherText, nil
}

func DecryptObject(obj []byte, aKey, kX1, kX2 string) ([]byte, error) {
	aesKey, err := kryptor.Decrypt(aKey, kX1, kX2)
	if err != nil {
		return nil, err
	}
	plainObj, err := decrypt(obj, aesKey)
	if err != nil {
		return plainObj, err
	}
	return plainObj, nil
	// decrypt object with KES (AES) engine, provided the aesKey, and return
}

func decrypt(obj, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create new cipher: %v", err)
	}

	if len(obj) < aes.BlockSize {
		return nil, fmt.Errorf("invalid obj block size")
	}

	iv := obj[:aes.BlockSize]
	obj = obj[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(obj, obj)

	return obj, nil
}

// get new base32 encoded aes key
func NewKey() []byte {
	b := make([]byte, 32)
	rand.Read(b)
	return b
}

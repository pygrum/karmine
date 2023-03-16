package kryptor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type DecryptTest struct {
	Input         string
	Key1          string
	Key2          string
	ExpectedValue string
	ExpectedError error
}

var (
	decryptTests = []DecryptTest{
		{
			Input:         "",
			Key1:          "",
			Key2:          "",
			ExpectedValue: "",
			ExpectedError: nil,
		},
	}
	decryptAllTestKey1 []byte
	decryptAllTestKey2 []byte
	decryptAllTests    []string
)

func TestDecrypt(t *testing.T) {
	for _, test := range decryptTests {
		v, err := Decrypt(test.Input, test.Key1, test.Key2)
		assert.Equal(t, test.ExpectedValue, string(v), "values must match")
		assert.Equal(t, test.ExpectedError, err, "errors must match")
	}
}

func TestDecryptAll(t *testing.T) {
	for _, test := range decryptTests {
		decryptAllTests = append(decryptAllTests, test.Input)
	}

	values := DecryptAll(decryptAllTestKey1, decryptAllTestKey2, decryptAllTests...)
	for i, v := range values {
		assert.Equal(t, decryptTests[i].ExpectedValue, string(v.Value), "values must match")
		assert.Equal(t, decryptTests[i].ExpectedError, v.Error, "errors must match")
	}
}

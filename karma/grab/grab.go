//go:build !windows
// +build !windows

// This package only exists for easy cross platform builds. It is irrelevant for platforms other than windows right now.

package grab

const (
	WinChromePwDBPath = ""
)

func NewUser() (*ChromeUser, error) { return nil, nil }

type ChromeUser struct {
	Username string
	OS       string
	HomeDir  string
	EncKey   []byte
}

func (cu *ChromeUser) GetChromeKey() error {
	return nil
}

func (cu *ChromeUser) DecryptDetails(pass string) (string, error) {
	return "", nil
}

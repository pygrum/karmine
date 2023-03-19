//go:build !windows
// +build !windows

package hide

import "os"

func HideF(filename string) error {
	err := os.Rename(filename, "."+filename)
	if err != nil {
		return err
	}
	return nil
}

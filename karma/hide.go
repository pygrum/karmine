//go:build !windows

package main

import "os"

func HideF(filename string) error {
	err := os.Rename(filename, "."+filename)
	if err != nil {
		return err
	}
	return nil
}

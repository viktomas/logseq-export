package main

import (
	"io"

	"github.com/spf13/afero"
)

func copy(appFS afero.Fs, src, dest string) error {
	srcFile, err := appFS.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := appFS.Create(dest) // creates if file doesn't exist
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

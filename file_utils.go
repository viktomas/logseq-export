package main

import (
	"io"
	"os"
)

func readFileToString(src string) (string, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()
	bytes, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func writeStringToFile(dest string, content string) error {
	err := os.WriteFile(dest, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func copy(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest) // creates if file doesn't exist
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

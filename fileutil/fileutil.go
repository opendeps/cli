package fileutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func GenerateStagingDir() string {
	tempDir, err := ioutil.TempDir(os.TempDir(), "mock")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created staging dir: %v\n", tempDir)
	return tempDir
}

func CopyFile(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error : %s", err.Error())
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest) // creates if file doesn't exist
	if err != nil {
		return fmt.Errorf("error : %s", err.Error())
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile) // check first var for number of bytes copied
	if err != nil {
		return fmt.Errorf("error : %s", err.Error())
	}

	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("error : %s", err.Error())
	}
	return nil
}

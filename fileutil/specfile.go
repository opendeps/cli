package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func getDefaultSearchFilenames() []string {
	return []string{
		"opendeps.yaml",
		"opendeps.yml",
	}
}

// FindSpecFile searches args for a path to an OpenDeps manifest file.
// If args is empty, the working directory is used and well-known
// manifest filenames are searched.
// If args is not empty, the path is made absolute, followed by
// a search for well-known filenames, or a fully qualified
// file path if specified.
func FindSpecFile(args []string) (specFile string, err error) {
	if len(args) == 0 {
		wd, _ := os.Getwd()
		return findSpecInDir(wd)
	} else {
		absPath, _ := filepath.Abs(args[0])
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("no opendeps manifest found at %v", absPath)
			} else {
				return "", fmt.Errorf("unable to stat %v: %v", absPath, err)
			}
		}
		if fileInfo.IsDir() {
			return findSpecInDir(absPath)
		}
		return absPath, nil
	}
}

func findSpecInDir(dir string) (specFile string, err error) {
	for _, defaultSearchFilename := range getDefaultSearchFilenames() {
		searchFilePath := filepath.Join(dir, defaultSearchFilename)
		if _, err := os.Stat(searchFilePath); err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return "", fmt.Errorf("unable to stat %v: %v", searchFilePath, err)
			}
		}
		return searchFilePath, nil
	}
	return "", fmt.Errorf("no opendeps manifest found at %v", dir)
}

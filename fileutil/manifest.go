/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

// FindManifestFile searches args for a path to an OpenDeps manifest file.
// If args is empty, the working directory is used and well-known
// manifest filenames are searched.
// If args is not empty, the path is made absolute, followed by
// a search for well-known filenames, or a fully qualified
// file path if specified.
func FindManifestFile(args []string) (manifestPath string, err error) {
	if len(args) == 0 {
		wd, _ := os.Getwd()
		return findManifestInDir(wd)
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
			return findManifestInDir(absPath)
		}
		return absPath, nil
	}
}

func findManifestInDir(dir string) (manifestPath string, err error) {
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

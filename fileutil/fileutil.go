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
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func GenerateStagingDir() string {
	tempDir, err := ioutil.TempDir(os.TempDir(), "mock")
	if err != nil {
		log.Fatal(err)
	}
	logrus.Debugf("created staging dir: %v\n", tempDir)
	return tempDir
}

// CopyContent retrieves the content of a file, based on
// its scheme, such as file:// or http://
// and writes it to destFile.
func CopyContent(source string, destFile string) error {
	logrus.Infof("copying from %v", source)
	logrus.Tracef("copying from %v to %v", source, destFile)

	content, err := ReadContent(source)
	if err != nil {
		return err
	}
	defer content.Close()

	err = writeToFile(content, destFile)
	if err != nil {
		return fmt.Errorf("failed to write to: %v: %v\n", destFile, err)
	}
	return nil
}

// ReadContent opens a ReadCloser for a file, based on
// its scheme, such as file:// or http://
func ReadContent(source string) (io.ReadCloser, error) {
	var content io.ReadCloser
	var err error
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		content, err = fetchHttp(source)
	} else {
		content, err = fetchFile(source)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from: %v: %v\n", source, err)
	}
	return content, err
}

func fetchHttp(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from URL [%v]: %v\n", url, err)
	} else {
		logrus.Debugf("%v returned: %s\n", url, resp.Status)
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("failed to read from URL [%v]: %s\n", url, resp.Status)
		}
		return resp.Body, nil
	}
}

func writeToFile(source io.Reader, dest string) error {
	destFile, err := os.Create(dest) // creates if file doesn't exist
	if err != nil {
		return fmt.Errorf("error : %s", err.Error())
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, source) // check first var for number of bytes copied
	if err != nil {
		return fmt.Errorf("error : %s", err.Error())
	}

	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("error : %s", err.Error())
	}
	return nil
}

func fetchFile(source string) (io.ReadCloser, error) {
	var normalisedPath string
	if strings.HasPrefix(source, "file://") {
		normalisedPath = strings.TrimPrefix(source, "file://")
	} else if strings.HasPrefix(source, "file:") {
		normalisedPath = strings.TrimPrefix(source, "file:")
	} else {
		normalisedPath = source
	}
	return os.Open(normalisedPath)
}

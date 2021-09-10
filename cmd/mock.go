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

package cmd

import (
	"fmt"
	"gatehill.io/imposter/engine"
	"gatehill.io/imposter/engine/docker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"opendeps.org/opendeps/fileutil"
	"opendeps.org/opendeps/model"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// mockCmd represents the mock command
var mockCmd = &cobra.Command{
	Use:   "mock OPENDEPS_FILE",
	Short: "Start live mocks of API dependencies",
	Long: `Starts a live mock of your API dependencies, based
on their OpenAPI specifications defined in the OpenDeps file.

This assumes that the specification URL is reachable
by this tool.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile, err := fileutil.FindSpecFile(args)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Debugf("reading opendeps manifest: %v\n", specFile)

		spec := model.Parse(specFile)
		stagingDir := generateMockConfig(specFile, spec)
		defer os.Remove(stagingDir)

		mockEngine := docker.BuildEngine(stagingDir, engine.StartOptions{
			Port:            8080,
			ImageTag:        "latest",
			ImagePullPolicy: engine.ImagePullIfNotPresent,
			LogLevel:        "DEBUG",
		})
		mockEngine.Start()
		trapExit(mockEngine)
		mockEngine.BlockUntilStopped()
	},
}

func init() {
	rootCmd.AddCommand(mockCmd)
}

func generateMockConfig(specFile string, spec *model.OpenDeps) string {
	stagingDir := fileutil.GenerateStagingDir()

	for _, dependency := range spec.Dependencies {
		openapiNormalisedPath := makeAbsoluteRelativeToFile(dependency.Spec, specFile)
		logrus.Debugf("copying openapi spec: %v\n", openapiNormalisedPath)

		openapiFilename := filepath.Base(openapiNormalisedPath)
		openapiDestFile := filepath.Join(stagingDir, openapiFilename)
		err := fileutil.CopyContent(openapiNormalisedPath, openapiDestFile)
		if err != nil {
			panic(err)
		}

		writeMockConfig(stagingDir, openapiFilename)
	}
	return stagingDir
}

func makeAbsoluteRelativeToFile(filePath string, specFile string) string {
	specDir := filepath.Dir(specFile)

	var openapiNormalisedPath string
	if strings.HasPrefix(filePath, "./") {
		openapiNormalisedPath = filepath.Join(specDir, strings.TrimPrefix(filePath, "."))
	} else {
		openapiNormalisedPath = filePath
	}
	return openapiNormalisedPath
}

func writeMockConfig(configDir string, openapiFilename string) {
	configFile, err := os.Create(filepath.Join(configDir, fmt.Sprintf("%v-config.yaml", openapiFilename)))
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	config := fmt.Sprintf(`---
plugin: openapi
specFile: "%v"
`, openapiFilename)

	_, err = configFile.WriteString(config)
	if err != nil {
		panic(err)
	}
}

// listen for an interrupt from the OS, then attempt engine cleanup
func trapExit(mockEngine engine.MockEngine) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		mockEngine.Stop()
		os.Exit(0)
	}()
}

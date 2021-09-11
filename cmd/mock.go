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
		manifestPath, err := fileutil.FindManifestFile(args)
		if err != nil {
			logrus.Fatal(err)
		}

		stagingDir := fileutil.GenerateStagingDir()
		defer os.Remove(stagingDir)

		logrus.Debugf("reading opendeps manifest: %v\n", manifestPath)
		manifest := model.Parse(manifestPath)

		bundleManifest(stagingDir, manifestPath)
		generateMockConfig(stagingDir, manifestPath, manifest)

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

func bundleManifest(stagingDir string, manifestPath string) {
	logrus.Debugf("bundling manifest: %v", manifestPath)
	err := fileutil.CopyContent(manifestPath, filepath.Join(stagingDir, "opendeps.yaml"))
	if err != nil {
		logrus.Fatal(err)
	}
	writeManifestOpenApiSpec(stagingDir)
}

func init() {
	rootCmd.AddCommand(mockCmd)
}

func generateMockConfig(stagingDir string, specFile string, manifest *model.OpenDeps) string {
	for _, dependency := range manifest.Dependencies {
		openapiNormalisedPath := makeAbsoluteRelativeToFile(dependency.Spec, specFile)
		logrus.Debugf("bundling openapi spec: %v\n", openapiNormalisedPath)

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

func writeManifestOpenApiSpec(configDir string) {
	specFileName := "opendeps-openapi-gen.yaml"
	specFile, err := os.Create(filepath.Join(configDir, specFileName))
	if err != nil {
		panic(err)
	}
	defer specFile.Close()

	spec := `---
openapi: "3.0.1"

info:
  title: OpenDeps Manifest endpoint
  version: "1.0.0"

paths:
  /.well-known/opendeps/manifest.yaml:
    get:
      responses:
        '200':
          description: Returns the OpenDeps manifest
          content:
            text/x-yaml:
              schema:
                type: object
`

	_, err = specFile.WriteString(spec)
	if err != nil {
		panic(err)
	}

	err = specFile.Sync()
	if err != nil {
		panic(err)
	}

	writeMockConfig(configDir, specFileName)
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

resources:
- path: /.well-known/opendeps/manifest.yaml
  method: GET
  response:
    staticFile: opendeps.yaml
    template: true
`, openapiFilename)

	_, err = configFile.WriteString(config)
	if err != nil {
		panic(err)
	}

	err = configFile.Sync()
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

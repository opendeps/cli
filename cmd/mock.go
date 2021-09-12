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
	"gatehill.io/imposter/engine"
	"gatehill.io/imposter/engine/docker"
	imposterfileutil "gatehill.io/imposter/fileutil"
	"gatehill.io/imposter/impostermodel"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"opendeps.org/opendeps/fileutil"
	"opendeps.org/opendeps/model"
	"os"
	"os/signal"
	"path/filepath"
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

		bundleManifest(stagingDir, manifestPath, flagForceOverwrite)
		bundleSpecs(stagingDir, manifestPath, manifest, flagForceOverwrite)

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

func bundleManifest(stagingDir string, manifestPath string, forceOverwrite bool) {
	logrus.Debugf("bundling manifest: %v", manifestPath)
	err := fileutil.CopyContent(manifestPath, filepath.Join(stagingDir, "opendeps.yaml"))
	if err != nil {
		logrus.Fatal(err)
	}

	specFileName := "opendeps-openapi-gen.yaml"
	writeManifestSpec(stagingDir, specFileName)

	resources := []impostermodel.Resource{
		{
			Path:   "/.well-known/opendeps/manifest.yaml",
			Method: "GET",
			Response: &impostermodel.ResponseConfig{
				StaticFile: "opendeps.yaml",
			},
		},
	}
	writeMockConfig(filepath.Join(stagingDir, specFileName), resources, forceOverwrite)
}

func bundleSpecs(stagingDir string, manifestPath string, manifest *model.OpenDeps, forceOverwrite bool) string {
	for _, dependency := range manifest.Dependencies {
		specNormalisedPath := fileutil.MakeAbsoluteRelativeToFile(dependency.Spec, manifestPath)
		logrus.Debugf("bundling openapi spec: %v\n", specNormalisedPath)

		specDestPath := filepath.Join(stagingDir, filepath.Base(specNormalisedPath))
		err := fileutil.CopyContent(specNormalisedPath, specDestPath)
		if err != nil {
			panic(err)
		}

		writeMockConfig(specDestPath, nil, forceOverwrite)
	}
	return stagingDir
}

// writeManifestSpec creates an OpenAPI spec describing the well known endpoint
// that serves the manifest.
func writeManifestSpec(configDir string, specFileName string) {
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
}

func writeMockConfig(specFilePath string, resources []impostermodel.Resource, forceOverwrite bool) {
	configFilePath := imposterfileutil.GenerateFilePathAdjacentToFile(specFilePath, "-config.yaml", forceOverwrite)
	configFile, err := os.Create(configFilePath)
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	config := impostermodel.GenerateConfig(specFilePath, resources, impostermodel.ConfigGenerationOptions{
		ScriptEngine: impostermodel.ScriptEngineNone,
	})

	_, err = configFile.Write(config)
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

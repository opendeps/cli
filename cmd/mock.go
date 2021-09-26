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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"opendeps.org/opendeps/fileutil"
	"opendeps.org/opendeps/manifest/bundler"
	"opendeps.org/opendeps/manifest/discovery"
	"opendeps.org/opendeps/manifest/model"
	"opendeps.org/opendeps/openapi"
	"os"
	"os/signal"
	"sync"
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
		manifestPath, err := discovery.FindManifestFile(args)
		if err != nil {
			logrus.Fatal(err)
		}

		stagingDir := fileutil.GenerateStagingDir()
		defer os.Remove(stagingDir)

		logrus.Debugf("reading opendeps manifest: %v\n", manifestPath)
		manifest := model.Parse(manifestPath)

		bundler.BundleManifest(stagingDir, manifestPath, flagForceOverwrite)
		openapi.BundleSpecs(stagingDir, manifestPath, manifest, flagForceOverwrite)

		mockEngine := docker.BuildEngine(stagingDir, engine.StartOptions{
			Port:           8080,
			Version:        "latest",
			PullPolicy:     engine.PullIfNotPresent,
			LogLevel:       "DEBUG",
			ReplaceRunning: true,
		})
		wg := &sync.WaitGroup{}
		mockEngine.Start(wg)

		trapExit(wg, mockEngine)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(mockCmd)
}

// listen for an interrupt from the OS, then attempt engine cleanup
func trapExit(wg *sync.WaitGroup, mockEngine engine.MockEngine) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		println()
		mockEngine.Stop(wg)
	}()
}

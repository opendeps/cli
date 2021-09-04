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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"opendeps.org/opendeps/fileutil"
	"opendeps.org/opendeps/model"
	"opendeps.org/opendeps/openapi"
	"os"
)

var quitIfDown, stopIfDown, failOptional bool

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test OPENDEPS_FILE",
	Short: "Tests the availability of dependencies",
	Long: `Invokes the availability endpoints of each dependency,
optionally ignoring failures if the dependency is not
marked as required.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile, err := fileutil.FindSpecFile(args)
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Info("testing dependencies")
		spec := model.Parse(specFile)

		unavailable := false
		for _, dep := range spec.Dependencies {
			err := testDependency(specFile, dep)
			if err != nil {
				if dep.Required || failOptional {
					logrus.Warnf("\u274C unavailable: %v: %v", dep.Summary, err)
					unavailable = true
					if stopIfDown {
						break
					}
				} else {
					logrus.Warnf("\u26A0 unavailable: %v: %v", dep.Summary, err)
				}
			} else {
				logrus.Infof("\u2705 available: %v", dep.Summary)
			}
		}

		// at least one dependency failed
		if unavailable && quitIfDown {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().BoolVarP(&quitIfDown, "quit-if-down", "q", false, "Exit with non-zero status if dependencies are down")
	testCmd.Flags().BoolVarP(&stopIfDown, "stop-if-down", "s", false, "Don't check further dependencies if one is down")
	testCmd.Flags().BoolVarP(&failOptional, "fail-optional", "o", false, "Fail if optional dependencies are down (default false)")
}

func testDependency(specFile string, dep model.Dependency) error {
	if "" != dep.Availability.Security {
		logrus.Warnf("security configuration for availability endpoints is not supported\n")
	}

	var url string
	if "" != dep.Availability.Url {
		// fully qualified
		url = dep.Availability.Url

	} else if "" != dep.Availability.Path {
		// relative - use openapi spec servers as base path
		basePath, err := determineBasePath(specFile, dep)
		if err != nil {
			return err
		}
		url = fmt.Sprintf("%v/%v", basePath, dep.Availability.Path)

	} else {
		panic(fmt.Errorf("No availability URL or path for %v\n", dep.Summary))
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to reach availability URL [%v]: %v\n", url, err)
	} else {
		logrus.Debugf("%v availability: %s\n", dep.Summary, resp.Status)
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("failed to reach availability URL [%v]: %s\n", url, resp.Status)
		}
	}
	return nil
}

func determineBasePath(specFile string, dependency model.Dependency) (string, error) {
	openapiNormalisedPath := makeAbsoluteRelativeToFile(dependency.Spec, specFile)
	openapiSpec, err := openapi.Parse(openapiNormalisedPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse spec [%v]: %v\n", openapiNormalisedPath, err)
	}
	if len(openapiSpec.Servers) == 0 {
		return "", fmt.Errorf("no servers found in spec [%v]\n", openapiNormalisedPath)
	} else if len(openapiSpec.Servers) > 1 {
		logrus.Warnf("more than 1 server found in spec [%v] - using first\n", openapiNormalisedPath)
	}
	serverUrl := openapiSpec.Servers[0].Url
	logrus.Debugf("determined server [%v] from openapi spec [%v]", serverUrl, openapiNormalisedPath)
	return serverUrl, nil
}

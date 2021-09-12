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
	"strings"
)

var flagNonZeroExit, flagContinueIfDown, flagRequireOptional bool
var flagServers map[string]string

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test OPENDEPS_FILE",
	Short: "Tests the availability of dependencies",
	Long: `Invokes the availability endpoints of each dependency,
optionally ignoring failures if the dependency is not
marked as required.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		manifestPath, err := fileutil.FindManifestFile(args)
		if err != nil {
			logrus.Fatal(err)
		}

		successful, tested := testDependencies(manifestPath)

		if successful == tested {
			logrus.Infof("all %d dependencies are available", tested)
		} else {
			// at least one dependency failed
			if flagNonZeroExit {
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().BoolVarP(&flagNonZeroExit, "non-zero-exit", "z", false, "Exit with non-zero status if dependencies are down")
	testCmd.Flags().BoolVarP(&flagContinueIfDown, "continue", "c", true, "Continue to check further dependencies if one or more is down")
	testCmd.Flags().BoolVarP(&flagRequireOptional, "require-optional", "o", false, "Require optional dependencies to be available")
	testCmd.Flags().StringToStringVarP(&flagServers, "server", "s", nil, "Override server base URL for a dependency (e.g. foo_service=https://example.com)")
}

func testDependencies(manifestPath string) (successful int, tested int) {
	logrus.Debugf("reading opendeps manifest: %v", manifestPath)
	manifest := model.Parse(manifestPath)

	logrus.Infof("testing %d dependencies", len(manifest.Dependencies))
	available := 0
	for depName, dep := range manifest.Dependencies {
		err := testDependency(manifestPath, depName, dep)
		if err != nil {
			if dep.Required || flagRequireOptional {
				logrus.Warnf("\u274C unavailable: %v: %v", dep.Summary, err)
				if !flagContinueIfDown {
					break
				}
			} else {
				logrus.Warnf("\u26A0 unavailable: %v: %v", dep.Summary, err)
			}
		} else {
			logrus.Infof("\u2705 available: %v", dep.Summary)
			available++
		}
	}
	return available, len(manifest.Dependencies)
}

func testDependency(manifestPath string, depName string, dep model.Dependency) error {
	if "" != dep.Availability.Security {
		logrus.Warnf("security configuration for availability endpoints is not supported\n")
	}

	var url string
	if "" != dep.Availability.Url {
		// fully qualified
		url = dep.Availability.Url

	} else if "" != dep.Availability.Path {
		// relative - use openapi spec servers as base path
		basePath, err := determineBasePath(manifestPath, depName, dep)
		if err != nil {
			return err
		}
		trimmedBasePath := strings.TrimSuffix(basePath, "/")
		trimmedPath := strings.TrimPrefix(dep.Availability.Path, "/")
		url = fmt.Sprintf("%v/%v", trimmedBasePath, trimmedPath)

	} else {
		panic(fmt.Errorf("No availability URL or path for %v\n", dep.Summary))
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to reach availability URL [%v]: %v\n", url, err)
	} else {
		logrus.Debugf("checked availability [%v]: %s\n", dep.Summary, resp.Status)
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("failed to reach availability URL [%v]: %s\n", url, resp.Status)
		}
	}
	return nil
}

func determineBasePath(manifestPath string, depName string, dependency model.Dependency) (string, error) {
	if serverUrl, found := flagServers[depName]; found {
		logrus.Debugf("determined server [%v] from overrides", serverUrl)
		return serverUrl, nil

	} else {
		specNormalisedPath := fileutil.MakeAbsoluteRelativeToFile(dependency.Spec, manifestPath)
		openapiSpec, err := openapi.Parse(specNormalisedPath)
		if err != nil {
			return "", fmt.Errorf("failed to parse spec [%v]: %v\n", specNormalisedPath, err)
		}
		if len(openapiSpec.Servers) == 0 {
			return "", fmt.Errorf("no servers found in spec [%v]\n", specNormalisedPath)
		} else if len(openapiSpec.Servers) > 1 {
			logrus.Warnf("more than 1 server found in spec [%v] - using first\n", specNormalisedPath)
		}
		serverUrl := openapiSpec.Servers[0].Url
		logrus.Debugf("determined server [%v] from openapi spec [%v]", serverUrl, specNormalisedPath)
		return serverUrl, nil
	}
}

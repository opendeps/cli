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
	"gatehill.io/imposter/fileutil"
	imposteropenapi "gatehill.io/imposter/openapi"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"opendeps.org/opendeps/model"
	"opendeps.org/opendeps/openapi"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var flagForceOverwrite bool

// scaffoldCmd represents the scaffold command
var scaffoldCmd = &cobra.Command{
	Use:   "scaffold DIR",
	Short: "Create an OpenDeps manifest based on OpenAPI files",
	Long: `Creates an OpenDeps manifest based on the OpenAPI specification files in a directory.

If DIR is not specified, the current working directory is used.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var specDir string
		if len(args) == 0 {
			specDir = "."
		} else {
			specDir = args[0]
		}
		specDir, err := filepath.Abs(specDir)
		if err != nil {
			logrus.Fatal(err)
		}

		scaffoldManifest(specDir, flagForceOverwrite)
	},
}

func init() {
	scaffoldCmd.Flags().BoolVarP(&flagForceOverwrite, "force-overwrite", "f", false, "Force overwrite of destination file(s) if already exist")
	rootCmd.AddCommand(scaffoldCmd)
}

func scaffoldManifest(specDir string, forceOverwrite bool) {
	openApiSpecs := imposteropenapi.DiscoverOpenApiSpecs(specDir)
	logrus.Infof("found %d OpenAPI spec(s)", len(openApiSpecs))

	manifest := model.OpenDeps{
		OpenDeps: model.OpenDepsSchemaVersion,
		Info: &model.Info{
			Title:   "OpenDeps manifest for " + filepath.Base(specDir),
			Version: "1.0.0",
		},
		Dependencies: make(map[string]model.Dependency),
	}

	for _, openApiSpec := range openApiSpecs {
		spec, err := openapi.Parse(openApiSpec)
		if err != nil {
			logrus.Fatalf("error parsing openapi spec: %v: %v", openApiSpec, err)
		}

		depName := strings.TrimSuffix(filepath.Base(openApiSpec), filepath.Ext(openApiSpec))
		manifest.Dependencies[depName] = model.Dependency{
			Summary: spec.Info.Title,
			Spec:    "./" + filepath.Base(openApiSpec),
			Version: spec.Info.Version,
			Availability: &model.Availability{
				Path: determineAvailabilityPath(spec),
			},
		}
	}

	manifestFileName := fileutil.GenerateFilenameAdjacentToFile(specDir, filepath.Join(specDir, "opendeps"), ".yaml", forceOverwrite)
	manifestPath := filepath.Join(specDir, manifestFileName)

	file, err := os.Create(manifestPath)
	if err != nil {
		logrus.Fatalf("error creating opendeps manifest file: %v: %v", manifestPath, err)
	}
	defer file.Close()

	marshalled, err := yaml.Marshal(manifest)
	_, err = file.Write(marshalled)
	if err != nil {
		logrus.Fatalf("error writing opendeps manifest file: %v: %v", manifestPath, err)
	}

	logrus.Infof("wrote OpenDeps manifest file: %v", manifestPath)
}

func determineAvailabilityPath(spec *openapi.PartialModel) string {
	for path, operations := range spec.Paths {
		if _, found := operations["get"]; found {
			return path
		}
	}
	return "/"
}

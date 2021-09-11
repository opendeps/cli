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
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"log"
	"opendeps.org/opendeps/fileutil"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate OPENDEPS_FILE",
	Short: "Validate a file against the OpenDeps schema",
	Long:  `Validates a YAML manifest file against the OpenDeps schema.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		manifestPath, err := fileutil.FindManifestFile(args)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("validating opendeps manifest: %v\n", manifestPath)

		json, err := loadSpecAsJson(manifestPath)
		if err != nil {
			log.Fatal(err)
		}

		validateSpec(json)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func loadSpecAsJson(manifestPath string) ([]byte, error) {
	y, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		log.Fatal(err)
	}

	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML at %v: %v\n", manifestPath, err)
	}
	return j, nil
}

func validateSpec(json []byte) {
	schemaLoader := gojsonschema.NewReferenceLoader("https://raw.githubusercontent.com/opendeps/specification/main/opendeps-specification.json")
	documentLoader := gojsonschema.NewBytesLoader(json)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Fatal(err.Error())
	}

	if result.Valid() {
		logrus.Infof("The document is valid\n")
	} else {
		logrus.Warnf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			logrus.Warnf("- %s\n", desc)
		}
	}
}

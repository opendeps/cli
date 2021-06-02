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
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate FILE",
	Short: "Validate a file against the OpenDeps schema",
	Long:  `Validates a YAML file against the OpenDeps schema.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		spec := args[0]
		fmt.Printf("validating %v\n", spec)

		json, err := loadSpecAsJson(spec)
		if err != nil {
			log.Fatal(err)
		}

		validateSpec(json)
	},
}

func loadSpecAsJson(specFile string) ([]byte, error) {
	y, err := ioutil.ReadFile(specFile)
	if err != nil {
		log.Fatal(err)
	}

	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML at %v: %v\n", specFile, err)
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
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

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

package model

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func Parse(manifestPath string) *OpenDeps {
	raw, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		logrus.Fatalln(err)
	}

	o := OpenDeps{}

	err = yaml.Unmarshal([]byte(raw), &o)
	if err != nil {
		logrus.Fatalf("error: %v\n", err)
	}

	logrus.Tracef("opendeps parsed:\n%v\n\n", o)
	return &o
}

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

package openapi

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"opendeps.org/opendeps/fileutil"
)

type Info struct {
	Title   string
	Version string
}

type Server struct {
	Url         string
	Description string
}

type PartialModel struct {
	Info    Info
	Paths   map[string]map[string]interface{}
	Servers []Server
}

func Parse(specFile string) (*PartialModel, error) {
	reader, err := fileutil.ReadContent(specFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	raw, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	o := PartialModel{}
	err = yaml.Unmarshal(raw, &o)
	if err != nil {
		return nil, fmt.Errorf("error: %v\n", err)
	}

	logrus.Tracef("openapi parsed:\n%v\n\n", o)
	return &o, nil
}

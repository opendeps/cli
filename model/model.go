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

var OpenDepsSchemaVersion = "0.1.0"

type Contact struct {
	Name  string `yaml:",omitempty"`
	Url   string `yaml:",omitempty"`
	Email string `yaml:",omitempty"`
}

type Info struct {
	Title       string   `yaml:",omitempty"`
	Description string   `yaml:",omitempty"`
	Contact     *Contact `yaml:",omitempty"`
	Version     string   `yaml:",omitempty"`
}

type Availability struct {
	Url      string `yaml:",omitempty"`
	Path     string `yaml:",omitempty"`
	Security string `yaml:",omitempty"`
}

type Dependency struct {
	Summary      string        `yaml:",omitempty"`
	Description  string        `yaml:",omitempty"`
	Spec         string        `yaml:",omitempty"`
	Version      string        `yaml:",omitempty"`
	Required     bool          `yaml:",omitempty"`
	Availability *Availability `yaml:",omitempty"`
}

type SecurityConfig struct {
	SecurityConfigType string   `yaml:"type"`
	Scheme             string   `yaml:",omitempty"`
	Headers            []string `yaml:",omitempty"`
}

type Components struct {
	SecurityConfigs map[string]SecurityConfig
}

type OpenDeps struct {
	OpenDeps     string
	Info         *Info
	Dependencies map[string]Dependency
	Components   *Components `yaml:",omitempty"`
}

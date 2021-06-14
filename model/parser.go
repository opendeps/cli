package model

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Contact struct {
	Name  string
	Url   string
	Email string
}

type Info struct {
	Title       string
	Description string
	Contact     Contact
	Version     string
}

type Availability struct {
	Url      string
	Path     string
	Security string
}

type Dependency struct {
	Summary      string
	Description  string
	Spec         string
	Version      string
	Required     bool
	Availability Availability
}

type SecurityConfig struct {
	SecurityConfigType string `yaml:"type"`
	Scheme             string
	Headers            []string
}

type OpenDeps struct {
	Info         Info
	Dependencies map[string]Dependency
	Components   struct {
		SecurityConfigs map[string]SecurityConfig
	}
}

func Parse(specFile string) *OpenDeps {
	raw, err := ioutil.ReadFile(specFile)
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

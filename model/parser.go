package model

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type InfoModel struct {
	Title       string
	Description string
	Contact     struct {
		Name  string
		Url   string
		Email string
	}
	Version string
}

type DependencyModel struct {
	Summary      string
	Description  string
	Spec         string
	Version      string
	Required     bool
	Availability struct {
		Url      string
		Path     string
		Security string
	}
}

type OpenDeps struct {
	Info         InfoModel
	Dependencies map[string]DependencyModel
	Components   struct {
		SecurityConfigs map[string]struct {
			SecurityConfigType string `yaml:"type"`
			Scheme             string
			Headers            []string
		}
	}
}

func Parse(specFile string) OpenDeps {
	y, err := ioutil.ReadFile(specFile)
	if err != nil {
		log.Fatal(err)
	}

	o := OpenDeps{}

	err = yaml.Unmarshal([]byte(y), &o)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	logrus.Tracef("--- t:\n%v\n\n", o)
	return o
}

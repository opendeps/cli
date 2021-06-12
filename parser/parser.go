package parser

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type OpenDeps struct {
	Info struct {
		Title       string
		Description string
		Contact     struct {
			Name  string
			Url   string
			Email string
		}
		Version string
	}
	Dependencies map[string]struct {
		Summary      string
		Description  string
		Spec         string
		Version      string
		Required     bool
		Availability struct {
			Url      string
			Security string
		}
	}
	Components struct {
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
	//fmt.Printf("--- t:\n%v\n\n", o)
	return o
}

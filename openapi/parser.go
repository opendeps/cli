package openapi

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"opendeps.org/opendeps/fileutil"
)

type Server struct {
	Url         string
	Description string
}

type PartialModel struct {
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

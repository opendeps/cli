package bundler

import (
	"gatehill.io/imposter/impostermodel"
	"github.com/sirupsen/logrus"
	"opendeps.org/opendeps/fileutil"
	"opendeps.org/opendeps/openapi"
	"os"
	"path/filepath"
)

func BundleManifest(stagingDir string, manifestPath string, forceOverwrite bool) {
	logrus.Debugf("bundling manifest: %v", manifestPath)
	err := fileutil.CopyContent(manifestPath, filepath.Join(stagingDir, "opendeps.yaml"))
	if err != nil {
		logrus.Fatal(err)
	}

	specFileName := "opendeps-openapi-gen.yaml"
	writeManifestSpec(stagingDir, specFileName)

	resources := []impostermodel.Resource{
		{
			Path:   "/.well-known/opendeps/manifest.yaml",
			Method: "GET",
			Response: &impostermodel.ResponseConfig{
				StaticFile: "opendeps.yaml",
			},
		},
	}
	openapi.WriteMockConfig(filepath.Join(stagingDir, specFileName), resources, forceOverwrite)
}

// writeManifestSpec creates an OpenAPI spec describing the well known endpoint
// that serves the manifest.
func writeManifestSpec(configDir string, specFileName string) {
	specFile, err := os.Create(filepath.Join(configDir, specFileName))
	if err != nil {
		panic(err)
	}
	defer specFile.Close()

	spec := `---
openapi: "3.0.1"

info:
  title: OpenDeps Manifest endpoint
  version: "1.0.0"

paths:
  /.well-known/opendeps/manifest.yaml:
    get:
      responses:
        '200':
          description: Returns the OpenDeps manifest
          content:
            text/x-yaml:
              schema:
                type: object
`

	_, err = specFile.WriteString(spec)
	if err != nil {
		panic(err)
	}
}

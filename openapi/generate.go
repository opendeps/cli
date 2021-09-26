package openapi

import (
	imposterfileutil "gatehill.io/imposter/fileutil"
	"gatehill.io/imposter/impostermodel"
	"github.com/sirupsen/logrus"
	"opendeps.org/opendeps/fileutil"
	"opendeps.org/opendeps/manifest/model"
	"os"
	"path/filepath"
)

func BundleSpecs(stagingDir string, manifestPath string, manifest *model.OpenDeps, forceOverwrite bool) string {
	for _, dependency := range manifest.Dependencies {
		specNormalisedPath := fileutil.MakeAbsoluteRelativeToFile(dependency.Spec, manifestPath)
		logrus.Debugf("bundling openapi spec: %v\n", specNormalisedPath)

		specDestPath := filepath.Join(stagingDir, filepath.Base(specNormalisedPath))
		err := fileutil.CopyContent(specNormalisedPath, specDestPath)
		if err != nil {
			panic(err)
		}

		WriteMockConfig(specDestPath, nil, forceOverwrite)
	}
	return stagingDir
}

func WriteMockConfig(specFilePath string, resources []impostermodel.Resource, forceOverwrite bool) {
	configFilePath := imposterfileutil.GenerateFilePathAdjacentToFile(specFilePath, "-config.yaml", forceOverwrite)
	configFile, err := os.Create(configFilePath)
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	config := impostermodel.GenerateConfig(specFilePath, resources, impostermodel.ConfigGenerationOptions{
		ScriptEngine: impostermodel.ScriptEngineNone,
	})

	_, err = configFile.Write(config)
	if err != nil {
		panic(err)
	}
}

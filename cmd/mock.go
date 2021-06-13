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
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"opendeps.org/opendeps/fileutil"
	"opendeps.org/opendeps/model"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// mockCmd represents the mock command
var mockCmd = &cobra.Command{
	Use:   "mock OPENDEPS_FILE",
	Short: "Start live mocks of API dependencies",
	Long: `Starts a live mock of your API dependencies, based
on their OpenAPI specifications defined in the OpenDeps file.

This assumes that the specification URL is reachable
by this tool.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile := args[0]

		logrus.Debugf("reading dependencies: %v\n", specFile)
		spec := model.Parse(specFile)

		stagingDir := generateMockConfig(specFile, spec)
		defer os.Remove(stagingDir)

		startMockEngine(stagingDir)
	},
}

func init() {
	rootCmd.AddCommand(mockCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mockCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mockCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func startMockEngine(stagingDir string) {
	logrus.Info("starting mock engine")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "docker.io/outofcoffee/imposter-openapi", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "outofcoffee/imposter-openapi",
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: stagingDir,
				Target: "/opt/imposter/config",
			},
		},
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "8080",
				},
			},
		},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	trapExit(cli, ctx, resp.ID)
	println("container engine started - press ctrl+c to stop")

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func generateMockConfig(specFile string, spec model.OpenDeps) string {
	specDir := filepath.Dir(specFile)
	stagingDir := fileutil.GenerateStagingDir()

	for _, dependency := range spec.Dependencies {
		var openapiNormalisedPath string
		if strings.HasPrefix(dependency.Spec, "file://") {
			openapiNormalisedPath = strings.TrimPrefix(dependency.Spec, "file://")
		} else if strings.HasPrefix(dependency.Spec, "file:") {
			openapiNormalisedPath = strings.TrimPrefix(dependency.Spec, "file:")
		} else if strings.HasPrefix(dependency.Spec, "./") {
			openapiNormalisedPath = filepath.Join(specDir, strings.TrimPrefix(dependency.Spec, "."))
		} else {
			logrus.Warnf("skipping unsupported spec url: %v\n", dependency.Spec)
			continue
		}

		logrus.Debugf("copying openapi spec: %v\n", openapiNormalisedPath)

		openapiFilename := filepath.Base(openapiNormalisedPath)
		openapiDestFile := filepath.Join(stagingDir, openapiFilename)
		err := fileutil.CopyFile(openapiNormalisedPath, openapiDestFile)
		if err != nil {
			panic(err)
		}

		writeMockConfig(stagingDir, openapiFilename)
	}
	return stagingDir
}

func writeMockConfig(configDir string, openapiFilename string) {
	configFile, err := os.Create(filepath.Join(configDir, fmt.Sprintf("%v-config.yaml", openapiFilename)))
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	config := fmt.Sprintf(`---
plugin: openapi
specFile: "%v"
`, openapiFilename)

	_, err = configFile.WriteString(config)
	if err != nil {
		panic(err)
	}
}

func stopMockEngine(cli *client.Client, ctx context.Context, containerID string) {
	logrus.Infof("\rstopping mock engine...\n")
	err := cli.ContainerStop(ctx, containerID, nil)
	if err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	println("container engine stopped")
}

// listen for an interrupt from the OS, then attempt engine cleanup
func trapExit(cli *client.Client, ctx context.Context, containerID string) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stopMockEngine(cli, ctx, containerID)
		os.Exit(0)
	}()
}

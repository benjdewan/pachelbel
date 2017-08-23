// Copyright © 2017 ben dewan <benj.dewan@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/benjdewan/pachelbel/connection"
	"github.com/ghodss/yaml"
)

// BuildClusterFilter accepts a list of cluster names to filter
// configuration data. Only deployments to clusters in the filter
// are returned by ReadFiles()
func BuildClusterFilter(clusters []string) {
	clusterFilter = make(map[string]struct{})

	for _, cluster := range clusters {
		clusterFilter[cluster] = struct{}{}
	}
}

// BuildDatacenterFilter accepts a list of datacenter slugs to filter
// configuration data by. Only deployments to datacenters in the filter
// are returned by ReadFiles()
func BuildDatacenterFilter(datacenters []string) {
	datacenterFilter = make(map[string]struct{})

	for _, datacenter := range datacenters {
		datacenterFilter[datacenter] = struct{}{}
	}
}

// ReadFiles works through a list of arguments to parse configuration
// data into deployment object. Both configuration files and directories
// of configuration files are valid arguments, but directories are not
// read recursively, only immediate child files are parsed.
func ReadFiles(args []string, verbose bool) ([]connection.Deployment, error) {
	deployments := []connection.Deployment{}
	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			return deployments, err
		}
		newDeployments := []DeploymentV1{}
		switch mode := info.Mode(); {
		case mode.IsDir():
			newDeployments, err = readDir(path, verbose)
		case mode.IsRegular():
			newDeployments, err = readFile(path, verbose)
		}
		if err != nil {
			return deployments, err
		}
		for _, d := range newDeployments {
			deployments = append(deployments, connection.Deployment(d))
		}
	}
	return deployments, nil
}

func readDir(root string, verbose bool) ([]DeploymentV1, error) {
	deployments := []DeploymentV1{}
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, readErr error) error {
		if readErr != nil {
			return readErr
		}
		if path == root {
			return nil
		}
		if info.IsDir() {
			if verbose {
				fmt.Printf("Skipping %v.\n", path)
				return nil
			}
		}
		newDeployments, err := readFile(path, verbose)
		if err != nil {
			return err
		}
		deployments = append(deployments, newDeployments...)
		return nil
	})
	return deployments, walkErr
}

func readFile(path string, verbose bool) ([]DeploymentV1, error) {
	deployments := []DeploymentV1{}
	file, err := os.Open(path)
	if err != nil {
		return deployments, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Our filepointer has become invalid or the
			// underlying syscall was interrupted
			panic(closeErr)
		}
	}()

	return readConfigs(file)
}

func readConfigs(r io.Reader) ([]DeploymentV1, error) {
	deployments := []DeploymentV1{}

	scanner := bufio.NewScanner(r)
	scanner.Split(splitYAMLObjects)
	for scanner.Scan() {
		deployment, err := readConfig(scanner.Bytes())
		if err != nil {
			return deployments, err
		}
		if filtered(deployment) {
			continue
		}
		deployments = append(deployments, deployment)
	}
	return deployments, nil
}

func readConfig(blob []byte) (DeploymentV1, error) {
	var deployment DeploymentV1
	if err := yaml.Unmarshal(blob, &deployment); err != nil {
		return deployment, err
	}
	err := validate(deployment, string(blob))
	return deployment, err
}

func splitYAMLObjects(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, []byte("\n---")); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return
}

var clusterFilter map[string]struct{}
var datacenterFilter map[string]struct{}

func filtered(deployment DeploymentV1) bool {
	if len(deployment.Cluster) > 0 {
		// At this point the deployment has already been validated, so
		// we can safely assume this means a cluster deployment
		return filterByCluster(deployment)
	}
	return filterByDatacenter(deployment)
}

func filterByCluster(d DeploymentV1) bool {
	if len(clusterFilter) == 0 {
		return false
	}
	_, ok := clusterFilter[d.Cluster]
	return !ok
}

func filterByDatacenter(d DeploymentV1) bool {
	if len(datacenterFilter) == 0 {
		return false
	}
	_, ok := datacenterFilter[d.Datacenter]
	return !ok
}

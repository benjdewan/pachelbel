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
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

func BuildClusterFilter(clusters []string) {
	clusterFilter = make(map[string]struct{})

	for _, cluster := range clusters {
		clusterFilter[cluster] = struct{}{}
	}
}

func ReadFiles(args []string, verbose bool) ([]DeploymentV1, error) {
	deployments := []DeploymentV1{}
	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			return []DeploymentV1{}, err
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

		deployments = append(deployments, newDeployments...)
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
	fmt.Printf("Reading configuration from %v\n", path)
	deployments := []DeploymentV1{}
	file, err := os.Open(path)
	if err != nil {
		return deployments, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(splitYamlObjects)
	for scanner.Scan() {
		var deployment DeploymentV1
		blob := scanner.Bytes()
		if err := yaml.Unmarshal(blob, &deployment); err != nil {
			return deployments, err
		}
		if err := Validate(deployment, string(blob)); err != nil {
			return deployments, err
		}
		if filtered(deployment) {
			if verbose {
				fmt.Printf("Not updating the '%s' deployment. Its cluster has been filtered out\n",
					deployment.Name)
			}
			continue
		}
		deployments = append(deployments, deployment)
	}
	return deployments, nil
}

func splitYamlObjects(data []byte, atEOF bool) (advance int, token []byte, err error) {
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

func filtered(deployment DeploymentV1) bool {
	if len(clusterFilter) == 0 {
		return false
	}
	_, ok := clusterFilter[deployment.Cluster]
	return !ok
}

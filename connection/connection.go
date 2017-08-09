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

package connection

import (
	"bytes"
	"fmt"
	"os"
	"time"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/ghodss/yaml"
)

type Deployment interface {
	GetCluster() string
	GetDatacenter() string
	GetName() string
	GetNotes() string
	GetScaling() int
	GetSSL() bool
	GetTeamRoles() map[string]([]string)
	GetTimeout() float64
	GetType() string
	GetVersion() string
	GetWiredTiger() bool
}

type Connection struct {
	client            *compose.Client
	accountID         string
	clusterIDsByName  map[string]string
	deploymentsByName map[string](*compose.Deployment)
	newDeploymentIDs  map[string]struct{}
	pollingInterval   time.Duration
}

func Init(apiKey string, pollingInterval int) (*Connection, error) {
	cxn := &Connection{
		newDeploymentIDs: make(map[string]struct{}),
		pollingInterval:  time.Duration(pollingInterval) * time.Second,
	}
	var err error

	cxn.client, err = createClient(apiKey)
	if err != nil {
		return cxn, err
	}

	cxn.accountID, err = fetchAccountID(cxn.client)
	if err != nil {
		return cxn, err
	}

	cxn.clusterIDsByName, err = fetchClusters(cxn.client)
	if err != nil {
		return cxn, err
	}

	cxn.deploymentsByName, err = fetchDeployments(cxn.client)

	return cxn, err
}

func Provision(cxn *Connection, deployment Deployment) error {
	if existing, ok := cxn.deploymentsByName[deployment.GetName()]; ok {
		return update(cxn, existing.ID, deployment)
	}
	return provision(cxn, deployment)
}

func (cxn *Connection) ConnectionStringsYAML(outFile string, verbose bool) error {
	fmt.Printf("Writing connection strings to '%v'\n", outFile)

	yamlObjects := []([]byte){}
	for deploymentID := range cxn.newDeploymentIDs {
		if verbose {
			fmt.Printf("Fetching latest metadata for deployment '%v'\n",
				deploymentID)
		}
		deployment, errs := cxn.client.GetDeployment(deploymentID)
		if len(errs) != 0 {
			return fmt.Errorf("Unable to resolve deployment '%v':\n%v\n",
				deploymentID, errsOut(errs))
		}
		cxnStrings := make(map[string][]string)
		cxnStrings[deployment.Name] = deployment.Connection.Direct
		yamlObj, err := yaml.Marshal(cxnStrings)
		if err != nil {
			return err
		}
		yamlObjects = append(yamlObjects, yamlObj)
	}

	handle, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := handle.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()
	_, err = handle.Write(bytes.Join(yamlObjects, []byte("\n---\n")))
	return err
}

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
	"errors"
	"fmt"

	compose "github.com/compose/gocomposeapi"
)

type Deployment interface {
	GetName() string
	GetCluster() string
	GetType() string
	GetNotes() string
	GetScaling() int
	GetSSL() bool
	GetTeams() map[string]([]string)
}

type Connection struct {
	client              *compose.Client
	accountID           string
	clusterIDsByName    map[string]string
	deploymentIDsByName map[string]string
}

func Init(apiKey string, verbose bool) (*Connection, error) {
	cxn := &Connection{}
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

	cxn.deploymentIDsByName, err = fetchDeployments(cxn.client)

	return cxn, err
}

func Provision(cxn *Connection, deployment Deployment, verbose bool) error {
	if id, ok := cxn.deploymentIDsByName[deployment.GetName()]; ok {
		return rescale(cxn, id, deployment, verbose)
	}
	return provision(cxn, deployment, verbose)
}

func rescale(cxn *Connection, deploymentID string, deployment Deployment, verbose bool) error {
	scalings, errs := cxn.client.GetScalings(deploymentID)
	if len(errs) != 0 {
		return fmt.Errorf("Unable get scaling status for '%s': %v",
			deployment.GetName(), errsOut(errs))
	}
	if scalings.AllocatedUnits == deployment.GetScaling() {
		if verbose {
			fmt.Printf("Nothing to do for '%s'\n", deployment.GetName())
		}
		return nil
	}

	sParams := compose.ScalingsParams{
		DeploymentID: deploymentID,
		Units:        deployment.GetScaling(),
	}
	_, errs = cxn.client.SetScalings(sParams)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to resize '%s': %v\n",
			deployment.GetName(), errsOut(errs))
	}
	return nil
}

func provision(cxn *Connection, deployment Deployment, verbose bool) error {
	if verbose {
		fmt.Printf("Provisioning '%s' in '%s'...\n", deployment.GetName(),
			deployment.GetCluster())
	}

	clusterID, ok := cxn.clusterIDsByName[deployment.GetCluster()]
	if !ok {
		return fmt.Errorf("Unable to provsion '%s'. The specified cluster name, '%s' does not map to a known cluster.",
			deployment.GetName(), deployment.GetCluster())
	}

	dParams := compose.DeploymentParams{
		Name:         deployment.GetName(),
		AccountID:    cxn.accountID,
		ClusterID:    clusterID,
		DatabaseType: deployment.GetType(),
		Notes:        deployment.GetNotes(),
		// Mising WiredTiger!!!
	}

	if deployment.GetScaling() > 1 {
		dParams.Units = deployment.GetScaling()
	}
	if deployment.GetSSL() {
		dParams.SSL = true
	}

	//This needs to be wrapped in retry logic
	newDeployment, errs := cxn.client.CreateDeployment(dParams)
	if errs != nil {
		return fmt.Errorf("Unable to create '%s': %s\n", errsOut(errs))
	}
	if verbose {
		fmt.Printf("Provision of '%s' is complete!\n", newDeployment.Name)
	}
	cxn.deploymentIDsByName[newDeployment.ID] = newDeployment.Name

	return nil
}

func createClient(apiKey string) (*compose.Client, error) {
	if len(apiKey) == 0 {
		return nil, errors.New("No API key found. Specify one using the --api-key flag or the COMPOSE_API_KEY environment variable")
	}

	return compose.NewClient(apiKey)
}

func fetchAccountID(client *compose.Client) (string, error) {
	account, errs := client.GetAccount()
	if len(errs) != 0 {
		return "", fmt.Errorf("Failed to get account id:\n%s", errsOut(errs))
	}

	return account.ID, nil
}

func fetchClusters(client *compose.Client) (map[string]string, error) {
	clusterIDsByName := make(map[string]string)

	clusters, errs := client.GetClusters()
	if len(errs) != 0 {
		return clusterIDsByName, fmt.Errorf("Failed to get cluster information:\n%s",
			errsOut(errs))
	}

	if clusters == nil {
		return clusterIDsByName, fmt.Errorf("No clusters found")
	}

	for _, cluster := range *clusters {
		clusterIDsByName[cluster.Name] = cluster.ID
	}
	return clusterIDsByName, nil
}

func fetchDeployments(client *compose.Client) (map[string]string, error) {
	deploymentIDsByName := make(map[string]string)
	deployments, errs := client.GetDeployments()
	if len(errs) != 0 {
		return deploymentIDsByName, fmt.Errorf("Failed to get deployments:\n%s",
			errsOut(errs))
	}

	if deployments == nil {
		// This is not necessarily an error.
		return deploymentIDsByName, nil
	}

	for _, deployment := range *deployments {
		deploymentIDsByName[deployment.Name] = deployment.ID
	}
	return deploymentIDsByName, nil
}

func errsOut(errs []error) string {
	var buf bytes.Buffer
	for _, err := range errs {
		if _, wErr := buf.WriteString(err.Error()); wErr != nil {
			panic(wErr)
		}
		if _, wErr := buf.WriteString("\n"); wErr != nil {
			panic(wErr)
		}
	}
	return buf.String()
}

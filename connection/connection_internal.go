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
	"time"

	compose "github.com/compose/gocomposeapi"
)

type cxnString struct {
	Name              string                    `json:"name"`
	ConnectionStrings compose.ConnectionStrings `json:"connection-strings"`
}

func rescale(cxn *Connection, deploymentID string, deployment Deployment, verbose bool) error {
	scalings, errs := cxn.client.GetScalings(deploymentID)
	if len(errs) != 0 {
		return fmt.Errorf("Unable get scaling status for '%s': %v",
			deployment.GetName(), errsOut(errs))
	}
	if scalings.AllocatedUnits == deployment.GetScaling() {
		fmt.Printf("Nothing to do for '%s'\n", deployment.GetName())
		cxn.newDeploymentIDs = append(cxn.newDeploymentIDs, deploymentID)
		return nil
	}

	if verbose {
		fmt.Printf("Rescaling deployment %v:\n\tCurrent scale: %v\n\tDesired scaling: %v\n",
			deployment.GetName(), scalings.AllocatedUnits, deployment.GetScaling())
	}

	sParams := compose.ScalingsParams{
		DeploymentID: deploymentID,
		Units:        deployment.GetScaling(),
	}
	recipe, errs := cxn.client.SetScalings(sParams)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to resize '%s': %v\n",
			deployment.GetName(), errsOut(errs))
	}

	err := cxn.waitOnRecipe(recipe.ID, deployment.GetTimeout(), verbose)
	if err == nil {
		cxn.newDeploymentIDs = append(cxn.newDeploymentIDs, recipe.DeploymentID)
	}
	return err
}

func provision(cxn *Connection, deployment Deployment, verbose bool) error {
	fmt.Printf("Provisioning '%s' in '%s'...\n", deployment.GetName(),
		deployment.GetCluster())

	dParams, err := deploymentParams(deployment, cxn)
	if err != nil {
		return err
	}

	//This needs to be wrapped in retry logic
	newDeployment, errs := cxn.client.CreateDeployment(dParams)
	if errs != nil {
		return fmt.Errorf("Unable to create '%s': %s\n",
			deployment.GetName(), errsOut(errs))
	}

	if err := cxn.waitOnRecipe(newDeployment.ProvisionRecipeID, deployment.GetTimeout(), verbose); err != nil {
		return err
	}
	fmt.Printf("Provision of '%s' is complete!\n", newDeployment.Name)
	cxn.newDeploymentIDs = append(cxn.newDeploymentIDs, newDeployment.ID)

	return nil
}

func (cxn *Connection) waitOnRecipe(recipeID string, timeout float64, verbose bool) error {
	fmt.Printf("Waiting for recipe %v to complete\n", recipeID)
	start := time.Now()
	for time.Since(start).Seconds() <= timeout {
		recipe, errs := cxn.client.GetRecipe(recipeID)
		if len(errs) != 0 {
			return fmt.Errorf("Error waiting on recipe %v:\n%v\n", recipeID, errsOut(errs))
		}
		if verbose {
			fmt.Printf("Recipe %v status: %v\n", recipeID,
				recipe.Status)
		}
		if recipe.Status == "complete" {
			return nil
		}
		time.Sleep(cxn.pollingInterval)
	}
	return fmt.Errorf("Timed out waiting on recipe %v to complete", recipeID)
}

func deploymentParams(deployment Deployment, cxn *Connection) (compose.DeploymentParams, error) {
	dParams := compose.DeploymentParams{
		Name:         deployment.GetName(),
		AccountID:    cxn.accountID,
		DatabaseType: deployment.GetType(),
		Notes:        deployment.GetNotes(),
	}

	clusterID, ok := cxn.clusterIDsByName[deployment.GetCluster()]
	if !ok {
		return dParams, fmt.Errorf("Unable to provsion '%s'. The specified cluster name, '%s' does not map to a known cluster.",
			deployment.GetName(), deployment.GetCluster())
	}

	dParams.ClusterID = clusterID

	if deployment.GetWiredTiger() {
		dParams.WiredTiger = true
	}

	if deployment.GetScaling() > 1 {
		dParams.Units = deployment.GetScaling()
	}

	if deployment.GetSSL() {
		dParams.SSL = true
	}

	return dParams, nil
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

func fetchDeployments(client *compose.Client) (map[string](*compose.Deployment), error) {
	deploymentsByName := make(map[string](*compose.Deployment))
	deployments, errs := client.GetDeployments()
	if len(errs) != 0 {
		return deploymentsByName, fmt.Errorf("Failed to get deployments:\n%s",
			errsOut(errs))
	}

	if deployments == nil {
		// This is not necessarily an error.
		return deploymentsByName, nil
	}

	for _, deployment := range *deployments {
		deploymentsByName[deployment.Name] = &deployment
	}
	return deploymentsByName, nil
}

func fetchAccountID(client *compose.Client) (string, error) {
	account, errs := client.GetAccount()
	if len(errs) != 0 {
		return "", fmt.Errorf("Failed to get account id:\n%s", errsOut(errs))
	}

	return account.ID, nil
}

func createClient(apiKey string) (*compose.Client, error) {
	if len(apiKey) == 0 {
		return nil, errors.New("No API key found. Specify one using the --api-key flag or the COMPOSE_API_KEY environment variable")
	}

	return compose.NewClient(apiKey)
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

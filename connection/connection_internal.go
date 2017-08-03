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

	compose "github.com/benjdewan/gocomposeapi"
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
		return addTeamRoles(cxn, deploymentID, deployment.GetTeamRoles(), verbose)
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
	if err != nil {
		return err
	}
	cxn.newDeploymentIDs = append(cxn.newDeploymentIDs, recipe.DeploymentID)
	return addTeamRoles(cxn, deploymentID, deployment.GetTeamRoles(), verbose)
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

func addTeamRoles(cxn *Connection, deploymentID string, teamRoles map[string][]string, verbose bool) error {
	existingRoles, errs := cxn.client.GetTeamRoles(deploymentID)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to retrieve team_role information for '%s':\n%v\n",
			deploymentID, errsOut(errs))
	}
	if len(teamRoles) == 0 {
		return nil
	}

	fmt.Printf("Setting up team roles for '%v'\n", deploymentID)

	for role, teams := range teamRoles {
		existingTeams := []compose.Team{}
		for _, existingTeamRoles := range *existingRoles {
			if existingTeamRoles.Name == role {
				existingTeams = existingTeamRoles.Teams
				break
			}
		}

		for _, teamID := range filterTeams(teams, existingTeams) {
			params := compose.TeamRoleParams{
				Name:   role,
				TeamID: teamID,
			}
			if verbose {
				fmt.Printf("Adding team '%v' to deployment '%v' with role '%v'\n",
					teamID, deploymentID, role)
			}

			_, createErrs := cxn.client.CreateTeamRole(deploymentID,
				params)
			if createErrs != nil {
				return fmt.Errorf("Unable to add team '%s' as '%s' to %s:\n%v\n",
					teamID, role, deploymentID,
					errsOut(createErrs))
			}
		}
	}
	return nil
}

func filterTeams(teams []string, filterList []compose.Team) []string {
	remainingTeams := []string{}
	filter := teamListToMap(filterList)
	for _, team := range teams {
		if _, ok := filter[team]; !ok {
			remainingTeams = append(remainingTeams, team)
		}
	}
	return remainingTeams
}

func teamListToMap(in []compose.Team) map[string]struct{} {
	out := make(map[string]struct{})
	for _, item := range in {
		out[item.ID] = struct{}{}
	}
	return out
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
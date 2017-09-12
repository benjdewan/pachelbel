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
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/golang-collections/go-datastructures/queue"
)

func (cxn *Connection) getDeploymentByName(name string) (*compose.Deployment, bool) {
	if item, ok := cxn.deploymentsByName.Load(name); ok {
		return item.(*compose.Deployment), true
	}
	return nil, false
}

func (cxn *Connection) waitOnRecipe(recipeID string, timeout float64) error {
	start := time.Now()
	for time.Since(start).Seconds() <= timeout {
		if recipe, errs := cxn.client.GetRecipe(recipeID); len(errs) != 0 {
			return fmt.Errorf("Error waiting on recipe %v:\n%v\n",
				recipeID, errsOut(errs))
		} else if recipe.Status == "complete" {
			return nil
		}
		time.Sleep(cxn.pollingInterval)
	}
	return fmt.Errorf("Timed out waiting on recipe %v to complete", recipeID)
}

func addTeamRoles(cxn *Connection, deploymentID string, teamRoles map[string][]string) error {
	existingRoles, errs := cxn.client.GetTeamRoles(deploymentID)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to retrieve team_role information for '%s':\n%v\n",
			deploymentID, errsOut(errs))
	}
	if len(teamRoles) == 0 {
		return nil
	}

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
	if len(errs) != 0 || clusters == nil {
		return clusterIDsByName, fmt.Errorf("Failed to get cluster information:\n%s",
			errsOut(errs))
	}

	for _, cluster := range *clusters {
		clusterIDsByName[cluster.Name] = cluster.ID
	}
	return clusterIDsByName, nil
}

func fetchDatacenters(client *compose.Client) (map[string]struct{}, error) {
	datacenters := make(map[string]struct{})

	datacenterObjs, errs := client.GetDatacenters()
	if len(errs) != 0 || datacenterObjs == nil {
		return datacenters, fmt.Errorf("Failed to get datacenter information:\n%s",
			errsOut(errs))
	}

	for _, datacenter := range *datacenterObjs {
		datacenters[datacenter.Slug] = struct{}{}
	}

	return datacenters, nil
}

func fetchDeployments(cxn *Connection) error {
	deployments, errs := cxn.client.GetDeployments()
	if len(errs) != 0 || deployments == nil {
		return fmt.Errorf("Failed to get deployments:\n%s", errsOut(errs))
	}

	var wg sync.WaitGroup
	q := queue.New(0)
	wg.Add(len(*deployments))
	for _, partial := range *deployments {
		go func(id string) {
			deployment, errs := cxn.client.GetDeployment(id)

			if len(errs) != 0 {
				msg := fmt.Errorf("Unable to get deployment information for '%s': %v", id, errs)
				enqueue(q, msg)
			}
			cxn.deploymentsByName.Store(deployment.Name, deployment)
			wg.Done()
		}(partial.ID)
	}
	wg.Wait()

	return flushErrors(q)
}

func fetchAccountID(client *compose.Client) (string, error) {
	account, errs := client.GetAccount()
	if len(errs) != 0 {
		return "", fmt.Errorf("Failed to get account id:\n%s", errsOut(errs))
	}

	return account.ID, nil
}

func createClient(apiKey string, w io.Writer) (*compose.Client, error) {
	if len(apiKey) == 0 {
		return nil, errors.New("No API key found. Specify one using the --api-key flag or the COMPOSE_API_KEY environment variable")
	}
	client, err := compose.NewClient(apiKey)
	if err != nil {
		// In the current version of gocomposeapi this is impossible
		panic(err)
	}
	return client.SetLogger(true, w), nil
}

func enqueue(q *queue.Queue, item interface{}) {
	if err := q.Put(item); err != nil {
		// This only happens if we are using a Queue after Dispose()
		// has been called on it.
		panic(err)
	}
}

func errsOut(errs []error) string {
	msgs := []string{}
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

func flushErrors(q *queue.Queue) error {
	if q.Empty() {
		q.Dispose()
		return nil
	}
	length := q.Len()
	items, qErr := q.Get(length)
	if qErr != nil {
		// Get() only returns an error if Dispose() has already
		// been called on the queue.
		panic(qErr)
	}
	q.Dispose()
	return fmt.Errorf("%d fatal error(s) occurred:\n%v", length, items)
}

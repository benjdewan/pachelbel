package connection

import (
	"errors"
	"fmt"
	"io"
	"strings"
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
		time.Sleep(5 * time.Second)
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

func enqueue(q *queue.Queue, items ...interface{}) {
	for _, item := range items {
		if err := q.Put(item); err != nil {
			// This only happens if we are using a Queue after Dispose()
			// has been called on it.
			panic(err)
		}
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

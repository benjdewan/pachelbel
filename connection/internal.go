package connection

import (
	"errors"
	"fmt"
	"io"
	"time"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/benjdewan/pachelbel/output"
	"github.com/masterminds/semver"
)

func (cxn *Connection) addToBuilder(id string, b *output.Builder) error {
	if output.IsFake(id) {
		return b.AddFake(id)
	}
	deployment, errs := cxn.client.GetDeployment(id)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get deployment information for '%s':\n%v", id, errs)
	}
	return b.Add(deployment)
}

func (cxn *Connection) existingDeployment(deployment compose.Deployment) (ExistingDeployment, error) {
	existing := ExistingDeployment{
		ID:      deployment.ID,
		Name:    deployment.Name,
		Type:    deployment.Type,
		Notes:   deployment.Notes,
		Version: deployment.Version,
	}

	transitions, errs := cxn.client.GetVersionsForDeployment(deployment.ID)
	if len(errs) != 0 {
		return existing, fmt.Errorf("Unable to get upgrade details for '%s'", deployment.Name)
	}
	if transitions != nil {
		existing.Upgrades = upgradeList(*transitions)
	}

	scalings, errs := cxn.client.GetScalings(deployment.ID)
	if len(errs) != 0 {
		return existing, fmt.Errorf("Unable to get scaling details for '%s'", deployment.Name)
	}

	existing.Scaling = scalings.AllocatedUnits
	return existing, nil
}

func upgradeList(transitions []compose.VersionTransition) []*semver.Version {
	versions := []*semver.Version{}
	for _, transition := range transitions {
		if transition.Method != "in_place" {
			continue
		}
		if version, err := semver.NewVersion(transition.ToVersion); err == nil {
			versions = append(versions, version)
		}
	}
	return versions
}

func buildDatabaseVersionMap(dbs []compose.Database) map[string][]string {
	databases := make(map[string][]string)
	for _, db := range dbs {
		databases[db.DatabaseType] = buildDatabaseVersionSlice(db.Embedded.Versions)
	}
	return databases
}

func buildDatabaseVersionSlice(vs []compose.Version) []string {
	versions := []string{}
	for _, v := range vs {
		versions = append(versions, v.Version)
	}
	return versions
}

func (cxn *Connection) wait(recipeID string, timeout float64) error {
	start := time.Now()
	for time.Since(start).Seconds() <= timeout {
		if recipe, errs := cxn.client.GetRecipe(recipeID); len(errs) != 0 {
			return fmt.Errorf("Error waiting on recipe %v:\n%v\n",
				recipeID, errs)
		} else if recipe.Status == "complete" {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("Timed out waiting on recipe %v to complete", recipeID)
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
		return "", fmt.Errorf("Failed to get account id:\n%v", errs)
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

package config

import (
	"fmt"

	"github.com/benjdewan/pachelbel/connection"
)

var validRoles = map[string]struct{}{
	"admin":     {},
	"developer": {},
	"manager":   {},
}

var (
	// Databases is a map of database types Compose supports to the versions of
	// those databases that Compose supports
	Databases map[string][]string
	// Clusters is a map of Cluster names to IDs to validate that the cluster
	// a user specifies exists
	Clusters map[string]string
	// Datacenters is a map of names to validate that the datacenter slugs a user
	// specifies exist
	Datacenters map[string]struct{}
	// CXN is a connection object for retrieving existing deployment information
	CXN *connection.Connection
)

func existingDeployment(idOrName string) (connection.ExistingDeployment, bool) {
	if CXN == nil {
		return connection.ExistingDeployment{}, false
	}
	d, err := CXN.ExistingDeployment(idOrName)
	return d, err == nil
}

func validateType(deploymentType string) []string {
	if _, ok := Databases[deploymentType]; !ok {
		return []string{fmt.Sprintf("'%s' is not a valid deployment type", deploymentType)}
	}
	return []string{}
}

func validateName(name string) []string {
	if len(name) == 0 {
		return []string{"The 'name' field is required"}
	}
	return []string{}
}

func validateScaling(scaling *int) []string {
	if scaling != nil && *scaling < 1 {
		return []string{"The 'scaling' field must be an integer >= 1"}
	}
	return []string{}
}

func validateTeams(teams []*TeamV1) []string {
	errs := []string{}
	if teams == nil {
		return errs
	}
	for _, team := range teams {
		if len(team.ID) == 0 {
			errs = append(errs, "Every team entry requires an ID")
		}
		if _, ok := validRoles[team.Role]; ok {
			continue
		}
		errs = append(errs,
			fmt.Sprintf("'%s' is not a valid team role", team.Role))
	}
	return errs
}

func validateWiredTiger(wiredTiger bool, deploymentType string) []string {
	if wiredTiger && deploymentType != "mongodb" {
		return []string{"The 'wired_tiger' field is only valid for the 'mongodb' deployment type"}
	}
	return []string{}
}

func validateCacheMode(cacheMode bool, deploymentType string) []string {
	if cacheMode && deploymentType != "redis" {
		return []string{"The 'cache_mode' field is only valid for the 'redis' deployment type"}
	}
	return []string{}
}

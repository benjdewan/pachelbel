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
	"fmt"
	"strings"
)

var validRolesV1 = map[string]struct{}{
	"admin":     {},
	"developer": {},
	"manager":   {},
}

var validTypes = map[string]struct{}{
	"mongodb":       {},
	"rethinkdb":     {},
	"elasticsearch": {},
	"redis":         {},
	"postgresql":    {},
	"rabbitmq":      {},
	"etcd":          {},
	"mysql":         {},
	"janusgraph":    {},
}

func validate(d deploymentV1, input string) error {
	errs := []string{}

	errs = validateConfigVersion(d.ConfigVersion, errs)
	errs = validateType(d.Type, errs)
	errs = validateDeploymentTarget(d.Cluster, d.Datacenter, d.Tags, errs)
	errs = validateName(d.Name, errs)
	errs = validateScaling(d.Scaling, errs)
	errs = validateWiredTiger(d.WiredTiger, d.Type, errs)
	errs = validateTeams(d.Teams, errs)

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("Errors occurred while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, strings.Join(errs, "\n"))
}

func validateConfigVersion(version int, errs []string) []string {
	if version != 1 {
		errs = append(errs,
			"Unsupported or missing 'config_version' field\n")
	}
	return errs
}

func validateType(deploymentType string, errs []string) []string {
	if len(deploymentType) == 0 {
		errs = append(errs, "The 'type' field is required\n")
	} else if _, ok := validTypes[deploymentType]; !ok {
		errs = append(errs,
			fmt.Sprintf("'%s' is not a valid deployment type.", deploymentType))
	}
	return errs
}

func validateDeploymentTarget(cluster, datacenter string, tags, errs []string) []string {
	if len(cluster) > 0 {
		if len(datacenter) > 0 || len(tags) > 0 {
			errs = append(errs,
				"Exactly one of the 'cluster', 'datacenter', or 'tags' fields must be provided for every deployment\n")
		}
	} else if len(datacenter) > 0 {
		if len(tags) > 0 {
			errs = append(errs,
				"Exactly one of the 'cluster', 'datacenter', or 'tags' fields must be provided for every deployment\n")
		}
	} else if len(tags) == 0 {
		errs = append(errs,
			"Exactly one of the 'cluster', 'datacenter', or 'tags' fields must be provided for every deployment\n")
	}

	return errs
}

func validateName(name string, errs []string) []string {
	if len(name) == 0 {
		errs = append(errs, "The 'name' field is required\n")
	}
	return errs
}

func validateScaling(scaling *int, errs []string) []string {
	if scaling != nil && *scaling < 1 {
		errs = append(errs, "The 'scaling' field must be an integer >= 1\n")
	}
	return errs
}

func validateWiredTiger(wiredTiger bool, deploymentType string, errs []string) []string {
	if wiredTiger && deploymentType != "mongodb" {
		errs = append(errs,
			"The 'wired_tiger' field is only valid for the 'mongodb' deployment type\n")
	}
	return errs
}

func validateTeams(teams []*TeamV1, errs []string) []string {
	if teams == nil {
		return errs
	}
	for _, team := range teams {
		if len(team.ID) == 0 {
			errs = append(errs, "Every team entry requires an ID\n")
		}
		if _, ok := validRolesV1[team.Role]; ok {
			continue
		}
		errs = append(errs,
			fmt.Sprintf("'%s' is not a valid team role\n", team.Role))
	}
	return errs
}

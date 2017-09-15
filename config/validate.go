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

import "fmt"

var validRoles = map[string]struct{}{
	"admin":     {},
	"developer": {},
	"manager":   {},
}

var validTypes = map[string]struct{}{
	"mongodb":        {},
	"rethinkdb":      {},
	"elastic_search": {},
	"redis":          {},
	"postgresql":     {},
	"rabbitmq":       {},
	"etcd":           {},
	"mysql":          {},
	"janusgraph":     {},
	"disque":         {},
}

func validateType(deploymentType string) []string {
	errs := []string{}
	if len(deploymentType) == 0 {
		errs = append(errs, "The 'type' field is required")
	} else if _, ok := validTypes[deploymentType]; !ok {
		errs = append(errs,
			fmt.Sprintf("'%s' is not a valid deployment type.", deploymentType))
	}
	return errs
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

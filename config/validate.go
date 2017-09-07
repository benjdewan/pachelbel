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

func validateTeams(teams []*TeamV1, errs []string) []string {
	if teams == nil {
		return errs
	}
	for _, team := range teams {
		if len(team.ID) == 0 {
			errs = append(errs, "Every team entry requires an ID\n")
		}
		if _, ok := validRoles[team.Role]; ok {
			continue
		}
		errs = append(errs,
			fmt.Sprintf("'%s' is not a valid team role\n", team.Role))
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

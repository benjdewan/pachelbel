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

type DeploymentV1 struct {
	ConfigVersion int         `json:"config_version"`
	Version       string      `json:"version"`
	Type          string      `json:"type"`
	Cluster       string      `json:"cluster"`
	Datacenter    string      `json:"datacenter"`
	Name          string      `json:"name"`
	Notes         string      `json:"notes"`
	SSL           bool        `json:"ssl"`
	Teams         [](*TeamV1) `json:"teams"`
	Scaling       *int        `json:"scaling"`
	WiredTiger    bool        `json:"wired_tiger"`
	Timeout       *int        `json:"timeout,omitempty"`
}

type TeamV1 struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

func (d DeploymentV1) GetName() string {
	return d.Name
}

func (d DeploymentV1) GetNotes() string {
	return d.Notes
}

func (d DeploymentV1) GetType() string {
	return d.Type
}

func (d DeploymentV1) GetCluster() string {
	return d.Cluster
}

func (d DeploymentV1) GetDatacenter() string {
	return d.Datacenter
}

func (d DeploymentV1) GetVersion() string {
	return d.Version
}

func (d DeploymentV1) GetScaling() int {
	if d.Scaling == nil {
		return 1
	}
	return *d.Scaling
}

func (d DeploymentV1) GetTimeout() float64 {
	if d.Timeout == nil {
		return float64(900)
	}
	return float64(*d.Timeout)
}

func (d DeploymentV1) GetWiredTiger() bool {
	return d.Type == "mongodb" && d.WiredTiger
}

func (d DeploymentV1) GetSSL() bool {
	return d.SSL
}

func (d DeploymentV1) GetTeamRoles() map[string]([]string) {
	teamIDsByRole := make(map[string]([]string))
	for _, team := range d.Teams {
		if _, ok := teamIDsByRole[team.Role]; ok {
			teamIDsByRole[team.Role] = append(teamIDsByRole[team.Role],
				team.ID)
		} else {
			teamIDsByRole[team.Role] = []string{team.ID}
		}
	}
	return teamIDsByRole
}

var validRolesV1 = map[string]struct{}{
	"admin":     struct{}{},
	"developer": struct{}{},
	"manager":   struct{}{},
}

var validTypes = map[string]struct{}{
	"mongodb":       struct{}{},
	"rethinkdb":     struct{}{},
	"elasticsearch": struct{}{},
	"redis":         struct{}{},
	"postgresql":    struct{}{},
	"rabbitmq":      struct{}{},
	"etcd":          struct{}{},
	"mysql":         struct{}{},
	"janusgraph":    struct{}{},
}

func Validate(d DeploymentV1, input string) error {
	errs := []string{}
	valid := true
	if d.ConfigVersion != 1 {
		valid = false
		errs = append(errs, "Unsupported or missing 'config_version' field\n")
	}

	if len(d.Type) == 0 {
		valid = false
		errs = append(errs, "The 'type' field is required\n")
	}

	if _, ok := validTypes[d.Type]; !ok {
		valid = false
		errs = append(errs,
			fmt.Sprintf("'%s' is not a valid deployment type.", d.Type))
	}

	if len(d.Cluster) == 0 && len(d.Datacenter) == 0 {
		valid = false
		errs = append(errs,
			"Either a 'cluster' or 'datacenter' must be provided for every deployment\n")
	}

	if len(d.Cluster) > 0 && len(d.Datacenter) > 0 {
		valid = false
		errs = append(errs,
			"A 'cluster' and 'datacenter' cannot be provided for a single deployment\n")
	}

	if len(d.Name) == 0 {
		valid = false
		errs = append(errs, "The 'name' field is required\n")
	}

	if d.Scaling != nil && *d.Scaling < 1 {
		valid = false
		errs = append(errs, "The 'scaling' field must be an integer >= 1\n")
	}

	if d.WiredTiger && d.Type != "mongodb" {
		valid = false
		errs = append(errs,
			"The 'wired_tiger' field is only valid for the 'mongodb' deployment type\n")
	}

	if d.Teams != nil {
		for _, team := range d.Teams {
			if len(team.ID) == 0 {
				valid = false
				errs = append(errs,
					"Every team entry requires an ID\n")
			}
			if _, ok := validRolesV1[team.Role]; !ok {
				valid = false
				errs = append(errs,
					fmt.Sprintf("'%s' is not a valid team role\n",
						team.Role))
			}
		}
	}

	if valid {
		return nil
	}

	return fmt.Errorf("Errors occured while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, strings.Join(errs, "\n"))
}

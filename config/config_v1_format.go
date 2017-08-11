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

// DeploymentV1 is the structure corresponding to version 1 of
// pachelbel's configuration YAML
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

// TeamV1 is the structure corresponding to version 1 of pachelbel's team_roles
// configuration YAML
type TeamV1 struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// GetName returns the name of the deployment
func (d DeploymentV1) GetName() string {
	return d.Name
}

// GetNotes returns any notes associated with the deployment
func (d DeploymentV1) GetNotes() string {
	return d.Notes
}

// GetType returns the type of deployment
func (d DeploymentV1) GetType() string {
	return d.Type
}

// GetCluster returns the cluster name the deployment should live in or
// nothing if the deployment should go to a datacenter
func (d DeploymentV1) GetCluster() string {
	return d.Cluster
}

// GetDatacenter returns the datacenter name the deployment should live in
// or nothing if the deployment should live in a cluster
func (d DeploymentV1) GetDatacenter() string {
	return d.Datacenter
}

// GetVersion returns the database version the deployment should be deploying
func (d DeploymentV1) GetVersion() string {
	return d.Version
}

// GetScaling returns the database scaling value for the deployment, or 1
// if it does not exist.
func (d DeploymentV1) GetScaling() int {
	if d.Scaling == nil {
		return 1
	}
	return *d.Scaling
}

// GetTimeout returns the maximum timeout in seconds to wait on recipes for
// this deployment. The default is 300 seconds
func (d DeploymentV1) GetTimeout() float64 {
	if d.Timeout == nil {
		return float64(300)
	}
	return float64(*d.Timeout)
}

// GetWiredTiger is true if the deployment type is 'mongodb' and
// the wired_tiger engine has been enabled.
func (d DeploymentV1) GetWiredTiger() bool {
	return d.Type == "mongodb" && d.WiredTiger
}

// GetSSL returns true if SSL should be enabled for a deployment.
func (d DeploymentV1) GetSSL() bool {
	return d.SSL
}

// TeamEntryCount returns the number of team roles to apply to
// a given deployment.
func (d DeploymentV1) TeamEntryCount() int {
	return len(d.Teams)
}

// GetTeamRoles returns a a map of arrays of team roles to apply keyed by
// the team ID for those roles.
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

// Validate returns an error enumerating any and all issues with the
// provided deployment object to help with debugging.
func Validate(d DeploymentV1, input string) error {
	errs := []string{}

	errs = validateConfigVersion(d.ConfigVersion, errs)
	errs = validateType(d.Type, errs)
	errs = validateDeploymentTarget(d.Cluster, d.Datacenter, errs)
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

func validateDeploymentTarget(cluster, datacenter string, errs []string) []string {
	if len(cluster) == 0 && len(datacenter) == 0 {
		errs = append(errs,
			"Either a 'cluster' or 'datacenter' must be provided for every deployment\n")
	} else if len(cluster) > 0 && len(datacenter) > 0 {
		errs = append(errs,
			"A 'cluster' and 'datacenter' cannot be provided for a single deployment\n")
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

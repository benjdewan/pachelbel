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

// deploymentV1 is the structure corresponding to version 1 of
// pachelbel's configuration YAML
// codebeat:disable[TOO_MANY_IVARS]
type deploymentV1 struct {
	ConfigVersion int         `json:"config_version"`
	Version       string      `json:"version"`
	Type          string      `json:"type"`
	Cluster       string      `json:"cluster"`
	Datacenter    string      `json:"datacenter"`
	Tags          []string    `json:"tags"`
	Name          string      `json:"name"`
	Notes         string      `json:"notes"`
	SSL           bool        `json:"ssl"`
	Teams         [](*TeamV1) `json:"teams"`
	Scaling       *int        `json:"scaling"`
	WiredTiger    bool        `json:"wired_tiger"`
	Timeout       *int        `json:"timeout,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

// TeamV1 is the structure corresponding to version 1 of pachelbel's team_roles
// configuration YAML
type TeamV1 struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// GetName returns the name of the deployment
func (d deploymentV1) GetName() string {
	return d.Name
}

// GetNotes returns any notes associated with the deployment
func (d deploymentV1) GetNotes() string {
	return d.Notes
}

// GetType returns the type of deployment
func (d deploymentV1) GetType() string {
	return d.Type
}

// GetCluster returns the cluster name the deployment should live in or
// nothing if the deployment should go to a datacenter
func (d deploymentV1) GetCluster() string {
	return d.Cluster
}

// ClusterDeployment returns true if this deployment should be inside a cluster
func (d deploymentV1) ClusterDeployment() bool {
	return len(d.Cluster) > 0
}

// TagDeployment returns true if this deployment should be made using
// provisioning tags
func (d deploymentV1) TagDeployment() bool {
	return len(d.Tags) > 0
}

// GetTags returns the provisioning tags for this deployment
func (d deploymentV1) GetTags() []string {
	return d.Tags
}

// GetDatacenter returns the datacenter name the deployment should live in
// or nothing if the deployment should live in a cluster
func (d deploymentV1) GetDatacenter() string {
	return d.Datacenter
}

// GetVersion returns the database version the deployment should be deploying
func (d deploymentV1) GetVersion() string {
	return d.Version
}

// GetScaling returns the database scaling value for the deployment, or 1
// if it does not exist.
func (d deploymentV1) GetScaling() int {
	if d.Scaling == nil {
		return 1
	}
	return *d.Scaling
}

// GetTimeout returns the maximum timeout in seconds to wait on recipes for
// this deployment. The default is 300 seconds
func (d deploymentV1) GetTimeout() float64 {
	if d.Timeout == nil {
		return float64(300)
	}
	return float64(*d.Timeout)
}

// GetWiredTiger is true if the deployment type is 'mongodb' and
// the wired_tiger engine has been enabled.
func (d deploymentV1) GetWiredTiger() bool {
	return d.Type == "mongodb" && d.WiredTiger
}

// GetSSL returns true if SSL should be enabled for a deployment.
func (d deploymentV1) GetSSL() bool {
	return d.SSL
}

// TeamEntryCount returns the number of team roles to apply to
// a given deployment.
func (d deploymentV1) TeamEntryCount() int {
	return len(d.Teams)
}

// GetTeamRoles returns a a map of arrays of team roles to apply keyed by
// the team ID for those roles.
func (d deploymentV1) GetTeamRoles() map[string]([]string) {
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

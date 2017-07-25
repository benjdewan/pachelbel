package config

import (
	"bytes"
	"fmt"
)

type DeploymentV1 struct {
	Version int         `yaml:"version"`
	Type    string      `yaml:"type"`
	Cluster string      `yaml:"cluster"`
	Name    string      `yaml:"name"`
	Notes   string      `yaml:"notes"`
	SSL     bool        `yaml:"ssl"`
	Teams   [](*TeamV1) `yaml:"teams"`
	Scaling int         `yaml:"scaling"`
}

type TeamV1 struct {
	ID   string `yaml:"id"`
	Role string `yaml:"role"`
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

func (d DeploymentV1) GetScaling() int {
	return d.Scaling
}

func (d DeploymentV1) GetSSL() bool {
	return d.SSL
}

func (d DeploymentV1) GetTeams() map[string]([]string) {
	teamRolesByID := make(map[string]([]string))
	for _, team := range d.Teams {
		if _, ok := teamRolesByID[team.ID]; ok {
			teamRolesByID[team.ID] = append(teamRolesByID[team.ID],
				team.Role)
		} else {
			teamRolesByID[team.ID] = []string{team.Role}
		}
	}
	return teamRolesByID
}

var validRolesV1 = map[string]struct{}{
	"admin":     struct{}{},
	"developer": struct{}{},
	"manager":   struct{}{},
}

func Validate(d DeploymentV1, input string) error {
	var buf bytes.Buffer
	valid := true
	if d.Version != 1 {
		valid = false
		addToBuf(&buf, "Unsupported or missing version field\n")
	}
	if len(d.Type) == 0 {
		valid = false
		addToBuf(&buf, "The 'type' field is required\n")
	}
	if len(d.Cluster) == 0 {
		valid = false
		addToBuf(&buf, "The 'cluster' field is required\n")
	}
	if len(d.Name) == 0 {
		valid = false
		addToBuf(&buf, "The 'name' field is required\n")
	}

	if d.Scaling < 0 {
		valid = false
		addToBuf(&buf, "The 'scaling' field must be an integer >= 1\n")
	}

	if d.Teams != nil {
		for _, team := range d.Teams {
			if len(team.ID) == 0 {
				valid = false
				addToBuf(&buf, "Every team entry requires an ID\n")
			}
			if _, ok := validRolesV1[team.Role]; !ok {
				valid = false
				addToBuf(&buf,
					fmt.Sprintf("'%s' is not a valid team role\n",
						team.Role))
			}
		}
	}

	if valid {
		return nil
	}

	return fmt.Errorf("Errors occured while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, buf.String())
}

func addToBuf(buf *bytes.Buffer, msg string) {
	if _, err := (*buf).WriteString(msg); err != nil {
		panic(err)
	}
}

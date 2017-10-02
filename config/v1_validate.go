package config

import (
	"fmt"
	"strings"
)

func validateV1(d deploymentV1, input string) (deploymentV1, error) {
	errs := []string{}

	errs = append(errs, validateConfigVersionV1(d.ConfigVersion)...)
	errs = append(errs, validateVersionByType(d.Version, d.Type)...)
	errs = append(errs, validateDeploymentTargetV1(d.Cluster, d.Datacenter, d.Tags)...)
	errs = append(errs, validateClusterV1(&d)...)
	errs = append(errs, validateDatacenterV1(d)...)
	errs = append(errs, validateName(d.Name)...)
	errs = append(errs, validateScaling(d.Scaling)...)
	errs = append(errs, validateWiredTiger(d.WiredTiger, d.Type)...)
	errs = append(errs, validateCacheMode(d.CacheMode, d.Type)...)
	errs = append(errs, validateTeams(d.Teams)...)

	if len(errs) == 0 {
		return d, nil
	}

	return d, fmt.Errorf("Errors occurred while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, strings.Join(errs, "\n"))
}

func validateConfigVersionV1(version int) []string {
	if version != 1 {
		return []string{"Unsupported or missing 'config_version' field"}
	}
	return []string{}
}

func validateClusterV1(d *deploymentV1) []string {
	if len(d.Cluster) == 0 {
		return []string{}
	} else if id, ok := Clusters[d.Cluster]; ok {
		d.Cluster = id
		return []string{}
	}

	return []string{fmt.Sprintf("Cannot find the specified cluster '%s'.",
		d.Cluster)}
}

func validateDatacenterV1(d deploymentV1) []string {
	if len(d.Datacenter) == 0 {
		return []string{}
	} else if _, ok := Datacenters[d.Datacenter]; ok {
		return []string{}
	}
	return []string{fmt.Sprintf("Cannot find the specified datacenter '%s'.",
		d.Datacenter)}
}

func validateDeploymentTargetV1(cluster, datacenter string, tags []string) []string {
	if xor3(len(cluster) > 0, len(datacenter) > 0, len(tags) > 0) {
		return []string{}
	}
	return []string{"Exactly one of the 'cluster', 'datacenter', or 'tags' fields must be provided for every deployment\n"}
}

func xor3(a, b, c bool) bool {
	return xor(xor(a, b), c) && !(a && b && c)
}

func xor(a, b bool) bool {
	return (!(a && b) && (!(!a && !b)))
}

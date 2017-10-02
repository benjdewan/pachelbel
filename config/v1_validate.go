package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/masterminds/semver"
)

func validateV1(d deploymentV1, input string) (deploymentV1, error) {
	errs := []string{}

	errs = append(errs, validateConfigVersionV1(d.ConfigVersion)...)
	errs = append(errs, validateVersionByTypeV1(&d)...)
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

func validateVersionByTypeV1(d *deploymentV1) []string {
	errs := []string{}
	if len(d.Type) == 0 {
		errs = append(errs, "The 'type' field is required")
	} else if versions, ok := Databases[d.Type]; ok {
		errs = append(errs, validateVersionV1(d, versions)...)
	} else {
		errs = append(errs,
			fmt.Sprintf("'%s' is not a valid deployment type", d.Type))
	}
	return errs
}

func validateVersionV1(d *deploymentV1, rawVersions []string) []string {
	errs := []string{}
	if len(d.Version) == 0 {
		return errs
	}

	constraint, versions, errs := toSemVer(d.Version, rawVersions)
	for _, version := range versions {
		if constraint.Check(version) {
			d.resolvedVersion = version.String()
			return errs
		}
	}
	return append(errs,
		fmt.Sprintf("'%s' has no version equalling or satisfying '%s'",
			d.Type, d.Version))
}

func toSemVer(rawConstraint string, rawVersions []string) (*semver.Constraints, []*semver.Version, []string) {
	errs := []string{}
	constraint, err := semver.NewConstraint(rawConstraint)
	if err != nil {
		return nil, nil, append(errs, fmt.Sprintf("Could not parse '%s': %v", rawConstraint, err))
	}

	versions := make([]*semver.Version, len(rawVersions))
	for i, raw := range rawVersions {
		version, err := semver.NewVersion(raw)
		if err != nil {
			return nil, nil, append(errs, fmt.Sprintf("Could not parse '%s': %v", version, err))
		}
		versions[i] = version
	}
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	return constraint, versions, errs
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

package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/benjdewan/pachelbel/connection"
	"github.com/benjdewan/pachelbel/runner"
	"github.com/masterminds/semver"
)

func validateV1(d deploymentV1, input string) (runner.Runner, error) {
	errs := []string{}

	errs = append(errs, validateConfigVersionV1(d.ConfigVersion)...)
	errs = append(errs, validateName(d.Name)...)
	errs = append(errs, validateTeams(d.Teams)...)
	errs = append(errs, validateDeploymentTargetV1(d.Cluster, d.Datacenter, d.Tags)...)
	errs = append(errs, validateClusterV1(&d)...)
	errs = append(errs, validateDatacenterV1(d)...)
	errs = append(errs, validateWiredTiger(d.WiredTiger, d.Type)...)
	errs = append(errs, validateCacheMode(d.CacheMode, d.Type)...)

	if existing, ok := existingDeployment(d.Name); ok {
		return validateExistingV1(d, existing, input, errs)
	}

	errs = append(errs, validateVersionByTypeV1(&d)...)
	errs = append(errs, validateScaling(d.Scaling)...)

	deploymentRunner := runner.Runner{
		Target: runner.Accessor(d),
		Action: runner.ActionCreate,
		Run:    runner.Create,
	}
	if len(errs) == 0 {
		return deploymentRunner, nil
	}

	return deploymentRunner, fmt.Errorf("Errors occurred while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, strings.Join(errs, "\n"))
}

func validateExistingScalingV1(d deploymentV1, existing connection.ExistingDeployment, errs []string) ([]string, []string) {
	actions := []string{}
	scaling := 0
	if d.Scaling == nil || *d.Scaling == existing.Scaling {
		d.Scaling = &scaling
	} else if *d.Scaling < existing.UtilizedScaling {
		d.Scaling = &scaling
		_, err := fmt.Fprintf(os.Stderr,
			"WARNING: %s currently utilizes %d units, but only %d are specified. Ignoring specified units",
			d.Name, existing.UtilizedScaling, *d.Scaling)
		if err != nil {
			errs = append(errs, fmt.Sprintf("Internal error while rendering configuration: %v", err))
		}
	} else {
		actions = append(actions, runner.ActionResize)
		errs = append(errs, validateScaling(d.Scaling)...)
	}
	return actions, errs
}

func validateExistingV1(d deploymentV1, existing connection.ExistingDeployment, input string, errs []string) (runner.Runner, error) {
	d.id = existing.ID

	actions, sErrs := validateExistingScalingV1(d, existing, errs)
	errs = append(errs, sErrs...)

	if d.Version == existing.Version {
		d.Version = ""
	} else if versionEquivalence(d.Version, existing.Version) {
		d.Version = ""
	} else {
		vErrs := validateVersionUpgradeV1(&d, existing.Upgrades)
		if len(vErrs) == 0 && !d.Upgradeable {
			fmt.Printf("A version upgrade for '%s' exists: '%s'\n", d.Name, d.Version)
			d.Version = ""
		} else {
			actions = append(actions, runner.ActionUpgrade)
		}
		errs = append(errs, vErrs...)
	}

	if d.Notes == existing.Notes {
		d.Notes = ""
	} else if d.Notes != "" {
		actions = append(errs, runner.ActionComment)
	}
	action, runFunc := toAction(actions)
	deploymentRunner := runner.Runner{
		Target: runner.Accessor(d),
		Action: action,
		Run:    runFunc,
	}
	if len(errs) == 0 {
		return deploymentRunner, nil
	}

	return deploymentRunner, fmt.Errorf("Errors occurred while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, strings.Join(errs, "\n"))
}

func versionEquivalence(requested, existing string) bool {
	if len(requested) == 0 {
		return true
	}

	constraint, err := semver.NewConstraint(requested)
	if err != nil {
		return false
	}

	existingVersion, err := semver.NewVersion(existing)
	if err != nil {
		return false
	}

	return constraint.Check(existingVersion)
}

func toAction(actions []string) (string, runner.RunFunc) {
	switch len(actions) {
	case 0:
		return runner.ActionLookup, runner.Lookup
	case 1:
		return actions[0], runner.Update
	default:
		action := strings.Join([]string{strings.Join(actions[:len(actions)-1], ", "), actions[len(actions)-1]}, " and ")
		return action, runner.Update
	}
}

func validateVersionUpgradeV1(d *deploymentV1, versions []*semver.Version) []string {
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	constraint, err := semver.NewConstraint(d.Version)
	if err != nil {
		return []string{fmt.Sprintf("Cannot parse '%s' as a version constraint: %v", d.Version, err)}
	}
	for _, version := range versions {
		if constraint.Check(version) {
			d.Version = version.String()
			return []string{}
		}
	}
	return []string{fmt.Sprintf("No valid version upgrades exist given '%s'", d.Version)}
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
			d.Version = version.String()
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

package connection

import (
	"fmt"
	"sort"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/masterminds/semver"
)

func update(cxn *Connection, accessor Accessor) error {
	deployment := accessor.(Deployment)
	existing, ok := cxn.getDeploymentByName(deployment.GetName())
	if !ok {
		return fmt.Errorf("Attempting to update '%s', but it doesn't exist",
			deployment.GetName())
	}

	if err := updateScalings(cxn, existing.ID, deployment); err != nil {
		return err
	}

	if err := cxn.attemptNotesUpdate(existing, deployment); err != nil {
		return err
	}

	if err := cxn.attemptVersionUpdate(existing, deployment); err != nil {
		return err
	}

	if err := addTeamRoles(cxn, existing.ID, deployment.GetTeamRoles()); err != nil {
		return err
	}

	// Changing versions and sizes can change the deployment ID. Ensure
	// we have the latest/live value
	updatedDeployment, errs := cxn.client.GetDeploymentByName(existing.Name)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get deployment information for '%s':\n%v",
			deployment.GetName(), errs)
	}
	cxn.newDeploymentIDs.Store(updatedDeployment.ID, struct{}{})
	return nil
}

func updateScalings(cxn *Connection, id string, deployment Deployment) error {
	existingScalings, errs := cxn.client.GetScalings(id)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get current scaling for '%s':\n%v",
			deployment.GetName(), errs)
	}

	if existingScalings.AllocatedUnits == deployment.GetScaling() {
		return nil
	}

	recipe, errs := cxn.client.SetScalings(compose.ScalingsParams{
		DeploymentID: id,
		Units:        deployment.GetScaling(),
	})
	if len(errs) != 0 {
		return fmt.Errorf("Unable to resize '%s':\n%v",
			deployment.GetName(), errs)
	}

	return cxn.waitOnRecipe(recipe.ID, deployment.GetTimeout())
}

func (cxn *Connection) attemptNotesUpdate(existing *compose.Deployment, deployment Deployment) error {
	if len(deployment.GetNotes()) == 0 || existing.Notes == deployment.GetNotes() {
		return nil
	}

	_, errs := cxn.client.PatchDeployment(compose.PatchDeploymentParams{
		DeploymentID: existing.ID,
		Notes:        deployment.GetNotes(),
	})
	if len(errs) != 0 {
		return fmt.Errorf("Unable to update notes for '%s':\n%v",
			existing.Name, errs)
	}
	return nil
}

func (cxn *Connection) attemptVersionUpdate(existing *compose.Deployment, deployment Deployment) error {
	if len(deployment.GetVersionConstraint()) == 0 {
		return nil
	}
	constraint, err := semver.NewConstraint(deployment.GetVersionConstraint())
	if err != nil {
		return fmt.Errorf("Unable to parse '%s': %v", deployment.GetVersionConstraint(), err)
	}

	if withinConstraint(constraint, existing.Version) {
		return nil
	}

	if version, ok := cxn.upgradeExistsForConstraint(constraint, existing.ID); ok {
		return cxn.updateVersion(version, existing.ID, deployment.GetTimeout())
	}

	return fmt.Errorf("There is no valid version transition for '%s' that satisfies '%s'", existing.Name, deployment.GetVersionConstraint())
}

func (cxn *Connection) updateVersion(version, id string, timeout float64) error {
	recipe, errs := cxn.client.UpdateVersion(id, version)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to upgrade %s to version %s:\n%v", id, version, errs)
	}
	return cxn.waitOnRecipe(recipe.ID, timeout)
}

func (cxn *Connection) upgradeExistsForConstraint(constraint *semver.Constraints, id string) (string, bool) {
	transitions, errs := cxn.client.GetVersionsForDeployment(id)
	if len(errs) != 0 || transitions == nil {
		return "", false
	}
	return transitionWithinConstraint(constraint, *transitions)
}

func transitionWithinConstraint(constraint *semver.Constraints, transitions []compose.VersionTransition) (string, bool) {
	possibles := []string{}
	for _, transition := range transitions {
		if transition.Method != "in_place" {
			continue
		}
		if withinConstraint(constraint, transition.ToVersion) {
			possibles = append(possibles, transition.ToVersion)
		}
	}
	return mostRecentVersion(possibles), len(possibles) > 0
}

func withinConstraint(constraint *semver.Constraints, raw string) bool {
	version, err := semver.NewVersion(raw)
	if err != nil {
		return false
	}
	return constraint.Check(version)
}

func mostRecentVersion(rawVersions []string) string {
	if len(rawVersions) == 0 {
		return ""
	}

	versions := make([]*semver.Version, len(rawVersions))
	for i, r := range rawVersions {
		version, err := semver.NewVersion(r)
		if err != nil {
			continue
		}
		versions[i] = version
	}
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	return versions[0].String()
}

func dryRunUpdate(cxn *Connection, accessor Accessor) error {
	deployment, ok := cxn.getDeploymentByName(accessor.GetName())
	if !ok {
		// This should never happen
		panic("map integrity failure")
	}
	cxn.newDeploymentIDs.Store(deployment.ID, struct{}{})
	return nil
}

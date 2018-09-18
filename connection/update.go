package connection

import (
	"fmt"

	compose "github.com/benjdewan/gocomposeapi"
)

// UpdateScaling does nothing if the provided scaling is 0, but
// otherwise will make an API call to rescale a deployment and
// then wait on the returned recipe until it completes or the
// timeout is exceeded.
func (cxn *Connection) UpdateScaling(deployment Deployment) error {
	if deployment.GetScaling() <= 1 {
		return nil
	}

	recipe, errs := cxn.client.SetScalings(compose.ScalingsParams{
		DeploymentID: deployment.GetID(),
		Units:        deployment.GetScaling(),
	})
	if len(errs) != 0 {
		return fmt.Errorf("Unable to resize '%s':\n%v", deployment.GetName(), errs)
	}

	return cxn.wait(recipe.ID, deployment.GetTimeout())
}

// UpdateNotes does nothing if the deployment notes field is blank,
// but otherwise makes an API call to replace the notes on the given
// deployment with the new ones
func (cxn *Connection) UpdateNotes(deployment Deployment) error {
	if len(deployment.GetNotes()) == 0 {
		return nil
	}

	_, errs := cxn.client.PatchDeployment(compose.PatchDeploymentParams{
		DeploymentID: deployment.GetID(),
		Notes:        deployment.GetNotes(),
	})
	if len(errs) != 0 {
		return fmt.Errorf("Unable to update notes on '%s':\n%v", deployment.GetName(), errs)
	}
	return nil
}

// UpdateVersion does nothing if the provided deployment version field
// is blank. Otherwise it will make an API call to Compose to trigger
// an in-place upgrade. Pachelbel can *only* perform in-place upgrades
func (cxn *Connection) UpdateVersion(deployment Deployment) error {
	if len(deployment.GetVersion()) == 0 {
		return nil
	}

	recipe, errs := cxn.client.UpdateVersion(deployment.GetID(), deployment.GetVersion())
	if len(errs) != 0 {
		return fmt.Errorf("Unable to upgrade %s to version %s:\n%v", deployment.GetName(), deployment.GetVersion(), errs)
	}
	return cxn.wait(recipe.ID, deployment.GetTimeout())

}

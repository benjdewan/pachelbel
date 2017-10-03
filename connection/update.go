package connection

import (
	"fmt"

	compose "github.com/benjdewan/gocomposeapi"
)

func (cxn *Connection) UpdateScaling(deployment Deployment) error {
	if deployment.GetScaling() == 0 {
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

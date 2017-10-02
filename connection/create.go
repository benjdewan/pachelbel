package connection

import (
	"fmt"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/benjdewan/pachelbel/output"
)

func create(cxn *Connection, accessor Accessor) error {
	deployment := accessor.(Deployment)

	newDeployment, errs := cxn.client.CreateDeployment(deploymentParams(deployment,
		cxn.accountID))
	if len(errs) != 0 {
		return fmt.Errorf("Unable to create '%s': %v\n",
			deployment.GetName(), errs)
	}

	if err := cxn.waitOnRecipe(newDeployment.ProvisionRecipeID, deployment.GetTimeout()); err != nil {
		return err
	}

	if err := addTeamRoles(cxn, newDeployment.ID, deployment.GetTeamRoles()); err != nil {
		return err
	}
	cxn.newDeploymentIDs.Store(newDeployment.ID, struct{}{})

	return nil
}

func deploymentParams(deployment Deployment, accountID string) compose.DeploymentParams {
	dParams := compose.DeploymentParams{
		Name:         deployment.GetName(),
		AccountID:    accountID,
		DatabaseType: deployment.GetType(),
		Notes:        deployment.GetNotes(),
		SSL:          deployment.GetSSL(),
		Version:      deployment.GetVersion(),
		CacheMode:    deployment.GetCacheMode(),
		WiredTiger:   deployment.GetWiredTiger(),
		Units:        deployment.GetScaling(),
	}

	return setDeploymentType(deployment, dParams)
}

func setDeploymentType(deployment Deployment, dParams compose.DeploymentParams) compose.DeploymentParams {
	if deployment.TagDeployment() {
		dParams.ProvisioningTags = deployment.GetTags()
	} else if deployment.ClusterDeployment() {
		dParams.ClusterID = deployment.GetCluster()
	} else {
		dParams.Datacenter = deployment.GetDatacenter()
	}

	return dParams
}

func dryRunCreate(cxn *Connection, accessor Accessor) error {
	cxn.newDeploymentIDs.Store(output.FakeID(accessor.GetType(), accessor.GetName()),
		struct{}{})
	return nil
}

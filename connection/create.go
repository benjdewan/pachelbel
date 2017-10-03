package connection

import (
	"fmt"

	compose "github.com/benjdewan/gocomposeapi"
)

// CreateDeployment creates a deployment in Compose and returns it on success
func (cxn *Connection) CreateDeployment(d Deployment) (*compose.Deployment, error) {
	newDeployment, errs := cxn.client.CreateDeployment(deploymentParams(d, cxn.accountID))
	if len(errs) != 0 {
		return nil, fmt.Errorf("Unable to create '%s': %v\n",
			d.GetName(), errs)
	}

	return newDeployment, cxn.wait(newDeployment.ProvisionRecipeID, d.GetTimeout())
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

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

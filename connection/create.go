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
)

func create(cxn *Connection, accessor Accessor) error {
	deployment := accessor.(Deployment)
	dParams, err := deploymentParams(deployment, cxn)
	if err != nil {
		return err
	}

	newDeployment, errs := cxn.client.CreateDeployment(dParams)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to create '%s': %v\n", dParams.Name, errs)
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

func deploymentParams(deployment Deployment, cxn *Connection) (compose.DeploymentParams, error) {
	dParams := compose.DeploymentParams{
		Name:         deployment.GetName(),
		AccountID:    cxn.accountID,
		DatabaseType: deployment.GetType(),
		Notes:        deployment.GetNotes(),
	}

	dParams, err := setDeploymentType(cxn, deployment, dParams)
	if err != nil {
		return dParams, err
	}

	if len(deployment.GetVersion()) > 0 {
		dParams.Version = deployment.GetVersion()
	}

	if deployment.GetWiredTiger() {
		dParams.WiredTiger = true
	}

	if deployment.GetScaling() > 1 {
		dParams.Units = deployment.GetScaling()
	}

	if deployment.GetSSL() {
		dParams.SSL = true
	}

	return dParams, nil
}

func setDeploymentType(cxn *Connection, deployment Deployment, dParams compose.DeploymentParams) (compose.DeploymentParams, error) {
	if deployment.TagDeployment() {
		//TODO fetch tags and validate that the provided tags exist.
		dParams.ProvisioningTags = deployment.GetTags()
		return dParams, nil
	}

	if deployment.ClusterDeployment() {
		clusterID, ok := cxn.clusterIDsByName[deployment.GetCluster()]
		if !ok {
			return dParams, fmt.Errorf("Unable to provsion '%s'. The specified cluster name, '%s' does not map to a known cluster.",
				deployment.GetName(), deployment.GetCluster())
		}
		dParams.ClusterID = clusterID
		return dParams, nil
	}

	datacenter := deployment.GetDatacenter()
	if _, ok := cxn.datacenters[datacenter]; !ok {
		return dParams, fmt.Errorf("Unable to provision '%s'. '%s' is not a known datacenter.",
			deployment.GetName(), datacenter)
	}
	dParams.Datacenter = datacenter
	return dParams, nil
}

func dryRunCreate(cxn *Connection, accessor Accessor) error {
	cxn.newDeploymentIDs.Store(fakeID(accessor.GetType(), accessor.GetName()),
		struct{}{})
	return nil
}

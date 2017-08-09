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
	"github.com/golang-collections/go-datastructures/queue"
)

func provision(cxn *Connection, deployment Deployment, errQueue *queue.Queue) {
	fmt.Printf("Provisioning '%s'...\n", deployment.GetName())

	dParams, err := deploymentParams(deployment, cxn)
	if err != nil {
		return
	}

	//This needs to be wrapped in retry logic
	newDeployment, errs := cxn.client.CreateDeployment(dParams)
	if len(errs) != 0 {
		enqueue(errQueue, fmt.Errorf("Unable to create '%s': %s\n", deployment.GetName(), errsOut(errs)))
		return
	}

	if err := cxn.waitOnRecipe(newDeployment.ProvisionRecipeID, deployment.GetTimeout()); err != nil {
		enqueue(errQueue, err)
		return
	}

	if err := addTeamRoles(cxn, newDeployment.ID, deployment.GetTeamRoles()); err != nil {
		enqueue(errQueue, err)
		return
	}

	fmt.Printf("Provision of '%s' is complete!\n", newDeployment.Name)
	cxn.newDeploymentIDs.Store(newDeployment.ID, struct{}{})

	return
}

func deploymentParams(deployment Deployment, cxn *Connection) (compose.DeploymentParams, error) {
	dParams := compose.DeploymentParams{
		Name:         deployment.GetName(),
		AccountID:    cxn.accountID,
		DatabaseType: deployment.GetType(),
		Notes:        deployment.GetNotes(),
	}

	if len(deployment.GetDatacenter()) > 0 {
		dParams.Datacenter = deployment.GetDatacenter()
	} else {
		clusterID, ok := cxn.clusterIDsByName[deployment.GetCluster()]
		if !ok {
			return dParams, fmt.Errorf("Unable to provsion '%s'. The specified cluster name, '%s' does not map to a known cluster.",
				deployment.GetName(), deployment.GetCluster())
		}
		dParams.ClusterID = clusterID
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

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
	"github.com/benjdewan/pachelbel/progress"
)

type deployFunc func(*Connection, Deployment) error

type composeDeployer struct {
	deployment Deployment
	run        deployFunc
}

func (cxn *Connection) listDeployers(deployments []Deployment) []composeDeployer {
	deployers := []composeDeployer{}
	for _, deployment := range deployments {
		deployers = append(deployers, cxn.newDeployer(deployment))
	}
	return deployers
}

func (cxn *Connection) newDeployer(deployment Deployment) composeDeployer {
	if _, ok := cxn.getDeploymentByName(deployment.GetName()); ok {
		return cxn.newUpdateDeployer(deployment)
	}
	return cxn.newCreateDeployer(deployment)
}

func (cxn *Connection) newCreateDeployer(deployment Deployment) composeDeployer {
	if cxn.dryRun {
		cxn.pb.AddBar(progress.ActionDryRunCreate, deployment.GetName())
		return composeDeployer{
			deployment: deployment,
			run:        dryRunCreate,
		}
	}
	cxn.pb.AddBar(progress.ActionCreate, deployment.GetName())
	return composeDeployer{
		deployment: deployment,
		run:        create,
	}
}

func (cxn *Connection) newUpdateDeployer(deployment Deployment) composeDeployer {
	if cxn.dryRun {
		cxn.pb.AddBar(progress.ActionDryRunUpdate, deployment.GetName())
		return composeDeployer{
			deployment: deployment,
			run:        dryRunUpdate,
		}
	}
	cxn.pb.AddBar(progress.ActionUpdate, deployment.GetName())
	return composeDeployer{
		deployment: deployment,
		run:        update,
	}
}

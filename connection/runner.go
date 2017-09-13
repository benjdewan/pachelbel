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
	compose "github.com/benjdewan/gocomposeapi"
	"github.com/benjdewan/pachelbel/progress"
)

type runFunc func(*Connection, Accessor) error

type cxnRunner struct {
	accessor Accessor
	run      runFunc
}

func (cxn *Connection) newRunners(accessors []Accessor) []cxnRunner {
	if cxn.dryRun {
		return cxn.newDryRunners(accessors)
	}
	runners := []cxnRunner{}
	for _, accessor := range accessors {
		runners = append(runners, cxnRunner{
			accessor: accessor,
			run:      cxn.assignRunFunc(accessor),
		})
	}
	return runners
}

func (cxn *Connection) newDryRunners(accessors []Accessor) []cxnRunner {
	runners := []cxnRunner{}
	for _, accessor := range accessors {
		runners = append(runners, cxnRunner{
			accessor: accessor,
			run:      cxn.assignDryRunFunc(accessor),
		})
	}
	return runners
}

func (cxn *Connection) assignRunFunc(accessor Accessor) runFunc {
	if !accessor.IsOwner() {
		cxn.pb.AddBar(progress.ActionLookup, accessor.GetName())
		return lookup
	}
	deployment, _ := cxn.client.GetDeploymentByName(accessor.GetName())
	return cxn.assignOwnerRunFunc(accessor.GetName(), deployment)

}

func (cxn *Connection) assignOwnerRunFunc(name string, deployment *compose.Deployment) runFunc {
	if deployment == nil {
		cxn.pb.AddBar(progress.ActionCreate, name)
		return create
	}
	cxn.pb.AddBar(progress.ActionUpdate, deployment.Name)
	// Cache this deployment struct for later reference
	cxn.deploymentsByName.Store(deployment.Name, deployment)
	return update
}

func (cxn *Connection) assignDryRunFunc(accessor Accessor) runFunc {
	if !accessor.IsOwner() {
		cxn.pb.AddBar(progress.ActionDryRunLookup, accessor.GetName())
		return dryRunLookup
	}
	deployment, _ := cxn.client.GetDeploymentByName(accessor.GetName())
	return cxn.assignOwnerDryRunFunc(accessor.GetName(), deployment)
}

func (cxn *Connection) assignOwnerDryRunFunc(name string, deployment *compose.Deployment) runFunc {
	if deployment == nil {
		cxn.pb.AddBar(progress.ActionDryRunCreate, name)
		return dryRunCreate
	}

	cxn.pb.AddBar(progress.ActionDryRunUpdate, deployment.Name)
	// Cache this deployment struct for later reference
	cxn.deploymentsByName.Store(deployment.Name, deployment)
	return dryRunUpdate
}

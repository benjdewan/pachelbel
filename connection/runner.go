package connection

import (
	"sync"

	compose "github.com/benjdewan/gocomposeapi"
)

const (
	actionUpdate            = "Updating"
	actionCreate            = "Creating"
	actionLookup            = "Looking up"
	actionDeprovision       = "Deprovisioning"
	actionDryRunUpdate      = "Pretending to update"
	actionDryRunCreate      = "Pretending to create"
	actionDryRunLookup      = "Pretending to lookup"
	actionDryRunDeprovision = "Pretending to deprovision"
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
	var wg sync.WaitGroup
	wg.Add(len(accessors))
	runners := make([]cxnRunner, len(accessors))

	for index, accessor := range accessors {
		go func(i int, a Accessor) {
			runners[i] = cxnRunner{
				accessor: a,
				run:      cxn.assignRunFunc(a),
			}
			wg.Done()
		}(index, accessor)
	}
	wg.Wait()
	return runners
}

func (cxn *Connection) newDryRunners(accessors []Accessor) []cxnRunner {
	var wg sync.WaitGroup
	wg.Add(len(accessors))
	runners := make([]cxnRunner, len(accessors))
	for index, accessor := range accessors {
		go func(i int, a Accessor) {
			runners[i] = cxnRunner{
				accessor: a,
				run:      cxn.assignDryRunFunc(a),
			}
			wg.Done()
		}(index, accessor)
	}
	wg.Wait()
	return runners
}

func (cxn *Connection) assignRunFunc(accessor Accessor) runFunc {
	if !accessor.IsOwner() {
		cxn.pb.AddBar(actionLookup, accessor.GetName())
		return lookup
	} else if accessor.IsDeleter() {
		cxn.pb.AddBar(actionDeprovision, accessor.GetName())
		return deprovision
	}
	deployment, _ := cxn.client.GetDeploymentByName(accessor.GetName())
	return cxn.assignOwnerRunFunc(accessor.GetName(), deployment)

}

func (cxn *Connection) assignOwnerRunFunc(name string, deployment *compose.Deployment) runFunc {
	if deployment == nil {
		cxn.pb.AddBar(actionCreate, name)
		return create
	}
	cxn.pb.AddBar(actionUpdate, deployment.Name)
	// Cache this deployment struct for later reference
	cxn.deploymentsByName.Store(deployment.Name, deployment)
	return update
}

func (cxn *Connection) assignDryRunFunc(accessor Accessor) runFunc {
	if !accessor.IsOwner() {
		cxn.pb.AddBar(actionDryRunLookup, accessor.GetName())
		return dryRunLookup
	} else if accessor.IsDeleter() {
		cxn.pb.AddBar(actionDryRunDeprovision, accessor.GetName())
		return dryRunDeprovision
	}
	deployment, _ := cxn.client.GetDeploymentByName(accessor.GetName())
	return cxn.assignOwnerDryRunFunc(accessor.GetName(), deployment)
}

func (cxn *Connection) assignOwnerDryRunFunc(name string, deployment *compose.Deployment) runFunc {
	if deployment == nil {
		cxn.pb.AddBar(actionDryRunCreate, name)
		return dryRunCreate
	}

	cxn.pb.AddBar(actionDryRunUpdate, deployment.Name)
	// Cache this deployment struct for later reference
	cxn.deploymentsByName.Store(deployment.Name, deployment)
	return dryRunUpdate
}

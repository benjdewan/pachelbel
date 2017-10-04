package runner

import (
	"log"
	"strings"

	"github.com/benjdewan/pachelbel/connection"
	"github.com/benjdewan/pachelbel/output"
)

const (
	// ActionLookup indicates we are doing a non-modifying lookup operation
	ActionLookup = "Looking up"
	// ActionCreate indicates we are creating a new deployment
	ActionCreate = "Creating"
	// ActionResize indicates we are resizing a deployment
	ActionResize = "Resizing"
	// ActionUpgrade indicates we are upgrading a deployment
	ActionUpgrade = "Upgrading"
	// ActionComment indicates we are updating the notes on a deployment
	ActionComment = "Commenting on"
	// ActionDeprovision indicates we are deprovisioning a deployment
	ActionDeprovision = "Deprovisioning"
)

// RunFunc is the signature of actions a Runner object can take like
// Create, Update, Lookup or Deprovision
type RunFunc func(*connection.Connection, Accessor) error

// Runner is an individual deployment operation
type Runner struct {
	Target Accessor
	Action string
	Run    RunFunc
}

// Deprovision is the RunFunc for deprovisioning a deployment
func Deprovision(cxn *connection.Connection, accessor Accessor) error {
	return cxn.Deprovision(accessor.(connection.Deprovision))
}

// Update is the RunFunc for updating a deployment if there is anything
// that can be updated
func Update(cxn *connection.Connection, accessor Accessor) error {
	deployment := accessor.(connection.Deployment)

	if err := cxn.UpdateScaling(deployment); err != nil {
		return err
	}

	if err := cxn.UpdateVersion(deployment); err != nil {
		return err
	}

	if err := cxn.UpdateNotes(deployment); err != nil {
		return err
	}

	if err := cxn.AddTeams(deployment.GetID(), deployment); err != nil {
		return err
	}

	return cxn.GetAndAdd(deployment.GetName())
}

// Create is a RunFunc for creating a new deployments
func Create(cxn *connection.Connection, accessor Accessor) error {
	deployment := accessor.(connection.Deployment)

	newDeployment, err := cxn.CreateDeployment(deployment)
	if err != nil {
		return err
	}

	if err := cxn.AddTeams(newDeployment.ID, deployment); err != nil {
		return err
	}
	cxn.Add(newDeployment.ID)
	return nil
}

// Lookup is the RunFunc for looking up existing deployments
func Lookup(cxn *connection.Connection, accessor Accessor) error {
	return cxn.GetAndAdd(accessor.GetName())
}

func dryRunLookup(cxn *connection.Connection, accessor Accessor) error {
	if err := cxn.GetAndAdd(accessor.GetName()); err != nil {
		cxn.Add(output.FakeID(accessor.GetType(), accessor.GetName()))
	}
	return nil
}

func dryRunCreate(cxn *connection.Connection, accessor Accessor) error {
	cxn.Add(output.FakeID(accessor.GetType(), accessor.GetName()))
	return nil
}

func dryRunUpdate(cxn *connection.Connection, accessor Accessor) error {
	cxn.Add(accessor.(connection.Deployment).GetID())
	return nil
}

func dryRunDeprovision(cxn *connection.Connection, accessor Accessor) error {
	return nil
}

func toDryRun(action string) RunFunc {
	switch action {
	case ActionLookup:
		return dryRunLookup
	case ActionCreate:
		return dryRunCreate
	case ActionDeprovision:
		return dryRunDeprovision
	default:
		if strings.Contains(action, ActionUpgrade) || strings.Contains(action, ActionResize) || strings.Contains(action, ActionComment) {
			return dryRunUpdate
		}
		log.Panicf("Unknown action: %s", action)
	}
	panic("unreachable code")
}

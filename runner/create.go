package runner

import (
	"github.com/benjdewan/pachelbel/connection"
	"github.com/benjdewan/pachelbel/output"
)

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

func DryRunCreate(cxn *connection.Connection, accessor Accessor) error {
	cxn.Add(output.FakeID(accessor.GetType(), accessor.GetName()))
	return nil
}

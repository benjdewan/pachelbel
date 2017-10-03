package runner

import (
	"github.com/benjdewan/pachelbel/connection"
)

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

func DryRunUpdate(cxn *connection.Connection, accessor Accessor) error {
	cxn.Add(accessor.(connection.Deployment).GetID())
	return nil
}

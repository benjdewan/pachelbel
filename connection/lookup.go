package connection

import (
	"fmt"

	"github.com/benjdewan/pachelbel/output"
)

func lookup(cxn *Connection, accessor Accessor) error {
	deployment, errs := cxn.client.GetDeploymentByName(accessor.GetName())
	if len(errs) != 0 {
		return fmt.Errorf("Failed to lookup '%s':\n%v", accessor.GetName(), errs)
	}
	cxn.newDeploymentIDs.Store(deployment.ID, struct{}{})
	return nil
}

func dryRunLookup(cxn *Connection, accessor Accessor) error {
	deployment, errs := cxn.client.GetDeploymentByName(accessor.GetName())
	if len(errs) != 0 {
		// This is a dry run, assume it's been created
		cxn.newDeploymentIDs.Store(output.FakeID(accessor.GetType(), accessor.GetName()), struct{}{})
		return nil
	}
	cxn.newDeploymentIDs.Store(deployment.ID, struct{}{})
	return nil
}

package runner

import (
	"github.com/benjdewan/pachelbel/connection"
	"github.com/benjdewan/pachelbel/output"
)

func Lookup(cxn *connection.Connection, accessor Accessor) error {
	return cxn.GetAndAdd(accessor.GetName())
}

func DryRunLookup(cxn *connection.Connection, accessor Accessor) error {
	if err := cxn.GetAndAdd(accessor.GetName()); err != nil {
		cxn.Add(output.FakeID(accessor.GetType(), accessor.GetName()))
	}
	return nil
}

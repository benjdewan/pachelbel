package runner

import "github.com/benjdewan/pachelbel/connection"

func Deprovision(cxn *connection.Connection, accessor Accessor) error {
	return cxn.Deprovision(accessor.(connection.Deprovision))
}

func DryRunDeprovision(cxn *connection.Connection, accessor Accessor) error {
	return nil
}

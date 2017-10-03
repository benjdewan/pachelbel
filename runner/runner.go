package runner

import (
	"log"
	"strings"

	"github.com/benjdewan/pachelbel/connection"
)

const (
	ActionLookup      = "Looking up"
	ActionCreate      = "Creating"
	ActionResize      = "Resizing"
	ActionUpgrade     = "Upgrading"
	ActionComment     = "Commenting on"
	ActionDeprovision = "Deprovisioning"
)

type RunFunc func(*connection.Connection, Accessor) error

// Runner is an individual deployment operation
type Runner struct {
	Target Accessor
	Action string
	Run    RunFunc
}

func toDryRun(action string) RunFunc {
	switch action {
	case ActionLookup:
		return DryRunLookup
	case ActionCreate:
		return DryRunCreate
	case ActionDeprovision:
		return DryRunDeprovision
	default:
		if strings.Contains(action, ActionUpgrade) || strings.Contains(action, ActionResize) || strings.Contains(action, ActionComment) {
			return DryRunUpdate
		} else {
			log.Panicf("Unknown action: %s", action)
		}
	}
	panic("unreachable code")
}

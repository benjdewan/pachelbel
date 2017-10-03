package config

import (
	"fmt"
	"strings"

	"github.com/benjdewan/pachelbel/runner"
)

func validateDeploymentClientV2(d deploymentClientV2, input string) error {
	errs := []string{}

	errs = append(errs, validateType(d.Type)...)
	errs = append(errs, validateName(d.Name)...)
	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("Errors occurred while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, strings.Join(errs, "\n"))
}

func validateDeprovisionV2(d deprovisionObjectV2, input string) (runner.Runner, bool, error) {
	var (
		deprovisioner runner.Runner
		errs          []string
		skip          bool
	)
	if len(d.ID) > 0 {
		deprovisioner, skip = validateDeprovisionByIDV2(d)
	} else if len(d.Name) > 0 {
		deprovisioner, skip = validateDeprovisionByNameV2(d)
	} else {
		errs = append(errs, "A 'name' or 'id' field is required")
	}

	if len(errs) == 0 {
		return deprovisioner, skip, nil
	}

	return deprovisioner, false, fmt.Errorf("Errors occurred while parsing the following deprovision object:\n%s\nErrors:\n%s", input, strings.Join(errs, "\n"))
}

func validateDeprovisionByIDV2(d deprovisionObjectV2) (runner.Runner, bool) {
	existing, ok := existingDeployment(d.ID)
	if !ok {
		return runner.Runner{}, true
	}
	d.Name = existing.Name
	d.dType = existing.Type
	return runner.Runner{
		Target: runner.Accessor(d),
		Action: runner.ActionDeprovision,
		Run:    runner.Deprovision,
	}, false
}

func validateDeprovisionByNameV2(d deprovisionObjectV2) (runner.Runner, bool) {
	existing, ok := existingDeployment(d.Name)
	if !ok {
		return runner.Runner{}, true
	}
	d.ID = existing.ID
	d.dType = existing.Type
	return runner.Runner{
		Target: runner.Accessor(d),
		Action: runner.ActionDeprovision,
		Run:    runner.Deprovision,
	}, false
}

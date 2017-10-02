package config

import (
	"fmt"
	"strings"
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

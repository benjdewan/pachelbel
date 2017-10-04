package connection

import "fmt"

// Deprovision makes an API call to compose to deprovision the specified
// deployment
func (cxn *Connection) Deprovision(deprovision Deprovision) error {
	recipe, errs := cxn.client.DeprovisionDeployment(deprovision.GetID())
	if len(errs) != 0 {
		return fmt.Errorf("Unable to deprovision '%s':\n%v",
			deprovision.GetName(), errs)
	}

	if deprovision.GetTimeout() == 0 {
		return nil
	}

	return cxn.wait(recipe.ID, deprovision.GetTimeout())
}

// Copyright © 2017 ben dewan <benj.dewan@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package connection

import (
	"fmt"

	compose "github.com/benjdewan/gocomposeapi"
)

func update(cxn *Connection, depID string, dep Deployment) error {
	timeout := dep.GetTimeout()
	if err := updateScalings(cxn, depID, dep.GetScaling(), timeout); err != nil {
		return err
	}

	if version := dep.GetVersion(); len(version) > 0 {
		if err := updateVersion(cxn, depID, dep.GetVersion(), timeout); err != nil {
			return err
		}
	} else {
		fmt.Printf("Not updating version of '%v'\n", depID)
	}

	if err := addTeamRoles(cxn, depID, dep.GetTeamRoles()); err != nil {
		return err
	}

	return nil
}

func updateScalings(cxn *Connection, depID string, newScale int, timeout float64) error {
	existingScalings, errs := cxn.client.GetScalings(depID)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get current scaling for '%s':\n%v",
			depID, errsOut(errs))
	}

	if existingScalings.AllocatedUnits == newScale {
		fmt.Printf("Existing deployment '%s' is the expected size '%v'\n",
			depID, newScale)
		cxn.newDeploymentIDs = append(cxn.newDeploymentIDs, depID)
		return nil
	}

	fmt.Printf("Rescaling '%s':\n\tCurrent scale: %v\n\tDesired scale: %v",
		depID, existingScalings.AllocatedUnits, newScale)

	params := compose.ScalingsParams{
		DeploymentID: depID,
		Units:        newScale,
	}

	recipe, errs := cxn.client.SetScalings(params)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to resize '%s':\n%v",
			depID, errsOut(errs))
	}

	err := cxn.waitOnRecipe(recipe.ID, timeout)
	if err != nil {
		return err
	}
	cxn.newDeploymentIDs = append(cxn.newDeploymentIDs, recipe.DeploymentID)
	return nil
}

func updateVersion(cxn *Connection, depID, newVersion string, timeout float64) error {
	deployment, errs := cxn.client.GetDeployment(depID)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to fetch current deployment information for '%s':\n%v",
			depID, errsOut(errs))
	}

	if deployment.Version == newVersion {
		fmt.Printf("Deployment '%s' is at version '%s'. Not upgrading\n",
			depID, newVersion)
		return nil
	}

	transitions, errs := cxn.client.GetVersionsForDeployment(depID)
	if len(errs) != 0 || transitions == nil {
		return fmt.Errorf("Error fetching upgrade information for '%s':\n%v",
			depID, errsOut(errs))
	}

	validTransition := false
	for _, transition := range *transitions {
		if transition.ToVersion == newVersion {
			validTransition = true
			break
		}
	}
	if !validTransition {
		return fmt.Errorf("Cannot upgrade '%s' to version '%s'.",
			depID, newVersion)
	}

	recipe, errs := cxn.client.UpgradeVersionForDeployment(depID, newVersion)
	if errs != nil {
		return fmt.Errorf("Unable to upgrade '%s' to verison '%s':\n%v",
			depID, newVersion, errsOut(errs))
	}

	err := cxn.waitOnRecipe(recipe.ID, timeout)
	if err != nil {
		return err
	}
	cxn.newDeploymentIDs = append(cxn.newDeploymentIDs, recipe.DeploymentID)
	return nil
}

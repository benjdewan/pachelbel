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
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/gosuri/uiprogress"
)

func update(cxn *Connection, depID string, dep Deployment, errQueue *queue.Queue) {
	timeout := dep.GetTimeout()
	bar, pollLength := newBar(timeout, cxn.pollingInterval, 2,
		dep.TeamEntryCount(), cxn.MaxNameLength, dep.GetName(),
		"updating")

	scaling := dep.GetScaling()
	if err := updateScalings(cxn, depID, scaling, timeout, bar); err != nil {
		enqueue(errQueue, err)
		return
	}
	setProgress(bar, pollLength)

	version := dep.GetVersion()
	if len(version) > 0 {
		if err := updateVersion(cxn, depID, version, timeout, bar); err != nil {
			enqueue(errQueue, err)
			return
		}
	}
	setProgress(bar, 2*pollLength)

	if err := addTeamRoles(cxn, depID, dep.GetTeamRoles(), bar); err != nil {
		enqueue(errQueue, err)
		return
	}
	setProgress(bar, bar.Total)

	// Changing versions and sizes can change the deployment ID. Ensure
	// we have the latest/live value
	updatedDeployment, errs := cxn.client.GetDeploymentByName(dep.GetName())
	if len(errs) != 0 {
		enqueue(errQueue, errsOut(errs))
		return
	}
	cxn.newDeploymentIDs.Store(updatedDeployment.ID, struct{}{})
	return
}

func updateScalings(cxn *Connection, depID string, newScale int, timeout float64, b *uiprogress.Bar) error {
	existingScalings, errs := cxn.client.GetScalings(depID)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get current scaling for '%s':\n%v",
			depID, errsOut(errs))
	}

	if existingScalings.AllocatedUnits == newScale {
		fmt.Printf("Existing deployment '%s' is the expected size '%v'\n",
			depID, newScale)
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

	err := cxn.waitOnRecipe(recipe.ID, timeout, b)
	if err != nil {
		return err
	}
	return nil
}

func updateVersion(cxn *Connection, depID, newVersion string, timeout float64, b *uiprogress.Bar) error {
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

	recipe, errs := cxn.client.UpdateVersion(depID, newVersion)
	if errs != nil {
		return fmt.Errorf("Unable to upgrade '%s' to version '%s':\n%v",
			depID, newVersion, errsOut(errs))
	}

	err := cxn.waitOnRecipe(recipe.ID, timeout, b)
	if err != nil {
		return err
	}
	return nil
}

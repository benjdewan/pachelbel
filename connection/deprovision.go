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

func dryRunDeprovision(cxn *Connection, accessor Accessor) error {
	return nil
}

func deprovision(cxn *Connection, accessor Accessor) error {
	deployment := accessor.(Deprovision)

	recipe, errs := cxn.client.DeprovisionDeployment(deployment.GetID())
	if len(errs) != 0 {
		return fmt.Errorf("Unable to deprovision '%s':\n%v",
			accessor.GetName(), errs)
	}

	if deployment.GetTimeout() == 0 {
		return nil
	}

	return cxn.waitOnRecipe(recipe.ID, deployment.GetTimeout())
}

type deprovisionObject struct {
	name           string
	id             string
	deploymentType string
	timeout        float64
}

func (d deprovisionObject) IsOwner() bool {
	return true
}

func (d deprovisionObject) IsDeleter() bool {
	return true
}

func (d deprovisionObject) GetName() string {
	return d.name
}

func (d deprovisionObject) GetTimeout() float64 {
	return d.timeout
}

func (d deprovisionObject) GetType() string {
	return d.deploymentType
}

func (d deprovisionObject) GetID() string {
	return d.id
}

func resolveDeprovisionObjects(client *compose.Client, idsAndNames []string, timeout float64) []Accessor {
	objs := []Accessor{}
	for _, idOrName := range idsAndNames {
		if obj := resolveId(client, idOrName, timeout); obj != nil {
			objs = append(objs, Accessor(*obj))
		}
	}
	return objs
}

func resolveId(client *compose.Client, idOrName string, timeout float64) *deprovisionObject {
	deployment, _ := client.GetDeployment(idOrName)
	if deployment != nil {
		return &deprovisionObject{
			name:           deployment.Name,
			id:             deployment.ID,
			deploymentType: deployment.Type,
			timeout:        timeout,
		}
	}
	deployment, _ = client.GetDeploymentByName(idOrName)
	if deployment != nil {
		return &deprovisionObject{
			name:           deployment.Name,
			id:             deployment.ID,
			deploymentType: deployment.Type,
			timeout:        timeout,
		}
	}
	return nil
}

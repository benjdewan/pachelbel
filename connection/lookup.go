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

import "fmt"

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
		cxn.newDeploymentIDs.Store(fakeID(accessor.GetType(), accessor.GetName()), struct{}{})
		return nil
	}
	cxn.newDeploymentIDs.Store(deployment.ID, struct{}{})
	return nil
}

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

package config

type endpointMapV2 struct {
	EndpointMap map[string]string `json:"endpoint_map"`
}

type deploymentClientV2 struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// IsDeleter will always return false because a deployment_client object does
// not have permission to modify deployments let alone delete them
func (d deploymentClientV2) IsDeleter() bool {
	return false
}

// IsOwner will always return false because a deployment_client object
// does not have permission to make changes to an existing deployment
func (d deploymentClientV2) IsOwner() bool {
	return false
}

// GetName returns the name of the deployment this object is a client of
func (d deploymentClientV2) GetName() string {
	return d.Name
}

// GetType returns the type of the deployment this object is a client of
func (d deploymentClientV2) GetType() string {
	return d.Type
}

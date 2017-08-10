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
	"time"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/golang-collections/go-datastructures/queue"
	"golang.org/x/sync/syncmap"
)

type Deployment interface {
	GetCluster() string
	GetDatacenter() string
	GetName() string
	GetNotes() string
	GetScaling() int
	GetSSL() bool
	GetTeamRoles() map[string]([]string)
	GetTimeout() float64
	GetType() string
	GetVersion() string
	GetWiredTiger() bool
}

type Connection struct {
	client            *compose.Client
	accountID         string
	clusterIDsByName  map[string]string
	deploymentsByName map[string](*compose.Deployment)
	newDeploymentIDs  *syncmap.Map
	pollingInterval   time.Duration
}

func Init(apiKey string, pollingInterval int) (*Connection, error) {
	cxn := &Connection{
		newDeploymentIDs: &syncmap.Map{},
		pollingInterval:  time.Duration(pollingInterval) * time.Second,
	}
	var err error

	cxn.client, err = createClient(apiKey)
	if err != nil {
		return cxn, err
	}

	cxn.accountID, err = fetchAccountID(cxn.client)
	if err != nil {
		return cxn, err
	}

	cxn.clusterIDsByName, err = fetchClusters(cxn.client)
	if err != nil {
		return cxn, err
	}

	cxn.deploymentsByName, err = fetchDeployments(cxn.client)

	return cxn, err
}

func Provision(cxn *Connection, deployment Deployment, errQueue *queue.Queue) {
	if existing, ok := cxn.deploymentsByName[deployment.GetName()]; ok {
		update(cxn, existing.ID, deployment, errQueue)
		return
	}
	provision(cxn, deployment, errQueue)
}

func (cxn *Connection) ConnectionStringsYAML(outFile string, errQueue *queue.Queue) {
	fmt.Printf("Writing connection strings to '%v'\n", outFile)

	yamlObjects := []([]byte){}
	cxn.newDeploymentIDs.Range(func(key, value interface{}) bool {
		var deploymentID string
		switch keyType := key.(type) {
		case string:
			deploymentID = keyType
		default:
			panic(fmt.Sprintf("Only deploymentIDs should be in this map"))
		}

		yamlObj, err := connectionStringsForDeployment(cxn, deploymentID)
		if err != nil {
			enqueue(errQueue, err)
			return false
		}
		yamlObjects = append(yamlObjects, yamlObj)
		return true
	})

	if err := writeConnectionStrings(yamlObjects, outFile); err != nil {
		enqueue(errQueue, err)
		return
	}
}

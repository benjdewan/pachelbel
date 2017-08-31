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
	"os"
	"sync"
	"time"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/benjdewan/pachelbel/progress"
	"github.com/golang-collections/go-datastructures/queue"
	"golang.org/x/sync/syncmap"
)

// Deployment is the interface for deployment objects that
// the Connection struct expects as input to Provision()
type Deployment interface {
	ClusterDeployment() bool
	TagDeployment() bool
	GetCluster() string
	GetTags() []string
	GetDatacenter() string
	GetName() string
	GetNotes() string
	GetScaling() int
	GetSSL() bool
	GetTeamRoles() map[string]([]string)
	TeamEntryCount() int
	GetTimeout() float64
	GetType() string
	GetVersion() string
	GetWiredTiger() bool
}

// Connection is the struct that manages the state of provisioning
// work done in Compose during an invocation of pachelbel.
// codebeat:disable[TOO_MANY_IVARS]
type Connection struct {
	// Internal fields
	client            *compose.Client
	logFile           *os.File
	dryRun            bool
	accountID         string
	clusterIDsByName  map[string]string
	datacenters       map[string]struct{}
	deploymentsByName *syncmap.Map
	newDeploymentIDs  *syncmap.Map
	pollingInterval   time.Duration
	pb                *progress.ProgressBars
}

// codebeat:enable[TOO_MANY_IVARS]

// Init creates a Connection struct that is used for provisioning
// Compose deployments. This struct is shared across every
// Provision call.
func Init(apiKey, logFile string, pollingInterval int, dryRun bool) (*Connection, error) {
	cxn := &Connection{
		newDeploymentIDs:  &syncmap.Map{},
		deploymentsByName: &syncmap.Map{},
		pollingInterval:   time.Duration(pollingInterval) * time.Second,
		pb:                progress.New(),
		dryRun:            dryRun,
	}
	cxn.pb.RefreshRate = cxn.pollingInterval
	var err error

	if len(logFile) > 0 {
		cxn.logFile, err = os.Create(logFile)
		if err != nil {
			return cxn, err
		}
	}

	cxn.client, err = createClient(apiKey, cxn.logFile)
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

	cxn.datacenters, err = fetchDatacenters(cxn.client)
	if err != nil {
		return cxn, err
	}

	err = fetchDeployments(cxn)

	return cxn, err
}

// Provision will create a new deployment or update an existing deployment
// to the size and version specified as well as ensure every team role listed
// is applied to that deployment.
func (cxn *Connection) Provision(deployments []Deployment, errQueue *queue.Queue) {
	deployers := cxn.listDeployers(deployments)

	var wg sync.WaitGroup
	wg.Add(len(deployers))

	cxn.pb.Start()
	for _, deployer := range deployers {
		go runDeployer(deployer, cxn, errQueue, &wg)
	}
	wg.Wait()
	cxn.pb.Stop()
}

// ConnectionStringsYAML writes out the connection strings for all the
// provisioned deployments as a YAML object to the provided file.
func (cxn *Connection) ConnectionYAML(outFile string, errQueue *queue.Queue) {
	fmt.Printf("Writing connection strings to '%v'\n", outFile)

	connections := []([]byte){}
	cxn.newDeploymentIDs.Range(func(key, value interface{}) bool {
		var err error
		connections, err = connectionYAMLByKey(cxn, connections, key)
		if err == nil {
			return true
		}
		enqueue(errQueue, err)
		return false
	})

	if err := writeConnectionYAML(connections, outFile); err != nil {
		enqueue(errQueue, err)
		return
	}
}

// Close closes any open connections and/or files possessed by the Connection
// instance.
func (cxn *Connection) Close() error {
	if cxn.logFile == nil {
		return nil
	}
	return cxn.logFile.Close()
}

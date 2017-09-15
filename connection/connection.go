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
	"github.com/benjdewan/pachelbel/output"
	"github.com/benjdewan/pachelbel/progress"
	"github.com/golang-collections/go-datastructures/queue"
	"golang.org/x/sync/syncmap"
)

// Accessor is the interface for any Compose Deployment information request.
// To make any deployment mutations (creating new deployments or updating
// existing ones), the provided object must also implement the Deployment
// interface.
type Accessor interface {
	IsOwner() bool
	GetName() string
	GetType() string
}

// Deployment is the interface for mutatable Compose deployments. For any
// Deployment object pachelbel will create or update a Compose deployment
// to match the name, type, size, location &c. to match the information provided
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
	GetCacheMode() bool
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

// New creates a new Connection struct, but does not initialize the Compose
// connection. Invoke Init() to do so.
func New(logFile string, pollingInterval int, dryRun bool) (*Connection, error) {
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
	}
	return cxn, err
}

// Init will establish the connection to Compose for the given Connection object
// and populate it with current information of existing deployments and clusters
func (cxn *Connection) Init(apiKey string) error {
	var err error
	cxn.client, err = createClient(apiKey, cxn.logFile)
	if err != nil {
		return err
	}

	cxn.accountID, err = fetchAccountID(cxn.client)
	if err != nil {
		return err
	}

	cxn.clusterIDsByName, err = fetchClusters(cxn.client)
	if err != nil {
		return err
	}

	cxn.datacenters, err = fetchDatacenters(cxn.client)
	return err
}

// Process reads through the provided slice of Accessors, creates or edits
// deployments where necessary or looks up existing deployments depending on
// whether a provided Accessor is an owner or not
func (cxn *Connection) Process(accessors []Accessor) error {
	runners := cxn.newRunners(accessors)

	var wg sync.WaitGroup
	wg.Add(len(runners))

	q := queue.New(0)
	cxn.pb.Start()
	for _, runner := range runners {
		go func(r cxnRunner) {
			if err := r.run(cxn, r.accessor); err != nil {
				cxn.pb.Error(r.accessor.GetName())
				enqueue(q, err)
			} else {
				cxn.pb.Done(r.accessor.GetName())
			}
			wg.Done()
		}(runner)
	}
	wg.Wait()
	cxn.pb.Stop()
	return flushErrors(q)
}

// ConnectionYAML writes out the connection strings for all the
// provisioned deployments as a YAML object to the provided file.
func (cxn *Connection) ConnectionYAML(endpointMap map[string]string, outFile string) error {
	q := queue.New(0)
	builder := output.New(endpointMap)
	cxn.newDeploymentIDs.Range(func(key, value interface{}) bool {
		if err := cxn.addToBuilder(key.(string), builder); err != nil {
			enqueue(q, err)
			return false
		}
		return true
	})

	if err := builder.Write(outFile); err != nil {
		enqueue(q, err)
	}
	return flushErrors(q)
}

func (cxn *Connection) addToBuilder(id string, b *output.Builder) error {
	if output.IsFake(id) {
		return b.AddFake(id)
	}
	deployment, errs := cxn.client.GetDeployment(id)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get deployment information for '%s':\n%v", id, errs)
	}
	return b.Add(deployment)
}

// Close closes any open connections and/or files possessed by the Connection
// instance.
func (cxn *Connection) Close() error {
	if cxn.logFile == nil {
		return nil
	}
	return cxn.logFile.Close()
}

package connection

import (
	"fmt"
	"os"
	"sync"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/benjdewan/pachelbel/output"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/masterminds/semver"
)

type ExistingDeployment struct {
	Scaling  int
	Notes    string
	Name     string
	ID       string
	Type     string
	Version  string
	Upgrades []*semver.Version
}

// Deployment is the interface for mutatable Compose deployments. For any
// Deployment object pachelbel will create or update a Compose deployment
// to match the name, type, size, location &c. to match the information provided
type Deployment interface {
	ClusterDeployment() bool
	TagDeployment() bool
	GetID() string
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

// Deprovision is the interface for deployment deprovision objects. To not wait
// for the deprovision recipe to complete for the given ID, ensure GetTimeout()
// returns 0
type Deprovision interface {
	GetID() string
	GetName() string
	GetTimeout() float64
}

// Connection is the struct that manages the state of provisioning
// work done in Compose during an invocation of pachelbel.
// codebeat:disable[TOO_MANY_IVARS]
type Connection struct {
	// Internal fields
	client           *compose.Client
	logFile          *os.File
	accountID        string
	newDeploymentIDs *sync.Map
}

// codebeat:enable[TOO_MANY_IVARS]

// New creates a new Connection struct, but does not initialize the Compose
// connection. Invoke Init() to do so.
func New(apiKey, logFile string) (*Connection, error) {
	cxn := &Connection{newDeploymentIDs: &sync.Map{}}
	var err error
	if len(logFile) > 0 {
		if cxn.logFile, err = os.Create(logFile); err != nil {
			return cxn, err
		}
	}
	cxn.client, err = createClient(apiKey, cxn.logFile)
	if err != nil {
		return cxn, err
	}

	cxn.accountID, err = fetchAccountID(cxn.client)
	return cxn, err
}

// AddTeams adds teams to the deployment specified by the ID with the roles provided
func (cxn *Connection) AddTeams(id string, deployment Deployment) error {
	teamRoles := deployment.GetTeamRoles()
	existingRoles, errs := cxn.client.GetTeamRoles(id)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to retrieve team_role information for '%s':\n%v\n",
			deployment.GetName(), errs)
	}
	if len(teamRoles) == 0 {
		return nil
	}

	for role, teams := range teamRoles {
		existingTeams := []compose.Team{}
		for _, existingTeamRoles := range *existingRoles {
			if existingTeamRoles.Name == role {
				existingTeams = existingTeamRoles.Teams
				break
			}
		}

		for _, teamID := range filterTeams(teams, existingTeams) {
			params := compose.TeamRoleParams{
				Name:   role,
				TeamID: teamID,
			}

			_, createErrs := cxn.client.CreateTeamRole(id, params)
			if createErrs != nil {
				return fmt.Errorf("Unable to add team '%s' as '%s' to %s:\n%v\n",
					teamID, role, deployment.GetName(),
					createErrs)
			}
		}
	}
	return nil
}

// Add a deployment ID to a connection object's internal deployment tracker
func (cxn *Connection) Add(id string) {
	cxn.newDeploymentIDs.Store(id, struct{}{})
}

// GetAndAdd retrieves the latest deployment information about the named
// deployment and stores its ID
func (cxn *Connection) GetAndAdd(name string) error {
	deployment, errs := cxn.client.GetDeploymentByName(name)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get the latest details of '%s':\n%v", name, errs)
	}
	cxn.newDeploymentIDs.Store(deployment.ID, struct{}{})
	return nil
}

// Clusters returns a map of cluster IDs by name and ID.
func (cxn *Connection) Clusters() (map[string]string, error) {
	clusters := make(map[string]string)

	clusterList, errs := cxn.client.GetClusters()
	if len(errs) != 0 || clusters == nil {
		return clusters, fmt.Errorf("Failed to get cluster information:\n%s", errs)
	}

	for _, cluster := range *clusterList {
		clusters[cluster.Name] = cluster.ID
		clusters[cluster.ID] = cluster.ID
	}
	return clusters, nil
}

// Datacenters returns a map of datacenters slugs to empty structs (for fast
// lookup).
func (cxn *Connection) Datacenters() (map[string]struct{}, error) {
	datacenters := make(map[string]struct{})

	datacenterObjs, errs := cxn.client.GetDatacenters()
	if len(errs) != 0 || datacenterObjs == nil {
		return datacenters, fmt.Errorf("Failed to get datacenter information:\n%v", errs)
	}

	for _, datacenter := range *datacenterObjs {
		datacenters[datacenter.Slug] = struct{}{}
	}

	return datacenters, nil
}

// SupportedDatabases returns a map of database types supported by Compose
// to the versions (both supported and deprecated) Compose works with.
// New deployments cannot be made using deprecated versions
func (cxn *Connection) SupportedDatabases() (map[string][]string, error) {
	dbs, errs := cxn.client.GetDatabases()
	if len(errs) != 0 {
		return nil, fmt.Errorf("Unable to enumerate supported database types:\n%v", errs)
	}
	return buildDatabaseVersionMap(*dbs), nil
}

func (cxn *Connection) ExistingDeployment(idOrName string) (ExistingDeployment, error) {
	deployment, errs := cxn.client.GetDeployment(idOrName)
	if len(errs) == 0 && deployment != nil {
		return cxn.existingDeployment(*deployment)
	}
	deployment, errs = cxn.client.GetDeploymentByName(idOrName)
	if len(errs) == 0 && deployment != nil {
		return cxn.existingDeployment(*deployment)
	}
	return ExistingDeployment{}, fmt.Errorf("Unable to resolve '%s' as a deployment id or name:\n%v", idOrName, errs)
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

// Close closes any open connections and/or files possessed by the Connection
// instance.
func (cxn *Connection) Close() error {
	if cxn.logFile == nil {
		return nil
	}
	return cxn.logFile.Close()
}

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

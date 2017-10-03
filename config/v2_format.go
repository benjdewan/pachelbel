package config

type endpointMapV2 struct {
	EndpointMap map[string]string `json:"endpoint_map"`
}

type deprovisionObjectV2 struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Timeout *int   `json:"timeout"`

	//internal fields
	dType string
}

func (d deprovisionObjectV2) GetName() string {
	return d.Name
}

func (d deprovisionObjectV2) GetTimeout() float64 {
	if d.Timeout == nil {
		return float64(300)
	}
	return float64(*d.Timeout)
}

func (d deprovisionObjectV2) GetType() string {
	return d.dType
}

func (d deprovisionObjectV2) GetID() string {
	return d.ID
}

type deploymentClientV2 struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetName returns the name of the deployment this object is a client of
func (d deploymentClientV2) GetName() string {
	return d.Name
}

// GetType returns the type of the deployment this object is a client of
func (d deploymentClientV2) GetType() string {
	return d.Type
}

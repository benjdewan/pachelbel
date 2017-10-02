package output

import (
	"bytes"
	"fmt"
	urlparser "net/url"
	"os"
	"strconv"
	"strings"

	compose "github.com/benjdewan/gocomposeapi"
	"github.com/ghodss/yaml"
)

type outputYAML struct {
	Type        string           `json:"type"`
	CACert      string           `json:"cacert,omitempty"`
	Version     string           `json:"version"`
	Connections []connectionYAML `json:"connections"`
}

// codebeat:disable[TOO_MANY_IVARS]
type connectionYAML struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     int    `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	Query    string `json:"query,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

// Builder is the stateful object used to build pachelbel's output files
type Builder struct {
	endpointMap map[string]string
	yml         [][]byte
}

// New returns an initialized Builder object
func New(endpointMap map[string]string) *Builder {
	return &Builder{
		endpointMap: endpointMap,
		yml:         [][]byte{},
	}
}

// Add takes a compose.Deployment object and converts its connection
// information into Builder's internal representation
func (b *Builder) Add(deployment *compose.Deployment) error {
	yamlObject, err := b.convert(deployment)
	b.yml = append(b.yml, yamlObject)
	return err
}

func (b *Builder) Write(file string) error {
	outBytes := bytes.Join(b.yml, []byte("\n"))
	handle, err := os.Create(file)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := handle.Close(); closeErr != nil {
			// Our fd became invalid, or the underlying
			// syscall was interrupted
			panic(closeErr)
		}
	}()

	_, err = handle.Write(outBytes)
	return err
}

func (b *Builder) convert(deployment *compose.Deployment) ([]byte, error) {
	connections, err := b.convertConnections(deployment.Connection.Direct)
	if err != nil {
		return []byte{}, err
	}

	outYAML := make(map[string]outputYAML)
	outYAML[deployment.Name] = outputYAML{
		Type:        deployment.Type,
		CACert:      deployment.CACertificateBase64,
		Version:     deployment.Version,
		Connections: connections,
	}
	return yaml.Marshal(outYAML)
}

func (b *Builder) convertConnections(cStrings []string) ([]connectionYAML, error) {
	connections := []connectionYAML{}
	for _, cString := range cStrings {
		newCXNs, err := b.convertConnection(cString)
		if err != nil {
			return connections, err
		}
		connections = append(connections, newCXNs...)
	}
	return connections, nil
}

func (b *Builder) convertConnection(cString string) ([]connectionYAML, error) {
	if strings.HasPrefix(cString, "mongodb") && strings.Contains(cString, ",") {
		return b.convertMongoConnection(cString)
	}
	url, err := urlparser.Parse(cString)
	if err != nil {
		return []connectionYAML{}, err
	}
	host, err := b.resolveHost(url.Hostname())
	if err != nil {
		return []connectionYAML{}, err
	}
	password, _ := url.User.Password()

	connection := connectionYAML{
		Scheme:   url.Scheme,
		Host:     host,
		Path:     url.Path,
		Query:    url.Query().Encode(),
		Username: url.User.Username(),
		Password: password,
	}
	if port, err := strconv.Atoi(url.Port()); err == nil {
		connection.Port = port
	}
	return []connectionYAML{connection}, nil
}

func (b *Builder) convertMongoConnection(mongoString string) ([]connectionYAML, error) {
	prefix, hosts, suffix := mongoSplit(mongoString)
	cStrings := []string{}
	for _, host := range strings.Split(hosts, ",") {
		cStrings = append(cStrings, mongoJoin(prefix, host, suffix))
	}
	return b.convertConnections(cStrings)
}

func mongoSplit(mongoStr string) (string, string, string) {
	splitPrefix := strings.Split(mongoStr, "@")
	prefix := splitPrefix[0]
	splitSuffix := strings.Split(splitPrefix[1], "/")
	return prefix, splitSuffix[0], splitSuffix[1]
}

func mongoJoin(prefix, host, suffix string) string {
	return strings.Join([]string{strings.Join([]string{prefix, host}, "@"), suffix}, "/")
}

func (b *Builder) resolveHost(host string) (string, error) {
	i := 0
	for {
		if i >= 8 {
			return host, fmt.Errorf("Recursive endpoint mapping is limited to 8 layers. '%s' exceeds that", host)
		}
		if rename, ok := b.endpointMap[host]; ok {
			host = rename
			i++
			continue
		}
		break
	}
	return host, nil
}

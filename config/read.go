package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/benjdewan/pachelbel/connection"
	"github.com/benjdewan/pachelbel/runner"
	"github.com/ghodss/yaml"
)

// Config returns parsed configuration objects to be
// consumed by pachelbel for provisioning
type Config struct {
	// Runners is a slice of all the deployments to be provisioned
	// and any that need to be looked up.
	//
	// The deployments to be provisioned in this slice have been
	// filtered by cluster and/or datacenter if those filters were
	// set.
	Runners []runner.Runner

	// EndpointMap is a list of mappings to perform to translate
	// the connection strings returned by Compose.io. This is most
	// useful with Enterprise accounts using clusters that expose
	// public endpoints, but the Compose API only returns the
	// private ones.
	EndpointMap map[string]string

	// Internal fields
	dNames map[string]struct{}
}

// BuildClusterFilter accepts a list of cluster names to filter
// configuration data. Only deployments to clusters in the filter
// are returned by ReadFiles()
func BuildClusterFilter(clusters []string) {
	clusterFilter = make(map[string]struct{})

	for _, cluster := range clusters {
		clusterFilter[cluster] = struct{}{}
	}
}

// BuildDatacenterFilter accepts a list of datacenter slugs to filter
// configuration data by. Only deployments to datacenters in the filter
// are returned by ReadFiles()
func BuildDatacenterFilter(datacenters []string) {
	datacenterFilter = make(map[string]struct{})

	for _, datacenter := range datacenters {
		datacenterFilter[datacenter] = struct{}{}
	}
}

// ReadFiles works through a list of arguments to parse configuration
// data into deployment object. Both configuration files and directories
// of configuration files are valid arguments, but directories are not
// read recursively, only immediate child files are parsed.
func ReadFiles(args []string) (*Config, error) {
	cfg := newConfig()

	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			return cfg, err
		}
		switch mode := info.Mode(); {
		case mode.IsDir():
			err = cfg.readDir(path)
		case mode.IsRegular():
			err = cfg.readFile(path)
		}
		if err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func newConfig() *Config {
	return &Config{
		Runners:     []runner.Runner{},
		EndpointMap: make(map[string]string),
		dNames:      make(map[string]struct{}),
	}
}

// I'm not certain there is a way to satisfy codebeat here that's actually
// more legibile
// codebeat:disable[BLOCK_NESTING]
func (cfg *Config) readDir(root string) error {
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, readErr error) error {
		if readErr != nil {
			return readErr
		} else if path == root || info.IsDir() {
			return nil
		} else if err := cfg.readFile(path); err != nil {
			return err
		}
		return nil
	})
	return walkErr
}

// codebeat:enable[BLOCK_NESTING]

func (cfg *Config) readFile(path string) error {
	file, err := os.Open(path) // #nosec
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Our filepointer has become invalid or the
			// underlying syscall was interrupted
			panic(closeErr)
		}
	}()

	return cfg.readConfigs(file)
}

func (cfg *Config) readConfigs(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(splitYAMLObjects)
	for scanner.Scan() {
		if err := cfg.readConfig(scanner.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type objectMetadata struct {
	ConfigVersion int    `json:"config_version"`
	ObjectType    string `json:"object_type"`
}

func (cfg *Config) readConfig(blob []byte) error {
	var metadata objectMetadata
	if err := yaml.Unmarshal(blob, &metadata); err != nil {
		return err
	}
	switch metadata.ConfigVersion {
	case 1:
		return cfg.readConfigV1(blob)
	case 2:
		return cfg.readConfigV2(metadata.ObjectType, blob)
	default:
		return fmt.Errorf("Expected `config_version` to be '1' or '2' but saw '%d'",
			metadata.ConfigVersion)

	}
}

func (cfg *Config) readConfigV2(objectType string, blob []byte) error {
	switch objectType {
	case "endpoint_map":
		return cfg.readEndpointMapV2(blob)
	case "deployment_client":
		return cfg.readDeploymentClientV2(blob)
	case "deprovision":
		return cfg.readDeprovisionV2(blob)
	default:
		return fmt.Errorf("'%s' is not a supported object_type", objectType)
	}
}

func (cfg *Config) readDeprovisionV2(blob []byte) error {
	var d deprovisionObjectV2
	if err := yaml.Unmarshal(blob, &d); err != nil {
		return err
	}
	deprovisioner, skip, err := validateDeprovisionV2(d, string(blob))
	if skip {
		return nil
	} else if err != nil {
		return err
	}
	name := deprovisioner.Target.GetName()
	if _, ok := cfg.dNames[name]; ok {
		return fmt.Errorf("Deployment names must be unique across all configuration objects, but '%s' is specified more than once", name)
	}
	cfg.dNames[name] = struct{}{}
	cfg.Runners = append(cfg.Runners, deprovisioner)
	return nil
}

func (cfg *Config) readDeploymentClientV2(blob []byte) error {
	var d deploymentClientV2
	if err := yaml.Unmarshal(blob, &d); err != nil {
		return err
	}
	if err := validateDeploymentClientV2(d, string(blob)); err != nil {
		return err
	}
	if _, ok := cfg.dNames[d.Name]; ok {
		return fmt.Errorf("Deployment names must be unique, but '%s' is specified more than once",
			d.Name)
	}
	cfg.dNames[d.Name] = struct{}{}
	cfg.Runners = append(cfg.Runners, runner.Runner{
		Target: runner.Accessor(d),
		Action: runner.ActionLookup,
		Run:    runner.Lookup,
	})
	return nil
}

func (cfg *Config) readEndpointMapV2(blob []byte) error {
	var e endpointMapV2
	if err := yaml.Unmarshal(blob, &e); err != nil {
		return err
	}

	for src, dst := range e.EndpointMap {
		if existing, ok := cfg.EndpointMap[src]; ok && existing != dst {
			return fmt.Errorf("Conflicting endpoint mappings:\n%s => %s\nand\n%s => %s", src, existing, src, dst)
		}
		cfg.EndpointMap[src] = dst
	}
	return nil
}

func (cfg *Config) readConfigV1(blob []byte) error {
	var (
		d   deploymentV1
		err error
	)
	if err = yaml.Unmarshal(blob, &d); err != nil {
		return err
	}
	deploymentRunner, err := validateV1(d, string(blob))
	if err != nil {
		return err
	}

	deployment := deploymentRunner.Target.(connection.Deployment)
	if filtered(deployment) {
		return nil
	}

	if _, ok := cfg.dNames[deployment.GetName()]; ok {
		return fmt.Errorf("Deployment names must be unique, but '%s' is specified more than once",
			deployment.GetName())
	}
	cfg.dNames[deployment.GetName()] = struct{}{}
	cfg.Runners = append(cfg.Runners, deploymentRunner)

	return nil
}

func splitYAMLObjects(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, []byte("\n---")); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return
}

var clusterFilter map[string]struct{}
var datacenterFilter map[string]struct{}

func filtered(deployment connection.Deployment) bool {
	if len(deployment.GetCluster()) > 0 {
		// At this point the deployment has already been validated, so
		// we can safely assume this means a cluster deployment
		return filterByCluster(deployment)
	}
	return filterByDatacenter(deployment)
}

func filterByCluster(d connection.Deployment) bool {
	if len(clusterFilter) == 0 {
		return false
	}
	_, ok := clusterFilter[d.GetCluster()]
	return !ok
}

func filterByDatacenter(d connection.Deployment) bool {
	if len(datacenterFilter) == 0 {
		return false
	}
	_, ok := datacenterFilter[d.GetDatacenter()]
	return !ok
}

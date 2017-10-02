package config

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"

	cxn "github.com/benjdewan/pachelbel/connection"
)

func TestSplitYAMLObjects(t *testing.T) {
	for i, test := range splitYAMLObjectsTests {
		actual := []string{}
		scanner := bufio.NewScanner(bytes.NewReader([]byte(test.input)))
		scanner.Split(splitYAMLObjects)
		for scanner.Scan() {
			actual = append(actual, scanner.Text())
		}
		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("Test #%d: Input '%s' expected [%v] but got [%v]",
				i, test.input, strings.Join(test.expected, ", "),
				strings.Join(actual, ", "))
		}
	}
}

func TestFiltered(t *testing.T) {
	setValidGlobals()
	for i, test := range filteredTests {
		clusterFilter = test.clusterFilter
		actual := filtered(cxn.Deployment(test.deployment))
		if actual != test.expected {
			t.Errorf("Test #%d: With filter: %v\nDeployment: %v\nExpected '%v' but saw '%v'",
				i, clusterFilter, test.deployment, test.expected,
				actual)
		}
	}
}

func TestReadConfig(t *testing.T) {
	setValidGlobals()
	for i, test := range readConfigTests {
		c := newConfig()
		err := c.readConfig([]byte(test.config))
		if test.valid {
			if err != nil {
				t.Errorf("Test #%d: Expected config to be valid, but got:\n%v",
					i, err)
			}
		} else {
			if err == nil {
				t.Errorf("Test #%d: %v\n should be invalid, but it wasn't caught",
					i, test.config)
			}
		}
	}
}

func setValidGlobals() {
	Databases = map[string][]string{
		"mongodb":        {},
		"rethink":        {},
		"elastic_search": {},
		"redis":          {"3.2.9"},
		"postgresql":     {"9.6.3", "9.6.4", "9.6.5"},
		"rabbitmq":       {},
		"etcd":           {},
		"mysql":          {},
		"janusgraph":     {},
		"scylla":         {},
		"disque":         {},
	}
	Clusters = map[string]string{
		"valid":      "123456789",
		"also-valid": "987654321",
	}
	Datacenters = map[string]struct{}{
		"aws:us-east-1":      {},
		"softlayer:dallas-1": {},
	}
}

var splitYAMLObjectsTests = []struct {
	input    string
	expected []string
}{
	{input: "foo", expected: []string{"foo"}},
	{input: "foo---bar", expected: []string{"foo---bar"}},
	{
		input:    "foo\n---bar",
		expected: []string{"foo", "---bar"},
	},
	{
		input:    "foo\n ---bar",
		expected: []string{"foo\n ---bar"},
	},
}

var emptyClusterFilter = make(map[string]struct{})
var oneClusterFilter = map[string]struct{}{
	"do-not-update": {},
}

var filteredTests = []struct {
	deployment    deploymentV1
	clusterFilter map[string]struct{}
	expected      bool
}{
	{
		deployment: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Cluster:       "valid",
			WiredTiger:    true,
		},
		clusterFilter: emptyClusterFilter,
		expected:      false,
	},
	{
		deployment: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Cluster:       "also-valid",
			WiredTiger:    true,
		},
		clusterFilter: oneClusterFilter,
		expected:      true,
	},
}
var readConfigTests = []struct {
	config string
	valid  bool
}{
	{
		config: "config_version: 0",
		valid:  false,
	},
	{
		config: "config_version: 1",
		valid:  false,
	},
	{
		config: `---
config_version: 1,
type: "invalid-type"
name: "names-are-not-validated"
datacenter: "aws:us-east-1"`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
cluster: "valid"`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
cluster: "also-valid"
datacenter: "aws:us-east-1"`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "aws:us-east-1"
scaling: 0`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"
scaling: 2`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"
wired_tiger: true`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"
cache_mode: true`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"
version: 3.2.9`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"
version: 11.0`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: postgresql
name: names-are-not-validated
datacenter: aws:us-east-1
version: 9.6.*`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: postgresql
name: names-are-not-validated
datacenter: aws:us-east-1
version: 9.X`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: postgresql
name: names-are-not-validated
datacenter: aws:us-east-1
version: "*"`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: postgresql
name: names-are-not-validated
datacenter: aws:us-east-1
version: 10.0`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "mongodb"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"
cache_mode: true`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "mongodb"
name: "names-are-not-validated"
datacenter: "aws:us-east-1"
wired_tiger: true`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "aws:us-east-1"
teams:
  - id: "team-ids-are-not-validated"
    role: "not-a-real-role"`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "softlayer:dallas-1"
teams:
  - id: "team-ids-are-not-validated"
    role: "admin"
  - id: "seriously-they-arent"
    role: "not-a-real-role"`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "aws:us-east-1"
teams:
  - id: "team-ids-are-not-validated"
    role: "admin"
  - id: "team-ids-are-not-validated"
    role: "manager"
  - id: "team-ids-are-not-validated"
    role: "developer"`,
		valid: true,
	},
	{
		config: `config_version: 2`,
		valid:  false,
	},
	{
		config: `---
config_version: 2`,
		valid: false,
	},
	{
		config: `---
config_version: 2
object_type: deployment`,
		valid: false,
	},
	{
		config: `---
config_version: 2
object_type: endpoint_map`,
		valid: true,
	},
	{
		config: `---
config_version: 2
object_type: endpoint_map
endpoint_map: `,
		valid: true,
	},
	{
		config: `---
config_version: 2
object_type: endpoint_map
endpoint_map:
  foo: bar
  baz: qux
  mumble: mamble`,
		valid: true,
	},
	{
		config: `---
config_version: 2
object_type: endpoint_map
endpoint_map:
  foo: bar
  bar: qux`,
		valid: true,
	},
	{
		config: `config_version: 2
object_type: deployment_client
name: foo
type: redis`,
		valid: true,
	},
	{
		config: `config_version: 2
object_type: deployment_client
name: foo
type: bar`,
		valid: false,
	},
	{
		config: `config_version: 2
object_type: deployment_client
name: foo`,
		valid: false,
	},
	{
		config: `config_version: 2
object_type: deployment_client
type: redis`,
		valid: false,
	},
}

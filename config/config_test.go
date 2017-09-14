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

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/benjdewan/pachelbel/connection"
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
	for i, test := range filteredTests {
		clusterFilter = test.clusterFilter
		actual := filtered(connection.Deployment(test.deployment))
		if actual != test.expected {
			t.Errorf("Test #%d: With filter: %v\nDeployment: %v\nExpected '%v' but saw '%v'",
				i, clusterFilter, test.deployment, test.expected,
				actual)
		}
	}
}

func TestReadConfig(t *testing.T) {
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
			Cluster:       "clusters-are-not-validated",
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
			Cluster:       "clusters-are-not-validated",
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
datacenter: "datacenters-are-not-validated"`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
cluster: "clusters-are-not-validated"`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "datacenters-are-not-validated"`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
cluster: "clusters-are-not-validated"
datacenter: "datacenters-are-not-validated"`,
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
datacenter: "datacenters-are-not-validated"
scaling: 0`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "datacenters-are-not-validated"
scaling: 2`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "datacenters-are-not-validated"
wired_tiger: true`,
		valid: false,
	},
	{
		config: `---
config_version: 1
type: "mongodb"
name: "names-are-not-validated"
datacenter: "datacenters-are-not-validated"
wired_tiger: true`,
		valid: true,
	},
	{
		config: `---
config_version: 1
type: "redis"
name: "names-are-not-validated"
datacenter: "datacenters-are-not-validated"
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
datacenter: "datacenters-are-not-validated"
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
datacenter: "datacenters-are-not-validated"
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
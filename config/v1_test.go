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

import "testing"

func TestValidateV1(t *testing.T) {
	for i, test := range configValidateV1Tests {
		err := validateV1(test.config, "ignored")
		if test.valid {
			if err != nil {
				t.Errorf("Test #%d: Expected\n%v\n to be valid, but saw: %v",
					i, test.config, err)
			}
		} else if err == nil {
			t.Errorf("Test #%d: Expected\n%v\n to be invalid, but there was no error thrown",
				i, test.config)
		}
	}
}

var (
	validScaling   int = 2
	invalidScaling int
)

var configValidateV1Tests = []struct {
	config deploymentV1
	valid  bool
}{
	{
		config: deploymentV1{ConfigVersion: 0},
		valid:  false,
	},
	{
		config: deploymentV1{ConfigVersion: 1},
		valid:  false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "invalid-type",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Cluster:       "clusters-are-not-validated",
		},
		valid: true,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
		},
		valid: true,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Tags:          []string{"tags-are-not-validated"},
		},
		valid: true,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Tags:          []string{"tags-are-not-validated"},
			Datacenter:    "datacenters-are-not-validated",
			Cluster:       "clusters-are-not-validated",
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Tags:          []string{"tags-are-not-validated"},
			Datacenter:    "datacenters-are-not-validated",
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Tags:          []string{"tags-are-not-validated"},
			Cluster:       "clusters-are-not-validated",
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Cluster:       "clusters-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
			Scaling:       &invalidScaling,
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
			Scaling:       &validScaling,
		},
		valid: true,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
			WiredTiger:    true,
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "mongodb",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
			WiredTiger:    true,
		},
		valid: true,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
			Teams: [](*TeamV1){
				&TeamV1{
					ID:   "team-ids-are-not-validated",
					Role: "not-a-real-role",
				},
			},
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
			Teams: [](*TeamV1){
				&TeamV1{
					ID:   "team-ids-are-not-validated",
					Role: "admin",
				},
				&TeamV1{
					ID:   "team-ids-are-not-validated",
					Role: "not-a-real-role",
				},
			},
		},
		valid: false,
	},
	{
		config: deploymentV1{
			ConfigVersion: 1,
			Type:          "redis",
			Name:          "names-are-not-validated",
			Datacenter:    "datacenters-are-not-validated",
			Teams: [](*TeamV1){
				&TeamV1{
					ID:   "team-ids-are-not-validated",
					Role: "admin",
				},
				&TeamV1{
					ID:   "team-ids-are-not-validated",
					Role: "developer",
				},
				&TeamV1{
					ID:   "team-ids-are-not-validated",
					Role: "manager",
				},
			},
		},
		valid: true,
	},
}

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
	"fmt"
	"strings"
)

func validateV1(d deploymentV1, input string) error {
	errs := []string{}

	errs = append(errs, validateConfigVersionV1(d.ConfigVersion)...)
	errs = append(errs, validateType(d.Type)...)
	errs = append(errs, validateDeploymentTargetV1(d.Cluster, d.Datacenter, d.Tags)...)
	errs = append(errs, validateName(d.Name)...)
	errs = append(errs, validateScaling(d.Scaling)...)
	errs = append(errs, validateWiredTiger(d.WiredTiger, d.Type)...)
	errs = append(errs, validateCacheMode(d.CacheMode, d.Type)...)
	errs = append(errs, validateTeams(d.Teams)...)

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("Errors occurred while parsing the following deployment object:\n%s\nErrors:\n%s",
		input, strings.Join(errs, "\n"))
}

func validateConfigVersionV1(version int) []string {
	if version != 1 {
		return []string{"Unsupported or missing 'config_version' field"}
	}
	return []string{}
}

func validateDeploymentTargetV1(cluster, datacenter string, tags []string) []string {
	if xor3(len(cluster) > 0, len(datacenter) > 0, len(tags) > 0) {
		return []string{}
	}
	return []string{"Exactly one of the 'cluster', 'datacenter', or 'tags' fields must be provided for every deployment\n"}
}

func xor3(a, b, c bool) bool {
	return xor(xor(a, b), c) && !(a && b && c)
}

func xor(a, b bool) bool {
	return (!(a && b) && (!(!a && !b)))
}

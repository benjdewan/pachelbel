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
	"math/rand"
	"strings"
)

var schemes = map[string]string{
	"postgresql":    "postgres",
	"redis":         "redis",
	"rabbitmq":      "amqps",
	"elasticsearch": "https",
	"etcd":          "https",
	"janus":         "https",
	"scylla":        "https",
	"mongodb":       "mongodb",
	"mysql":         "mysql",
}

var userpass = []string{
	"mario:175Am31",
	"luigi:gr33nM4r10",
	"zelda:br347h0ft3hw1ld",
	"gfreeman:1,2...",
	"doomguy:s3cr37_r00m5",
	"bjblazkowicz:h3lm3t5t4ck5",
	"admin:admin",
	"alice:Ez57510qVFnK7obJYKr3",
}

func dryRunCreate(cxn *Connection, deployment Deployment) error {
	cxn.newDeploymentIDs.Store(fakeID(deployment), struct{}{})
	return nil
}

func dryRunUpdate(cxn *Connection, deployment Deployment) error {
	existing, ok := cxn.getDeploymentByName(deployment.GetName())
	if !ok {
		return fmt.Errorf("Attempting to update '%s', but it doesn't exist",
			deployment.GetName())
	}
	cxn.newDeploymentIDs.Store(existing.ID, struct{}{})
	return nil
}

func fakeConnectionString(id string) ([]byte, error) {
	return []byte(fmt.Sprintf("%s://%s@pachelbel-dry-run.compose.direct:%d",
		schemes[strings.Split(id, "::")[0]],
		userpass[rand.Intn(len(userpass))],
		rand.Intn(6497)+3)), nil
}

func fakeID(deployment Deployment) string {
	return fmt.Sprintf("%s::%s", deployment.GetType(), deployment.GetName())
}

func isFake(id string) bool {
	return strings.Contains(id, "::")
}

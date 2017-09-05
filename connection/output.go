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
	"net/url"
	"os"
	"strconv"

	"github.com/ghodss/yaml"
)

type outputYAML struct {
	Type        string           `json:"type"`
	CACert      string           `json:"cacert,omitempty"`
	Connections []connectionYAML `json:"connections"`
}

// codebeat:disable[TOO_MANY_IVARS]
type connectionYAML struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     int    `json:"port,omitempty"`
	Path     string `json:"path,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

func connectionYAMLByID(cxn *Connection, connections []map[string]outputYAML, id string) ([]map[string]outputYAML, error) {
	var cxnYAML map[string]outputYAML
	var err error
	if cxn.dryRun && isFake(id) {
		cxnYAML = fakeOutputYAML(id)
	} else {
		cxnYAML, err = getOutputYAML(cxn, id)
	}
	return append(connections, cxnYAML), err
}

func getOutputYAML(cxn *Connection, id string) (map[string]outputYAML, error) {
	deployment, errs := cxn.client.GetDeployment(id)
	if len(errs) != 0 {
		return make(map[string]outputYAML), fmt.Errorf("%v", errsOut(errs))
	}
	connections, err := extractConnectionsYAML(deployment.Connection.Direct)
	if err != nil {
		return make(map[string]outputYAML), err
	}
	cxnYAML := make(map[string]outputYAML)
	cxnYAML[deployment.Name] = outputYAML{
		Type:        deployment.Type,
		CACert:      deployment.CACertificateBase64,
		Connections: connections,
	}
	return cxnYAML, nil
}

func extractConnectionsYAML(connectionStrings []string) ([]connectionYAML, error) {
	cxnYAMLs := []connectionYAML{}
	for _, cString := range connectionStrings {
		uri, err := url.Parse(cString)
		if err != nil {
			return cxnYAMLs, err
		}
		cxnYAMLs = append(cxnYAMLs, newConnectionYAML(uri))
	}
	return cxnYAMLs, nil
}

func newConnectionYAML(u *url.URL) connectionYAML {
	password, _ := u.User.Password()
	cxnYAML := connectionYAML{
		Scheme:   u.Scheme,
		Host:     u.Hostname(),
		Path:     u.Path,
		Username: u.User.Username(),
		Password: password,
	}
	if port, err := strconv.Atoi(u.Port()); err == nil {
		cxnYAML.Port = port
	}
	return cxnYAML
}

func writeConnectionYAML(cxnYAMLs []map[string]outputYAML, file string) error {
	outBytes, err := yaml.Marshal(cxnYAMLs)
	if err != nil {
		return err
	}
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

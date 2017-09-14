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
	"bytes"
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

func (cxn *Connection) connectionYAMLByID(endpointMap map[string]string, id string) ([]byte, error) {
	if cxn.dryRun && isFake(id) {
		return fakeOutputYAML(id)
	}
	return cxn.getOutputYAML(endpointMap, id)
}

func (cxn *Connection) getOutputYAML(endpointMap map[string]string, id string) ([]byte, error) {
	deployment, errs := cxn.client.GetDeployment(id)
	if len(errs) != 0 {
		return []byte{}, fmt.Errorf("%v", errsOut(errs))
	}
	connections, err := extractConnectionsYAML(endpointMap,
		deployment.Connection.Direct)
	if err != nil {
		return []byte{}, err
	}
	cxnYAML := make(map[string]outputYAML)
	cxnYAML[deployment.Name] = outputYAML{
		Type:        deployment.Type,
		CACert:      deployment.CACertificateBase64,
		Connections: connections,
	}
	return yaml.Marshal(cxnYAML)
}

func extractConnectionsYAML(endpointMap map[string]string, connectionStrings []string) ([]connectionYAML, error) {
	cxnYAMLs := []connectionYAML{}
	for _, cString := range connectionStrings {
		uri, err := url.Parse(cString)
		if err != nil {
			return cxnYAMLs, err
		}
		cxnYAMLs = append(cxnYAMLs, newConnectionYAML(endpointMap, uri))
	}
	return cxnYAMLs, nil
}

func newConnectionYAML(endpointMap map[string]string, u *url.URL) connectionYAML {
	password, _ := u.User.Password()
	host := u.Hostname()

	// host renaming can be recursive, but cycle chasing is difficult, so for
	// now do at most 8 layers of mapping and no more
	for i := 0; i < 7; i++ {
		if i == 15 {
			panic("Endpoint mapping is cyclic or deeper than 7 layers")
		}
		if rename, ok := endpointMap[host]; ok {
			host = rename
			continue
		}
		break
	}
	cxnYAML := connectionYAML{
		Scheme:   u.Scheme,
		Host:     host,
		Path:     u.Path,
		Username: u.User.Username(),
		Password: password,
	}
	if port, err := strconv.Atoi(u.Port()); err == nil {
		cxnYAML.Port = port
	}
	return cxnYAML
}

func writeConnectionYAML(cxnYAML [][]byte, file string) error {
	outBytes := bytes.Join(cxnYAML, []byte("\n"))
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

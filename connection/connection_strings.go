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
	"os"

	"github.com/ghodss/yaml"
)

func connectionStringsForDeployment(cxn *Connection, id string) ([]byte, error) {
	deployment, errs := cxn.client.GetDeployment(id)
	if len(errs) != 0 {
		return []byte{}, fmt.Errorf("%v", errsOut(errs))
	}

	cxnStringsObject := make(map[string][]string)
	cxnStringsObject[deployment.Name] = deployment.Connection.Direct

	return yaml.Marshal(cxnStringsObject)
}

func writeConnectionStrings(cxnStrings [][]byte, file string) error {
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

	_, err = handle.Write(bytes.Join(cxnStrings, []byte("\n---\n")))
	return err
}

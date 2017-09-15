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

package output

import (
	"reflect"
	"testing"
)

func TestConvertConnection(t *testing.T) {
	b := Builder{endpointMap: make(map[string]string)}
	for i, test := range convertConnectionTests {
		actual, err := b.convertConnection(test.input)
		if err != nil && test.valid {
			t.Errorf("Test #%d: '%s' is valid, but saw: %v", i, test.input, err)
			continue
		}
		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("Test #%d: Expected '%v' but saw '%v", i, test.expected, actual)
		}
	}
}

var convertConnectionTests = []struct {
	input    string
	expected []connectionYAML
	valid    bool
}{
	{
		input: "postgres://user:pass@compose.com:12345/compose",
		expected: []connectionYAML{
			{
				Scheme:   "postgres",
				Host:     "compose.com",
				Port:     12345,
				Path:     "/compose",
				Username: "user",
				Password: "pass",
			},
		},
		valid: true,
	},
	{
		input: "mongodb://user:pass@compose.com:12345,compose-2.com:9876/admin?ssl=true",
		expected: []connectionYAML{
			{
				Scheme:   "mongodb",
				Host:     "compose.com",
				Port:     12345,
				Path:     "/admin",
				Query:    "ssl=true",
				Username: "user",
				Password: "pass",
			},
			{
				Scheme:   "mongodb",
				Host:     "compose-2.com",
				Port:     9876,
				Path:     "/admin",
				Query:    "ssl=true",
				Username: "user",
				Password: "pass",
			},
		},
		valid: true,
	},
}

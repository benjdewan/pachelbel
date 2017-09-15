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

package progress

import "testing"

func TestCenter(t *testing.T) {
	for _, test := range centerTests {
		actual := center(test.str, test.width)
		if actual != test.expected {
			t.Errorf("Expected '%s' (%d), but saw '%s', (%d)",
				test.expected, len(test.expected), actual, len(actual))
		}
	}
}

var centerTests = []struct {
	str      string
	width    int
	expected string
}{
	{str: "DONE", width: 4, expected: "DONE"},
	{str: "ERROR", width: 4, expected: "ERRO"},
	{str: "DONE", width: 15, expected: "     DONE      "},
	{str: "", width: 10, expected: "          "},
}

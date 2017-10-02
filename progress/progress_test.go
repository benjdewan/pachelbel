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

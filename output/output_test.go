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

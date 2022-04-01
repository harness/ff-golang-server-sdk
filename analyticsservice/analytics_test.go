package analyticsservice

import "testing"

func Test_convertInterfaceToString(t *testing.T) {
	testCases := map[string]struct {
		input    interface{}
		expected string
	}{
		"Given I input a string": {
			input:    "123",
			expected: "123",
		},
		"Given I input a bool with the value false": {
			input:    false,
			expected: "false",
		},
		"Given I input a bool with the value true": {
			input:    true,
			expected: "true",
		},
		"Given I input an int64": {
			input:    int64(123),
			expected: "123",
		},
		"Given I input an int": {
			input:    123,
			expected: "123",
		},
		"Given I input a float64": {
			input:    float64(2.5),
			expected: "2.5",
		},
		"Given I input a float32": {
			input:    float32(2.5),
			expected: "2.5",
		},
		"Given I input a nil value": {
			input:    nil,
			expected: "nil",
		},
	}

	for desc, tc := range testCases {
		tc := tc
		t.Run(desc, func(t *testing.T) {

			actual := convertInterfaceToString(tc.input)
			if actual != tc.expected {
				t.Errorf("(%s): expected %s, actual %s", desc, tc.expected, actual)
			}
		})
	}
}

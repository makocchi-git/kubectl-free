package util

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValidateThreshold(t *testing.T) {

	var tests = []struct {
		description string
		warn        int64
		crit        int64
		expected    error
	}{
		{"warn:0 crit:0", 0, 0, nil},
		{"warn:10 crit:20", 10, 20, nil},
		{"warn:20 crit:10", 20, 10, fmt.Errorf("can not set critical threshold less than warn threshold (warn:20 crit:10)")},
	}

	for _, test := range tests {

		t.Run(test.description, func(t *testing.T) {
			actual := ValidateThreshold(test.warn, test.crit)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected(%v) differ (got: %v)", test.expected, actual)
			}
		})
	}
}

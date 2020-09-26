package tutil

import (
	"testing"
)

func TestDefaultString(t *testing.T) {
	tests := []struct {
		in     string
		def    string
		expect string
	}{
		{"", "root", "root"},
		{" ", "root", "root"},
		{"grooke", "root", "grooke"},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			out := DefaultString(test.in, test.def)
			if test.expect != out {
				t.Errorf("expect: '%s', got '%s'", test.expect, out)
			}
		})
	}
}

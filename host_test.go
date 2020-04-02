package gossh

import (
	"fmt"
	"testing"
)

func TestNSpaces(t *testing.T) {

	var tests = []struct {
		in     int
		expect string
	}{
		{-9, ""},
		{-2, ""},
		{-1, ""},
		{0, ""},
		{1, " "},
		{2, " │"},
		{9, " │ │ │ │ "},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%d", test.in), func(t *testing.T) {

			got := nSpaces(test.in)

			if got != test.expect {
				t.Errorf("value: got \"%s\" - expect \"%s\"", got, test.expect)
			}

		})
	}
}

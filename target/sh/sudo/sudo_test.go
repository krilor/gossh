package sudo

import (
	"testing"
)

func TestEscape(t *testing.T) {
	tests := []struct {
		in     string
		expect string
	}{
		{`'`, `'\''`},
		{`awk '{ print $1 }'`, `awk '\''{ print $1 }'\''`},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			out := Escape(test.in)
			if test.expect != out {
				t.Errorf("expect: %s, got %s", test.expect, out)
			}
		})
	}
}

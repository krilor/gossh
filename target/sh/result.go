package sh

import (
	"bytes"
	"strings"
)

// Result represents a reponse from a command
type Result struct {
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	ExitStatus int
}

// TrimOut returns Stdout, with trimmed ends
func (r Result) TrimOut() string {
	return strings.TrimSpace(r.Stdout.String())
}

// TrimErr returns stderr, with trimmed ends
func (r Result) TrimErr() string {
	return strings.TrimSpace(r.Stderr.String())
}

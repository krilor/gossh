package sh

import (
	"bytes"
	"strings"
)

// Response represents a reponse from a command
type Response struct {
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	ExitStatus int
}

// Out returns Stdout, with trimmed ends
func (r Response) Out() string {
	return strings.Trim(r.Stdout.String(), " \n")
}

// Err returns stderr, with trimmed ends
func (r Response) Err() string {
	return strings.Trim(r.Stderr.String(), " \n")
}

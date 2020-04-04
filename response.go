package gossh

import "bytes"

// Response contains the response from a remotely run cmd
type Response struct {
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	ExitStatus int
}

// ExitStatusSuccess is a convenience method to check if an exit code is either 0 or BlockedByValidate.
// It is provided as a form of syntactic sugar.
func (r Response) ExitStatusSuccess() bool {
	// TODO add exitStatuses ...int to allow for including more exit statuses as ok.
	return r.ExitStatus == 0 || r.ExitStatus == BlockedByValidate
}

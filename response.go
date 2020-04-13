package gossh

// Response contains the response from a remotely run cmd
type Response struct {
	Stdout     string
	Stderr     string
	ExitStatus int
}

// Success is a convenience method to check if an exit code is either 0 or BlockedByValidate.
// It is provided as a form of syntactic sugar.
func (r Response) Success() bool {
	// TODO add exitStatuses ...int to allow for including more exit statuses as ok.
	return r.ExitStatus == 0 || r.ExitStatus == BlockedByValidate
}

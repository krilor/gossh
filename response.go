package gossh

import (
	"regexp"
	"strings"
)

// Response contains the response from a remotely run cmd
type Response struct {
	Stdout     string
	Stderr     string
	ExitStatus int
}

// ExitStatusSuccess is a convenience method to check if an exit code is either 0 or BlockedByValidate.
// It is provided as a form of syntactic sugar.
func (r Response) ExitStatusSuccess() bool {
	// TODO add exitStatuses ...int to allow for including more exit statuses as ok.
	return r.ExitStatus == 0 || r.ExitStatus == BlockedByValidate
}

// sudopattern matches sudo prompt
var sudopattern *regexp.Regexp = regexp.MustCompile(`\[sudo\] password for [^:]+: `)

// scrubStd cleans an out/err string. Removes trailing newline and sudo prompt.
func scrubStd(in string) string {
	return sudopattern.ReplaceAllString(strings.Trim(in, "\n"), "")
}

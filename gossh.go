package gossh

// Package gossh provides interfaces and functionality for declarative IT automation on target vms or containers.

//go:generate stringer -type=Status

// Status indicates the status of a Rule.
type Status int

const (
	// StatusUndefined is set as the zero-value. Should only be used when error is returned, but consider using StatusFailed instead.
	StatusUndefined Status = iota

	// StatusSkipped means that the rule is skipped.
	// Useful for conditional (e.g. os-specific) rules.
	StatusSkipped

	// StatusSatisfied means rule was allready adhered to and no changes had to be made.
	StatusSatisfied

	// StatusNotOK means that a check was done, and it was not ok.
	// The only use for this status is when changes are blocked on the target - when Rules are applied in check-only mode.
	StatusNotOK

	// StatusChanged means that the rule did changes to the target to ensure that the declared rule was satisfied.
	// The changes was successful.
	StatusChanged

	// StatusFailed means someting went wrong. Usually returned when error is also returned.
	StatusFailed
)

// OK reports if the status should be considered as OK.
// Returns true if status is one of StatusSkipped, StatusSatisfied, StatusNotOK, StatusChanged
func (s Status) OK() bool {
	return s == StatusSkipped || s == StatusSatisfied || s == StatusNotOK || s == StatusChanged
}

// Target is a target for commands and rules
//
// Targets can be in a validate state, that does not allow for commands that RunChange the state of the system to run.
type Target interface {
	// Apply checks and ensures that the Target adheres to Rule r.
	//
	// String name should be unique within the immediate context, short and descriptive.
	Apply(name string, r Rule) (Status, error)

	// AllowChange reports if the target allows changes to be done or not.
	// False will be returned if target is in check-only mode.
	//
	// The function can be used by Rule implementers in cases where it makes more sense than running RunQuery and getting a blocked response.
	AllowChange() bool

	// RunChange runs the command cmd on the Target.
	//
	// RunChange does the same as RunQuery, but must ONLY be used for cmd's that modify state on the Target.
	//
	// Callers must handle a returned Response with ExitStatus BlockedByValidate, indicating that changes cannot be done to the target.
	// When BlockedByValidate is returned, stdout and stderr will be empty string and err will be nil.
	//
	// If user is not the connected user, sudo/su will be applied to the command.
	// Empty user means connected user. '-' is interpreted as 'root'.
	//
	// Stdin can be used to add stdin to the commmand.
	RunChange(cmd string, stdin string, user string) (Response, error)

	// RunQuery runs the command cmd on the Target.
	//
	// RunQuery does the same as RunChange, but must ONLY be used for cmd's that doesn't modify any state on the Target.
	//
	// If user is not the connected user, sudo/su will be applied to the command.
	// Empty user means connected user. '-' is interpreted as 'root'.
	//
	// Stdin can be used to add stdin to the commmand.
	RunQuery(cmd string, stdin string, user string) (Response, error)

	// Log can be used by clients to surface info-type logs.
	//
	// Errors should not be logged using this method, but rater retured to the caller.
	//
	// The msg argument should be used to add some constant description to
	// the log line. The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and values
	//
	// Log is inspired by the go-logr interface, but only for strings.
	Log(msg string, keysAndValues ...string)
}

// BlockedByValidate is an ExitCode used to indicate that the command was not run, but rather blocked because the target does not allow any modifications.
// The intended use for this is when validating rules.
const BlockedByValidate int = 81549300

// Rule is an interface that wraps the Ensure method
//
// Ensure runs commands on Target t to ensure that a declared state is adhered to.
//
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Rule interface {
	Ensure(t Target) (status Status, err error)
}

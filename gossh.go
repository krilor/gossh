package gossh

// Package gossh provides interfaces and functionality for declarative IT automation on remote hosts or containers.

//go:generate stringer -type=Status

// Status indicates the state of the rule.
type Status int

const (
	// StatusUndefined is set as the zero-value. Should only be used when error is returned, but consider using Failed instead.
	StatusUndefined Status = iota
	// StatusSkipped means that the rule should be skipped.
	// Useful only for conditional or os-specific rules
	// Should not be returned if rules are checked ok. In that case, Statisfied should be returned.
	StatusSkipped
	// StatusSatisfied means rule was allready adhered to, i.e. check was OK.
	StatusSatisfied
	// StatusCheckNotOK means that a check was done, and it was not ok.
	StatusCheckNotOK
	// StatusChanged means that the rule Ensure ran, and exited with success.
	StatusChanged
	// StatusFailed means someting went wrong. Usually returned when error is also returned.
	StatusFailed
)

// OK reports if the status should be considered as OK.
// Returns true if status is one of
// StatusSkipped,
// StatusSatisfied,
// StatusCheckNotOK,
// StatusChanged
func (s Status) OK() bool {
	return s == StatusSkipped || s == StatusSatisfied || s == StatusCheckNotOK || s == StatusChanged
}

// Target is a target for commands and rules
//
// Targets can be in a validate state, that does not allow for commands that RunChange the state of the system to run.
type Target interface {
	// Apply checks and ensures that the Target adheres to the rule r.
	//
	// String name should be unique within a rule, short and descriptive.
	Apply(name string, r Rule) (Status, error)

	// AllowChange reports if the machine allows changes to be done.
	// It can be used by rule implementers in cases where it makes more sense than running RunQuery and getting blocked response.
	AllowChange() bool

	// RunChange runs the command cmd on the Target.
	//
	// RunChange does the same as RunQuery, but must ONLY be used for cmd's that modify state on the Target.
	//
	// Callers must handle that when Target is in validate state,
	// RunChange will return Response with ExitStatus BlockedByValidate,
	// empty stdout/err and a nil error.
	//
	// Target will use sudo to change to user if it is not the connected user.
	// If user is an empty string, the connected user will be used.
	RunChange(cmd string, stdin string, user string) (Response, error)

	// RunQuery runs the command cmd on the Target.
	//
	// RunQuery does the same as RunChange, but must ONLY be used for cmd's that doesn't modify any state on the Target.
	//
	// Target will use sudo to change to user if it is not the connected user.
	// If user is an empty string, the connected user will be used.
	RunQuery(cmd string, stdin string, user string) (Response, error)

	// Log can be used by clients to surface info-type logs.
	//""
	// It is inspired by the go-logr interface, but only for strings.
	//
	// Errors should not be logged using this method, but rater retured to the caller.
	//
	// The msg argument should be used to add some constant description to
	// the log line. The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should RunChangenate string
	// keys and values
	Log(msg string, keysAndValues ...string)
}

// BlockedByValidate is an ExitCode used to indicate that the command was not run, but rather blocked because the target does not allow any modifications.
// The intended use for this is when validating rules.
const BlockedByValidate int = 81549300

// Rule is the interface that groups the Check and Ensure methods
//
// The main purpose of this combines interface is to have a Rule that conditionally run Ensure based on Check
//
// In go-speak, it should have been called a CheckEnsurer
type Rule interface {
	Ensure(t Target) (status Status, err error)
}

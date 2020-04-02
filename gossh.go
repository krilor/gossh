package gossh

// Package gossh provides interfaces and functionality for declarative IT automation on remote hosts or containers.

// Target is a target for commands and rules
//
// Targets can be in a validate state, that does not allow for commands that RunChange the state of the system to run.
type Target interface {
	// Apply checks and ensures that the Target adheres to the rule r.
	//
	// String name should be unique within a rule, short and descriptive.
	Apply(trace Trace, name string, r Rule) error

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
	RunChange(trace Trace, cmd string, user string) (Response, error)

	// RunQuery runs the command cmd on the Target.
	//
	// RunQuery does the same as RunChange, but must ONLY be used for cmd's that doesn't modify any state on the Target.
	//
	// Target will use sudo to change to user if it is not the connected user.
	// If user is an empty string, the connected user will be used.
	RunQuery(trace Trace, cmd string, user string) (Response, error)

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
	Log(trace Trace, msg string, keysAndValues ...string)
}

// BlockedByValidate is an ExitCode used to indicate that the command was not run, but rather blocked because the target does not allow any modifications.
// The intended use for this is when validating rules.
const BlockedByValidate int = 81549300

// Checker is the interface that wraps the Check method.
//
// Check runs commands on t to and reports to ok wether or not the rule is adhered to or not.
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Checker interface {
	Check(trace Trace, t Target) (ok bool, err error)
}

// Ensurer is the interface that wraps the Ensure method
//
// Ensure runs commands on t to ensure that a specified state is adhered to.
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Ensurer interface {
	Ensure(trace Trace, t Target) error
}

// Rule is the interface that groups the Check and Ensure methods
//
// The main purpose of this combines interface is to have a Rule that conditionally run Ensure based on Check
//
// In go-speak, it should have been called a CheckEnsurer
type Rule interface {
	Checker
	Ensurer
}

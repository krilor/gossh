package gossh

import (
	"io"
	"os"
)

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

	// StatusNotSatisfied means that a check was done, and it was not ok.
	// The only use for this status is when changes are blocked on the target - when Rules are applied in check-only mode.
	StatusNotSatisfied

	// StatusEnforced means that the rule did changes to the target to ensure that the declared rule was satisfied.
	// The changes was successful.
	StatusEnforced

	// StatusFailed means someting went wrong. Usually returned when error is also returned.
	StatusFailed
)

// OK reports if the status should be considered as OK.
// Returns true if status is one of StatusSkipped, StatusSatisfied, StatusNotSatisfied, StatusEnforced
func (s Status) OK() bool {
	return s == StatusSkipped || s == StatusSatisfied || s == StatusNotSatisfied || s == StatusEnforced
}

// FileSystem is an interface that wraps filesystem-related methods
//
// The underlying implementation of the filesystem can entail SCP, SFTP, shell commands etc.
// Since methods would most frequently have to do network calls or reads, the
// returned interfaces for most methods are pretty restricted.
//
// Most methods have duplicate method suffixed "As", which allows the caller to specify the user.
type FileSystem interface {

	// Create creates the named file mode 0666 (before umask), truncating it if it already exists.
	// The file is opened as write only. ( os.O_WRONLY|os.O_CREATE|os.O_TRUNC )
	Create(path string) (io.WriteCloser, error)
	CreateAs(user, path string) (io.WriteCloser, error)

	// Open opens the named file for reading.
	Open(path string) (io.ReadCloser, error)
	OpenAs(user, path string) (io.ReadCloser, error)

	// Chown changes the user and group of the named path.
	Chown(path, username, groupname string) error
	ChownAs(user, path, username, groupname string) error

	// Chmod changes the mode of the file to mode.
	Chmod(path string, mode os.FileMode) error
	ChmodAs(user, path string, mode os.FileMode) error

	// Append could be something to implement?

	// Mkdir creates the specified directory.
	// An error will be returned if a file or directory with the specified path already exists,
	// or if the directory's parent folder does not exist (the method cannot create complete paths).
	Mkdir(path string, perm os.FileMode) error
	MkdirAs(user, path string, perm os.FileMode) error

	// MkdirAll creates a directory named path, along with any necessary parents,
	// and returns nil, or else returns an error.
	// If path is already a directory, MkdirAll does nothing and returns nil.
	// If path contains a regular file, an error is returned
	MkdirAll(path string, perm os.FileMode) error
	MkdirAllAs(user, path string, perm os.FileMode) error

	// Lstat returns a FileInfo structure describing the file specified by path.
	// If path is a symbolic link, the returned FileInfo structure describes the symbolic link.
	Lstat(path string) (os.FileInfo, error)
	LstatAs(user, path string) (os.FileInfo, error)

	// Stat returns a FileInfo structure describing the file specified by path.
	// If path is a symbolic link, the returned FileInfo structure describes the referent file.
	Stat(path string) (os.FileInfo, error)
	StatAs(user, path string) (os.FileInfo, error)

	// Link creates newname as a hard link to the oldname file.
	Link(oldname, newname string) error
	LinkAs(user, oldname, newname string) error

	// Symlink creates newname as a symbolic link to oldname.
	Symlink(oldname, newname string) error
	SymlinkAs(user, oldname, newname string) error

	// ReadDirNames reads the contents of a directory and returns a list of directory entries
	ReadDirNames(path string) ([]string, error)
	ReadDirNamesAs(user, path string) ([]string, error)

	// Remove removes the named file or (empty) directory.
	Remove(path string) error
	RemoveAs(user, path string) error

	// RemoveAll removes path and any children it contains. It removes everything it can but returns the first error it encounters.
	// If the path does not exist, RemoveAll returns nil (no error).
	RemoveAll(path string) error
	RemoveAllAs(user, path string) error

	// Rename renames (moves) oldpath to newpath.
	// If newpath already exists and is not a directory, Rename replaces it.
	Rename(oldpath, newpath string) error
	RenameAs(user, oldpath, newpath string) error

	// Tempfile returns a temp filename that can be used by the current user
	TempFile() string
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
	// The function can be used by Rule implementers in cases where it makes more sense than running RunCheck and getting a blocked response.
	AllowChange() bool

	// RunChange runs the command cmd on the Target.
	//
	// RunChange does the same as RunCheck, but must ONLY be used for cmd's that modify state on the Target.
	//
	// Callers must handle a returned Response with ExitStatus BlockedByValidate, indicating that changes cannot be done to the target.
	// When BlockedByValidate is returned, stdout and stderr will be empty string and err will be nil.
	//
	// If user is not the connected user, sudo/su will be applied to the command.
	// Empty user means connected user. '-' is interpreted as 'root'.
	//
	// Stdin can be used to add stdin to the commmand.
	RunChange(cmd string, stdin string, user string) (Response, error)

	// RunCheck runs the command cmd on the Target.
	//
	// RunCheck does the same as RunChange, but must ONLY be used for cmd's that doesn't modify any state on the Target.
	//
	// If user is not the connected user, sudo/su will be applied to the command.
	// Empty user means connected user. '-' is interpreted as 'root'.
	//
	// Stdin can be used to add stdin to the commmand.
	RunCheck(cmd string, stdin string, user string) (Response, error)

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
// Ensure runs commands on Target t to check and enforce that a declared state is adhered to.
//
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Rule interface {
	Ensure(t Target) (status Status, err error)
}

package gossh

import (
	"fmt"
	"io"
	"os"

	"github.com/krilor/gossh/sh"
)

// Target is an interface that contains all methods that can be done to a host.
type Target interface {
	// Targets must implement fmt.Stringer
	fmt.Stringer

	// Close closes all underlying connections on Target.
	Close() error

	// As returns a new target with user as the active user.
	As(user string)

	// User returns the currently active user
	User() string

	// Run runs cmd on Target, using bash.
	//
	// Stdin can be used to add stdin to the commmand.
	// Targets should handle that stdin is nil, i.e. no stdin.
	Run(cmd string, stdin io.Reader) (sh.Response, error) // TODO response should be interface..

	// Mkdir creates the specified directory.
	// An error will be returned if a file or directory with the specified path already exists,
	// or if the directory's parent folder does not exist (the method cannot create complete paths).
	Mkdir(path string, perm os.FileMode) error

	// MkdirAll creates a directory named path, along with any necessary parents,
	// and returns nil, or else returns an error.
	// If path is already a directory, MkdirAll does nothing and returns nil.
	// If path contains a regular file, an error is returned
	// MkdirAll(path string, perm os.FileMode) error

	// Create creates the named file mode 0666 (before umask), truncating it if it already exists.
	// The file is opened as write only. ( os.O_WRONLY|os.O_CREATE|os.O_TRUNC )
	// Create(path string) (io.WriteCloser, error)

	// Open opens the named file for reading.
	// Open(path string) (io.ReadCloser, error)

	// Chown changes the user and group of the named path.
	// Chown(path, username, groupname string) error

	// Chmod changes the mode of the file to mode.
	// Chmod(path string, mode os.FileMode) error

	// Append returns a writer that appends to path.
	// Append(path string) (io.WriteCloser, error)

	// Lstat returns a FileInfo structure describing the file specified by path.
	// If path is a symbolic link, the returned FileInfo structure describes the symbolic link.
	// Lstat(path string) (os.FileInfo, error)

	// Stat returns a FileInfo structure describing the file specified by path.
	// If path is a symbolic link, the returned FileInfo structure describes the referent file.
	// Stat(path string) (os.FileInfo, error)

	// Link creates newname as a hard link to the oldname file.
	// Link(oldname, newname string) error

	// Symlink creates newname as a symbolic link to oldname.
	// Symlink(oldname, newname string) error

	// ReadDirNames reads the contents of a directory and returns a list of directory entries
	// ReadDirNames(path string) ([]string, error)

	// Remove removes the named file or (empty) directory.
	// Remove(path string) error

	// RemoveAll removes path and any children it contains. It removes everything it can but returns the first error it encounters.
	// If the path does not exist, RemoveAll returns nil (no error).
	// RemoveAll(path string) error

	// Rename renames (moves) oldpath to newpath.
	// If newpath already exists and is not a directory, Rename replaces it.
	// Rename(oldpath, newpath string) error

	// Tempfile returns a temp filename that can be used by the current user
	// TempFile() (string, error)

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
	// Log(msg string, keysAndValues ...string)
}

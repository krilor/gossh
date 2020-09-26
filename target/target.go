package target

import (
	"fmt"
	"io"

	"github.com/krilor/gossh/target/sh"
)

// Target is an interface that contains all basic methods that can be done to a host
type Target interface {
	// Targets must implement fmt.Stringer
	fmt.Stringer

	// Close closes all underlying connections on Target.
	Close() error

	// As returns a new target with user as the active user.
	As(user string)

	// User returns the connected user
	User() string

	// ActiveUser returns the currently active user on the target
	ActiveUser() string

	// Run runs cmd on target, using bash.
	//
	// Stdin can be used to add stdin to the commmand.
	// Targets should handle that stdin is nil, i.e. no stdin.
	Run(cmd string, stdin io.Reader) (sh.Result, error)

	// Create creates the named file mode 0666 (before umask), truncating it if it already exists.
	// The file is opened as write only. ( os.O_WRONLY|os.O_CREATE|os.O_TRUNC )
	Create(path string) (io.WriteCloser, error)

	// Open opens the named file for reading.
	Open(path string) (io.ReadCloser, error)

	// Append returns a writer that appends to path.
	Append(path string) (io.WriteCloser, error)
}

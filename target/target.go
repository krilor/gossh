package target

import (
	"fmt"
	"io"
	"os"

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

	// Put creates the named file in path with perm (before umask), truncating it if it already exists.
	// Data is written to the file. Modelled after ioutil.Writefile
	Put(filename string, data []byte, perm os.FileMode) error

	// Get reads the file named by path.
	// A successful call returns err == nil, not err == EOF. Because ReadFile reads the whole file, it does not treat an EOF from Read as an error to be reported.
	// Modelled after ioutil.Readfile
	Get(filename string) ([]byte, error)
}

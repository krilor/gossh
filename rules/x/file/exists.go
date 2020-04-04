package file

import (
	"fmt"

	"github.com/krilor/gossh"
	"github.com/pkg/errors"
)

// Exists is a struct that implements the rule to check if a file exists or not
type Exists struct {
	Path string
	User string
}

// Check if file exists
func (e Exists) Check(trace gossh.Trace, t gossh.Target) (bool, error) {

	cmd := fmt.Sprintf("stat %s", e.Path)

	r, err := t.RunQuery(trace, cmd, "", e.User)

	if err != nil {
		return false, errors.Wrap(err, "stat errored")
	}

	if r.ExitStatus != 0 {
		return false, nil
	}

	return true, nil
}

// Ensure that file exists
func (e Exists) Ensure(trace gossh.Trace, t gossh.Target) error {

	cmd := fmt.Sprintf("touch %s", e.Path)
	r, err := t.RunChange(trace, cmd, "", e.User)

	if err != nil {
		return errors.Wrap(err, "could not ensure file")
	}

	if !r.ExitStatusSuccess() {
		return errors.New("something went wrong with touch")
	}
	return nil
}

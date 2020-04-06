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

// check if file exists
func (e Exists) check(t gossh.Target) (bool, error) {

	cmd := fmt.Sprintf("stat %s", e.Path)

	r, err := t.RunCheck(cmd, "", e.User)

	if err != nil {
		return false, errors.Wrap(err, "stat errored")
	}

	if r.ExitStatus != 0 {
		return false, nil
	}

	return false, nil
}

// Ensure that file exists
func (e Exists) Ensure(t gossh.Target) (gossh.Status, error) {

	ok, err := e.check(t)

	if ok {
		return gossh.StatusSatisfied, nil
	}

	cmd := fmt.Sprintf("touch %s", e.Path)
	r, err := t.RunChange(cmd, "", e.User)

	if err != nil {
		return gossh.StatusFailed, errors.Wrap(err, "could not ensure file")
	}

	if !r.Success() {
		return gossh.StatusFailed, errors.New("something went wrong with touch")
	}
	return gossh.StatusEnforced, nil
}

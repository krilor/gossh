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
func (e Exists) check(h *gossh.Host) (bool, error) {

	cmd := fmt.Sprintf("stat %s", e.Path)

	r, err := h.RunCheck(cmd, "", e.User)

	if err != nil {
		return false, errors.Wrap(err, "stat errored")
	}

	if r.ExitStatus != 0 {
		return false, nil
	}

	return false, nil
}

// Ensure that file exists
func (e Exists) Ensure(h *gossh.Host) (gossh.Status, error) {

	ok, err := e.check(h)

	if ok {
		return gossh.StatusSatisfied, nil
	}

	cmd := fmt.Sprintf("touch %s", e.Path)
	r, err := h.RunChange(cmd, "", e.User)

	if err != nil {
		return gossh.StatusFailed, errors.Wrap(err, "could not ensure file")
	}

	if !r.Success() {
		return gossh.StatusFailed, fmt.Errorf("something went wrong with touch %d", r.ExitStatus)
	}
	return gossh.StatusEnforced, nil
}

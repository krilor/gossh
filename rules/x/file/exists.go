package file

import (
	"fmt"

	"github.com/krilor/gossh"
	"github.com/pkg/errors"
)

// Exists is a struct that implements the rule to check if a file exists or not
type Exists string

// Check if file exists
func (e Exists) Check(trace gossh.Trace, m *gossh.Host) (bool, error) {

	cmd := fmt.Sprintf("stat %s", string(e))

	r, err := m.Run(trace, cmd, false)

	if err != nil {
		return false, errors.Wrap(err, "stat errored")
	}

	if r.ExitStatus != 0 {
		return false, nil
	}

	return true, nil
}

// Ensure that file exists
func (e Exists) Ensure(trace gossh.Trace, m *gossh.Host) error {
	cmd := fmt.Sprintf("touch %s", string(e))
	r, err := m.Run(trace, cmd, false)
	if err != nil {
		return errors.Wrap(err, "could not ensure file")
	}
	if r.ExitStatus != 0 {
		return errors.New("something went wrong with touch")
	}
	return nil
}

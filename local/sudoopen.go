package local

import (
	"io"
	"os/exec"

	"github.com/krilor/gossh/sh"
	"github.com/pkg/errors"
)

// sudoOpen returns a sudoReadCloser for the path
func sudoOpen(path, user, pwd string) (io.ReadCloser, error) {

	s := sudoReadCloser{}

	sudo := sh.NewSudo("cat "+path, user, pwd, nil)
	s.cmd = exec.Command("sudo", sudo.Args()...)

	var err error
	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get stdout")
	}

	err = s.cmd.Start()

	if err != nil {
		return nil, errors.Wrap(err, "could not start cat")
	}

	return &s, nil
}

// sudoReadCloser is a WriteCloser that that uses cat to pipe stdin data to a specified file.
type sudoReadCloser struct {
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func (s *sudoReadCloser) Read(p []byte) (n int, err error) {
	return s.stdout.Read(p)
}

func (s *sudoReadCloser) Close() error {
	err := s.stdout.Close()
	if err != nil {
		return errors.Wrap(err, "unable to close stdin")
	}

	return s.cmd.Wait()
}

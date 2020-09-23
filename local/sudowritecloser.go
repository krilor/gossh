package local

import (
	"io"
	"os/exec"

	"github.com/krilor/gossh/sh"
	"github.com/pkg/errors"
)

// sudoOpenFile returns a sudoWriteCloser for the path
func sudoOpenFile(path, user, pwd string) (io.WriteCloser, error) {

	s := sudoWriteCloser{}

	sudo := sh.NewSudo("cat > "+path, user, pwd, nil)
	s.cmd = exec.Command("sudo", sudo.Args()...)

	var err error
	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get stdout")
	}
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get stdin")
	}

	err = s.cmd.Start()

	if err != nil {
		return nil, errors.Wrap(err, "could not start cat")
	}

	return &s, nil
}

// sudoWriteCloser is a WriteCloser that that uses cat to pipe stdin data to a specified file.
type sudoWriteCloser struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func (s *sudoWriteCloser) Write(p []byte) (n int, err error) {
	return s.stdin.Write(p)
}

func (s *sudoWriteCloser) Close() error {
	err := s.stdin.Close()
	if err != nil {
		return errors.Wrap(err, "unable to close stdin")
	}

	return s.cmd.Wait()
}

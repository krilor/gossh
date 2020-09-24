package local

import (
	"io"
	"os/exec"

	"github.com/krilor/gossh/sh"
	"github.com/pkg/errors"
)

type sudoOpenMode string

const (
	modeCreate sudoOpenMode = ">"
	modeAppend sudoOpenMode = ">>"
	modeRead   sudoOpenMode = ""
)

// sudoOpen returns a sudoWriteCloser for the path
func sudoOpen(path, user, pwd string, mode sudoOpenMode) (*sudoReadWriteCloser, error) {

	s := sudoReadWriteCloser{}

	sudo := sh.NewSudo("cat "+string(mode)+" "+path, user, pwd, nil)
	s.cmd = exec.Command("sudo", sudo.Args()...)

	var err error

	if mode == modeRead {
		s.stdout, err = s.cmd.StdoutPipe()
		if err != nil {
			return nil, errors.Wrap(err, "could not get stdin")
		}
	} else {
		s.stdin, err = s.cmd.StdinPipe()
		if err != nil {
			return nil, errors.Wrap(err, "could not get stdin")
		}
	}

	err = s.cmd.Start()

	if err != nil {
		return nil, errors.Wrap(err, "could not start cat")
	}

	return &s, nil
}

// sudoWriteCloser is a WriteCloser that that uses cat to pipe stdin data to a specified file.
type sudoReadWriteCloser struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func (s *sudoReadWriteCloser) Write(p []byte) (n int, err error) {
	return s.stdin.Write(p)
}

func (s *sudoReadWriteCloser) Read(p []byte) (n int, err error) {
	return s.stdout.Read(p)
}

func (s *sudoReadWriteCloser) Close() error {

	var err error
	if s.stdin != nil {
		err = s.stdin.Close()
	} else {
		err = s.stdout.Close()
	}

	if err != nil {
		return errors.Wrap(err, "unable to close stdin")
	}

	return s.cmd.Wait()
}

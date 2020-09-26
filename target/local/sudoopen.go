package local

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/krilor/gossh/target/sh/sudo"
	"github.com/pkg/errors"
)

type sudoOpenMode string

const (
	modeCreate sudoOpenMode = "tee >"
	modeAppend sudoOpenMode = "tee >>"
	modeRead   sudoOpenMode = "cat"
)

// sudoOpen returns a sudoWriteCloser for the path
func sudoOpen(path, user, pwd string, mode sudoOpenMode) (*sudoReadWriteCloser, error) {

	s := sudoReadWriteCloser{mode: mode}
	var stdin io.Reader
	stdin, s.stdinpipe = io.Pipe()

	s.sudo = sudo.New(string(mode)+" "+path, user, pwd, stdin)
	s.cmd = exec.Command("sudo", s.sudo.Args()...)

	var err error

	s.cmd.Stderr = s.sudo
	s.sudo.Stderr = &s.stderr

	s.sudo.StdinPipe, err = s.cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get stdinpipe")
	}

	s.stdoutpipe, err = s.cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get stdoutpipe")
	}

	err = s.cmd.Start()

	if err != nil {
		return nil, errors.Wrap(err, "could not start cat")
	}

	return &s, nil
}

// sudoWriteCloser is a WriteCloser that that uses cat to pipe stdin data to a specified file.
type sudoReadWriteCloser struct {
	mode       sudoOpenMode
	stdinpipe  io.WriteCloser
	stdoutpipe io.ReadCloser
	stderr     bytes.Buffer
	cmd        *exec.Cmd
	sudo       *sudo.Sudo
}

func (s *sudoReadWriteCloser) Write(p []byte) (n int, err error) {
	return s.stdinpipe.Write(p)
}

func (s *sudoReadWriteCloser) Read(p []byte) (n int, err error) {
	return s.stdoutpipe.Read(p)
}

func (s *sudoReadWriteCloser) Close() error {

	var err error

	err = s.stdinpipe.Close()
	if err != nil {
		return errors.Wrap(err, "could not close stdinpipe")
	}

	if s.mode != modeRead {
		err = s.stdoutpipe.Close()
		if err != nil {
			return errors.Wrap(err, "could not close stdinpipe")
		}
	}

	if err != nil {
		return errors.Wrap(err, "could not close stdinpipe")
	}

	err = s.cmd.Wait()

	if err != nil {
		return errors.Wrap(err, s.stderr.String())
	}

	return nil
}

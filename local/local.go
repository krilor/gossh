package local

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/krilor/gossh/sh"
	"github.com/pkg/errors"
)

// Package local contains functionality for doing work on localhost.

// Local implements a localhost Target
type Local struct {
	user       string
	activeuser string
	sudopass   string
}

// New returns a instance of Local
func New(sudopass string) (Local, error) {
	l := Local{
		sudopass: sudopass,
	}
	who := exec.Command("whoami")
	buf, err := who.Output()

	if err != nil {
		return l, errors.Wrap(err, "could not identify user")
	}
	l.user = strings.Trim(string(buf), " \n")
	l.activeuser = l.user
	return l, nil
}

// Close does nothing. It is just there to satisfy the gossh.Target interface.
func (l Local) Close() error {
	return nil
}

// As returns an instance of Local where user is the active user
func (l Local) As(user string) Local {
	l.activeuser = user
	return l
}

// Sudo reports if sudo is required
func (l Local) Sudo() bool {
	return l.user != l.activeuser
}

// Run runs cmd
func (l Local) Run(cmd string, stdin io.Reader) (sh.Response, error) {
	if l.Sudo() {
		return l.runsudo(cmd, stdin)
	}

	return l.run(cmd, stdin)
}

// runsudo runs cmd as activeuser using sudo
func (l Local) runsudo(cmd string, stdin io.Reader) (sh.Response, error) {

	resp := sh.Response{}

	sudo := sh.NewSudo(cmd, l.activeuser, l.sudopass, stdin)
	command := exec.Command("sudo", sudo.Args()...)

	var err error
	command.Stdout = &resp.Stdout
	command.Stderr = sudo
	sudo.StdinPipe, err = command.StdinPipe()
	sudo.Stderr = &resp.Stderr

	err = command.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			resp.ExitStatus = exitError.ExitCode()
		} else {
			resp.ExitStatus = -1
			return resp, errors.Wrapf(err, "could not run command \"%s\"", cmd)
		}
	}

	return resp, nil
}

// runsudo runs cmd as activeuser using sudo
func (l Local) run(cmd string, stdin io.Reader) (sh.Response, error) {

	resp := sh.Response{}

	command := exec.Command("bash", "-c", cmd)

	var err error
	command.Stdout = &resp.Stdout
	command.Stderr = &resp.Stderr
	command.Stdin = stdin

	err = command.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			resp.ExitStatus = exitError.ExitCode()
		} else {
			resp.ExitStatus = -1
			return resp, errors.Wrapf(err, "could not run command \"%s\"", cmd)
		}
	}

	return resp, nil
}

// Mkdir creates the specified directory
// Permission bits are set to 0666 before umask.
func (l Local) Mkdir(path string) error {
	if l.Sudo() {
		cmd := fmt.Sprintf("mkdir %s", path)
		_, err := l.Run(cmd, nil)
		return errors.Wrap(err, "mkdir failed")
	}

	return os.Mkdir(path, 0666)
}

// Create creates the file in path
func (l Local) Create(path string) (io.WriteCloser, error) {
	if l.Sudo() {
		return sudoOpenFile(path, l.activeuser, l.sudopass)
	}

	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}

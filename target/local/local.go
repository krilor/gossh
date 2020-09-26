package local

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/krilor/gossh/target/sh"
	"github.com/krilor/gossh/target/sh/sudo"
	"github.com/pkg/errors"
)

// Package local contains functionality for doing work on localhost.

// Local implements a localhost Target
type Local struct {
	user       string
	activeUser string
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
	l.activeUser = l.user
	return l, nil
}

// Close does nothing. It is just there to satisfy the gossh.Target interface.
func (l Local) Close() error {
	return nil
}

// As returns an instance of Local where user is the active user
func (l Local) As(user string) Local {
	l.activeUser = user
	return l
}

// User returns the connected user
func (l Local) User() string {
	return l.user
}

// ActiveUser returns the currently active user
func (l Local) ActiveUser() string {
	return l.activeUser
}

// sudo reports if sudo is required
func (l Local) sudo() bool {
	return l.user != l.activeUser
}

// String implements fmt.Stringer
func (l Local) String() string {
	return fmt.Sprintf("%s@local", l.activeUser)
}

// Run runs cmd
func (l Local) Run(cmd string, stdin io.Reader) (sh.Result, error) {
	if l.sudo() {
		return l.runsudo(cmd, stdin)
	}

	return l.run(cmd, stdin)
}

// runsudo runs cmd as activeUser using sudo
func (l Local) runsudo(cmd string, stdin io.Reader) (sh.Result, error) {

	resp := sh.Result{}

	sudo := sudo.New(cmd, l.activeUser, l.sudopass, stdin)
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

// runsudo runs cmd as activeUser using sudo
func (l Local) run(cmd string, stdin io.Reader) (sh.Result, error) {

	resp := sh.Result{}

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

// Create creates the file in path
func (l Local) Create(path string) (io.WriteCloser, error) {
	if l.sudo() {
		return sudoOpen(path, l.activeUser, l.sudopass, modeCreate)
	}

	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}

// Open opens a file in path
func (l Local) Open(path string) (io.ReadCloser, error) {
	if l.sudo() {
		return sudoOpen(path, l.activeUser, l.sudopass, modeRead)
	}

	return os.Open(path)
}

// Append opens a file in path for appending
func (l Local) Append(path string) (io.WriteCloser, error) {
	if l.sudo() {
		return sudoOpen(path, l.activeUser, l.sudopass, modeAppend)
	}

	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
}

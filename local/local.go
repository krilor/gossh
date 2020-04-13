package local

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/krilor/gossh"
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

// sudo reports if sudo is required
func (l Local) sudo() bool {
	return l.user != l.activeuser
}

// run cmd
// refs:
// https://stackoverflow.com/a/30329351 - shell
// https://stackoverflow.com/a/24095983 - sudo
// https://stackoverflow.com/a/55055100 - exit status
func (l Local) run(cmd string, stdin string, sudo bool, user string) (gossh.Response, error) {

	r := gossh.Response{}
	var command *exec.Cmd
	if sudo {
		// -k is used to reset previous sudo timestamps
		// -S reads password from stdin
		// -u sets the user
		// -E preserve user environment when running command
		command = exec.Command("sudo", "-kSE", "-u", user, "bash", "-c", cmd)
		command.Stdin = strings.NewReader(l.sudopass + "\n" + stdin + "\n")
	} else {
		command = exec.Command("bash", "-c", cmd)
		command.Stdin = strings.NewReader(stdin + "\n")
	}
	o := bytes.Buffer{}
	e := bytes.Buffer{}

	command.Stdout = &o
	command.Stderr = &e

	err := command.Run()

	r.Stdout = o.String()
	r.Stderr = e.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			r.ExitStatus = exitError.ExitCode()
		} else {
			r.ExitStatus = -1
			return r, errors.Wrapf(err, "could not run command \"%s\"", cmd)
		}
	}

	return r, nil
}

// Mkdir creates the specified directory
// Permission bits are set to 0666 before umask.
func (l Local) Mkdir(path string) error {
	if l.sudo() {
		cmd := fmt.Sprintf("mkdir %s", path)
		_, err := l.run(cmd, "", true, l.activeuser)
		return errors.Wrap(err, "mkdir failed")
	}

	return os.Mkdir(path, 0666)
}

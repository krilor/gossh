package docker

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/lithammer/shortuuid"
	"github.com/pkg/errors"
)

// Package docker provides functionality to create throwaway docker containers in test

// findPort finds and returns a free port on the host machine
// inspired by https://github.com/Konstantin8105/FreePort/blob/master/freeport.go
func findPort() (port int, err error) {
	ln, err := net.Listen("tcp", "[::]:0")

	if err != nil {
		return 0, err
	}
	defer ln.Close()

	return ln.Addr().(*net.TCPAddr).Port, nil
}

// Container represents a docker container
type Container struct {
	name   string
	port   int
	killed bool
}

// Kill issues 'docker kill' on the container
func (c *Container) Kill() error {

	cmd := exec.Command("docker", "kill", c.name)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "could not stop throwaway container")
	}

	return nil
}

// Exec issues 'docker exec' to the container.
// Returns stdout, stderr and exitcode if found.
// Leading and trailing spaces and newlines are trimmed from stdout and stderr.
func (c *Container) Exec(cmd string) (string, string, int, error) {

	command := exec.Command("docker", "exec", c.name, "bash", "-c", cmd)

	o := bytes.Buffer{}
	e := bytes.Buffer{}
	var s int = 0

	command.Stdout = &o
	command.Stderr = &e

	err := command.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			s = exitError.ExitCode()
			err = nil
		} else {
			s = -1
		}
	}

	return strings.Trim(o.String(), " \n"), strings.Trim(e.String(), " \n"), s, err
}

// Port gets the port that SSH is listening on
func (c *Container) Port() int {
	return c.port
}

// New creates runs a docker container with ssh enabled
// Returns generated ID and port
func New() (*Container, error) {

	c := Container{}

	cmd := exec.Command("docker", "build", "-t", "gossh_throwaway_ubuntu", "-")
	cmd.Stdin = strings.NewReader(string(dockerFileUbuntu))
	b, err := cmd.CombinedOutput()

	if err != nil {
		return &c, errors.Wrapf(err, "could not create throwaway container: %s", string(b))
	}

	c.name = shortuuid.New()
	c.port, err = findPort()
	if err != nil {
		return &c, errors.Wrap(err, "could not get free port")
	}

	cmd = exec.Command("docker", "run", "-d", "--rm", "--name", c.name, "-p", fmt.Sprintf("%d:22", c.port), "gossh_throwaway_ubuntu")
	b, err = cmd.CombinedOutput()

	if err != nil {
		return &c, errors.Wrapf(err, "could not run throwaway container %s on port %d: %s", c.name, c.port, string(b))
	}

	return &c, nil
}

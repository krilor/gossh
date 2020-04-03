package docker

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/lithammer/shortuuid"
	"github.com/pkg/errors"
)

// Package docker provides functionality to create throwaway

//
// inspired by https://github.com/Konstantin8105/FreePort/blob/master/freeport.go
func findPort() (port int, err error) {
	ln, err := net.Listen("tcp", "[::]:0")

	if err != nil {
		return 0, err
	}
	defer ln.Close()

	return ln.Addr().(*net.TCPAddr).Port, nil
}

// New creates runs a docker container with ssh enabled
// Returns generated ID and port
func New() (string, int, error) {

	cmd := exec.Command("docker", "build", "-t", "gossh_throwaway_ubuntu", "-")
	cmd.Stdin = strings.NewReader(string(dockerFileUbuntu))
	b, err := cmd.CombinedOutput()

	if err != nil {
		return "", 0, errors.Wrapf(err, "could not create throwaway container: %s", string(b))
	}

	id := shortuuid.New()
	port, err := findPort()
	if err != nil {
		return "", 0, errors.Wrap(err, "could not get free port")
	}

	cmd = exec.Command("docker", "run", "-d", "--rm", "--name", id, "-p", fmt.Sprintf("%d:22", port), "gossh_throwaway_ubuntu")
	b, err = cmd.CombinedOutput()

	if err != nil {
		return "", 0, errors.Wrapf(err, "could not run throwaway container %s on port %d: %s", id, port, string(b))
	}

	return id, port, nil
}

// Stop kills and removes the docker container with id ID
func Stop(ID string) error {

	cmd := exec.Command("docker", "stop", ID)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "could not stop throwaway container")
	}

	return nil
}

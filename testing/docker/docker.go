package docker

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Package docker provides functionality to create throwaway docker containers in test
//
// The package goal is to provide ssh-enabled (password) containers with the following user setup
//
//     | User   | Password  | Sudo rights  | Sudo lecture | SSH keys |
//     |--------|-----------|------------- |--------------|----------|
//     | root   | rootpwd   | ALL          | N/A          | N/A      |
//     | gossh  | gosshpwd  | ALL          | never        | TODO     |
//     | hobgob | hobgobpwd | NOPASSWD:ALL | never        | TODO     |
//     | joxter | joxterpwd | ALL          | allways      | TODO     |
//     | groke  | grokepwd  | NOPASSWD:ALL | allways      | TODO     |
//     | stinky | stinkypwd | NO           | N/A          | TODO     |
//

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

// NewSSHClient returns a new ssh client for the container
// user must be eiter root, gossh or hobgob
func (c *Container) NewSSHClient(user string) (*ssh.Client, error) {

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(user + "pwd"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", fmt.Sprintf("localhost:%d", c.port), config)
}

// New creates runs a docker container with ssh enabled.
func New(image Image) (*Container, error) {

	c := Container{}

	cmd := exec.Command("docker", "build", "-t", image.Name(), "-")
	cmd.Stdin = strings.NewReader(string(image.Dockerfile()))
	b, err := cmd.CombinedOutput()

	if err != nil {
		return &c, errors.Wrapf(err, "could not create throwaway container: %s", string(b))
	}

	c.name = shortuuid.New()
	c.port, err = findPort()
	if err != nil {
		return &c, errors.Wrap(err, "could not get free port")
	}

	cmd = exec.Command("docker", "run", "-d", "--rm", "--name", c.name, "-p", fmt.Sprintf("%d:22", c.port), image.Name())
	b, err = cmd.CombinedOutput()

	if err != nil {
		return &c, errors.Wrapf(err, "could not run throwaway container %s on port %d: %s", c.name, c.port, string(b))
	}

	cc := ssh.ClientConfig{
		User: "gossh",
		Auth: []ssh.AuthMethod{
			ssh.Password("gosshpwd"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// This is to wait for sshd to be ready to accept connections
	for i := 0; i < 6; i++ {
		time.Sleep(time.Duration(int(time.Millisecond) * 150 * i))

		client, err := ssh.Dial("tcp", fmt.Sprintf(":%d", c.port), &cc)

		if err == nil {
			client.Close()
			break
		}
	}

	return &c, err
}

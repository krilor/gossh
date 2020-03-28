package machine

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Machine is a remote machine one is trying to connect
type Machine struct {
	client   *ssh.Client
	addr     string
	port     int
	user     string
	sudopass string
}

// New returns a Machine based on address, port and user
// It will connect to the SSH agent to get any ssh keys
func New(addr string, port int, user string, sudopass string) (*Machine, error) {
	m := Machine{
		addr:     addr,
		port:     port,
		user:     user,
		sudopass: sudopass,
	}
	var err error

	a, err := getAgentAuths()
	auths := []ssh.AuthMethod{
		ssh.Password(sudopass),
	}
	if err == nil {
		auths = append(auths, a)
	}

	cc := ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO
	}

	m.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", addr, port), &cc)
	if err != nil {
		return &m, errors.Wrapf(err, "unable to establish connection to %s:%d", addr, port)
	}

	return &m, nil
}

// String implements io.Stringer for a Machine
func (m *Machine) String() string {
	return fmt.Sprintf("%s@%s:%d", m.user, m.addr, m.port)
}

// isReady reports if the machine is ready, i.e. if it has been initialized using New()
func (m Machine) isReady() bool {
	return m.client != nil
}

// Run runs cmd on machine, as sudo or not, and returns the response
func (m Machine) Run(cmd string, sudo bool) (Response, error) {
	if !m.isReady() {
		return Response{}, errors.New("machines must be initialized using machine.New()")
	}
	session, err := m.client.NewSession()
	r := Response{}

	if err != nil {
		return r, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	session.Stdout = &r.Stdout
	session.Stderr = &r.Stderr

	if sudo {
		session.Stdin = strings.NewReader(m.sudopass + "\n")
		err = session.Run("sudo -S " + cmd)
	} else {
		err = session.Run(cmd)
	}

	if err != nil {

		switch t := err.(type) {
		case *ssh.ExitError:
			r.ExitStatus = t.Waitmsg.ExitStatus()
		case *ssh.ExitMissingError:
			r.ExitStatus = -1
		default:
			return r, errors.Wrap(err, "run of command failed")
		}

	} else {
		r.ExitStatus = 0
	}

	return r, nil
}

// Response contains the response from a remotely run cmd
type Response struct {
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	ExitStatus int
}

func getAgentAuths() (ssh.AuthMethod, error) {

	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open SSH_AUTH_SOCK")
	}

	agentClient := agent.NewClient(conn)

	return ssh.PublicKeysCallback(agentClient.Signers), nil
}

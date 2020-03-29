package gossh

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Host is a remote host one is trying to connect
type Host struct {
	client   *ssh.Client
	addr     string
	port     int
	user     string
	sudopass string
}

// NewHost returns a Host based on address, port and user
// It will connect to the SSH agent to get any ssh keys
func NewHost(addr string, port int, user string, sudopass string) (*Host, error) {
	m := Host{
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

// String implements io.Stringer for a Host
func (m *Host) String() string {
	return fmt.Sprintf("%s@%s:%d", m.user, m.addr, m.port)
}

// isReady reports if h is ready, i.e. if it has been initialized using New()
func (m Host) isReady() bool {
	return m.client != nil
}

// Log is logging on the host
// TODO this should probably use logr with keyValues... or some kind of JSON logging
func (m *Host) Log(trace Trace, key string, value string) {
	log.Println(m.String(), trace.ID(), trace.Prev(), key, value)
}

// Run runs cmd on host, as sudo or not, and returns the response
func (m Host) Run(trace Trace, cmd string, sudo bool) (Response, error) {

	trace = trace.Span()
	m.Log(trace, "run", "start")
	defer m.Log(trace, "run", "end")

	m.Log(trace, "cmd", cmd)
	m.Log(trace, "sudo", fmt.Sprintf("%v", sudo))

	if !m.isReady() {
		return Response{}, errors.New("hosts must be initialized using gossh.NewHost()")
	}

	session, err := m.client.NewSession()
	r := Response{}

	if err != nil {
		return r, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	m.Log(trace, "session", "ready")

	session.Stdout = &r.Stdout
	session.Stderr = &r.Stderr

	if sudo {

		session.Stdin = strings.NewReader(m.sudopass + "\n")
		sudocmd := "sudo -S " + cmd
		m.Log(trace, "run", sudocmd)
		err = session.Run(sudocmd)

	} else {

		m.Log(trace, "run", cmd)
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

	m.Log(trace, "stdout", r.Stdout.String())
	m.Log(trace, "stderr", r.Stderr.String())
	m.Log(trace, "exitstatus", fmt.Sprintf("%d", r.ExitStatus))

	return r, nil
}

// Apply applies Rule r on m, i.e. runs Check and conditionally runs Ensure
// Id must be unique string. //TODO - how to explain this
// TODO - maybe use ... on r to allow specification of multiple rules at once
func (m *Host) Apply(id string, trace Trace, r Rule) error {

	trace = trace.Span()
	m.Log(trace, "apply", "start")
	defer m.Log(trace, "apply", "end")

	span := trace.Span()
	m.Log(span, "check", "start")
	ok, err := r.Check(span, m)
	m.Log(span, "check", "end")

	if err != nil {
		return errors.Wrapf(err, "could not check rule %v on machinve %v", r, m)
	}

	if ok {
		return nil
	}

	m.Log(trace, "check", fmt.Sprintf("%v", ok))

	span = trace.Span()
	m.Log(span, "ensure", "start")
	err = r.Ensure(span, m)
	m.Log(span, "ensure", "end")

	if err != nil {
		return errors.Wrapf(err, "could not ensure rule %v on host %v", r, m)
	}

	return nil
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

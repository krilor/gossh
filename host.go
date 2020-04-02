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

	// Validate is used to indicate if the host only allows RunQuery, that does not alter the state of the system
	Validate bool
}

// NewHost returns a Host based on address, port and user
// It will connect to the SSH agent to get any ssh keys
func NewHost(addr string, port int, user string, sudopass string) (*Host, error) {
	h := Host{
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

	h.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", addr, port), &cc)
	if err != nil {
		return &h, errors.Wrapf(err, "unable to establish connection to %s:%d", addr, port)
	}

	return &h, nil
}

// String implements io.Stringer for a Host
func (h *Host) String() string {
	return fmt.Sprintf("%s@%s:%d", h.user, h.addr, h.port)
}

// isReady reports if h is ready, i.e. if it has been initialized using New()
func (h *Host) isReady() bool {
	return h.client != nil
}

// Log is logging
func (h *Host) Log(trace Trace, msg string, keyAndValues ...string) {
	log.Println(h.String(), trace.id, trace.prev, msg, keyAndValues)
}

// RunChange are used to run cmd's that RunChanges the state on m
func (h *Host) RunChange(trace Trace, cmd string, user string) (Response, error) {
	h.Log(trace, "runchange", "invoked")
	if h.Validate {
		h.Log(trace, "runchange", "blocked by validate")
		return Response{
			Stdout:     *bytes.NewBuffer([]byte{}),
			Stderr:     *bytes.NewBuffer([]byte{}),
			ExitStatus: BlockedByValidate,
		}, nil
	}
	return h.run(trace, cmd, user)
}

// RunQuery are used to run cmd's that doet not modify anything on m
func (h *Host) RunQuery(trace Trace, cmd string, user string) (Response, error) {
	h.Log(trace, "runquery", "invoked")
	return h.run(trace, cmd, user)
}

// Run runs cmd on host, as sudo or not, and returns the response
func (h *Host) run(trace Trace, cmd string, user string) (Response, error) {

	trace = trace.Span()
	h.Log(trace, "run", "start")
	defer h.Log(trace, "run", "end")

	h.Log(trace, "cmd", cmd)
	h.Log(trace, "sudo", fmt.Sprintf("%v", user))

	if !h.isReady() {
		return Response{}, errors.New("hosts must be initialized using gossh.NewHost()")
	}

	session, err := h.client.NewSession()
	r := Response{}

	if err != nil {
		return r, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	h.Log(trace, "session", "ready")

	session.Stdout = &r.Stdout
	session.Stderr = &r.Stderr

	if user != "" && user != h.user {

		session.Stdin = strings.NewReader(h.sudopass + "\n")
		sudocmd := fmt.Sprintf("sudo -S -u %s %s", user, cmd)
		h.Log(trace, "run", sudocmd)
		err = session.Run(sudocmd)

	} else {

		h.Log(trace, "run", cmd)
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

	h.Log(trace, "stdout", r.Stdout.String())
	h.Log(trace, "stderr", r.Stderr.String())
	h.Log(trace, "exitstatus", fmt.Sprintf("%d", r.ExitStatus))

	return r, nil
}

// nSpaces is a little utility to get n spaces and lines
func nSpaces(n int) string {
	b := strings.Builder{}
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			b.Write([]byte(" "))
		} else {

			b.Write([]byte("│"))
		}
	}

	return b.String()
}

// Apply applies Rule r on m, i.e. runs Check and conditionally runs Ensure
// Id must be unique string. //TODO - how to explain this
// TODO - maybe use ... on r to allow specification of multiple rules at once
func (h *Host) Apply(trace Trace, name string, r Rule) error {

	h.Log(trace, "apply", "name", name)

	trace = trace.Span()
	fmt.Printf("%s┌ %s %s\n", nSpaces(trace.level), name, "Start")

	h.Log(trace, "apply", "start")
	defer h.Log(trace, "apply", "end")

	span := trace.Span()
	h.Log(span, "check", "start")
	ok, err := r.Check(span, h)
	h.Log(span, "check", "end")

	if err != nil {
		fmt.Printf("%s└ %s %s: %v\n", nSpaces(trace.level), name, "ERROR", err)
		return errors.Wrapf(err, "could not check rule %v on machinve %v", r, h)
	}

	if ok {
		fmt.Printf("%s└ %s %s\n", nSpaces(trace.level), name, "OK")
		return nil
	}

	h.Log(trace, "check", fmt.Sprintf("%v", ok))

	span = trace.Span()
	h.Log(span, "ensure", "start")
	err = r.Ensure(span, h)
	h.Log(span, "ensure", "end")

	if err != nil {
		fmt.Printf("%s└ %s %s\n", nSpaces(trace.level), name, "CHANGED")
		return errors.Wrapf(err, "could not ensure rule %v on host %v", r, h)
	}

	fmt.Printf("%s└ %s %s\n", nSpaces(trace.level), name, "OK")
	return nil
}

// Response contains the response from a remotely run cmd
type Response struct {
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	ExitStatus int
}

// ExitStatusSuccess is a convenience method to check if an exit code is either 0 or BlockedByValidate.
// It is provided as a form of syntactic sugar.
func (r Response) ExitStatusSuccess() bool {
	// TODO add exitStatuses ...int to allow for including more exit statuses as ok.
	return r.ExitStatus == 0 || r.ExitStatus == BlockedByValidate
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

package gossh

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// runner is the interface for running arbritrary shell commands
// used for implementing different kinds of hosts, such as localhost, ssh, docker, remote docker
type runner interface {
	user() string
	run(cmd string, stdin string, sudo bool, user string) (Response, error)
}

// local is the runner for local tasks
type local struct {
	sudopass string
	usr      string
}

func (l *local) String() string {
	return "localhost"
}

func (l *local) user() string {
	if l.usr == "" {
		u, _ := user.Current()
		l.usr = u.Username
	}
	return l.usr
}

// run cmd
// refs:
// https://stackoverflow.com/a/30329351 - shell
// https://stackoverflow.com/a/24095983 - sudo
// https://stackoverflow.com/a/55055100 - exit status
func (l *local) run(cmd string, stdin string, sudo bool, user string) (Response, error) {

	r := Response{}
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

	r.Stdout = scrubStd(o.String())
	r.Stderr = scrubStd(e.String())

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

type remote struct {
	client   *ssh.Client
	addr     string
	port     int
	usr      string
	sudopass string
}

func (r *remote) user() string {
	return r.usr
}

func (r *remote) String() string {
	return fmt.Sprintf("%s@%s:%d", r.usr, r.addr, r.port)
}

func newRemote(addr string, port int, user string, sudopass string) (*remote, error) {
	r := remote{
		addr:     addr,
		port:     port,
		usr:      user,
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

	r.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", addr, port), &cc)
	if err != nil {
		return &r, errors.Wrapf(err, "unable to establish connection to %s:%d", addr, port)
	}

	return &r, nil

}

// run runs cmd on remote
func (r *remote) run(cmd string, stdin string, sudo bool, user string) (Response, error) {

	session, err := r.client.NewSession()
	resp := Response{}

	if err != nil {
		return resp, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	o := bytes.Buffer{}
	e := bytes.Buffer{}

	session.Stdout = &o
	session.Stderr = &e

	// TODO - consider using session.Shell - http://networkbit.ch/golang-ssh-client/#multiple_commands
	if sudo {
		session.Stdin = strings.NewReader(r.sudopass + "\n" + stdin + "\n")
		if user == "" || user == "-" {
			user = "root"
		}
		sudocmd := fmt.Sprintf(`sudo -k -S -u %s bash -c "%s"`, user, cmd)
		err = session.Run(sudocmd)

	} else {
		session.Stdin = strings.NewReader(stdin + "\n")
		err = session.Run(cmd)
	}

	resp.Stdout = scrubStd(o.String())
	resp.Stderr = scrubStd(e.String())

	if err != nil {

		switch t := err.(type) {
		case *ssh.ExitError:
			resp.ExitStatus = t.Waitmsg.ExitStatus()
		case *ssh.ExitMissingError:
			resp.ExitStatus = -1
		default:
			return resp, errors.Wrap(err, "run of command failed")
		}

	} else {
		resp.ExitStatus = 0
	}

	return resp, nil
}

// Host is a remote host one is trying to connect
type Host struct {
	r runner
	// Validate is used to indicate if the host only allows RunQuery, that does not alter the state of the system
	Validate bool
	t        trace
}

// NewRemoteHost returns a Host based on address, port and user
// It will connect to the SSH agent to get any ssh keys
func NewRemoteHost(addr string, port int, user string, sudopass string) (*Host, error) {
	r, err := newRemote(addr, port, user, sudopass)

	if err != nil {
		return &Host{}, err
	}

	return &Host{r, false, newTrace()}, nil
}

// NewLocalHost resturns a host that represents a local host
func NewLocalHost() *Host {
	return &Host{&local{}, false, newTrace()}
}

// String implements io.Stringer for a Host
func (h *Host) String() string {
	return fmt.Sprintf("%v", h.r)
}

// isReady reports if h is ready, i.e. if it has been initialized using New()
func (h *Host) isReady() bool {
	return h.r != nil
}

// AllowChange reports if the host will allow calls to RunChange
func (h *Host) AllowChange() bool {
	return !h.Validate
}

// Log is logging
func (h *Host) Log(msg string, keyAndValues ...string) {
	log.Println(h.String(), h.t.id, h.t.prev, msg, keyAndValues)
}

// RunChange are used to run cmd's that RunChanges the state on m
func (h *Host) RunChange(cmd string, stdin string, user string) (Response, error) {
	h.Log("runchange", "invoked")
	if h.Validate {
		h.Log("runchange", "blocked by validate")
		return Response{ExitStatus: BlockedByValidate}, nil
	}

	sudo := false

	if user == "-" {
		user = "root"
	}

	if user != "" && user != h.r.user() {
		sudo = true
	}

	return h.run(cmd, stdin, sudo, user)
}

// RunQuery are used to run cmd's that doet not modify anything on m
func (h *Host) RunQuery(cmd string, stdin string, user string) (Response, error) {
	h.Log("runquery", "invoked")

	sudo := false

	if user == "-" {
		user = "root"
	}

	if user != "" && user != h.r.user() {
		sudo = true
	}

	return h.run(cmd, stdin, sudo, user)
}

// Run runs cmd on host, as sudo or not, and returns the response
func (h *Host) run(cmd string, stdin string, sudo bool, user string) (Response, error) {

	h = h.fork()
	h.Log("run", "start")
	defer h.Log("run", "end")

	h.Log("cmd", cmd)
	h.Log("sudo", fmt.Sprintf("%v", user))

	if !h.isReady() {
		return Response{}, errors.New("hosts not initialized. use New...Host")
	}

	r, err := h.r.run(cmd, stdin, sudo, user)

	if err != nil {
		return r, err
	}

	h.Log("stdout", r.Stdout)
	h.Log("stderr", r.Stderr)
	h.Log("exitstatus", fmt.Sprintf("%d", r.ExitStatus))

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
func (h *Host) Apply(name string, r Rule) (Status, error) {
	// TODO - maybe use ... on r to allow specification of multiple rules at once

	h.Log("apply", "name", name)

	fmt.Printf("%s┌ %s %s\n", nSpaces(h.t.level), name, "Start")

	h.Log("apply", "start")
	defer h.Log("apply", "end")

	h.Log("ensure", "start")
	status, err := r.Ensure(h)
	h.Log("ensure", "end")

	if err != nil {
		fmt.Printf("%s└ %s %s\n", nSpaces(h.t.level), name, "CHANGED")
		return StatusFailed, errors.Wrapf(err, "could not ensure rule %v on host %v", r, h)
	}

	fmt.Printf("%s└ %s %s\n", nSpaces(h.t.level), name, "OK")
	return status, nil
}

// fork is in tracing to enable forking a host
func (h *Host) fork() *Host {
	new := *h
	new.t = new.t.span()
	return &new
}

// getAgentAuths is a helper function to get SSH keys from an ssh agent
func getAgentAuths() (ssh.AuthMethod, error) {

	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open SSH_AUTH_SOCK")
	}

	agentClient := agent.NewClient(conn)

	return ssh.PublicKeysCallback(agentClient.Signers), nil
}

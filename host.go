package gossh

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"regexp"
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

// Host is a remote host one is trying to connect
type Host struct {
	r runner
	// Validate is used to indicate if the host only allows RunCheck, that does not alter the state of the system
	Validate bool
	t        trace
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

// RunCheck are used to run cmd's that doet not modify anything on m
func (h *Host) RunCheck(cmd string, stdin string, user string) (Response, error) {
	h.Log("runcheck", "invoked")

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
	var err error
	//status, err := r.Ensure(h)
	h.Log("ensure", "end")

	if err != nil {
		fmt.Printf("%s└ %s %s\n", nSpaces(h.t.level), name, "CHANGED")
		return StatusFailed, errors.Wrapf(err, "could not ensure rule %v on host %v", r, h)
	}

	fmt.Printf("%s└ %s %s\n", nSpaces(h.t.level), name, "OK")
	return 1, nil
}

// fork is in tracing to enable forking a host
func (h *Host) fork() *Host {
	new := *h
	new.t = new.t.span()
	return &new
}

// sudopattern matches sudo prompt
var sudopattern *regexp.Regexp = regexp.MustCompile(`\[sudo\] password for [^:]+: `)

// scrubStd cleans an out/err string. Removes trailing newline and sudo prompt.
func scrubStd(in string) string {
	return sudopattern.ReplaceAllString(strings.Trim(in, "\n"), "")
}

// AgentAuths is a helper function to get SSH keys from an ssh agent.
// If any errors occur, an empty PublicKeys ssh.AuthMethod will be returned.
func AgentAuths() ssh.AuthMethod {

	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return ssh.PublicKeys()
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)

	// TODO how do we close these clients?
	return ssh.PublicKeysCallback(agentClient.Signers)
}

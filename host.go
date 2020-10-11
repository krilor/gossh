package gossh

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/krilor/gossh/target"
	"github.com/krilor/gossh/target/local"
	"github.com/krilor/gossh/target/rmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Host is where we do all the things
type Host struct {
	t target.Target
	// AllowChange controls if it is allowed to do any changes on the host
	AllowChange bool
}

// New returns a host based on a target
func New(target target.Target) *Host {
	return &Host{t: target}
}

// NewLocalHost returns a new host that is pointing to localhost
func NewLocalHost(sudopass string) (*Host, error) {

	h := Host{}
	var err error

	h.t, err = local.New(sudopass)

	return &h, err
}

// NewRemoteHost returns a new remote host
func NewRemoteHost(addr string, user string, sudopass string, hostkeycallback ssh.HostKeyCallback, auths ...ssh.AuthMethod) (*Host, error) {

	h := Host{}
	var err error

	h.t, err = rmt.New(addr, user, sudopass, hostkeycallback, auths...)

	return &h, err

}

// String implements io.Stringer for a Host
func (h *Host) String() string {
	return h.t.String()
}

// isReady reports if h is ready, i.e. if it has been initialized using New*()
func (h *Host) isReady() bool {
	return h.t != nil
}

// Log is logging
func (h *Host) Log(msg string, keyAndValues ...string) {
	log.Println(h.String(), msg, keyAndValues)
}

// RunChange are used to run cmd's that RunChanges the state on m
func (h *Host) RunChange(cmd string, stdin string, user string) (Response, error) {
	return h.run(cmd, stdin, user)
}

// RunCheck are used to run cmd's that does not modify anything on m
func (h *Host) RunCheck(cmd string, stdin string, user string) (Response, error) {
	return h.run(cmd, stdin, user)
}

// Run runs cmd on host, as sudo or not, and returns the response
func (h *Host) run(cmd string, stdin string, user string) (Response, error) {
	if user != "" {
		h.t.As(user)
	}

	r, err := h.t.Run(cmd, bytes.NewBufferString(stdin))

	res := Response{
		Stderr:     r.Stderr.String(),
		Stdout:     r.Stdout.String(),
		ExitStatus: r.ExitStatus,
	}
	if err != nil {
		return res, err
	}

	return res, nil
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

	h.Log("apply", "start")
	defer h.Log("apply", "end")

	h.Log("ensure", "start")
	var err error
	_, err = r.Ensure(h)
	h.Log("ensure", "end")

	if err != nil {
		fmt.Printf("%s└ %s\n", name, "CHANGED")
		return StatusFailed, errors.Wrapf(err, "could not ensure rule %v on host %v", r, h)
	}

	fmt.Printf("%s└ %s\n", name, "OK")
	return 1, nil
}

// sudopattern matches sudo prompt
var sudopattern *regexp.Regexp = regexp.MustCompile(`\[sudo\] password for [^:]+: `)

// scrubStd cleans an out/err string. Removes trailing newline and sudo prompt.
func scrubStd(in string) string {
	return sudopattern.ReplaceAllString(strings.Trim(in, "\n"), "")
}

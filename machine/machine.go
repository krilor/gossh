package machine

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/lithammer/shortuuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Inventory is a list of Machines
type Inventory []*Machine

// Add adds m to i
func (i *Inventory) Add(m *Machine) {
	l := append(*i, m)
	*i = l
	return
}

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

	runID := shortuuid.New()

	log.Println(m.String(), runID, "cmd", cmd, "sudo", sudo)

	if !m.isReady() {
		return Response{}, errors.New("machines must be initialized using machine.New()")
	}

	session, err := m.client.NewSession()
	r := Response{}

	if err != nil {
		return r, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	log.Println(m.String(), runID, "session", "ready")

	session.Stdout = &r.Stdout
	session.Stderr = &r.Stderr

	if sudo {
		session.Stdin = strings.NewReader(m.sudopass + "\n")
		sudocmd := "sudo -S " + cmd
		log.Println(m.String(), runID, "run", sudocmd)
		err = session.Run(sudocmd)
	} else {
		log.Println(m.String(), runID, "run", cmd)
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

	log.Println(m.String(), runID, "stdout", r.Stdout.String())
	log.Println(m.String(), runID, "stderr", r.Stderr.String())
	log.Println(m.String(), runID, "exitstatus", r.ExitStatus)

	return r, nil
}

// Apply applies Rule r on m, i.e. runs Check and conditionally runs Ensure
// TODO - maybe use ... on r to allow specification of multiple rules at once
func (m *Machine) Apply(r Rule) error {

	ok, err := r.Check(m, false)
	if err != nil {
		return fmt.Errorf("could not check rule %v on machinve %v", r, m)
	}

	if ok {
		log.Printf("Rule check %v was ok", r)
		return nil
	}

	log.Printf("Rule check %v was NOT ok", r)

	err = r.Ensure(m, false)
	if err != nil {
		return fmt.Errorf("could not ensure rule %v on machine %v", r, m)
	}

	return nil
}

// Check runs Checker c on m
func (m *Machine) Check(c Checker) (bool, error) {
	return c.Check(m, false)
}

// Ensure runs Ensurer e on m
func (m *Machine) Ensure(c Ensurer) error {
	return c.Ensure(m, false)
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

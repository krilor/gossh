package rmt

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/krilor/gossh/target/rmt/suftp"
	"github.com/krilor/gossh/target/sh"
	"github.com/krilor/gossh/target/sh/sudo"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Package rmt contains functionality for *Remote targets
// A Remote target is a target that is connected to using SSH

// Remote represents a Remote target, connected to over SSH
type Remote struct {
	addr string

	// auth
	conn     *ssh.Client
	connuser string
	// sudopass is connusers sudo password
	sudopass string

	// the user currently operating as
	activeUser string

	// sftp holds all sftp connections. key is username. Pointer?
	sftp map[string]*sftp.Client
}

// New returns a new Remote target from connection details
func New(addr string, user string, sudopass string, hostkeycallback ssh.HostKeyCallback, auths ...ssh.AuthMethod) (*Remote, error) {

	r := Remote{
		addr:       addr,
		connuser:   user,
		sudopass:   sudopass,
		activeUser: user,
		sftp:       map[string]*sftp.Client{},
	}

	cc := ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostkeycallback,
	}

	var err error
	r.conn, err = ssh.Dial("tcp", addr, &cc)
	if err != nil {
		return &r, errors.Wrapf(err, "unable to establish ssh connection to %s", addr)
	}

	return &r, nil

}

// Close closes all underlying connections
func (r *Remote) Close() error {
	for _, c := range r.sftp {
		c.Close()
	}

	return r.conn.Close()
}

// sftpClient returns a sftp client for r.activeUser
// if client does not exist, it will be created
func (r *Remote) sftpClient() (*sftp.Client, error) {
	var c *sftp.Client
	var err error
	var ok bool
	c, ok = r.sftp[r.activeUser]
	if ok {
		return c, nil
	}
	// need to create a new connection
	if r.sudo() {
		c, err = suftp.NewSudoClient(r.conn, r.activeUser, r.sudopass)
	} else {
		c, err = sftp.NewClient(r.conn)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "could not start sftp connection for %s", r.activeUser)
	}
	r.sftp[r.activeUser] = c
	return c, nil
}

// sudo reports if operations must be done using sudo, i.e. if active user is not the connected user.
func (r *Remote) sudo() bool {
	return r.activeUser != r.connuser
}

// As returns a new Remote that will use the same underlying connections, but all operations will be done as user.
//
// No tests are done in this method. If user does not exist or does not have sudo rights, that will only be evident when trying to use methods on the returned object.
func (r *Remote) As(user string) {
	r.activeUser = user
	return
}

// User returns the connected user
func (r *Remote) User() string {
	return r.connuser
}

// ActiveUser returns the currently active user
func (r *Remote) ActiveUser() string {
	return r.activeUser
}

// String implements fmt.Stringer
func (r *Remote) String() string {
	return fmt.Sprintf("%s@%s", r.connuser, r.addr)
}

// Run executes cmd on Remote with the currently active user and returns the response.
// Reader stdin is used to add stdin.
func (r *Remote) Run(cmd string, stdin io.Reader) (sh.Result, error) {
	if r.sudo() {
		return r.runsudo(cmd, stdin)
	}
	return r.run(cmd, stdin)
}

// run run cmd on remote
func (r *Remote) run(cmd string, stdin io.Reader) (sh.Result, error) {
	session, err := r.conn.NewSession()
	resp := sh.Result{}

	if err != nil {
		return resp, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	session.Stdout = &resp.Stdout
	session.Stderr = &resp.Stderr
	session.Stdin = stdin

	err = session.Run(cmd)

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

// runsudo runs cmd on Remote as sudo / activeUser
func (r *Remote) runsudo(cmd string, stdin io.Reader) (sh.Result, error) {

	session, err := r.conn.NewSession()
	resp := sh.Result{}

	if err != nil {
		return resp, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	sudo := sudo.New(cmd, r.activeUser, r.sudopass, stdin)

	session.Stdout = &resp.Stdout
	session.Stderr = sudo
	sudo.StdinPipe, err = session.StdinPipe()
	sudo.Stderr = &resp.Stderr

	err = session.Run(sudo.Cmd())

	if err != nil {

		switch t := err.(type) {
		case *ssh.ExitError:
			resp.ExitStatus = t.Waitmsg.ExitStatus()
		case *ssh.ExitMissingError:
			resp.ExitStatus = -1
		default:
			return resp, errors.Wrap(err, "run of command failed")
		}

		if sudo.WrongPwd() {
			return resp, errors.Wrap(err, "wrong sudo password")
		}

	} else {
		resp.ExitStatus = 0
	}

	return resp, nil
}

// Put implements target.Put
func (r *Remote) Put(filename string, data []byte, perm os.FileMode) error {
	sftp, err := r.sftpClient()
	if err != nil {
		return errors.Wrap(err, "could not get sftp client")
	}

	f, err := sftp.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
	}
	defer f.Close()

	n, err := f.Write(data)

	if err != nil {
		return errors.Wrap(err, "write error")
	}

	if n != len(data) {
		return fmt.Errorf("wrote %d of %d bytes to file", n, len(data))
	}

	err = f.Chmod(perm)

	if err != nil {
		return errors.Wrap(err, "chmod error")
	}

	return nil
}

// Get retrieves the contents of the named
func (r *Remote) Get(filename string) ([]byte, error) {
	sftp, err := r.sftpClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not get sftp client")
	}

	b := &bytes.Buffer{}

	f, err := sftp.Open(filename)
	if err != nil {
		return []byte{}, errors.Wrap(err, "unable to open file")
	}
	defer f.Close()

	_, err = f.WriteTo(b)
	if err != nil {
		return b.Bytes(), errors.Wrap(err, "reading failed")
	}

	return b.Bytes(), nil
}

// scput puts the contents of a Reader on a path on the Remote machine, using scp
// TODO - consider (re)moving this
//
// Inspiration:
// https://github.com/laher/scp-go/blob/master/scp/toRemote.go
// https://gist.github.com/jedy/3357393
//
// SCP notes:
// https://web.archive.org/web/20170215184048/https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
// https://en.wikipedia.org/wiki/Secure_copy#cite_note-Pechanec-2
func (r *Remote) scput(content io.Reader, size int64, path string, mode uint32) error {

	session, err := r.conn.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create scp session")
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		// header message has the format C<mode> <size> <filename>
		fmt.Fprintf(w, "C%04o %d %s\n", mode, size, filepath.Base(path))

		io.Copy(w, content)

		// transfer end with \x00
		fmt.Fprint(w, "\x00")
	}()

	if b, err := session.CombinedOutput(fmt.Sprintf("/usr/bin/scp -tr %s", path)); err != nil {
		return errors.Wrapf(err, "unable to copy content: %s", string(b))
	}

	return nil
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

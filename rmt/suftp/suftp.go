package suftp

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Package suftp contains handy utility methods to create non-standard SFTP clients for github.com/pkg/sftp.
//
// The package if mostly useful for sudo-based sftp or when the remote sshd does not have an sftp subsystem configured.

// NewSubsystemClient creates a new SFTP client on conn, using zero or more option
// functions.
//
// Subsystem must be defined in targets sshd_config.
func NewSubsystemClient(conn *ssh.Client, subsystem string, opts ...sftp.ClientOption) (*sftp.Client, error) {

	s, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	if err := s.RequestSubsystem(subsystem); err != nil {
		return nil, err
	}
	pw, err := s.StdinPipe()
	if err != nil {
		return nil, err
	}
	pr, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return sftp.NewClientPipe(pr, pw, opts...)
}

// NewSubsystemCommandClient creates a new SFTP client on conn, using a custom subsystem command and zero or more option
// functions.
//
// Command cmd can be used to specify a custom subsystem command, similar to the -s option for sftp cli.
//
// Specify subsystem as a path when the remote sshd does not have an sftp Subsystem configured.
// Specify subsystem as "sudo -u [user] /path/to/sftp-server" to get SFTP with another user.
//
// Sudo must have NOPASSWD for the sftp-server binary.
func NewSubsystemCommandClient(conn *ssh.Client, cmd string, opts ...sftp.ClientOption) (*sftp.Client, error) {

	s, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	// TODO - use s.Start()
	ok, err := s.SendRequest("exec", true, ssh.Marshal(struct{ Command string }{cmd}))
	if err == nil && !ok {
		err = fmt.Errorf("sftp: command %v failed", cmd)
	}
	if err != nil {
		return nil, err
	}
	pw, err := s.StdinPipe()
	if err != nil {
		return nil, err
	}
	pr, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return sftp.NewClientPipe(pr, pw, opts...)
}

// NewSudoClient creates a new SFTP client on conn, using user and sudopass for the conn user and zero or more option
// functions.
//
// The user is the user to get an sftp client for. Sudopwd is the password for the user on conn.
//
// NOTE - There is a known sudo issue where sudo emits an error that makes this method not work.
// See https://bugzilla.redhat.com/show_bug.cgi?id=1773148
// Error returned will be "wrong sudo password or sudo error"
func NewSudoClient(conn *ssh.Client, user, sudopwd string, opts ...sftp.ClientOption) (*sftp.Client, error) {

	s, err := conn.NewSession()
	if err != nil {
		return nil, err
	}

	promptpwd := "gimmeyourpwdnow"
	promptsuccess := "thesudopwdwasok"
	promptlength := len(promptpwd)

	// serverpaths are the most likely paths to the sftp-server binary, ordered from most likely less likely
	// paths are from https://winscp.net/eng/docs/faq_su#fn2
	serverpaths := []string{
		"/usr/libexec/openssh/sftp-server",
		"/usr/lib/openssh/sftp-server",
		"/usr/lib/sftp-server",
		"/usr/bin/sftp-server",
		"/bin/sftp-server",
		"sftp-server",
	}

	if user == "" || user == "-" {
		user = "root"
	}

	// this command is where most of the magic happens
	//
	// sudo -p specifies the password prompt from sudo. The prompt is sent to stderr
	// sudo -S dictates that password should be read from stdin
	// sudo -u specifies the user
	// sh -c 'cmd' passes cmd to sh
	// >&2 echo "%s" echos %s to stderr ( see https://stackoverflow.com/a/23550347 )
	// & %s is the final command to be executed, which is the path to the sftp-server binary. It is a list of possible binary paths and it will pick the first one that exists. At the end, it will look in $PATH.
	// In total this means that we can read from stderr and write to stdin to
	// TODO - this should probably be a separate, tested function
	cmd := fmt.Sprintf(`sudo -u '%s' -p '%s' -S sh -c '>&2 echo "%s" & %s'`, user, promptpwd+"\n", promptsuccess, strings.Join(serverpaths, " 2> /dev/null || "))

	stdin, err := s.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := s.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = s.Start(cmd)

	if err != nil {
		return nil, err
	}

	// Sudo might output a lecture, so looping for either password or success prompt
	var prompt string
	for true {
		prompt, err = getPrompt(stderr, promptlength)
		if err != nil {
			return nil, errors.Wrap(err, "first stderr read failed")
		}
		if prompt == promptpwd || prompt == promptsuccess {
			break
		}
	}

	switch prompt {
	case promptpwd:
		stdin.Write([]byte(sudopwd + "\n"))
	case promptsuccess:
		return sftp.NewClientPipe(stdout, stdin, opts...)
	}

	// Second read should be either be success or something else. If it is not success, wrong sudo pasword is the most likely scenario.
	prompt, err = getPrompt(stderr, promptlength)
	if err != nil {
		return nil, errors.Wrap(err, "second stderr read failed")
	}

	if prompt != promptsuccess {
		return nil, fmt.Errorf("wrong sudo password or sudo error")
	}

	return sftp.NewClientPipe(stdout, stdin, opts...)
}

// getPrompt is a handy method for reading a prompt from stderr
// prompt should be at the end of the read, minus a newline
// TODO - test?
func getPrompt(rd io.Reader, length int) (string, error) {
	buf := make([]byte, 2048)

	n, err := rd.Read(buf)
	if err != nil {
		return "", err
	}

	var prompt string
	if n <= length+1 {
		prompt = string(buf[:n-1])
	} else {
		prompt = string(buf[n-length-1 : n-1])
	}

	return prompt, nil
}

package sudo

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	sudoPwdPrompt string = "SHOWMETHEMONEY"
	sudoSuccess   string = "ITISALLGOODNOW"
	sudoFailed    string = "OHMYTHISISBAAD"
)

// New returns a new Sudo
func New(cmd, user, pwd string, stdin io.Reader) *Sudo {
	return &Sudo{
		cmd:    cmd,
		user:   user,
		pwd:    pwd,
		stdin:  stdin,
		Stderr: &bytes.Buffer{},
	}
}

// Sudo can be used to connect to stderr and stdin to detect and respond to sudo password prompts.
type Sudo struct {
	cmd   string
	user  string
	pwd   string
	stdin io.Reader // stdin to cmd

	Stderr io.Writer // where to pass stderr once sudo is done

	StdinPipe io.WriteCloser // pipe to the cmds stdin

	pwdprompts int
	success    bool
}

// Cmd returns a command that contains the command
// Useful for ssh.Session.Run
func (s *Sudo) Cmd() string {
	return fmt.Sprintf(`sudo -p %s -S -u %s bash -c '>&2 printf %s & %s'`,
		sudoPwdPrompt,
		s.user,
		sudoSuccess,
		Escape(s.cmd),
	)
}

// Args returns all args to the sudo command
// Useful for os.Exec
func (s *Sudo) Args() []string {
	return []string{
		`-p`,
		sudoPwdPrompt,
		`-S`,
		`-u`,
		s.user,
		`bash`,
		`-c`,
		fmt.Sprintf(`>&2 printf %s & %s`, sudoSuccess, s.cmd),
	}
}

// WrongPwd reports if password prompt was recieved more than once
func (s *Sudo) WrongPwd() bool {
	return s.pwdprompts > 1
}

// Write implements writer
func (s *Sudo) Write(p []byte) (int, error) {
	if s.success {
		return s.Stderr.Write(p)
	}

	switch string(p) {
	case sudoPwdPrompt:
		s.pwdprompts += s.pwdprompts
		s.StdinPipe.Write([]byte(s.pwd + "\n"))
	case sudoSuccess:
		s.success = true
		if s.stdin != nil {
			io.Copy(s.StdinPipe, s.stdin)
		}
		s.StdinPipe.Close()
	}

	return len(p), nil
}

// Escape escapes a cmd so that it can be used inside a single-quoted argument.
// The intended purpose e.g. when strings are used as input to sh -c '%s'
// The method assumes that the outer, surrounding quote is a singlequote.
// The surrounding quote must not be part of cmd.
func Escape(cmd string) string {
	// nice ref on stack: https://stackoverflow.com/questions/1250079/how-to-escape-single-quotes-within-single-quoted-strings
	return strings.ReplaceAll(cmd, `'`, `'\''`)
}

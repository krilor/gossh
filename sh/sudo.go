package sh

import (
	"bytes"
	"fmt"
	"io"
)

const (
	sudoPwdPrompt string = "SHOWMETHEMONEY"
	sudoSuccess   string = "ITSALLGOOD"
)

// NewSudo returns a new Sudo
func NewSudo(cmd, user, pwd string, stdin io.Reader) *Sudo {
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
func (s *Sudo) Cmd() string {
	return fmt.Sprintf(`sudo -p "%s" -S -u %s bash -c '%s & %s'`,
		sudoPwdPrompt,
		s.user,
		Escape(`>&2 printf "%s" "`+sudoSuccess+`"`),
		Escape(s.cmd),
	)
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
		io.Copy(s.StdinPipe, s.stdin)
		s.StdinPipe.Close()
	}

	return len(p), nil
}

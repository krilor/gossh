package gossh

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"testing"
)

var sudopass string
var currentuser string

func init() {
	sudopass = os.Getenv("SUDOPASS")
	if sudopass == "" {
		fmt.Println("###### Remember to set the env var SUDOPASS using \" export SUDOPASS=pwd\"")
	}
	u, _ := user.Current()
	currentuser = u.Username
}

func TestNSpaces(t *testing.T) {

	var tests = []struct {
		in     int
		expect string
	}{
		{-9, ""},
		{-2, ""},
		{-1, ""},
		{0, ""},
		{1, " "},
		{2, " │"},
		{9, " │ │ │ │ "},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%d", test.in), func(t *testing.T) {

			got := nSpaces(test.in)

			if got != test.expect {
				t.Errorf("value: got \"%s\" - expect \"%s\"", got, test.expect)
			}

		})
	}
}

func TestLocal(t *testing.T) {

	l := local{}

	var tests = []struct {
		cmd    string
		sudo   bool
		user   string
		stdin  string
		expect Response
	}{
		{
			cmd: `echo "hello"`,
			expect: Response{
				Stdout:     *bytes.NewBufferString("hello\n"),
				Stderr:     *bytes.NewBufferString(""),
				ExitStatus: 0,
			},
		},
		{
			cmd: `echo -n "hello"`,
			expect: Response{
				Stdout:     *bytes.NewBufferString("hello"),
				Stderr:     *bytes.NewBufferString(""),
				ExitStatus: 0,
			},
		},
		{
			cmd: `somecommandthatdoesnotexist`,
			expect: Response{
				Stdout:     *bytes.NewBufferString(""),
				Stderr:     *bytes.NewBufferString("bash: somecommandthatdoesnotexist: command not found\n"),
				ExitStatus: 127,
			},
		},
		{
			cmd: `cat filethatdoesntexist`,
			expect: Response{
				Stdout:     *bytes.NewBufferString(""),
				Stderr:     *bytes.NewBufferString("cat: filethatdoesntexist: No such file or directory\n"),
				ExitStatus: 1,
			},
		},
		{
			cmd:   `sed s/a/X/ | sed s/c/Z/`,
			stdin: "abc",
			expect: Response{
				Stdout:     *bytes.NewBufferString("XbZ\n"),
				Stderr:     *bytes.NewBufferString(""),
				ExitStatus: 0,
			},
		},
		{
			cmd:   `sed s/a/X/ | sed s/c/Z/`,
			sudo:  true,
			user:  "root",
			stdin: "abc",
			expect: Response{
				Stdout:     *bytes.NewBufferString("XbZ\n"),
				Stderr:     *bytes.NewBufferString("[sudo] password for " + currentuser + ": "),
				ExitStatus: 0,
			},
		},
		{
			cmd:  `ls /root`,
			sudo: true,
			expect: Response{
				Stdout:     *bytes.NewBufferString(""),
				Stderr:     *bytes.NewBufferString("[sudo] password for " + currentuser + ": "),
				ExitStatus: 0,
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s %v %s", test.cmd, test.stdin, test.sudo, test.user), func(t *testing.T) {

			got, err := l.run(test.cmd, test.stdin, test.sudo, test.user, sudopass)
			if err != nil {
				t.Errorf("errored: %v", err)
			}

			if got.Stdout.String() != test.expect.Stdout.String() {
				t.Errorf("stdout: got \"%s\" - expect \"%s\"", got.Stdout.String(), test.expect.Stdout.String())
			}
			if got.Stderr.String() != test.expect.Stderr.String() {
				t.Errorf("stderr: got \"%s\" - expect \"%s\"", got.Stderr.String(), test.expect.Stderr.String())
			}
			if got.ExitStatus != test.expect.ExitStatus {
				t.Errorf("exitstatus: got \"%d\" - expect \"%d\"", got.ExitStatus, test.expect.ExitStatus)
			}

		})
	}
}

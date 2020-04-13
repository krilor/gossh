package rmt

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/krilor/gossh/testing/docker"
)

func TestRemote(t *testing.T) {

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
				Stdout:     "hello",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd: `echo -n "hello"`,
			expect: Response{
				Stdout:     "hello",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd: `somecommandthatdoesnotexist`,
			expect: Response{
				Stdout:     "",
				Stderr:     "bash: somecommandthatdoesnotexist: command not found",
				ExitStatus: 127,
			},
		},
		{
			cmd: `cat filethatdoesntexist`,
			expect: Response{
				Stdout:     "",
				Stderr:     "cat: filethatdoesntexist: No such file or directory",
				ExitStatus: 1,
			},
		},
		{
			cmd:   `sed s/a/X/ | sed s/c/Z/`,
			stdin: "abc",
			expect: Response{
				Stdout:     "XbZ",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd:   `sed s/a/X/ | sed s/c/Z/`,
			sudo:  true,
			user:  "root",
			stdin: "abc",
			expect: Response{
				Stdout:     "XbZ",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd:  `ls -l /root | grep total | awk '{print \$1}'`, // BUG TOOD - there is a bug here. Why do we have to escape the dollar sign?
			sudo: true,
			expect: Response{
				Stdout:     "total",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd:  `echo "test" | sed s/t/b/`,
			sudo: true,
			expect: Response{
				Stdout:     "best",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
	}

	for _, img := range []docker.Image{
		docker.Ubuntu("bionic"),
		docker.CentOS(7),
	} {

		c, err := docker.New(img)
		if err != nil {
			log.Fatalf("could not get throwaway container: %v", err)
		}
		defer c.Kill()

		r, err := newRemote("localhost", c.Port(), "gossh", "gosshpwd")

		if err != nil {
			log.Fatalf("could not connect to throwaway container %v", err)
		}

		for _, test := range tests {
			t.Run(fmt.Sprintf("%s %s %s %v %s", img.Name(), test.cmd, test.stdin, test.sudo, test.user), func(t *testing.T) {

				got, err := r.run(test.cmd, test.stdin, test.sudo, test.user)
				if err != nil {
					t.Errorf("errored: %v", err)
				}

				if got.Stdout != test.expect.Stdout {
					t.Errorf("stdout: got \"%s\" - expect \"%s\"", got.Stdout, test.expect.Stdout)
				}
				if got.Stderr != test.expect.Stderr {
					t.Errorf("stderr: got \"%s\" - expect \"%s\"", got.Stderr, test.expect.Stderr)
				}
				if got.ExitStatus != test.expect.ExitStatus {
					t.Errorf("exitstatus: got \"%d\" - expect \"%d\"", got.ExitStatus, test.expect.ExitStatus)
				}

			})
		}
	}
}

func TestRemotePut(t *testing.T) {
	var tests = []string{"lionking"}

	for _, img := range []docker.Image{
		docker.Ubuntu("bionic"),
		docker.CentOS(7),
	} {

		c, err := docker.New(img)
		if err != nil {
			log.Fatalf("could not get throwaway container: %v", err)
		}
		defer c.Kill()

		r, err := newRemote("localhost", c.Port(), "gossh", "gosshpwd")

		if err != nil {
			log.Fatalf("could not connect to throwaway container %v", err)
		}

		for _, test := range tests {
			t.Run(test+img.Name(), func(t *testing.T) {
				err := r.put(strings.NewReader(test), int64(len(test)), "/tmp/"+test, 644)

				if err != nil {
					t.Error("scp put failed", err)
				}

				o, e, s, err := c.Exec("cat /tmp/" + test)

				if err != nil {
					t.Error("put failed", o, e, s, err)
				}

				if o != test {
					t.Errorf("file not equal: got: %s, expected: %s", o, test)
				}
			})
		}
	}
}

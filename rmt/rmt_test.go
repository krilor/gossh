package rmt

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/krilor/gossh/testing/docker"
	"golang.org/x/crypto/ssh"
)

var containers []*docker.Container

func TestMain(m *testing.M) {

	// need to parse flags for testing.Short()
	flag.Parse()

	// setup
	var imgs []docker.Image
	if testing.Short() {
		imgs = docker.Bench
	} else {
		imgs = docker.FullBench
	}

	for _, img := range imgs {
		log.Println("creating container: ", img.Name())
		c, err := docker.New(img)
		if err != nil {
			log.Fatalf("unable to create container %s", img.Name())
		}
		containers = append(containers, c)
	}

	// run tests
	code := m.Run()

	// teardown
	for _, c := range containers {
		log.Println("killing container:", c.Image())
		c.Kill()
	}

	os.Exit(code)

}

func TestMkdir(t *testing.T) {

	tests := []struct {
		activeuser string
		path       string
		shouldfail bool
	}{
		{"root", "/root/testmkdir1", false},
		{"hobgob", "/root/testmkdir2", true},
		{"gossh", "/home/gossh/testmkdir3", false},
	}

	for _, c := range containers {
		r, err := New(c.Addr(), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))
		if err != nil {
			t.Fatal("could not connect to container:", err)
		}
		defer r.Close()
		for _, test := range tests {
			t.Run(fmt.Sprintf("%s:%s:%s:%v", c.Image(), test.activeuser, test.path, test.shouldfail), func(t *testing.T) {

				r.activeuser = test.activeuser
				err = r.Mkdir(test.path)
				if err != nil && !test.shouldfail {
					t.Error("test file creation errored", err)
				} else if err == nil && test.shouldfail {
					t.Error("test didn't erorr as expected")
				}

				if !test.shouldfail {
					o, _, _, _ := c.Exec("stat --format='%U' " + test.path)
					if o != test.activeuser {
						t.Errorf("wrong ownership. got %s", o)
					}
				}
			})
		}
	}

}

func TestRemote(t *testing.T) {

	/*if testing.Short() {
		//	t.Skip("skipping in short mode")
	}*/

	type resp struct {
		Stdout     string
		Stderr     string
		ExitStatus int
	}

	var tests = []struct {
		cmd    string
		sudo   bool
		user   string
		stdin  string
		expect resp
	}{
		{
			cmd: `printf "hello"`,
			expect: resp{
				Stdout:     "hello",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd: `printf "hello"`,
			expect: resp{
				Stdout:     "hello",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd: `somecommandthatdoesnotexist`,
			expect: resp{
				Stdout:     "",
				Stderr:     "bash: somecommandthatdoesnotexist: command not found",
				ExitStatus: 127,
			},
		},
		{
			cmd: `cat filethatdoesntexist`,
			expect: resp{
				Stdout:     "",
				Stderr:     "cat: filethatdoesntexist: No such file or directory",
				ExitStatus: 1,
			},
		},
		{
			cmd:   `sed s/a/X/ | sed s/c/Z/`,
			stdin: "abc",
			expect: resp{
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
			expect: resp{
				Stdout:     "XbZ",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd:  `ls -l /root | grep total | awk '{print $1}'`,
			sudo: true,
			expect: resp{
				Stdout:     "total",
				Stderr:     "",
				ExitStatus: 0,
			},
		},
		{
			cmd:  `echo "test" | sed s/t/b/`,
			sudo: true,
			expect: resp{
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

		r, err := New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

		if err != nil {
			log.Fatalf("could not connect to throwaway container %v", err)
		}

		for _, test := range tests {
			t.Run(fmt.Sprintf("%s %s %s %v %s", img.Name(), test.cmd, test.stdin, test.sudo, test.user), func(t *testing.T) {

				if test.user != "" {
					r.activeuser = test.user
				}
				got, err := r.Run(test.cmd, strings.NewReader(test.stdin))
				if err != nil {
					t.Errorf("errored: %v", err)
				}

				if got.Out() != test.expect.Stdout {
					t.Errorf("stdout: got \"%s\" - expect \"%s\"", got.Stdout.String(), test.expect.Stdout)
				}
				if got.Err() != test.expect.Stderr {
					t.Errorf("stderr: got \"%s\" - expect \"%s\"", got.Stderr.String(), test.expect.Stderr)
				}
				if got.ExitStatus != test.expect.ExitStatus {
					t.Errorf("exitstatus: got \"%d\" - expect \"%d\"", got.ExitStatus, test.expect.ExitStatus)
				}

			})
		}
	}
}

func TestRemotePut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
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

		r, err := New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

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

func TestAs(t *testing.T) {
	original := Remote{
		connuser:   "jon",
		activeuser: "jon",
	}
	super := original.As("root")

	if super.activeuser != "root" {
		t.Errorf("super.activeuser error: expect 'root', got '%s'", super.activeuser)
	}

	if original.activeuser != "jon" {
		t.Errorf("original.activeuser error: expect 'jon', got '%s'", original.activeuser)
	}
}

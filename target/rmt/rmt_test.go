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
			log.Fatal("unable to create container", img.Name(), err.Error())

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

func TestRun(t *testing.T) {

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

	for _, c := range containers {

		r, err := New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

		if err != nil {
			log.Fatalf("could not connect to throwaway container %v", err)
		}

		for _, test := range tests {
			t.Run(fmt.Sprintf("%s %s %s %v %s", c.Image(), test.cmd, test.stdin, test.sudo, test.user), func(t *testing.T) {

				if test.user != "" {
					r.activeUser = test.user
				}
				got, err := r.Run(test.cmd, strings.NewReader(test.stdin))
				if err != nil {
					t.Errorf("errored: %v", err)
				}

				if got.TrimOut() != test.expect.Stdout {
					t.Errorf("stdout: got \"%s\" - expect \"%s\"", got.Stdout.String(), test.expect.Stdout)
				}
				if got.TrimErr() != test.expect.Stderr {
					t.Errorf("stderr: got \"%s\" - expect \"%s\"", got.Stderr.String(), test.expect.Stderr)
				}
				if got.ExitStatus != test.expect.ExitStatus {
					t.Errorf("exitstatus: got \"%d\" - expect \"%d\"", got.ExitStatus, test.expect.ExitStatus)
				}

			})
		}
	}
}

func TestScput(t *testing.T) {
	var tests = []string{"lionking"}

	for _, c := range containers {

		r, err := New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

		if err != nil {
			log.Fatal("could not connect to throwaway container:", err)
		}

		for _, test := range tests {
			t.Run(test+c.Image(), func(t *testing.T) {
				err := r.scput(strings.NewReader(test), int64(len(test)), "/tmp/"+test, 644)

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
		activeUser: "jon",
	}
	original.As("root")

	if original.activeUser != "root" {
		t.Errorf("super.activeUser error: expect 'root', got '%s'", original.activeUser)
	}

	if original.connuser != "jon" {
		t.Errorf("original.connuser error: expect 'jon', got '%s'", original.activeUser)
	}
}

func TestPut(t *testing.T) {
	var tests = []struct {
		path    string
		user    string
		content string
	}{
		{"/home/gossh/testcreate", "gossh", "filecontent\ntwo lines"},
		{"/home/stinky/testcreate", "stinky", "filecontent\nthree\n lines"},
	}

	for _, c := range containers {
		r, err := New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

		if err != nil {
			log.Fatal("could not connect to throwaway container:", err)
		}
		for _, test := range tests {
			t.Run(test.path+c.Image(), func(t *testing.T) {

				r.activeUser = test.user

				err := r.Put(test.path, []byte(test.content), 0644)
				if err != nil {
					t.Fatal("could not create file", err)
				}

				out, _, _, _ := c.Exec(fmt.Sprintf("cat %s", test.path))
				if out != test.content {
					t.Errorf(`wrong content: expect "%s", got "%s"`, test.content, out)
				}

				out, _, _, _ = c.Exec("stat --format='%U' " + test.path)
				if out != test.user {
					t.Errorf("wrong ownership. got %s", out)
				}

			})
		}
	}
}

func TestGet(t *testing.T) {
	var tests = []struct {
		path    string
		user    string
		content string
	}{
		{"/home/gossh/testopen", "gossh", "filecontent"},
		{"/home/stinky/testopen", "stinky", "filecontent"},
	}

	for _, c := range containers {
		r, err := New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

		if err != nil {
			log.Fatal("could not connect to throwaway container:", err)
		}
		for _, test := range tests {
			t.Run(test.path+c.Image(), func(t *testing.T) {

				r.activeUser = test.user

				c.Exec(fmt.Sprintf("echo -n \"%s\" > %s && chown %s:%s %s %% && chmod 600 %s", test.content, test.path, test.user, test.user, test.path, test.path))

				b, err := r.Get(test.path)
				if err != nil {
					t.Fatal("could not open file", err)
				}

				out := string(b)
				if out != test.content {
					t.Errorf(`wrong content: expect "%s", got "%s"`, test.content, out)
				}

			})
		}
	}
}

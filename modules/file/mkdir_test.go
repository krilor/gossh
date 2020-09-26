package file

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/lithammer/shortuuid"
	"golang.org/x/crypto/ssh"
)

func TestMkdir(t *testing.T) {

	l, err := New(testsudopass)

	tests := []struct {
		activeUser string
		path       string
	}{
		{l.user, testdir + "/gossh_testmkdir2-" + shortuuid.New()},
		{"root", testdir + "/gossh_testmkdir1-" + shortuuid.New()},
	}

	if err != nil {
		t.Fatal("could not get local:", err)
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {

			l.activeUser = test.activeUser
			err = l.Mkdir(test.path)
			if err != nil {
				t.Error("test file creation errored", err)
			}

			stat := exec.Command("stat", "--format=%U", test.path)
			o, err := stat.Output()
			if err != nil {
				t.Error("stat failed", err)
			}

			owner := strings.Trim(string(o), " \n")
			if owner != test.activeUser {
				t.Errorf("wrong ownership. got '%s'", o)
			}

		})
	}

}

func TestMkdir(t *testing.T) {

	tests := []struct {
		activeUser string
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
			t.Run(fmt.Sprintf("%s:%s:%s:%v", c.Image(), test.activeUser, test.path, test.shouldfail), func(t *testing.T) {

				r.activeUser = test.activeUser
				err = r.Mkdir(test.path)
				if err != nil && !test.shouldfail {
					t.Error("test file creation errored", err)
				} else if err == nil && test.shouldfail {
					t.Error("test didn't erorr as expected")
				}

				if !test.shouldfail {
					o, _, _, _ := c.Exec("stat --format='%U' " + test.path)
					if o != test.activeUser {
						t.Errorf("wrong ownership. got %s", o)
					}
				}
			})
		}
	}

}

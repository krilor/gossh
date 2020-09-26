package file

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"

	"github.com/krilor/gossh/testing/tutil"
	"golang.org/x/crypto/ssh"
)

func TestChownCmd(t *testing.T) {
	tests := []struct {
		path      string
		username  string
		groupname string
		expect    string
		err       bool
	}{
		{"", "root", "root", "", true},
		{"/tmp/path", "", "", "", true},
		{"/tmp/path", "uname", "gname", "chown uname:gname /tmp/path", false},
		{"/tmp/path", "", "gname ", "chgrp gname /tmp/path", false},
		{"/tmp/path", "uname ", "", "chown uname /tmp/path", false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s:%s %s", test.username, test.groupname, test.path), func(t *testing.T) {
			out, err := ChownCmd(test.path, test.username, test.groupname)
			if test.err && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !test.err && err != nil {
				t.Errorf("did not expect error, but got one")
			}

			if test.expect != out {
				t.Errorf("expect: %s, got %s", test.expect, out)
			}

		})
	}
}

func TestChown(t *testing.T) {

	l, err := New(testsudopass)
	if err != nil {
		t.Fatal("could not get local:", err)
	}

	var tests = []struct {
		path      string
		user      string // the active user
		username  string
		groupname string
	}{
		{"/tmp/test1", "root", l.user, l.user},
		{"/tmp/test1", "root", l.user, "root"},
		{"/tmp/test1", "root", "root", l.user},
		{"/tmp/test3", "root", l.user, ""},
		{"/tmp/test4", "root", "", l.user},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {

			l.activeUser = test.user

			exec.Command("touch", test.path).Run()

			err := l.Chown(test.path, test.username, test.groupname)
			if err != nil {
				t.Fatal("could not chown", err)
			}

			out, err := exec.Command("stat", "--format", "%U:%G", test.path).Output()
			if err != nil {
				t.Fatal("could not stat", err)
			}

			expect := fmt.Sprintf("%s:%s", tutil.DefaultString(test.username, l.user), tutil.DefaultString(test.groupname, l.user))

			if strings.TrimSpace(string(out)) != expect {
				t.Errorf("wrong ownership. expect: %s, got: %s", expect, out)
			}

		})
	}

}

func TestChown(t *testing.T) {
	var tests = []struct {
		path      string
		user      string // the active user
		username  string
		groupname string
	}{
		{"/tmp/test1", "root", "gossh", "groke"},
		{"/tmp/test2", "root", "groke", "stinky"},
		{"/tmp/test3", "root", "groke", ""},
		{"/tmp/test4", "root", "", "stinky"},
	}

	for _, c := range containers {
		r, err := New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

		if err != nil {
			log.Fatal("could not connect to throwaway container:", err)
		}
		for _, test := range tests {
			t.Run(test.path+c.Image(), func(t *testing.T) {

				r.activeUser = test.user
				c.Exec("touch " + test.path)

				err := r.Chown(test.path, test.username, test.groupname)
				if err != nil {
					t.Fatal("could not chown", err)
				}

				out, _, _, err := c.Exec("stat --format='%U:%G' " + test.path)
				if err != nil {
					t.Fatal("could not stat", err)
				}

				expect := fmt.Sprintf("%s:%s", tutil.DefaultString(test.username, "root"), tutil.DefaultString(test.groupname, "root"))

				if out != expect {
					t.Errorf("wrong ownership. expect: %s, got: %s", expect, out)
				}

			})
		}
	}
}

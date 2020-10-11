package local

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/lithammer/shortuuid"
)

var testsudopass string
var testdir string

func TestMain(m *testing.M) {
	// set sudopass
	testsudopass = os.Getenv("GOSSH_SUDOPASS")
	if testsudopass == "" {
		log.Fatal("missing GOSSH_SUDOPASS env variable")
	}

	exec.Command("sudo", "-k").Run()

	testdir = fmt.Sprintf("%s/gossh-%s", os.TempDir(), shortuuid.New())
	err := os.Mkdir(testdir, 0777)
	if err != nil {
		log.Fatal("could not create testdir:", testdir, "-", err)
	}

	code := m.Run()

	err = os.RemoveAll(testdir)
	if err != nil {
		log.Fatal("could not clear testdir", testdir, "-", err)
	}

	os.Exit(code)
}

func TestPut(t *testing.T) {

	l, err := New(testsudopass)
	if err != nil {
		t.Fatal("could not get local:", err)
	}

	tests := []struct {
		activeUser string
		path       string
		content    string
	}{
		{l.user, testdir + "/testcreate1", "some\ncontent"},
		{"root", testdir + "/testcreate2", "content"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {

			l.activeUser = test.activeUser

			err := l.Put(test.path, []byte(test.content), 0644)
			if err != nil {
				t.Fatal("put errored", err)
			}

			stat := exec.Command("stat", "--format=%U", test.path)
			o, err := stat.CombinedOutput()
			if err != nil {
				t.Error("stat failed", err, string(o))
			}

			owner := strings.Trim(string(o), " \n")
			if owner != test.activeUser {
				t.Errorf("wrong ownership. got '%s'", o)
			}

			content, err := ioutil.ReadFile(test.path)
			if err != nil {
				t.Error("could not read file", err)
			}

			if string(content) != test.content {
				t.Errorf(`file content wrong - expect: "%s", got : "%s"`, test.content, string(content))
			}

		})
	}

}

func TestGet(t *testing.T) {

	l, err := New(testsudopass)
	if err != nil {
		t.Fatal("could not get local:", err)
	}

	tests := []struct {
		activeUser string
		path       string
		content    string
	}{
		{l.user, testdir + "/testopen1", "some\ncontent"},
		{"root", testdir + "/testopen2", "content"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {

			l.activeUser = test.activeUser

			l.Put(test.path, []byte(test.content), 0644)

			content, err := l.Get(test.path)
			if err != nil {
				t.Fatal("open errored", err)
			}

			if string(content) != test.content {
				t.Errorf(`file content wrong - expect: "%s", got : "%s"`, test.content, string(content))
			}

		})
	}

}

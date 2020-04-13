package local

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/lithammer/shortuuid"
)

var testsudopass string

func TestMain(m *testing.M) {
	// set sudopass
	testsudopass = os.Getenv("GOSSH_SUDOPASS")
	if testsudopass == "" {
		log.Fatal("missing GOSSH_SUDOPASS env variable")
	}

	os.Exit(m.Run())
}

func TestMkdir(t *testing.T) {

	l, err := New(testsudopass)

	tests := []struct {
		activeuser string
		path       string
	}{
		{"root", "/tmp/gossh_testmkdir1-" + shortuuid.New()},
		{l.user, "/tmp/gossh_testmkdir2-" + shortuuid.New()},
	}

	if err != nil {
		t.Fatal("could not get local:", err)
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {

			l.activeuser = test.activeuser
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
			if owner != test.activeuser {
				t.Errorf("wrong ownership. got '%s'", o)
			}

		})
	}

}

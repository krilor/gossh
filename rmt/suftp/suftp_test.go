package suftp

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/krilor/gossh/testing/docker"
)

func TestSudoSftp(t *testing.T) {

	tests := []struct {
		user    string // ssh user
		sudo    string // the user to sudo to
		sudopwd string // users sudopassword
		file    string // file path to create
		errend  string // the end of a error string. empty if no error.
	}{
		{"gossh", "hobgob", "gosshpwd", "/home/hobgob/somefile", ""},
		{"gossh", "hobgob", "incorrectpassword", "/home/hobgob/somefile", "wrong sudo password"},
		{"gossh", "root", "gosshpwd", "/root/somefile", ""},
		{"hobgob", "gossh", "", "/home/gossh/somefile", ""},
		{"hobgob", "gossh", "hobgobpwd", "/home/gossh/somefile2", ""},
		{"hobgob", "root", "", "/root/anotherfile", ""},
		{"hobgob", "", "", "/root/anotherfile2", ""},
		{"joxter", "stinky", "joxterpwd", "/home/stinky/joxterfile", ""},
		{"stinky", "gossh", "stinkypwd", "/home/gossh/stinkyfile", "sudo failed or no sudo rights"}, // stinky does not have sudo rights
	}

	for _, img := range docker.FullBench {

		c, err := docker.New(img)
		if err != nil {
			log.Fatalf("could not get throwaway container: %v", err)
		}
		defer c.Kill()

		for _, test := range tests {
			t.Run(fmt.Sprintf("%s - %v", img.Name(), test), func(t *testing.T) {

				conn, err := c.NewSSHClient(test.user)
				if err != nil {
					t.Fatalf("could not connect to throwaway container %v", err)
				}
				defer conn.Close()

				sftp, err := NewSudoClient(conn, test.sudo, test.sudopwd)
				if err != nil {
					if test.errend != "" {
						// error was expected
						if strings.HasSuffix(err.Error(), test.errend) {
							// correct error - skip the test of the test
							t.SkipNow()
						} else {
							t.Fatalf("wrong error string: expect end as %s, got %s", test.errend, err.Error())
						}
					} else {
						// errors was not expected
						t.Fatal("could not get sudo connection:", err)
					}
				}
				defer sftp.Close()

				err = sftp.Mkdir(test.file)
				if err != nil {
					t.Fatal("could not create dir in hobgob home", err)
				}

				o, _, s, err := c.Exec("stat --format='%U' " + test.file)

				sudo := test.sudo
				if sudo == "" || sudo == "-" {
					sudo = "root"
				}

				if o != sudo || s != 0 {
					t.Errorf("owner: expect %s:%d, got %s:%d", test.sudo, 0, o, s)
				}
			})
		}

	}
}

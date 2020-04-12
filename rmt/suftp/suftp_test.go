package suftp

import (
	"fmt"
	"log"
	"testing"

	"github.com/krilor/gossh/testing/docker"
)

func TestSudoSftp(t *testing.T) {

	tests := []struct {
		user    string // ssh user
		sudo    string // the user to sudo to
		sudopwd string // users sudopassword
		file    string // file path to create
	}{
		{"gossh", "hobgob", "gosshpwd", "/home/hobgob/somefile"},
		{"gossh", "root", "gosshpwd", "/root/somefile"},
		{"hobgob", "gossh", "", "/home/gossh/somefile"},
		{"hobgob", "gossh", "hobgobpwd", "/home/gossh/somefile2"},
		{"hobgob", "root", "", "/root/anotherfile"},
		{"joxter", "stinky", "joxterpwd", "/home/stinky/joxterfile"},
		// TODO negative test cases
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
					t.Fatal("could not get sudo connection:", err)
				}
				defer sftp.Close()

				err = sftp.Mkdir(test.file)
				if err != nil {
					t.Fatal("could not create dir in hobgob home", err)
				}
			})
		}

	}
}

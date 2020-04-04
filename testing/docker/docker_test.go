package docker

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

func TestDockerThrowaway(t *testing.T) {

	imgs := []Image{
		NewDebianImage("ubuntu", "bionic"),
		NewRHELImage("centos", "7"),
	}

	for _, img := range imgs {
		t.Run(img.Name(), func(t *testing.T) {

			c, err := New(img)

			if err != nil {
				t.Error(err)
			}

			o, e, s, err := c.Exec("whoami | sed s/o/u/g") // root -> [replace o with u] -> ruut

			if err != nil {
				t.Error(err)
			} else if o != "ruut" || e != "" || s != 0 {
				t.Errorf("got %s %s %d", o, e, s)
			}

			for _, u := range []string{"gossh", "hobgob"} {
				err = testSSHConnection(c.Port(), u, u+"pwd")
				if err != nil {
					t.Errorf("ssh for %s failed: %v", u, err)
				}
			}

			err = c.Kill()
			if err != nil {
				t.Error("kill failed", err)
			}
		})
	}

}

func testSSHConnection(port int, user string, password string) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%d", port), config)
	if err != nil {
		return errors.Wrap(err, "failed to dial")
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create session")
	}
	defer session.Close()

	b, err := session.CombinedOutput(`echo "` + password + `" | sudo -S -k ls /root`)

	if err != nil {
		return errors.Wrapf(err, "failed to sudo: %s", string(b))
	}

	return nil
}

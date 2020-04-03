package docker

import (
	"testing"
)

func TestDockerThrowaway(t *testing.T) {

	t.Run("docker", func(t *testing.T) {

		c, err := New()

		if err != nil {
			t.Error(err)
		}

		o, e, s, err := c.Exec("whoami | sed s/o/u/g") // root -> [replace o with u] -> ruut

		if err != nil {
			t.Error(err)
		} else if o != "ruut" || e != "" || s != 0 {
			t.Errorf("got %s %s %d", o, e, s)
		}

		err = c.Kill()
		if err != nil {
			t.Error(err)
		}
	})
}

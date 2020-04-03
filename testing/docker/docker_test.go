package docker

import (
	"testing"
)

func TestDockerThrowaway(t *testing.T) {

	t.Run("docker", func(t *testing.T) {

		ID, _, err := New()

		if err != nil {
			t.Error(err)
		}

		err = Stop(ID)
		if err != nil {
			t.Error(err)
		}
	})
}

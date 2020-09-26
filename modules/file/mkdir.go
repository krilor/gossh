package file

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// Mkdir creates the specified directory
// Permission bits are set to 0666 before umask.
func (l Local) Mkdir(path string) error {
	if l.Sudo() {
		cmd := fmt.Sprintf("mkdir %s", path)
		_, err := l.Run(cmd, nil)
		return errors.Wrap(err, "mkdir failed")
	}

	return os.Mkdir(path, 0666)
}

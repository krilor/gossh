package file

import (
	"os"

	"github.com/krilor/gossh"
	"github.com/pkg/errors"
)

// Info hold information about a file
type Info struct {
	Name  string
	Path  string
	User  string
	UID   int
	Group string
	GID   int
	Mode  os.FileMode
}

// Dir is a wrapper for i.Mode.IsDir()
func (i Info) Dir() bool {
	return i.Mode.IsDir()
}

// Stat returns stat info
func Stat(t gossh.Target, abspath string, root bool) (Info, error) {
	user := ""
	if root {
		user = "root"
	}
	_, err := t.RunQuery("stat "+abspath, "", user)

	if err != nil {
		return Info{}, errors.Wrapf(err, "stat %s failed", abspath)
	}

	return Info{}, nil

}

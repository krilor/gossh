package gossh

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Fact is a key
type Fact int

//go:generate stringer -type=Fact

const (
	// OS is the operating system
	OS Fact = iota
	// OSFamily is the family of operating systems
	OSFamily
	// OSVersion is the os version
	OSVersion
)

// Facts are gathered from the machine
type Facts struct {
	kv       map[Fact]string
	gathered bool
}

// Gather gathers facts about the machine
func (f *Facts) Gather(m *Host) error {

	if f.kv == nil {
		f.kv = map[Fact]string{}
	}

	r, err := m.RunCheck(`cat /etc/*release | tr '[:upper:]' '[:lower:]'`, "", "")

	if err != nil {
		return errors.Wrap(err, "getting release info errored")
	}

	if !r.Success() {
		return fmt.Errorf("getting release info failed: %s", r.Stderr)
	}

	kv := parseINI(r.Stdout)

	for _, item := range []struct {
		key  string
		fact Fact
	}{
		{"id", OS},
		{"id_like", OSFamily},
		{"version_id", OSVersion},
	} {
		value, ok := kv[item.key]
		if ok {
			f.kv[item.fact] = value
		}
	}

	f.gathered = true

	return nil

}

// parseINI is used to parse single-level ini-type files
func parseINI(in string) map[string]string {
	kv := map[string]string{}

	lines := strings.Split(in, "\n")

	for _, l := range lines {

		if len(l) < 1 || l[0] == '#' {
			continue
		}

		parts := strings.SplitN(l, "=", 2)

		if len(parts) != 2 {
			continue
		}

		kv[parts[0]] = strings.Trim(parts[1], `"`)
	}

	return kv
}

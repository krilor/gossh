package main

import (
	"fmt"
	"log"

	"github.com/krilor/gossh/apt"
	"github.com/krilor/gossh/base"
	"github.com/krilor/gossh/file"
	"github.com/krilor/gossh/machine"
	"github.com/pkg/errors"
)

func main() {

	inventory := machine.Inventory{}

	// Add a machine to the inventory
	// As of now, it's hardcoded to a docker container on localhost
	m, err := machine.New("localhost", 2222, "gossh", "gosshpwd")
	if err != nil {
		fmt.Printf("could not get new machine %v: %v\n", m, err)
		return
	}

	inventory.Add(m)

	// TODO - add inventory from files, e.g.:
	// machine.NewInventoryFromFile("./inventory.json")

	bootstrap := base.Multi{}

	// file.Exists is not a very helpful rule, it just creates a empty file if it does not exist
	bootstrap.Add(file.Exists("/tmp/hello.nothing2"))

	// apt.Package installs/uninstalls a apt package
	bootstrap.Add(apt.Package{
		Name:   "tree",
		Status: apt.StatusInstalled,
	})

	// This rule does nothing useful, but just shows off the use of a simple cmd based rule
	// This will allways run
	bootstrap.Add(base.Cmd{
		CheckCmd:  "false",
		EnsureCmd: "ls",
	})

	// This rule is a meta-rule used to construct other rules on the fly

	// This is where it starts to get hairy. The Meta rule is used to create a custom rule on the fly.
	// The example is quite simple and not very useful, but shows how to use commands directly on m,
	//  as well as reusing the Ensure command of another Rule
	filename := "somefile.txt"

	bootstrap.Add(base.Meta{
		CheckFunc: func(trace machine.Trace, m *machine.Machine) (bool, error) {
			cmd := fmt.Sprintf("ls -1 /tmp | grep %s", filename)
			r, err := m.Run(trace, cmd, false)
			if err != nil {
				return false, errors.Wrap(err, "could not check for somefile")
			}
			if r.ExitStatus == 0 {
				return true, nil
			}

			return false, nil
		},
		EnsureFunc: func(trace machine.Trace, m *machine.Machine) error {
			return file.Exists("/tmp/"+filename).Ensure(trace, m)
		},
	})

	fmt.Println("len", len(bootstrap))

	// TODO Instead of Apply, one could also do Plan (terraform style)
	for _, m := range inventory {
		log.Println("doing machine", m)
		err = m.Apply("bootstrap", machine.NewTrace(), bootstrap)
		if err != nil {
			fmt.Println("apply of bootstrap gone wrong", err)
		}
	}

}

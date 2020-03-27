package main

import (
	"fmt"
	"github.com/krilor/gossh/apt"
	"github.com/krilor/gossh/file"
	"github.com/krilor/gossh/machine"
	"github.com/krilor/gossh/rule"
	"github.com/krilor/gossh/state"
	"github.com/pkg/errors"
)

func GetState() (state.State, error) {

	newState := state.New()

	// Add a machine to the state
	m, err := machine.New("localhost", 22, "myUsername", "myPassword")
	if err != nil {
		fmt.Printf("could not get new machine %v: %v\n", m, err)
		return newState, err
	}
	newState.AddMachine(m)

	// TODO - add inventory, e.g.:
	// state.AddInventory("./inventory.json")

	// Adding predefined rules from separate packages by using state.AddRule

	// file.Exists is not a very helpful rule, it just creates a empty file if it does not exist
	newState.AddRule(file.Exists("/tmp/hello.nothing2"))

	// apt.Package installs/uninstalls a apt package
	newState.AddRule(apt.Package{
		Name:   "tree",
		Status: apt.StatusInstalled,
	})

	// This rule does nothing useful, but just shows off the use of a simple cmd based rule
	// This will allways run
	newState.AddRule(rule.Cmd{
		CheckCmd:  "false",
		EnsureCmd: "ls",
	})

	// This rule is a meta-rule used to construct other rules on the fly

	// This is where it starts to get hairy. The Meta rule is used to create a custom rule on the fly.
	// The example is quite simple and not very useful, but shows how to use commands directly on m,
	//  as well as reusing the Ensure command of another Rule
	filename := "somefile.txt"

	newState.AddRule(rule.Meta{
		CheckFunc: func(m *machine.Machine, sudo bool) (bool, error) {
			cmd := fmt.Sprintf("ls -1 /tmp | grep %s", filename)
			r, err := m.Run(cmd, false)
			if err != nil {
				return false, errors.Wrap(err, "could not check for somefile")
			}
			if r.ExitStatus == 0 {
				return true, nil
			}

			return false, nil
		},
		EnsureFunc: func(m *machine.Machine, sudo bool) error {
			return file.Exists("/tmp/"+filename).Ensure(m, false)
		},
	})

	return newState, nil
}

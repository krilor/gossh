package main

import (
	"fmt"
	"log"
	"os"

	"github.com/krilor/gossh"
	"github.com/krilor/gossh/rules/x/apt"
	"github.com/krilor/gossh/rules/x/base"
	"github.com/krilor/gossh/rules/x/file"
	"github.com/pkg/errors"
)

func main() {

	f, err := os.OpenFile("random.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	inventory := gossh.Inventory{}

	// Add a host to the inventory
	// As of now, it's hardcoded to a docker container on localhost
	m, err := gossh.NewHost("localhost", 2222, "gossh", "gosshpwd")
	if err != nil {
		fmt.Printf("could not get new host %v: %v\n", m, err)
		return
	}

	inventory.Add(m)

	// TODO - add inventory from files, e.g.:
	// gossh.NewInventoryFromFile("./inventory.json")

	bootstrap := base.Multi{}

	// file.Exists is not a very helpful rule, it just creates a empty file if it does not exist
	bootstrap.Add(file.Exists{Path: "/tmp/hello.nothing2"})

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
		CheckFunc: func(trace gossh.Trace, t gossh.Target) (bool, error) {
			cmd := fmt.Sprintf("ls -1 /tmp | grep %s", filename)
			r, err := t.RunQuery(trace, cmd, "")
			if err != nil {
				return false, errors.Wrap(err, "could not check for somefile")
			}
			if r.ExitStatus == 0 {
				return true, nil
			}

			return false, nil
		},
		EnsureFunc: func(trace gossh.Trace, t gossh.Target) error {
			return file.Exists{Path: "/tmp/" + filename}.Ensure(trace, m)
		},
	})

	for _, m := range inventory {
		log.Println("doing host", m)
		err = m.Apply(gossh.NewTrace(), "bootstrap", bootstrap)
		if err != nil {
			fmt.Println("apply of bootstrap gone wrong", err)
		}
	}

}

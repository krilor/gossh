package state

import (
	"fmt"
	"log"

	"github.com/krilor/gossh/machine"
	"github.com/krilor/gossh/rule"
)

// State is where you register all machines and rules
type State struct {
	machines []*machine.Machine
	rules    []rule.Rule
}

// New returns a new state
func New() State {
	return State{
		machines: []*machine.Machine{},
		rules:    []rule.Rule{},
	}
}

// AddMachine adds a machine to the state
func (s *State) AddMachine(m *machine.Machine) {
	s.machines = append(s.machines, m)
}

// AddRule adds a rule to the state
func (s *State) AddRule(r rule.Rule) {
	s.rules = append(s.rules, r)
}

// Apply runs all rules on all machines
func (s State) Apply() error {
	for _, m := range s.machines {
		log.Printf("Running on machine %v", m)

		for _, r := range s.rules {
			log.Printf("Running rule %v", r)

			ok, err := r.Check(m, false)
			if err != nil {
				return fmt.Errorf("could not check rule %v on machinve %v", r, m)
			}

			if ok {
				log.Printf("Rule check %v was ok", r)
				continue
			}

			log.Printf("Rule check %v was NOT ok", r)

			err = r.Ensure(m, false)
			if err != nil {
				return fmt.Errorf("could not ensure rule %v on machine %v", r, m)
			}
		}
	}
	return nil
}

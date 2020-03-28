package state

import (
	"github.com/krilor/gossh/machine"
)

// State is where you register all machines and rules
type State struct {
	machines []*machine.Machine
	rules    []machine.Rule
}

// New returns a new state
func New() State {
	return State{
		machines: []*machine.Machine{},
		rules:    []machine.Rule{},
	}
}

// AddMachine adds a machine to the state
func (s *State) AddMachine(m *machine.Machine) {
	s.machines = append(s.machines, m)
}

// AddRule adds a rule to the state
func (s *State) AddRule(r machine.Rule) {
	s.rules = append(s.rules, r)
}

// Apply runs all rules on all machines
func (s State) Apply() error {
	return nil
}

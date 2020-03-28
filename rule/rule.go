package rule

import (
	"fmt"
	"log"

	"github.com/krilor/gossh/machine"
	"github.com/pkg/errors"
)

// Checker is the interface that wraps the Check method.
//
// Check runs commands on m to and reports to ok wether or not the rule is adhered to or not.
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Checker interface {
	Check(m *machine.Machine, sudo bool) (ok bool, err error)
}

// Ensurer is the interface that wraps the Ensure method
//
// Ensure runs commands on m to ensure that a specified state is adhered to.
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Ensurer interface {
	Ensure(m *machine.Machine, sudo bool) error
}

// Rule is the interface that groups the Check and Ensure methods
//
// The main purpose of this combines interface is to have a Rule that conditionally run Ensure based on Check
//
// In go-speak, it should have been called a CheckEnsurer
type Rule interface {
	Checker
	Ensurer
}

// Cmd can be used to do a simple command based rule
type Cmd struct {
	CheckCmd  string
	EnsureCmd string
}

// Check will pass if Cmd's CheckCmd return ExitStatus equals 0
func (c Cmd) Check(m *machine.Machine, sudo bool) (bool, error) {
	r, err := m.Run(c.CheckCmd, sudo)

	if err != nil {
		return false, errors.Wrapf(err, "command %s failed", c.CheckCmd)
	}

	return r.ExitStatus == 0, nil
}

// Ensure simply runs Cmd's EnsureCmd
func (c Cmd) Ensure(m *machine.Machine, sudo bool) error {
	_, err := m.Run(c.EnsureCmd, sudo)

	if err != nil {
		return errors.Wrapf(err, "command %s failed", c.EnsureCmd)
	}

	return nil
}

// Meta is a rule that can be used to write your own rules
type Meta struct {
	CheckFunc  func(m *machine.Machine, sudo bool) (bool, error)
	EnsureFunc func(m *machine.Machine, sudo bool) error
}

// Check runs CheckFunc
func (ma Meta) Check(m *machine.Machine, sudo bool) (bool, error) {
	return ma.CheckFunc(m, sudo)
}

// Ensure runs EnsureFunc
func (ma Meta) Ensure(m *machine.Machine, sudo bool) error {
	return ma.EnsureFunc(m, sudo)
}

// NewMeta can be used to create a new Meta rule
func NewMeta(check func(m *machine.Machine, sudo bool) (bool, error), ensure func(m *machine.Machine, sudo bool) error) Meta {
	return Meta{check, ensure}
}

// Multi is a rule that consists of a list of rules
//
// Check will allways return false and nil and only serves as a building block for nested rules.
type Multi []Rule

// Check implements Checker
// Check will allways return false and nil
func (mi Multi) Check(m *machine.Machine, sudo bool) (bool, error) {
	return false, nil
}

// Ensure implements Ensurer
//
// Ensure runs Check and Ensure on all rules in the list.
//
// Multi will stop executing and return an error if encountering an error from any Check or Ensure method.
func (mi Multi) Ensure(m *machine.Machine, sudo bool) error {

	for _, r := range mi {
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

	return nil
}

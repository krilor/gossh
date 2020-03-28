package base

import (
	"log"

	"github.com/krilor/gossh/machine"
	"github.com/pkg/errors"
)

// Package base provides simple rules that can be used as building blocks in other rules

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
type Multi []machine.Rule

// Check implements Checker
// Check will allways return false and nil
func (p Multi) Check(m *machine.Machine, sudo bool) (bool, error) {
	return false, nil
}

// Ensure implements Ensurer
//
// Ensure runs Check and Ensure on all rules in the list.
//
// Multi will stop executing and return an error if encountering an error from any Check or Ensure method.
func (p Multi) Ensure(m *machine.Machine, sudo bool) error {

	for _, r := range p {
		log.Printf("Running rule %v", r)
		m.Apply(r)
	}

	return nil
}

// Add adds a rule to Multi p
func (p *Multi) Add(r machine.Rule) {
	l := append(*p, r)
	*p = l
	return
}

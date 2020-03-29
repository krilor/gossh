package base

import (
	"fmt"

	"github.com/krilor/gossh"
	"github.com/pkg/errors"
)

// Package base provides simple rules that can be used as building blocks in other rules

// Cmd can be used to do a simple command based rule
type Cmd struct {
	CheckCmd  string
	EnsureCmd string
}

// Check will pass if Cmd's CheckCmd return ExitStatus equals 0
func (c Cmd) Check(trace gossh.Trace, m *gossh.Host) (bool, error) {
	r, err := m.Run(trace, c.CheckCmd, false)

	if err != nil {
		return false, errors.Wrapf(err, "command %s failed", c.CheckCmd)
	}

	return r.ExitStatus == 0, nil
}

// Ensure simply runs Cmd's EnsureCmd
func (c Cmd) Ensure(trace gossh.Trace, m *gossh.Host) error {
	_, err := m.Run(trace, c.EnsureCmd, false)

	if err != nil {
		return errors.Wrapf(err, "command %s failed", c.EnsureCmd)
	}

	return nil
}

// Meta is a rule that can be used to write your own rules, on the fly
type Meta struct {
	CheckFunc  func(trace gossh.Trace, m *gossh.Host) (bool, error)
	EnsureFunc func(trace gossh.Trace, m *gossh.Host) error
}

// Check runs CheckFunc
func (ma Meta) Check(trace gossh.Trace, m *gossh.Host) (bool, error) {
	return ma.CheckFunc(trace, m)
}

// Ensure runs EnsureFunc
func (ma Meta) Ensure(trace gossh.Trace, m *gossh.Host) error {
	return ma.EnsureFunc(trace, m)
}

// NewMeta can be used to create a new Meta rule
func NewMeta(check func(trace gossh.Trace, m *gossh.Host) (bool, error), ensure func(trace gossh.Trace, m *gossh.Host) error) Meta {
	return Meta{check, ensure}
}

// Multi is a rule that consists of a list of rules
//
// Check will allways return false and nil and only serves as a building block for nested rules.
type Multi []gossh.Rule

// Check implements Checker
// Check will allways return false and nil
func (p Multi) Check(trace gossh.Trace, m *gossh.Host) (bool, error) {
	return false, nil
}

// Ensure implements Ensurer
//
// Ensure runs Check and Ensure on all rules in the list.
//
// Multi will stop executing and return an error if encountering an error from any Check or Ensure method.
func (p Multi) Ensure(trace gossh.Trace, m *gossh.Host) error {

	for i, r := range p {
		m.Apply(fmt.Sprintf("multi%d", i), trace, r)
	}

	return nil
}

// Add adds a rule to Multi p
func (p *Multi) Add(r gossh.Rule) {
	l := append(*p, r)
	*p = l
	return
}

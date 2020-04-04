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
	User      string
}

// Check will pass if Cmd's CheckCmd return ExitStatus equals 0
func (c Cmd) Check(trace gossh.Trace, t gossh.Target) (bool, error) {
	r, err := t.RunQuery(trace, c.CheckCmd, "", c.User)

	if err != nil {
		return false, errors.Wrapf(err, "command %s failed", c.CheckCmd)
	}

	return r.ExitStatus == 0, nil
}

// Ensure simply runs Cmd's EnsureCmd
func (c Cmd) Ensure(trace gossh.Trace, t gossh.Target) error {
	r, err := t.RunChange(trace, c.EnsureCmd, "", c.User)

	if err != nil || !r.ExitStatusSuccess() {
		return errors.Wrapf(err, "command %s failed", c.EnsureCmd)
	}

	return nil
}

// Meta is a rule that can be used to write your own rules, on the fly
type Meta struct {
	CheckFunc  func(trace gossh.Trace, t gossh.Target) (bool, error)
	EnsureFunc func(trace gossh.Trace, t gossh.Target) error
}

// Check runs CheckFunc
func (ma Meta) Check(trace gossh.Trace, t gossh.Target) (bool, error) {
	return ma.CheckFunc(trace, t)
}

// Ensure runs EnsureFunc
func (ma Meta) Ensure(trace gossh.Trace, t gossh.Target) error {
	return ma.EnsureFunc(trace, t)
}

// NewMeta can be used to create a new Meta rule
func NewMeta(check func(trace gossh.Trace, t gossh.Target) (bool, error), ensure func(trace gossh.Trace, t gossh.Target) error) Meta {
	return Meta{check, ensure}
}

// Multi is a rule that consists of a list of rules
//
// Check will allways return false and nil and only serves as a building block for nested rules.
type Multi []gossh.Rule

// Check implements Checker
// Check will allways return false and nil
func (p Multi) Check(trace gossh.Trace, t gossh.Target) (bool, error) {
	return false, nil
}

// Ensure implements Ensurer
//
// Ensure runs Check and Ensure on all rules in the list.
//
// Multi will stop executing and return an error if encountering an error from any Check or Ensure method.
func (p Multi) Ensure(trace gossh.Trace, t gossh.Target) error {

	for i, r := range p {
		name := fmt.Sprintf("multi%d", i)
		err := t.Apply(trace, name, r)
		if err != nil {
			return errors.Wrapf(err, "%s - %s failed to apply", name, r)
		}
	}

	return nil
}

// Add adds a rule to Multi p
func (p *Multi) Add(r gossh.Rule) {
	l := append(*p, r)
	*p = l
	return
}

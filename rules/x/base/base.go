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

// Ensure simply runs Cmd's CheckCmd, then EnsureCmd
func (c Cmd) Ensure(h *gossh.Host) (gossh.Status, error) {

	r, err := h.RunCheck(c.CheckCmd, "", c.User)

	if err != nil {
		return gossh.StatusFailed, errors.Wrapf(err, "command %s failed", c.CheckCmd)
	}

	if r.ExitStatus == 0 {
		return gossh.StatusSatisfied, nil
	}

	r, err = h.RunChange(c.EnsureCmd, "", c.User)

	if err != nil || !r.Success() {
		return gossh.StatusFailed, errors.Wrapf(err, "command %s failed", c.EnsureCmd)
	}

	return gossh.StatusEnforced, nil
}

// Meta is a rule that can be used to write your own rules, on the fly
type Meta struct {
	EnsureFunc func(h *gossh.Host) (gossh.Status, error)
}

// Ensure runs EnsureFunc
func (ma Meta) Ensure(h *gossh.Host) (gossh.Status, error) {
	return ma.EnsureFunc(h)
}

// NewMeta can be used to create a new Meta rule
func NewMeta(ensure func(h *gossh.Host) (gossh.Status, error)) Meta {
	return Meta{ensure}
}

// Multi is a rule that consists of a list of rules
//
// Check will allways return false and nil and only serves as a building block for nested rules.
type Multi []gossh.Rule

// Ensure implements Ensurer
//
// Ensure runs Check and Ensure on all rules in the lish.
//
// Multi will stop executing and return an error if encountering an error from any Check or Ensure method.
func (p Multi) Ensure(h *gossh.Host) (gossh.Status, error) {

	var status gossh.Status

	for i, r := range p {
		name := fmt.Sprintf("multi%d - %v", i, r)
		s, err := h.Apply(name, r)
		if s > status {
			status = s
		}
		if err != nil {
			return status, errors.Wrapf(err, "%s - %s failed to apply", name, r)
		}
	}

	return status, nil
}

// Add adds a rule to Multi p
func (p *Multi) Add(r gossh.Rule) {
	l := append(*p, r)
	*p = l
	return
}

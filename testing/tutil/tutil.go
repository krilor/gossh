package tutil

import "strings"

// Package tutil provides testing utility functions

// DefaultString returns def if in is empty
func DefaultString(in, def string) string {
	if strings.Trim(in, " \t\n") == "" {
		return def
	}

	return in
}

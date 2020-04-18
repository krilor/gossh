package sh

import (
	"strings"
)

// Package sh contains (ba)sh related methods and types

// Escape escapes a cmd so that it can be used inside a single-quoted argument.
// The intended purpose e.g. when strings are used as input to sh -c '%s'
// The method assumes that the outer, surrounding quote is a singlequote.
// The surrounding quote must not be part of cmd.
func Escape(cmd string) string {
	// nice ref on stack: https://stackoverflow.com/questions/1250079/how-to-escape-single-quotes-within-single-quoted-strings
	return strings.ReplaceAll(cmd, `'`, `'\''`)
}

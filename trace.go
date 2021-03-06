package gossh

import "github.com/lithammer/shortuuid"

// Trace is a type used to trace Apply and Run on Hosts
type trace struct {
	id    string
	prev  string
	level int
}

// String implements fmt.Stringer
func (t trace) String() string {
	return t.id
}

// span gets a new random ID and creates a span within that
func (t trace) span() trace {
	t.prev = t.id
	t.id = newUniqueID()
	t.level = t.level + 1
	return t
}

// newTrace returns a new trace with id as the first element
// If id is an empty string, then a new id is generated
func newTrace() trace {
	return trace{
		id:   newUniqueID(),
		prev: "",
	}
}

// newUniqueID is a simple wrapper for the id generation implementation/package
func newUniqueID() string {
	return shortuuid.New()
}

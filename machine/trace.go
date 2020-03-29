package machine

import "github.com/lithammer/shortuuid"

// Trace is a type used to trace Apply and Run on Machines
type Trace struct {
	id   string
	prev string
}

// String implements fmt.Stringer
func (t Trace) String() string {
	return t.id
}

// ID returns the ID
func (t Trace) ID() string {
	return t.id
}

// Prev returns the second-last trace id
func (t Trace) Prev() string {
	return t.prev
}

// Span gets a new random ID and creates a span within that
func (t Trace) Span() Trace {
	t.prev = t.id
	t.id = newUniqueID()
	return t
}

// NewTrace returns a new trace with id as the first element
// If id is an empty string, then a new id is generated
func NewTrace() Trace {
	return Trace{
		id:   newUniqueID(),
		prev: "",
	}
}

// newUniqueID is a simple wrapper for the id generation implementation/package
func newUniqueID() string {
	return shortuuid.New()
}

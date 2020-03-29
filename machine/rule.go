package machine

// Checker is the interface that wraps the Check method.
//
// Check runs commands on m to and reports to ok wether or not the rule is adhered to or not.
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Checker interface {
	Check(trace Trace, m *Machine) (ok bool, err error)
}

// Ensurer is the interface that wraps the Ensure method
//
// Ensure runs commands on m to ensure that a specified state is adhered to.
// If anything goes wrong, error err is returned. Otherwise err is nil.
type Ensurer interface {
	Ensure(trace Trace, m *Machine) error
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

package gossh

// Inventory is a list of Hosts
type Inventory []*Host

// Add adds m to i
func (i *Inventory) Add(m *Host) {
	l := append(*i, m)
	*i = l
	return
}

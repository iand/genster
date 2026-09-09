package genbio

import "hash/fnv"

// Variant selects phrasing deterministically so that one subject reads the
// same on every generation while different subjects read differently.
type Variant struct {
	state uint64
}

// NewVariant returns a Variant seeded from s, typically a subject identifier.
func NewVariant(s string) *Variant {
	h := fnv.New64a()
	h.Write([]byte(s))
	return &Variant{state: h.Sum64()}
}

// Pick returns one of the alternatives and advances the internal state so that
// successive picks vary independently. It returns the empty string when no
// alternatives are supplied.
func (v *Variant) Pick(alternatives ...string) string {
	if len(alternatives) == 0 {
		return ""
	}
	return alternatives[v.next()%uint64(len(alternatives))]
}

// next returns the next value in the xorshift sequence and updates the state.
func (v *Variant) next() uint64 {
	v.state ^= v.state << 13
	v.state ^= v.state >> 7
	v.state ^= v.state << 17
	return v.state
}

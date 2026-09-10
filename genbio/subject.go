package genbio

import "github.com/iand/genster/model"

// Subject is the input to Bio, carrying the facts a biography is drawn from.
// It currently embeds a model.Person; individual fields will be promoted onto
// Subject as the package is decoupled from the genster model.
type Subject struct {
	*model.Person
}

// FromPerson returns a Subject drawn from p.
func FromPerson(p *model.Person) Subject {
	return Subject{Person: p}
}

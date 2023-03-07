package tree

import (
	"fmt"
	"strings"

	"github.com/iand/genster/model"
)

type Overrides struct {
	person map[string][]FV
	place  map[string][]FV
}

func (o *Overrides) AddOverride(kind string, id string, field string, value string) error {
	switch kind {
	case "person":
		if _, ok := personOverrides[field]; !ok {
			return fmt.Errorf("unsupported override field for person: %s", field)
		}
		if o.person == nil {
			o.person = make(map[string][]FV, 0)
		}
		o.person[id] = append(o.person[id], FV{Field: strings.ToLower(field), Value: value})
	case "place":
		if _, ok := placeOverrides[field]; !ok {
			return fmt.Errorf("unsupported override field for place: %s", field)
		}
		if o.place == nil {
			o.place = make(map[string][]FV, 0)
		}
		o.place[id] = append(o.place[id], FV{Field: strings.ToLower(field), Value: value})
	default:
		return fmt.Errorf("unsupported override kind: %s", kind)
	}

	return nil
}

func (o *Overrides) ApplyPerson(p *model.Person) error {
	fvs, ok := o.person[p.ID]
	if !ok {
		return nil
	}

	for _, fv := range fvs {
		if fn, ok := personOverrides[fv.Field]; ok {
			fn(p, fv.Value)
		}
	}

	return nil
}

func (o *Overrides) ApplyPlace(p *model.Place) error {
	fvs, ok := o.place[p.ID]
	if !ok {
		return nil
	}

	for _, fv := range fvs {
		if fn, ok := placeOverrides[fv.Field]; ok {
			fn(p, fv.Value)
		}
	}

	return nil
}

type FV struct {
	Field string
	Value string
}

// all possible person overrides. use a map so the names of the overrides could be
// printed in a help command.
var personOverrides = map[string]func(p *model.Person, v string){
	"nickname": func(p *model.Person, v string) { p.NickName = v },
	"olb":      func(p *model.Person, v string) { p.Olb = v },
}

// all possible place overrides
var placeOverrides = map[string]func(p *model.Place, v string){
	"preferredname": func(p *model.Place, v string) { p.PreferredName = v },
}

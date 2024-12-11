package narrative

import (
	"fmt"

	"github.com/iand/genster/model"
)

type NameChooser interface {
	FirstUse(any) string   // name to use for first occurrence
	Subsequent(any) string // name to use for subsequent occurrences
}

type DefaultNameChooser struct{}

var _ NameChooser = DefaultNameChooser{}

func (c DefaultNameChooser) FirstUse(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredUniqueName
	case *model.Place:
		return vt.PreferredUniqueName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c DefaultNameChooser) Subsequent(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredFamiliarName
	case *model.Place:
		return vt.PreferredName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

// FullNameChooser always returns the full name
type FullNameChooser struct{}

var _ NameChooser = FullNameChooser{}

func (c FullNameChooser) FirstUse(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredFullName
	case *model.Place:
		return vt.PreferredFullName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c FullNameChooser) Subsequent(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredFullName
	case *model.Place:
		return vt.PreferredFullName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

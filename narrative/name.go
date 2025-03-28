package narrative

import (
	"fmt"

	"github.com/iand/genster/model"
)

type NameChooser interface {
	FirstUse(any) string                                            // name to use for first occurrence
	Subsequent(any) string                                          // name to use for subsequent occurrences
	FirstUseSplit(m any, pov *model.POV) (string, string, string)   // prefix, name and suffix to use for first occurrence
	SubsequentSplit(m any, pov *model.POV) (string, string, string) // prefix, name and suffix to use for subsequent occurrences
}

type DefaultNameChooser struct {
	POV *model.POV
}

var _ NameChooser = DefaultNameChooser{}

func (c DefaultNameChooser) FirstUse(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredUniqueName
	case *model.Place:
		return vt.ProseName
	case *model.Family:
		return vt.PreferredUniqueName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c DefaultNameChooser) FirstUseSplit(v any, pov *model.POV) (string, string, string) {
	switch vt := v.(type) {
	case *model.Person:
		return "", vt.PreferredUniqueName, ""
	case *model.Place:
		prefix := vt.DescriptivePrefix()
		if prefix != "" {
			prefix += " "
		}

		suffix := vt.FullContext
		if suffix != "" {
			suffix = " in " + suffix
		}
		return prefix + " ", vt.Name, suffix
	case *model.Family:
		return "", vt.PreferredUniqueName, ""
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c DefaultNameChooser) Subsequent(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredFamiliarName
	case *model.Place:
		return vt.NameWithDistrict
	case *model.Family:
		return vt.PreferredFamiliarName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c DefaultNameChooser) SubsequentSplit(v any, pov *model.POV) (string, string, string) {
	switch vt := v.(type) {
	case *model.Person:
		return "", vt.PreferredFamiliarName, ""
	case *model.Place:
		return "", vt.NameWithDistrict, ""
	case *model.Family:
		return "", vt.PreferredFamiliarName, ""
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
		return vt.FullName
	case *model.Family:
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
		return vt.FullName
	case *model.Family:
		return vt.PreferredFullName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c FullNameChooser) FirstUseSplit(v any, pov *model.POV) (string, string, string) {
	switch vt := v.(type) {
	case *model.Person:
		return "", vt.PreferredFullName, ""
	case *model.Place:
		return "", vt.FullName, ""
	case *model.Family:
		return "", vt.PreferredFullName, ""
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c FullNameChooser) SubsequentSplit(v any, pov *model.POV) (string, string, string) {
	switch vt := v.(type) {
	case *model.Person:
		return "", vt.PreferredFullName, ""
	case *model.Place:
		return "", vt.FullName, ""
	case *model.Family:
		return "", vt.PreferredFullName, ""
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

type TimelineNameChooser struct{}

var _ NameChooser = TimelineNameChooser{}

func (c TimelineNameChooser) FirstUse(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredUniqueName
	case *model.Place:
		return vt.FullName
	case *model.Family:
		return vt.PreferredUniqueName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c TimelineNameChooser) Subsequent(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		return vt.PreferredFamiliarName
	case *model.Place:
		return vt.NameWithDistrict
	case *model.Family:
		return vt.PreferredFamiliarName
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c TimelineNameChooser) FirstUseSplit(v any, pov *model.POV) (string, string, string) {
	switch vt := v.(type) {
	case *model.Person:
		return "", vt.PreferredUniqueName, ""
	case *model.Place:
		if pov != nil && vt.SameCountry(pov.Place) {
			return "", vt.NameWithRegion, ""
		}
		return "", vt.FullName, ""
	case *model.Family:
		return "", vt.PreferredUniqueName, ""
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

func (c TimelineNameChooser) SubsequentSplit(v any, pov *model.POV) (string, string, string) {
	switch vt := v.(type) {
	case *model.Person:
		return "", vt.PreferredFamiliarName, ""
	case *model.Place:
		return "", vt.NameWithDistrict, ""
	case *model.Family:
		return "", vt.PreferredFamiliarName, ""
	default:
		panic(fmt.Sprintf("unexpected object type in name chooser: %T", v))
	}
}

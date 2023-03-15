package place

import (
// "strings"
)

type Hint interface {
	ID() string // a unique identifier for the hint
}

var _ Hint = MaybeCountry{}

type MaybeCountry struct{}

func (MaybeCountry) ID() string {
	return "maybecountry"
}

func (MaybeCountry) Match(ph *PlaceHierarchy) (bool, []Hint) {
	if ph == nil || ph.Name.Name == "" {
		return false, nil
	}
	if ph.Parent != nil {
		return false, nil
	}

	pn, ok := LookupCountryName(ph.Name.Name)
	if !ok {
		return false, nil
	}
	ph.Name = pn
	ph.Kind = PlaceKindCountry

	return true, pn.ChildHints
}

var _ Hint = MaybeUKNation{}

type MaybeUKNation struct{}

func (MaybeUKNation) ID() string {
	return "maybeuknation"
}

func (MaybeUKNation) Match(ph *PlaceHierarchy) (bool, []Hint) {
	if ph == nil || ph.Name.Name == "" {
		return false, nil
	}
	if ph.Parent != nil && ph.Parent.Kind != PlaceKindCountry && ph.Parent.Name.Name != "United Kingdom" {
		return false, nil
	}

	pn, ok := LookupUKNationName(ph.Name.Name)
	if !ok {
		return false, nil
	}
	ph.Name = pn
	ph.Kind = PlaceKindUKNation

	return true, pn.ChildHints
}

// // A hint attempts to match a place from an input string. It returns a place
// // with a confidence score ranging from 0 (no match) to 1 (confident match)
// type Hint func(pc PlaceHierarchy) (Place, float64)

// var InUK Hint = func(pc PlaceHierarchy) (Place, float64) {
// 	if len(pc.Hierarchy) == 0 {
// 		return nil, 0
// 	}

// 	last := pc.Hierarchy[len(pc.Hierarchy)-1]
// 	if strings.EqualFold(last, "wales") {
// 		return InWales(pc.TrimHierarchy(1))
// 	} else if strings.EqualFold(last, "scotland") {
// 		return InScotland(pc.TrimHierarchy(1))
// 	} else if strings.EqualFold(last, "england") {
// 		return InEngland(pc.TrimHierarchy(1))
// 	} else if strings.EqualFold(last, "northern ireland") {
// 		return InNorthernIreland(pc.TrimHierarchy(1))
// 	}

// 	return matchBest(pc, InEngland, InWales, InScotland, InNorthernIreland)
// }

// var InEngland Hint = func(pc PlaceHierarchy) (Place, float64) {
// 	pl := &TownCountyCountry{CountryName: "England"}
// 	if len(pc.Hierarchy) == 0 {
// 		return pl, 0.5
// 	}

// 	pl.County = pc.Hierarchy[len(pc.Hierarchy)-1]
// 	if len(pc.Hierarchy) == 1 {
// 		return pl, 0.6
// 	}

// 	pl.Town = pc.Hierarchy[len(pc.Hierarchy)-2]
// 	return pl, 0.7
// }

// var (
// 	InScotland        Hint = InEngland
// 	InWales           Hint = InEngland
// 	InNorthernIreland Hint = InEngland
// )

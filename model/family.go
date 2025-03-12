package model

import (
	"strconv"
	"strings"
)

const (
	FamilyBondMarried         = "married"
	FamilyBondUnmarried       = "unmarried"
	FamilyBondLikelyMarried   = "likely married"
	FamilyBondLikelyUnmarried = "likely unmarried"
	FamilyBondUnknown         = "unknown"

	FamilyEndReasonUnknown  = "unknown"
	FamilyEndReasonDivorce  = "divorce"
	FamilyEndReasonDeath    = "death"
	FamilyEndReasonAnulment = "anulment"
)

type Family struct {
	ID                    string   // canonical id
	Tags                  []string // tags to add to the family's page
	PreferredUniqueName   string   // name used to identify the family
	PreferredFullName     string   // full name using full legal names of parents in family
	PreferredFamiliarName string   // name that can be used in prose, usually just the first name of each parent
	Father                *Person
	Mother                *Person
	Children              []*Person
	BestStartEvent        TimelineEvent // event that best represents the start of the family unit if the bond is a marriage type
	BestEndEvent          TimelineEvent // event that best represents the end of the family unit if the bond is a marriage type

	BestStartDate *Date // date that best represents the start of the family unit
	BestEndDate   *Date // date that best represents the end of the family unit

	Timeline       []TimelineEvent
	Bond           string  // the kind of bond between the parents in the family
	EndReason      string  // the reason the family unit ended
	EndDeathPerson *Person // the person whose death ended the family unit, if any

	NumberOfChildren NumberOfChildren
	AllChildrenKnown bool // true if the list of children should be considered a complete list

	PublishChildren bool // true if this family's children should always be included in the publish set
	Unknown         bool
}

func (f *Family) SameAs(other *Family) bool {
	if f == nil || other == nil {
		return false
	}
	return f == other || (f.ID != "" && f.ID == other.ID)
}

func (f *Family) OtherParent(p *Person) *Person {
	if p.SameAs(f.Father) {
		return f.Mother
	}
	if p.SameAs(f.Mother) {
		return f.Father
	}
	return UnknownPerson()
}

func (f *Family) IsUnknown() bool {
	if f == nil {
		return true
	}
	return f.Unknown
}

type NumberOfChildren string

const NumberOfChildrenUnknown NumberOfChildren = ""

func ParseNumberOfChildren(v string) (NumberOfChildren, error) {
	if v == "" {
		return NumberOfChildrenUnknown, nil
	}

	n := NumberOfChildren(v)
	if strings.HasSuffix(v, "+") {
		v = v[:len(v)-1]
	}
	if _, err := strconv.Atoi(v); err != nil {
		return NumberOfChildrenUnknown, err
	}

	return n, nil
}

func (n NumberOfChildren) IsUnknown() bool {
	return n == NumberOfChildrenUnknown
}

func (n NumberOfChildren) IsZero() bool {
	return n == "0"
}

func (n NumberOfChildren) Number() int {
	if n == NumberOfChildrenUnknown {
		return 0
	}
	if strings.HasSuffix(string(n), "+") {
		n = n[:len(n)-1]
	}
	if v, err := strconv.Atoi(string(n)); err == nil {
		return v
	}
	return 0
}

func (n NumberOfChildren) IsLowerBound() bool {
	if n == NumberOfChildrenUnknown {
		return false
	}
	return strings.HasSuffix(string(n), "+")
}

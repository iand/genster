package model

import (
	"fmt"
	"testing"
)

func TestPersonCanBeNamed(t *testing.T) {
	testCases := []struct {
		name string
		p    *Person
		want bool
	}{
		{name: "nil", p: nil, want: false},
		{name: "unknown", p: UnknownPerson(), want: false},
		{name: "named", p: &Person{PreferredFullName: "Mary Brown"}, want: true},
		{name: "unidentified", p: &Person{Unidentified: true}, want: false},
		{name: "redacted", p: &Person{Redacted: true}, want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.CanBeNamed(); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// A partner the person was not married to is their partner, whatever children they had.
func TestPersonRelationToUnmarriedPartner(t *testing.T) {
	for _, children := range []int{0, 1, 2} {
		for _, bond := range []string{FamilyBondUnmarried, FamilyBondLikelyUnmarried, FamilyBondUnknown} {
			t.Run(fmt.Sprintf("%s with %d children", bond, children), func(t *testing.T) {
				man := &Person{ID: "m", Gender: GenderMale}
				woman := &Person{ID: "w", Gender: GenderFemale}
				f := &Family{ID: "f", Bond: bond, Father: man, Mother: woman}
				for i := range children {
					f.Children = append(f.Children, &Person{ID: fmt.Sprintf("c%d", i), Gender: GenderMale})
				}
				man.Families = []*Family{f}
				woman.Families = []*Family{f}

				if got := woman.RelationTo(man, Year(1900)); got != "partner" {
					t.Errorf("woman to man: got %q, want %q", got, "partner")
				}
				if got := man.RelationTo(woman, Year(1900)); got != "partner" {
					t.Errorf("man to woman: got %q, want %q", got, "partner")
				}
			})
		}
	}
}

func TestPersonUnknownNameDescription(t *testing.T) {
	testCases := []struct {
		gender Gender
		want   string
	}{
		{gender: GenderMale, want: "a man whose name is not known"},
		{gender: GenderFemale, want: "a woman whose name is not known"},
		{gender: GenderUnknown, want: "a person whose name is not known"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.gender), func(t *testing.T) {
			p := &Person{Gender: tc.gender, Unidentified: true}
			if got := p.UnknownNameDescription(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

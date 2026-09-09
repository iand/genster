package genbio

import (
	"testing"

	"github.com/iand/genster/model"
)

var _ phraser = (*Variant)(nil)

// fixedChooser always selects the alternative at idx, giving predictable output.
type fixedChooser struct{ idx int }

func (c fixedChooser) Pick(alternatives ...string) string {
	if len(alternatives) == 0 {
		return ""
	}
	return alternatives[c.idx%len(alternatives)]
}

func place(fullName string) *model.Place {
	return &model.Place{FullName: fullName}
}

func birth(p *model.Person, d *model.Date, pl *model.Place) *model.BirthEvent {
	return &model.BirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: d, Place: pl},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
}

func death(p *model.Person, d *model.Date, pl *model.Place) *model.DeathEvent {
	return &model.DeathEvent{
		GeneralEvent:           model.GeneralEvent{Date: d, Place: pl},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
}

// fullMale returns a person with birth, marriage, children and death all known.
func fullMale() *model.Person {
	p := &model.Person{
		ID:                "p1",
		PreferredFullName: "John Smith",
		Gender:            model.GenderMale,
	}
	p.BestBirthlikeEvent = birth(p, model.PreciseDate(1820, 3, 15), place("Manchester, Lancashire"))
	p.BestDeathlikeEvent = death(p, model.PreciseDate(1878, 6, 20), place("Salford, Lancashire"))
	spouse := &model.Person{ID: "p1w", PreferredFamiliarFullName: "Mary Jones", Gender: model.GenderFemale}
	p.Children = []*model.Person{{ID: "c1"}, {ID: "c2"}, {ID: "c3"}}
	p.Families = []*model.Family{
		{ID: "f1", Bond: model.FamilyBondMarried, Father: p, Mother: spouse, BestStartDate: model.Year(1845)},
	}
	return p
}

func TestBioGolden(t *testing.T) {
	// A female known only by birth year and a count of children.
	female := &model.Person{ID: "p2", PreferredFullName: "Mary Ann Jones", Gender: model.GenderFemale}
	female.BestBirthlikeEvent = birth(female, model.Year(1900), nil)
	female.Children = []*model.Person{{ID: "d1"}, {ID: "d2"}}

	// A person of unknown gender with a birthplace and a year of death.
	they := &model.Person{ID: "p3", PreferredFullName: "Alex Doe", Gender: model.GenderUnknown}
	they.BestBirthlikeEvent = birth(they, nil, place("Yorkshire"))
	they.BestDeathlikeEvent = death(they, model.Year(1950), nil)

	testCases := []struct {
		name string
		p    *model.Person
		want string
	}{
		{
			name: "full_male",
			p:    fullMale(),
			want: "John Smith was born on 15 Mar, 1820 in Manchester, Lancashire. He married Mary Jones in 1845 and had three children. He died on 20 Jun, 1878 in Salford, Lancashire, aged 58.",
		},
		{
			name: "female_partial",
			p:    female,
			want: "Mary Ann Jones was born in 1900. She had two children.",
		},
		{
			name: "unknown_gender",
			p:    they,
			want: "Alex Doe was born in Yorkshire. They died in 1950.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := bio(FromPerson(tc.p), fixedChooser{0})
			if got != tc.want {
				t.Errorf("got:\n  %q\nwant:\n  %q", got, tc.want)
			}
		})
	}
}

func TestBioDeterministic(t *testing.T) {
	p := fullMale()
	first := BioFromPerson(p)
	for range 5 {
		if got := BioFromPerson(p); got != first {
			t.Fatalf("Bio is not deterministic: %q then %q", first, got)
		}
	}
	if first == "" {
		t.Error("Bio returned empty string")
	}
}

func TestBioVariesAcrossSubjects(t *testing.T) {
	seen := map[string]bool{}
	for i := range 50 {
		p := fullMale()
		p.ID = "seed-" + string(rune('a'+i))
		seen[BioFromPerson(p)] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected phrasing to vary across subjects, saw only %d distinct biographies", len(seen))
	}
}

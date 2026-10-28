package narrative

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

func parish(name, context string) *model.Place {
	return &model.Place{ID: "pl-" + name, Name: name, FullContext: context, PlaceType: model.PlaceTypeParish}
}

func checkContains(t *testing.T, got string, want string, notWant ...string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("output does not contain %q\ngot:\n%s", want, got)
	}
	for _, nw := range notWant {
		if strings.Contains(got, nw) {
			t.Errorf("output contains %q\ngot:\n%s", nw, got)
		}
	}
}

// A sentence about the children of a marriage starts a new sentence with no space before
// the full stop, and a parish is named with a single space after "parish of".
func TestFamilyStatementMarriedChildrenSentence(t *testing.T) {
	for _, seq := range []int{0, 2} {
		husband := &model.Person{ID: "h", PreferredGivenName: "Robert", PreferredFamiliarName: "Robert", Gender: model.GenderMale}
		wife := &model.Person{ID: "w", PreferredUniqueName: "Sarah Whitney", PreferredFamiliarName: "Sarah", Gender: model.GenderFemale}
		marriage := &model.MarriageEvent{
			GeneralEvent:      model.GeneralEvent{Date: model.PreciseDate(1830, 9, 1), Place: parish("Sharnbrook", "Bedfordshire, England")},
			GeneralUnionEvent: model.GeneralUnionEvent{Husband: husband, Wife: wife},
		}
		f := &model.Family{
			ID: "fam", Bond: model.FamilyBondMarried, Father: husband, Mother: wife,
			BestStartDate: marriage.Date, BestStartEvent: marriage,
			Children: []*model.Person{
				{ID: "c1", PreferredFamiliarName: "Ann", Gender: model.GenderFemale},
				{ID: "c2", PreferredFamiliarName: "John", Gender: model.GenderMale},
				{ID: "c3", PreferredFamiliarName: "Mary", Gender: model.GenderFemale},
			},
		}
		husband.Families = []*model.Family{f}

		s := &FamilyStatement[md.Text]{Principal: husband, Family: f}
		intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: husband}}
		enc := &md.Content{}
		s.RenderDetail(seq, intro, enc, &DefaultNameChooser{})
		got := enc.String()

		checkContains(t, got, "in the parish of Sharnbrook in Bedfordshire, England.", " .", "of  ")
	}
}

// A birth clause is followed by a comma before the parentage; "X was the son of" is not.
func TestIntroStatementParentageComma(t *testing.T) {
	father := &model.Person{ID: "f", PreferredUniqueName: "Henry Dunmore", PreferredFamiliarName: "Henry", Gender: model.GenderMale}

	p := &model.Person{ID: "p", PreferredGivenName: "John", PreferredFamiliarName: "John", Gender: model.GenderMale, Father: father}
	p.BestBirthlikeEvent = &model.BirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1817, 5, 4), Place: parish("Great Bowden", "Leicestershire, England")},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}

	s := &IntroStatement[md.Text]{Principal: p}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

	checkContains(t, enc.String(), "Leicestershire, England, the son of Henry Dunmore.")
}

// The twin sentence in a child summary is a sentence of its own.
func TestPersonBirthSummaryTwinSentence(t *testing.T) {
	george := &model.Person{ID: "g", PreferredFamiliarName: "George", Gender: model.GenderMale}
	p := &model.Person{ID: "s", PreferredFamiliarName: "Sarah", Gender: model.GenderFemale}
	p.BestBirthlikeEvent = &model.BaptismEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1816, 1, 18), Place: parish("Runwell", "Essex, England")},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
	p.Associations = []model.Association{{Kind: model.AssociationKindTwin, Other: george}}

	enc := &md.Content{}
	got := PersonBirthSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("Sarah"), true, true, false, false).String()

	checkContains(t, got, "England. She was the twin to George.", "England She")
}

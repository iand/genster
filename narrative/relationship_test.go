package narrative

import (
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// endedByDeathOf returns a family of p and other with the given bond that ended with the death
// of other in 1927.
func endedByDeathOf(p, other *model.Person, bond string, children ...*model.Person) *model.Family {
	death := &model.DeathEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1927, 6, 18)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: other}}
	other.BestDeathlikeEvent = death
	f := &model.Family{ID: "f", Bond: bond, Children: children, EndReason: model.FamilyEndReasonDeath, BestEndDate: death.Date, BestEndEvent: death, EndDeathPerson: other}
	if p.Gender.IsMale() {
		f.Father, f.Mother = p, other
	} else {
		f.Father, f.Mother = other, p
	}
	p.Families = append(p.Families, f)
	return f
}

func renderFamilyEnd(p *model.Person, f *model.Family) string {
	s := &FamilyEndStatement[md.Text]{Principal: p, Family: f}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})
	return enc.String()
}

func herbert() *model.Person {
	return &model.Person{ID: "p", PreferredGivenName: "Herbert", PreferredFamiliarName: "Herbert", Gender: model.GenderMale}
}

func annie() *model.Person {
	return &model.Person{ID: "o", PreferredFullName: "Annie Louisa Tew", PreferredFamiliarName: "Annie", Gender: model.GenderFemale}
}

var nonSpouseLabels = []string{"his wife", "her husband", "children's mother", "son's mother", "partner"}

// Shaped like I973: a relationship with no start event and no children is not dated by the
// subject's own death.
func TestUndatedRelationship(t *testing.T) {
	newHenry := func() (*model.Person, *model.Family) {
		p := &model.Person{ID: "p", PreferredGivenName: "Henry", PreferredFamiliarName: "Henry", Gender: model.GenderMale}
		hannah := &model.Person{ID: "h", PreferredUniqueName: "Hannah Gibbard", PreferredFamiliarName: "Hannah", Gender: model.GenderFemale}
		f := &model.Family{ID: "f2", Bond: model.FamilyBondUnknown, Father: p, Mother: hannah, BestStartDate: model.BeforeYear(1877)}
		return p, f
	}

	t.Run("only relationship", func(t *testing.T) {
		p, f := newHenry()
		p.Families = []*model.Family{f}
		checkContains(t, renderFamilyStatement(p, f), "He had a relationship with Hannah Gibbard.", "1877", "met")
	})

	t.Run("after a marriage", func(t *testing.T) {
		p, f := newHenry()
		frances := &model.Person{ID: "w", PreferredUniqueName: "Frances Higgs", PreferredFamiliarName: "Frances", Gender: model.GenderFemale}
		p.Families = []*model.Family{marriedFamily("f1", p, frances, model.Year(1814)), f}
		checkContains(t, renderFamilyStatement(p, f), "He also had a relationship with Hannah Gibbard.", "1877", "met")
	})
}

func TestFamilyEndLikelyMarried(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "Elizabeth", PreferredFamiliarName: "Elizabeth", Gender: model.GenderFemale}
	other := &model.Person{ID: "o", PreferredFullName: "Thomas Ginn", PreferredFamiliarName: "Thomas", Gender: model.GenderMale}
	f := endedByDeathOf(p, other, model.FamilyBondLikelyMarried, &model.Person{ID: "c1"}, &model.Person{ID: "c2"})

	checkContains(t, renderFamilyEnd(p, f), "Thomas, her husband, died")
}

func TestFamilyEndUnmarriedOneChild(t *testing.T) {
	p, other := herbert(), annie()
	child := &model.Person{ID: "c", PreferredFamiliarName: "Bertie", Gender: model.GenderMale}
	f := endedByDeathOf(p, other, model.FamilyBondUnmarried, child)

	checkContains(t, renderFamilyEnd(p, f), "Annie Louisa Tew, the mother of his son Bertie, died", nonSpouseLabels...)
}

func TestFamilyEndUnmarriedSeveralChildren(t *testing.T) {
	p, other := herbert(), annie()
	f := endedByDeathOf(p, other, model.FamilyBondUnmarried, &model.Person{ID: "c1"}, &model.Person{ID: "c2"})

	checkContains(t, renderFamilyEnd(p, f), "Annie Louisa Tew, the mother of his children, died", nonSpouseLabels...)
}

func TestFamilyEndUnknownBondNoChildren(t *testing.T) {
	p, other := herbert(), annie()
	f := endedByDeathOf(p, other, model.FamilyBondUnknown)

	checkContains(t, renderFamilyEnd(p, f), "Annie Louisa Tew died", append(nonSpouseLabels, "the mother of")...)
}

func TestFamilyEndUnmarriedRedactedChild(t *testing.T) {
	p, other := herbert(), annie()
	f := endedByDeathOf(p, other, model.FamilyBondUnmarried, redacted("c", model.GenderMale))

	checkContains(t, renderFamilyEnd(p, f), "Annie Louisa Tew, the mother of his son, died", append(nonSpouseLabels, "living or recently deceased")...)
}

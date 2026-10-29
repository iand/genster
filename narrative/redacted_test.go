package narrative

import (
	"testing"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// redacted returns a person named as the tree names a redacted person.
func redacted(id string, g model.Gender) *model.Person {
	p := &model.Person{ID: id, Gender: g, Redacted: true}
	p.RedactNames("(living or recently deceased person)")
	return p
}

var redactedForms = []string{"(living or recently deceased person)", "living or recently died", "1915"}

// marriedFamily returns a married family of p and other that started on dt.
func marriedFamily(id string, p, other *model.Person, dt *model.Date) *model.Family {
	f := &model.Family{ID: id, Bond: model.FamilyBondMarried, BestStartDate: dt}
	if dt != nil {
		f.BestStartEvent = &model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: dt}}
	}
	if p.Gender.IsMale() {
		f.Father, f.Mother = p, other
	} else {
		f.Father, f.Mother = other, p
	}
	return f
}

func renderFamilyStatement(p *model.Person, f *model.Family) string {
	s := &FamilyStatement[md.Text]{Principal: p, Family: f}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})
	return enc.String()
}

func TestRedactedSpouseInMarriageStatement(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "John", PreferredFamiliarName: "John", Gender: model.GenderMale}
	spouse := redacted("s", model.GenderFemale)
	f := marriedFamily("f", p, spouse, model.Year(1903))
	p.Families = []*model.Family{f}

	checkContains(t, renderFamilyStatement(p, f), "He married in 1903.", redactedForms...)
}

func TestRedactedSpouseUndatedOnlyMarriage(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "John", PreferredFamiliarName: "John", Gender: model.GenderMale}
	spouse := redacted("s", model.GenderFemale)
	f := marriedFamily("f", p, spouse, nil)
	p.Families = []*model.Family{f}

	checkContains(t, renderFamilyStatement(p, f), "He married.", append(redactedForms, "unknown date")...)
}

func TestRedactedSpouseUndatedFurtherMarriage(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "John", PreferredFamiliarName: "John", Gender: model.GenderMale}
	first := &model.Person{ID: "w1", PreferredUniqueName: "Mary Brown", PreferredFamiliarName: "Mary", Gender: model.GenderFemale}
	spouse := redacted("s", model.GenderFemale)
	f1 := marriedFamily("f1", p, first, model.Year(1900))
	f2 := marriedFamily("f2", p, spouse, nil)
	p.Families = []*model.Family{f1, f2}

	checkContains(t, renderFamilyStatement(p, f2), "He also married.", append(redactedForms, "unknown date")...)
}

func TestRedactedPartnerMet(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "Margery", PreferredFamiliarName: "Margery", Gender: model.GenderFemale}
	husband := &model.Person{ID: "h", PreferredUniqueName: "Thomas Peak", PreferredFamiliarName: "Thomas", Gender: model.GenderMale}
	partner := redacted("o", model.GenderMale)
	f1 := marriedFamily("f1", p, husband, model.Year(1912))
	f2 := &model.Family{ID: "f2", Bond: model.FamilyBondUnknown, Father: partner, Mother: p, BestStartDate: model.BeforeYear(1935)}
	p.Families = []*model.Family{f1, f2}

	checkContains(t, renderFamilyStatement(p, f2), "She had another partner before 1935.", redactedForms...)
}

func TestRedactedSpousesInMarriageSummary(t *testing.T) {
	p := &model.Person{ID: "p", PreferredFamiliarName: "Alice", Gender: model.GenderFemale}
	p.Families = []*model.Family{
		marriedFamily("f1", p, redacted("s1", model.GenderMale), model.Year(1932)),
		marriedFamily("f2", p, redacted("s2", model.GenderMale), model.Year(1951)),
	}

	enc := &md.Content{}
	got := PersonMarriageSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("Alice"), false, true).String()

	checkContains(t, got, "Alice married in 1932 and again in 1951.", redactedForms...)
}

func TestRedactedSpouseAfterNamedInMarriageSummary(t *testing.T) {
	p := &model.Person{ID: "p", PreferredFamiliarName: "Alice", Gender: model.GenderFemale}
	named := &model.Person{ID: "s1", PreferredFamiliarFullName: "Arthur Cox", Gender: model.GenderMale}
	p.Families = []*model.Family{
		marriedFamily("f1", p, named, model.Year(1932)),
		marriedFamily("f2", p, redacted("s2", model.GenderMale), model.Year(1951)),
	}

	enc := &md.Content{}
	got := PersonMarriageSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("Alice"), false, true).String()

	checkContains(t, got, "Alice married Arthur Cox in 1932 and again in 1951.", redactedForms...)
}

func TestRedactedSpouseUndatedInMarriageSummary(t *testing.T) {
	p := &model.Person{ID: "p", PreferredFamiliarName: "James", Gender: model.GenderMale}
	p.Families = []*model.Family{marriedFamily("f1", p, redacted("s1", model.GenderFemale), nil)}

	enc := &md.Content{}
	got := PersonMarriageSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("James"), false, true).String()

	checkContains(t, got, "James married.", redactedForms...)
}

func TestRedactedTwin(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "Bernard", PreferredFamiliarName: "Bernard", Gender: model.GenderMale}
	twin := redacted("t", model.GenderMale)
	p.Associations = []model.Association{{Kind: model.AssociationKindTwin, Other: twin}}
	birth := &model.BirthEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1901, 3, 1)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
	p.BestBirthlikeEvent = birth
	p.Timeline = []model.TimelineEvent{birth}

	t.Run("intro", func(t *testing.T) {
		intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
		enc := &md.Content{}
		(&IntroStatement[md.Text]{Principal: p}).RenderDetail(0, intro, enc, &DefaultNameChooser{})
		checkContains(t, enc.String(), "He was a twin.", append(redactedForms, "twin to")...)
	})

	t.Run("birth summary", func(t *testing.T) {
		enc := &md.Content{}
		got := PersonBirthSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("Bernard"), false, true, false, true).String()
		checkContains(t, got, "He was a twin.", append(redactedForms, "twin to")...)
	})
}

func TestRedactedChildOfSingleParent(t *testing.T) {
	for _, dt := range []*model.Date{model.Year(1915), model.PreciseDate(1915, 6, 2)} {
		t.Run(dt.String(), func(t *testing.T) {
			p := &model.Person{ID: "p", PreferredGivenName: "Ethel", PreferredFamiliarName: "Ethel", Gender: model.GenderFemale}
			c := redacted("c", model.GenderFemale)
			c.BestBirthlikeEvent = &model.BirthEvent{GeneralEvent: model.GeneralEvent{Date: dt}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c}}
			f := &model.Family{ID: "f", Bond: model.FamilyBondUnmarried, Father: model.UnknownPerson(), Mother: p, Children: []*model.Person{c}}
			p.Families = []*model.Family{f}

			checkContains(t, renderFamilyStatement(p, f), "a daughter with an unknown father.", redactedForms...)
		})
	}
}

func TestRedactedOtherParentWithOneChild(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "Henry", PreferredFamiliarName: "Henry", Gender: model.GenderMale}
	wife := &model.Person{ID: "w", PreferredUniqueName: "Mary Brown", PreferredFamiliarName: "Mary", Gender: model.GenderFemale}
	other := redacted("o", model.GenderFemale)
	c := redacted("c", model.GenderMale)
	f1 := marriedFamily("f1", p, wife, model.Year(1920))
	f2 := &model.Family{ID: "f2", Bond: model.FamilyBondUnmarried, Father: p, Mother: other, Children: []*model.Person{c}}
	p.Families = []*model.Family{f1, f2}

	checkContains(t, renderFamilyStatement(p, f2), "by another relationship.", append(redactedForms, " with ")...)
}

func TestRedactedSolePartnerWithChildren(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "Lilian", PreferredFamiliarName: "Lilian", Gender: model.GenderFemale}
	other := redacted("o", model.GenderMale)
	f := &model.Family{ID: "f", Bond: model.FamilyBondUnmarried, Father: other, Mother: p, Children: []*model.Person{redacted("c1", model.GenderMale), redacted("c2", model.GenderFemale)}}
	p.Families = []*model.Family{f}

	checkContains(t, renderFamilyStatement(p, f), "had two children.", append(redactedForms, " with ", "another relationship")...)
}

func TestRedactedChildrenSummary(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "John", PreferredFamiliarName: "John", Gender: model.GenderMale}
	wife := &model.Person{ID: "w", PreferredUniqueName: "Mary Brown", PreferredFamiliarName: "Mary", Gender: model.GenderFemale}
	named := &model.Person{ID: "c1", PreferredGivenName: "Ann", PreferredFamiliarName: "Ann", Gender: model.GenderFemale}
	named.BestBirthlikeEvent = &model.BirthEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1905)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: named}}
	f := marriedFamily("f", p, wife, model.Year(1903))
	f.Children = []*model.Person{named, redacted("c2", model.GenderMale), redacted("c3", model.GenderFemale)}
	p.Families = []*model.Family{f}

	checkContains(t, renderFamilyStatement(p, f), "They also had two other children.", redactedForms...)
}

func TestRedactedDeathInformant(t *testing.T) {
	p := &model.Person{ID: "p", PreferredUniqueName: "Ethel Sweetland", Gender: model.GenderFemale}
	son := redacted("s", model.GenderMale)
	p.Children = []*model.Person{son}
	death := &model.DeathEvent{
		GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1983, 2, 1)},
		GeneralIndividualEvent: model.GeneralIndividualEvent{
			Principal: p,
		},
	}
	death.Participants = []*model.EventParticipant{{Person: p, Role: model.EventRolePrincipal}, {Person: son, Role: model.EventRoleInformant}}

	enc := &md.Content{}
	fmtr := NewNarrativeTimelineEntryFormatter[md.Text](&model.POV{Person: p}, enc, logging.Default(), &TimelineNameChooser{}, false)
	checkContains(t, fmtr.Title(0, death), "The informant was her son.", redactedForms...)
}

func TestRedactedSpouseInTimeline(t *testing.T) {
	p := &model.Person{ID: "p", PreferredUniqueName: "John Sharp", Gender: model.GenderMale}
	s := redacted("s", model.GenderFemale)
	ev := &model.MarriageEvent{GeneralUnionEvent: model.GeneralUnionEvent{
		Husband: p, Wife: s,
		Participants: []*model.EventParticipant{{Person: p, Role: model.EventRoleHusband}, {Person: s, Role: model.EventRoleWife}},
	}}

	enc := &md.Content{}
	fmtr := NewNarrativeTimelineEntryFormatter[md.Text](&model.POV{Person: p}, enc, logging.Default(), &TimelineNameChooser{}, false)
	checkContains(t, fmtr.Title(0, ev), "Married.", redactedForms...)
}

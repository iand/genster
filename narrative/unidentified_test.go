package narrative

import (
	"testing"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// unidentified returns a person whose forename and surname are both unknown, named as the tree
// names such people.
func unidentified(id string, g model.Gender) *model.Person {
	return &model.Person{
		ID:                        id,
		Unidentified:              true,
		Gender:                    g,
		PreferredUniqueName:       "–?– –?–",
		PreferredFullName:         "an unidentified person",
		PreferredGivenName:        "unidentified",
		PreferredFamiliarName:     "unidentified",
		PreferredFamiliarFullName: "an unidentified person",
		PreferredFamilyName:       "–?–",
	}
}

var unidentifiedForms = []string{"–?–", "unidentified", "Unidentified"}

func TestUnidentifiedSpouseInMarriageStatement(t *testing.T) {
	husband := &model.Person{ID: "h", PreferredGivenName: "John", PreferredFamiliarName: "John", Gender: model.GenderMale}
	wife := unidentified("w", model.GenderFemale)
	m := &model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: model.BeforeYear(1805)}, GeneralUnionEvent: model.GeneralUnionEvent{Husband: husband, Wife: wife}}
	f := &model.Family{ID: "f", Bond: model.FamilyBondMarried, Father: husband, Mother: wife, BestStartDate: m.Date, BestStartEvent: m}
	husband.Families = []*model.Family{f}

	s := &FamilyStatement[md.Text]{Principal: husband, Family: f}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: husband}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

	checkContains(t, enc.String(), "He married a woman whose name is not known before 1805.", unidentifiedForms...)
}

func TestUnidentifiedParentInIntro(t *testing.T) {
	father := &model.Person{ID: "f", PreferredUniqueName: "Robert Paybody", PreferredFamiliarName: "Robert", Gender: model.GenderMale}
	mother := unidentified("m", model.GenderFemale)
	p := &model.Person{ID: "p", PreferredGivenName: "Robert", PreferredFamiliarName: "Robert", Gender: model.GenderMale, Father: father, Mother: mother}

	s := &IntroStatement[md.Text]{Principal: p}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

	checkContains(t, enc.String(), "Robert was the son of Robert Paybody and a woman whose name is not known.", unidentifiedForms...)
}

func TestUnidentifiedSpouseInMarriageSummary(t *testing.T) {
	p := &model.Person{ID: "p", PreferredFamiliarName: "Alice", Gender: model.GenderFemale}
	spouse := unidentified("s", model.GenderMale)
	m := &model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1764)}}
	p.Families = []*model.Family{{ID: "f", Bond: model.FamilyBondMarried, Father: spouse, Mother: p, BestStartDate: m.Date, BestStartEvent: m}}

	enc := &md.Content{}
	got := PersonMarriageSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("Alice"), false, true).String()

	checkContains(t, got, "Alice married a man whose name is not known in 1764.", unidentifiedForms...)
}

func TestUnidentifiedSpouseInFamilyEnd(t *testing.T) {
	husband := &model.Person{ID: "h", PreferredGivenName: "William", PreferredFamiliarName: "William", Gender: model.GenderMale}
	wife := unidentified("w", model.GenderFemale)
	wife.BestDeathlikeEvent = &model.DeathEvent{GeneralEvent: model.GeneralEvent{Date: model.BeforeYear(1798)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: wife}}
	f := &model.Family{ID: "f", Bond: model.FamilyBondMarried, Father: husband, Mother: wife, EndReason: model.FamilyEndReasonDeath, BestEndDate: model.BeforeYear(1798), BestEndEvent: wife.BestDeathlikeEvent}
	husband.Families = []*model.Family{f}

	s := &FamilyEndStatement[md.Text]{Principal: husband, Family: f}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: husband}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

	checkContains(t, enc.String(), "His wife died before 1798", unidentifiedForms...)
}

func TestUnidentifiedPersonOwnPage(t *testing.T) {
	p := unidentified("p", model.GenderFemale)

	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	(&IntroStatement[md.Text]{Principal: p}).RenderDetail(0, intro, enc, &DefaultNameChooser{})
	(&DeathStatement[md.Text]{Principal: p}).RenderDetail(1, intro, enc, &DefaultNameChooser{})
	got := enc.String()

	checkContains(t, got, "Nothing is known about her early life.", unidentifiedForms...)
	checkContains(t, got, "The time and place of her death is not known.")
}

func TestUnidentifiedSpouseInTimeline(t *testing.T) {
	p := &model.Person{ID: "p", PreferredUniqueName: "John Sharp", Gender: model.GenderMale}
	s := unidentified("s", model.GenderFemale)
	ev := &model.MarriageEvent{GeneralUnionEvent: model.GeneralUnionEvent{
		Husband: p, Wife: s,
		Participants: []*model.EventParticipant{{Person: p, Role: model.EventRoleHusband}, {Person: s, Role: model.EventRoleWife}},
	}}

	enc := &md.Content{}
	fmtr := NewNarrativeTimelineEntryFormatter(&model.POV{Person: p}, enc, logging.Default(), &TimelineNameChooser{}, false)
	checkContains(t, fmtr.Title(0, ev), "Married a woman whose name is not known.", unidentifiedForms...)
}

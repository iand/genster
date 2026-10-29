package narrative

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// child returns a child whose best birth-like event is a baptism or birth with the given date
// and place.
func child(id, name string, baptised bool, dt *model.Date, pl *model.Place) *model.Person {
	c := &model.Person{ID: id, PreferredGivenName: name, PreferredFamiliarName: name, Gender: model.GenderFemale}
	ge := model.GeneralEvent{Date: dt, Place: pl}
	gi := model.GeneralIndividualEvent{Principal: c}
	if baptised {
		c.BestBirthlikeEvent = &model.BaptismEvent{GeneralEvent: ge, GeneralIndividualEvent: gi}
	} else {
		c.BestBirthlikeEvent = &model.BirthEvent{GeneralEvent: ge, GeneralIndividualEvent: gi}
	}
	c.Timeline = []model.TimelineEvent{c.BestBirthlikeEvent}
	return c
}

func estimatedYearDate(y int) *model.Date {
	d := model.Year(y)
	d.Derivation = model.DateDerivationEstimated
	return d
}

func renderChildren(children ...*model.Person) string {
	p := &model.Person{ID: "p", PreferredGivenName: "Joseph", PreferredFamiliarName: "Joseph", Gender: model.GenderMale}
	wife := &model.Person{ID: "w", PreferredUniqueName: "Elizabeth Gooch", PreferredFamiliarName: "Elizabeth", Gender: model.GenderFemale}
	f := marriedFamily("f", p, wife, model.PreciseDate(1766, 7, 28))
	children[0].Gender = model.GenderMale // so the list is of children, not daughters
	f.Children = children
	p.Families = []*model.Family{f}
	return renderFamilyStatement(p, f)
}

// Shaped like I29: a place shared by most children is stated once before the list.
func TestChildListSharedPlace(t *testing.T) {
	fress := parish("Fressingfield", "Suffolk, England")
	got := renderChildren(
		child("c1", "Joseph", true, model.PreciseDate(1767, 5, 1), fress),
		child("c2", "Mary", true, model.PreciseDate(1769, 2, 4), fress),
		child("c3", "John", true, model.PreciseDate(1771, 3, 14), fress),
		child("c4", "Sarah", true, model.PreciseDate(1775, 3, 8), parish("Metfield", "Suffolk, England")),
	)

	checkContains(t, got, "They had four children, all baptised in the parish of Fressingfield in Suffolk, England (unless noted):–")
	checkContains(t, got, "Mary</em> baptised on 4 Feb 1769.")
	checkContains(t, got, "Metfield")
	if n := strings.Count(got, "Fressingfield"); n != 1 {
		t.Errorf("Fressingfield appears %d times, want 1:\n%s", n, got)
	}
}

// When every child shares the place, there is nothing to note.
func TestChildListPlaceSharedByAll(t *testing.T) {
	fress := parish("Fressingfield", "Suffolk, England")
	got := renderChildren(
		child("c1", "Joseph", true, model.PreciseDate(1767, 5, 1), fress),
		child("c2", "Mary", true, model.PreciseDate(1769, 2, 4), fress),
		child("c3", "John", true, model.PreciseDate(1771, 3, 14), fress),
	)

	checkContains(t, got, "They had three children, all baptised in the parish of Fressingfield in Suffolk, England:–", "unless noted")
}

// A note is needed when one of the shared details is not shared by every child, even when the
// other is.
func TestChildListSharedPlaceAllEstimatesSome(t *testing.T) {
	gn := parish("Greens Norton", "Northamptonshire, England")
	got := renderChildren(
		child("c1", "Susanna", false, estimatedYearDate(1800), gn),
		child("c2", "Abigale", false, estimatedYearDate(1801), gn),
		child("c3", "Henry", false, estimatedYearDate(1803), gn),
		child("c4", "John", false, model.PreciseDate(1811, 4, 2), gn),
	)

	checkContains(t, got, "They had four children, all born in the parish of Greens Norton in Northamptonshire, England, with birth years estimated (unless noted):–")
}

// A place shared by only some children is given on each of them.
func TestChildListPlaceNotShared(t *testing.T) {
	fress := parish("Fressingfield", "Suffolk, England")
	got := renderChildren(
		child("c1", "Joseph", true, model.PreciseDate(1767, 5, 1), fress),
		child("c2", "Mary", true, model.PreciseDate(1769, 2, 4), fress),
		child("c3", "John", true, model.PreciseDate(1771, 3, 14), parish("Metfield", "Suffolk, England")),
		child("c4", "Sarah", true, model.PreciseDate(1775, 3, 8), parish("Mendham", "Suffolk, England")),
	)

	checkContains(t, got, "They had four children:–", "unless noted")
}

// Shaped like I2076: estimated birth years shared by most children are noted once.
func TestChildListEstimatedBirths(t *testing.T) {
	got := renderChildren(
		child("c1", "Susanna", false, estimatedYearDate(1800), nil),
		child("c2", "Abigale", false, estimatedYearDate(1801), nil),
		child("c3", "Henry", false, estimatedYearDate(1803), nil),
		child("c4", "John", false, model.PreciseDate(1811, 4, 2), nil),
	)

	checkContains(t, got, "They had four children, with birth years estimated (unless noted):–", "probably")
	checkContains(t, got, "Susanna</em> born in 1800.")
	checkContains(t, got, "John</em> born on 2 Apr 1811.")
}

// A shared place and shared estimated birth years are noted together.
func TestChildListSharedPlaceAndEstimatedBirths(t *testing.T) {
	gn := parish("Greens Norton", "Northamptonshire, England")
	got := renderChildren(
		child("c1", "Susanna", false, estimatedYearDate(1800), gn),
		child("c2", "Abigale", false, estimatedYearDate(1801), gn),
		child("c3", "Henry", false, estimatedYearDate(1803), gn),
	)

	checkContains(t, got, "They had three children, all born in the parish of Greens Norton in Northamptonshire, England, with birth years estimated:–", "probably", "unless noted")
	checkContains(t, got, "Susanna</em> born in 1800.")
}

// Shaped like I1455: a single child listed in a bullet is not also named in the sentence that
// introduces it.
func TestSingleChildNotNamedTwice(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "John", PreferredFamiliarName: "John", Gender: model.GenderMale}
	p.BestBirthlikeEvent = &model.BaptismEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1635, 5, 24)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
	c := child("c", "Samuel", false, model.Year(1675), nil)
	c.Gender = model.GenderMale
	f := &model.Family{ID: "f", Bond: model.FamilyBondUnknown, Father: p, Mother: model.UnknownPerson(), Children: []*model.Person{c}}
	p.Families = []*model.Family{f}

	got := renderFamilyStatement(p, f)
	if n := strings.Count(got, "Samuel"); n != 1 {
		t.Errorf("Samuel appears %d times, want 1:\n%s", n, got)
	}
	checkContains(t, got, "had a child:–")
}

// Shaped like I92: the time of a birth relative to the previous statement does not repeat the
// date that the clause describing the birth gives.
func TestSingleChildDateNotRepeated(t *testing.T) {
	p := &model.Person{ID: "p", PreferredGivenName: "Frances", PreferredFamiliarName: "Frances", Gender: model.GenderFemale}
	c := child("c", "Eliza", true, model.PreciseDate(1837, 7, 23), nil)
	c.PreferredFullName = "Eliza Snowling"
	f := &model.Family{ID: "f", Bond: model.FamilyBondUnknown, Father: model.UnknownPerson(), Mother: p, Children: []*model.Person{c}}
	p.Families = []*model.Family{f}

	s := &FamilyStatement[md.Text]{Principal: p, Family: f}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}, LastIntroDate: model.PreciseDate(1836, 7, 5)}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

	checkContains(t, enc.String(), "baptised on 23 Jul 1837", "on 23 July")
}

// A time relative to the previous statement that includes the date separates them with a space.
func TestRelativeTimeSpacing(t *testing.T) {
	for seq := range 4 {
		intro := &IntroGenerator[md.Text]{POV: &model.POV{}, LastIntroDate: model.PreciseDate(1837, 2, 1)}
		got := intro.RelativeTime(seq, model.PreciseDate(1837, 7, 23), true)
		if strings.Contains(got, "yearon") || strings.Contains(got, "lateron") || strings.Contains(got, "afteron") {
			t.Errorf("seq %d: missing space before the date: %q", seq, got)
		}
		if !strings.Contains(got, "on 23 July") {
			t.Errorf("seq %d: date missing: %q", seq, got)
		}
	}
}

// Shaped like I92: a single child who died young has the date and place of death in the sentence
// that says so, rather than in a second sentence.
func TestSingleChildDiedYoungOneSentence(t *testing.T) {
	newFamily := func(deathlike func(c *model.Person) model.IndividualTimelineEvent) (*model.Person, *model.Family) {
		p := &model.Person{ID: "p", PreferredGivenName: "Frances", PreferredFamiliarName: "Frances", Gender: model.GenderFemale}
		c := child("c", "Eliza", true, model.PreciseDate(1837, 7, 23), nil)
		c.PreferredFullName = "Eliza Snowling"
		c.BestDeathlikeEvent = deathlike(c)
		c.Timeline = append(c.Timeline, c.BestDeathlikeEvent)
		f := &model.Family{ID: "f", Bond: model.FamilyBondUnknown, Father: model.UnknownPerson(), Mother: p, Children: []*model.Person{c}}
		p.Families = []*model.Family{f}
		return p, f
	}
	fress := parish("Fressingfield", "Suffolk, England")

	t.Run("death", func(t *testing.T) {
		p, f := newFamily(func(c *model.Person) model.IndividualTimelineEvent {
			return &model.DeathEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1837, 10, 17), Place: fress}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c}}
		})
		checkContains(t, renderFamilyStatement(p, f), "Eliza</em> died in infancy, on 17 Oct 1837 in the parish of Fressingfield in Suffolk, England.", "She died")
	})

	t.Run("burial", func(t *testing.T) {
		p, f := newFamily(func(c *model.Person) model.IndividualTimelineEvent {
			return &model.BurialEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1837, 10, 19), Place: fress}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c}}
		})
		checkContains(t, renderFamilyStatement(p, f), "Eliza</em> died in infancy. She was buried on 19 Oct 1837", "died in infancy, on")
	})
}

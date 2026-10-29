package infer

import (
	"testing"

	"github.com/iand/genster/model"
)

func birthOf(p *model.Person, d *model.Date) *model.BirthEvent {
	return &model.BirthEvent{GeneralEvent: model.GeneralEvent{Date: d}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
}

func baptismOf(p *model.Person, d *model.Date) *model.BaptismEvent {
	return &model.BaptismEvent{GeneralEvent: model.GeneralEvent{Date: d}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
}

func deathOf(p *model.Person, d *model.Date) *model.DeathEvent {
	return &model.DeathEvent{GeneralEvent: model.GeneralEvent{Date: d}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
}

func burialOf(p *model.Person, d *model.Date) *model.BurialEvent {
	return &model.BurialEvent{GeneralEvent: model.GeneralEvent{Date: d}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
}

// person returns a person with the given name and gender.
func person(id, name string, g model.Gender) *model.Person {
	return &model.Person{ID: id, PreferredFullName: name, Gender: g}
}

// withBirth makes ev the person's best birth-like event and adds it to their timeline.
func withBirth(p *model.Person, ev model.IndividualTimelineEvent) *model.Person {
	p.BestBirthlikeEvent = ev
	p.Timeline = append(p.Timeline, ev)
	return p
}

// withDeath makes ev the person's best death-like event and adds it to their timeline.
func withDeath(p *model.Person, ev model.IndividualTimelineEvent) *model.Person {
	p.BestDeathlikeEvent = ev
	p.Timeline = append(p.Timeline, ev)
	return p
}

// family links parents and children as the tree does.
func family(id string, father, mother *model.Person, children ...*model.Person) *model.Family {
	f := &model.Family{ID: id, Father: father, Mother: mother, Children: children}
	if father == nil {
		f.Father = model.UnknownPerson()
	} else {
		father.Families = append(father.Families, f)
	}
	if mother == nil {
		f.Mother = model.UnknownPerson()
	} else {
		mother.Families = append(mother.Families, f)
	}
	for _, c := range children {
		c.Father = f.Father
		c.Mother = f.Mother
	}
	return f
}

// anomalyTexts returns the texts of the anomalies on p with the given context.
func anomalyTexts(p *model.Person, context string) []string {
	var texts []string
	for _, a := range p.Anomalies {
		if a.Context == context {
			if a.Category != model.AnomalyCategoryEvent {
				continue
			}
			texts = append(texts, a.Text)
		}
	}
	return texts
}

func checkAnomalies(t *testing.T, p *model.Person, context string, want ...string) {
	t.Helper()
	got := anomalyTexts(p, context)
	if len(got) != len(want) {
		t.Fatalf("%s: got %d %q anomalies %q, want %d %q", p.ID, len(got), context, got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: anomaly %d: got %q, want %q", p.ID, i, got[i], want[i])
		}
	}
}

func TestCheckChronologyBurialBeforeDeath(t *testing.T) {
	testCases := []struct {
		name     string
		death    *model.Date
		burials  []*model.Date
		wantText string // empty means no burial-before-death anomaly
	}{
		{
			name:     "burial before death",
			death:    model.PreciseDate(1900, 6, 17),
			burials:  []*model.Date{model.PreciseDate(1900, 1, 31)},
			wantText: "Burial dated 31 Jan 1900 is before death dated 17 Jun 1900",
		},
		{
			name:     "one burial before death and one after",
			death:    model.PreciseDate(1900, 6, 17),
			burials:  []*model.Date{model.PreciseDate(1900, 1, 31), model.PreciseDate(1900, 6, 20)},
			wantText: "Burial dated 31 Jan 1900 is before death dated 17 Jun 1900",
		},
		{
			name:    "burial after death",
			death:   model.PreciseDate(1900, 6, 17),
			burials: []*model.Date{model.PreciseDate(1900, 6, 20)},
		},
		{
			name:    "burial on day of death",
			death:   model.PreciseDate(1913, 1, 12),
			burials: []*model.Date{model.PreciseDate(1913, 1, 12)},
		},
		{
			name:    "year-only burial overlapping death",
			death:   model.PreciseDate(1900, 1, 30),
			burials: []*model.Date{model.Year(1900)},
		},
		{
			name:    "unknown burial date",
			death:   model.PreciseDate(1900, 6, 17),
			burials: []*model.Date{model.UnknownDate()},
		},
		{
			name:    "unknown death date",
			death:   model.UnknownDate(),
			burials: []*model.Date{model.PreciseDate(1900, 1, 31)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "p1"}
			withDeath(p, deathOf(p, tc.death))
			for _, d := range tc.burials {
				p.Timeline = append(p.Timeline, burialOf(p, d))
			}

			CheckChronology(p)

			if tc.wantText == "" {
				checkAnomalies(t, p, "Burial before death")
				return
			}
			checkAnomalies(t, p, "Burial before death", tc.wantText)
		})
	}
}

// Shaped like I724: censuses recorded fifteen and twenty-five years after burial.
func TestCheckChronologyCensusAfterBurial(t *testing.T) {
	p := person("p", "Anne Hinson", model.GenderFemale)
	withBirth(p, baptismOf(p, model.PreciseDate(1766, 8, 20)))
	withDeath(p, burialOf(p, model.PreciseDate(1826, 12, 27)))
	for _, d := range []*model.Date{model.PreciseDate(1841, 6, 6), model.PreciseDate(1851, 3, 30)} {
		p.Timeline = append(p.Timeline, &model.CensusEvent{
			GeneralEvent: model.GeneralEvent{Date: d},
			Entries:      []*model.CensusEntry{{Principal: p}},
		})
	}

	CheckChronology(p)

	checkAnomalies(t, p, "Event after death",
		"Census dated 6 Jun 1841 is more than a year after burial dated 27 Dec 1826",
		"Census dated 30 Mar 1851 is more than a year after burial dated 27 Dec 1826",
	)
}

// Shaped like I2160: a baptism dated forty years after death, with a birth that agrees with
// the death. The best birth-like event is the baptism, as genster chooses it for its precision.
func TestCheckChronologyBaptismAfterDeath(t *testing.T) {
	p := person("p", "Marte Halvorsdatter", model.GenderFemale)
	p.Timeline = append(p.Timeline, birthOf(p, model.Year(1784)))
	withBirth(p, baptismOf(p, model.PreciseDate(1874, 12, 5)))
	withDeath(p, deathOf(p, model.Year(1834)))

	CheckChronology(p)

	checkAnomalies(t, p, "Event after death", "Baptism dated 5 Dec 1874 is more than a year after death dated 1834")
}

// Shaped like I2160: ages are measured from a firm birth event rather than from a baptism
// chosen as the best birth-like event, so a mistyped baptism gives one anomaly, not one for
// every parent and child.
func TestCheckChronologyAgesUseBirthOverBaptism(t *testing.T) {
	father := person("f", "Halvor Halvorsen", model.GenderMale)
	withBirth(father, birthOf(father, model.Year(1759)))
	mother := person("m", "Berte Nilsdatter", model.GenderFemale)
	withBirth(mother, birthOf(mother, model.Year(1763)))

	p := person("p", "Marte Halvorsdatter", model.GenderFemale)
	p.Timeline = append(p.Timeline, birthOf(p, model.Year(1784)))
	withBirth(p, baptismOf(p, model.PreciseDate(1874, 12, 5)))
	withDeath(p, deathOf(p, model.Year(1834)))
	family("fp", father, mother, p)

	child := person("c", "Hans Larsen", model.GenderMale)
	withBirth(child, birthOf(child, model.PreciseDate(1815, 8, 28)))
	family("fc", nil, p, child)

	for _, q := range []*model.Person{father, mother, p, child} {
		CheckChronology(q)
	}

	checkAnomalies(t, p, "Event after death", "Baptism dated 5 Dec 1874 is more than a year after death dated 1834")
	for _, q := range []*model.Person{father, mother, p, child} {
		checkAnomalies(t, q, "Parent too young at child's birth")
		checkAnomalies(t, q, "Parent too old at child's birth")
	}
}

// Shaped like I2837: a possible birth dated after burial.
func TestCheckChronologyPossibleBirthAfterBurial(t *testing.T) {
	p := person("p", "Joanna Smith", model.GenderFemale)
	p.Timeline = append(p.Timeline, &model.PossibleBirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.Year(1782)},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	})
	withDeath(p, burialOf(p, model.PreciseDate(1730, 12, 28)))

	CheckChronology(p)

	checkAnomalies(t, p, "Event after death", "Possible birth dated 1782 is more than a year after burial dated 28 Dec 1730")
}

// Events within a year of death, after-death events such as probate, and qualified dates that
// may fall before death are not anomalies. An occupation can be recorded after death, as a
// child's marriage register records the father's occupation.
func TestCheckChronologyEventsNearDeathAreNotAnomalies(t *testing.T) {
	p := person("p", "John Smith", model.GenderMale)
	withBirth(p, birthOf(p, model.Year(1800)))
	withDeath(p, deathOf(p, model.PreciseDate(1860, 3, 1)))
	p.Timeline = append(p.Timeline,
		&model.OccupationEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1861, 1, 10)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}},
		&model.OccupationEvent{GeneralEvent: model.GeneralEvent{Date: model.BeforeYear(1870)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}},
		&model.ProbateEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1865)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}},
		&model.OccupationEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1872, 10, 28)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}},
		burialOf(p, model.PreciseDate(1860, 3, 5)),
	)

	CheckChronology(p)

	checkAnomalies(t, p, "Event after death")
}

// Shaped like I594: a marriage dated two years after baptism.
func TestCheckChronologyMarriageTooYoung(t *testing.T) {
	husband := person("h", "Robert Doughty", model.GenderMale)
	withBirth(husband, baptismOf(husband, model.PreciseDate(1807, 7, 13)))
	wife := person("w", "Sarah Whitney", model.GenderFemale)
	withBirth(wife, baptismOf(wife, model.PreciseDate(1790, 1, 1)))
	f := family("f", husband, wife)
	f.Bond = model.FamilyBondMarried
	f.BestStartEvent = &model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1809, 9, 1)}, GeneralUnionEvent: model.GeneralUnionEvent{Husband: husband, Wife: wife}}
	f.BestStartDate = f.BestStartEvent.GetDate()

	CheckChronology(husband)
	CheckChronology(wife)

	checkAnomalies(t, husband, "Marriage at a young age", "Marriage to Sarah Whitney dated 1 Sep 1809 is 2 years after baptism dated 13 Jul 1807")
	checkAnomalies(t, wife, "Marriage at a young age")
}

// A family start date inferred from a child's birth is not a marriage date.
func TestCheckChronologyInferredFamilyStartIsNotMarriage(t *testing.T) {
	husband := person("h", "Richard Lovelace", model.GenderMale)
	withBirth(husband, baptismOf(husband, model.Year(1540)))
	wife := person("w", "Alice Croade", model.GenderFemale)
	withBirth(wife, baptismOf(wife, model.Year(1560)))
	f := family("f", husband, wife)
	f.Bond = model.FamilyBondMarried
	f.BestStartDate = model.BeforeYear(1570)

	CheckChronology(wife)

	checkAnomalies(t, wife, "Marriage at a young age")
}

// Shaped like I2757 and I1532: a child baptised when the mother was ten, and a child baptised
// 228 years after the father.
func TestCheckChronologyParentAgeAtChildsBirth(t *testing.T) {
	t.Run("mother too young", func(t *testing.T) {
		mother := person("m", "Alice Croade", model.GenderFemale)
		withBirth(mother, baptismOf(mother, model.Year(1560)))
		child := person("c", "Agnes Lovelace", model.GenderFemale)
		withBirth(child, baptismOf(child, model.Year(1570)))
		family("f", nil, mother, child)

		CheckChronology(mother)

		want := "Baptism of Agnes Lovelace dated 1570 is 10 years after the baptism of Alice Croade dated 1560"
		checkAnomalies(t, mother, "Parent too young at child's birth", want)
		checkAnomalies(t, child, "Parent too young at child's birth", want)
	})

	t.Run("father too old", func(t *testing.T) {
		father := person("f", "Robert Paybody", model.GenderMale)
		withBirth(father, baptismOf(father, model.PreciseDate(1632, 4, 29)))
		child := person("c", "Samuel Paybody", model.GenderMale)
		withBirth(child, baptismOf(child, model.PreciseDate(1860, 12, 27)))
		family("f", father, nil, child)

		CheckChronology(father)

		want := "Baptism of Samuel Paybody dated 27 Dec 1860 is 228 years after the baptism of Robert Paybody dated 29 Apr 1632"
		checkAnomalies(t, father, "Parent too old at child's birth", want)
		checkAnomalies(t, child, "Parent too old at child's birth", want)
	})

	t.Run("mother over 55", func(t *testing.T) {
		mother := person("m", "Mary Brown", model.GenderFemale)
		withBirth(mother, birthOf(mother, model.Year(1800)))
		child := person("c", "Ann Brown", model.GenderFemale)
		withBirth(child, birthOf(child, model.Year(1858)))
		family("f", nil, mother, child)

		CheckChronology(mother)

		checkAnomalies(t, mother, "Parent too old at child's birth", "Birth of Ann Brown dated 1858 is 58 years after the birth of Mary Brown dated 1800")
	})

	t.Run("father at 70 and mother at 45 are plausible", func(t *testing.T) {
		father := person("f", "John Brown", model.GenderMale)
		withBirth(father, birthOf(father, model.Year(1780)))
		mother := person("m", "Mary Brown", model.GenderFemale)
		withBirth(mother, birthOf(mother, model.Year(1805)))
		child := person("c", "Ann Brown", model.GenderFemale)
		withBirth(child, birthOf(child, model.Year(1850)))
		family("f", father, mother, child)

		CheckChronology(father)
		CheckChronology(mother)

		for _, p := range []*model.Person{father, mother, child} {
			checkAnomalies(t, p, "Parent too young at child's birth")
			checkAnomalies(t, p, "Parent too old at child's birth")
		}
	})

	t.Run("estimated dates are not checked", func(t *testing.T) {
		mother := person("m", "Mary Brown", model.GenderFemale)
		est := model.Year(1840)
		est.Derivation = model.DateDerivationEstimated
		withBirth(mother, birthOf(mother, est))
		child := person("c", "Ann Brown", model.GenderFemale)
		withBirth(child, birthOf(child, model.Year(1845)))
		family("f", nil, mother, child)

		CheckChronology(mother)

		checkAnomalies(t, mother, "Parent too young at child's birth")
	})
}

func TestCheckChronologyChildBornAfterParentsDeath(t *testing.T) {
	t.Run("after mother's death", func(t *testing.T) {
		mother := person("m", "Mary Brown", model.GenderFemale)
		withBirth(mother, birthOf(mother, model.Year(1800)))
		withDeath(mother, deathOf(mother, model.PreciseDate(1830, 5, 1)))
		child := person("c", "Ann Brown", model.GenderFemale)
		withBirth(child, birthOf(child, model.PreciseDate(1831, 2, 1)))
		family("f", nil, mother, child)

		CheckChronology(mother)

		want := "Birth of Ann Brown dated 1 Feb 1831 is after the death of Mary Brown dated 1 May 1830"
		checkAnomalies(t, mother, "Child born after parent's death", want)
		checkAnomalies(t, child, "Child born after parent's death", want)
	})

	t.Run("posthumous child of father", func(t *testing.T) {
		father := person("f", "John Brown", model.GenderMale)
		withBirth(father, birthOf(father, model.Year(1800)))
		withDeath(father, deathOf(father, model.PreciseDate(1830, 5, 1)))
		child := person("c", "Ann Brown", model.GenderFemale)
		withBirth(child, birthOf(child, model.PreciseDate(1831, 2, 1)))
		family("f", father, nil, child)

		CheckChronology(father)

		checkAnomalies(t, father, "Child born after parent's death")
	})

	t.Run("more than a year after father's death", func(t *testing.T) {
		father := person("f", "John Brown", model.GenderMale)
		withBirth(father, birthOf(father, model.Year(1800)))
		withDeath(father, deathOf(father, model.PreciseDate(1830, 5, 1)))
		child := person("c", "Ann Brown", model.GenderFemale)
		withBirth(child, birthOf(child, model.PreciseDate(1832, 2, 1)))
		family("f", father, nil, child)

		CheckChronology(father)

		checkAnomalies(t, father, "Child born after parent's death", "Birth of Ann Brown dated 1 Feb 1832 is more than a year after the death of John Brown dated 1 May 1830")
	})

	t.Run("baptism after mother's death is plausible", func(t *testing.T) {
		mother := person("m", "Mary Brown", model.GenderFemale)
		withBirth(mother, birthOf(mother, model.Year(1800)))
		withDeath(mother, deathOf(mother, model.PreciseDate(1830, 5, 1)))
		child := person("c", "Ann Brown", model.GenderFemale)
		withBirth(child, baptismOf(child, model.PreciseDate(1830, 6, 1)))
		family("f", nil, mother, child)

		CheckChronology(mother)

		checkAnomalies(t, mother, "Child born after parent's death")
	})
}

// A family listed twice on a parent, as the loaders can produce before duplicates are removed,
// gives one anomaly.
func TestCheckChronologyDuplicateFamilyGivesOneAnomaly(t *testing.T) {
	father := person("f", "Robert Paybody", model.GenderMale)
	withBirth(father, baptismOf(father, model.Year(1632)))
	child := person("c", "Samuel Paybody", model.GenderMale)
	withBirth(child, baptismOf(child, model.Year(1860)))
	f := family("f", father, nil, child)
	father.Families = append(father.Families, f)

	CheckChronology(father)

	checkAnomalies(t, father, "Parent too old at child's birth", "Baptism of Samuel Paybody dated 1860 is 228 years after the baptism of Robert Paybody dated 1632")
}

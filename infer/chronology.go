package infer

import (
	"fmt"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

// Limits used by CheckChronology. An event outside them is implausible rather than unusual,
// so it points to an error in the data.
const (
	minMarriageAge         = 12  // youngest plausible age at marriage
	minParentAge           = 12  // youngest plausible age of a parent at a child's birth
	maxMotherAge           = 55  // oldest plausible age of a mother at a child's birth
	maxFatherAge           = 80  // oldest plausible age of a father, or a parent of unknown gender
	daysAllowedAfterDeath  = 365 // own events dated within this many days after death are allowed
	daysAllowedAfterFather = 365 // a child born within this many days after the father's death is allowed
)

// CheckChronology records anomalies for events in a person's life whose dates cannot all be
// right, so the data can be corrected. Each anomaly names the event most likely to be wrong.
// Anomalies about a parent and child are recorded on both. It expects best birth and death
// events and family start events to have been chosen.
func CheckChronology(p *model.Person) {
	checkEventsAfterDeath(p)
	checkBurialBeforeDeath(p)
	checkMarriageAge(p)
	checkChildren(p)
}

// checkEventsAfterDeath flags events of the person's own life dated more than a year after
// their death or burial.
func checkEventsAfterDeath(p *model.Person) {
	dev := p.BestDeathlikeEvent
	if dev == nil || !isCheckable(dev) {
		return
	}

	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) || !isCheckable(ev) || canFollowDeath(ev) {
			continue
		}
		if dev.GetDate().DefinitelyBeforeBy(ev.GetDate(), daysAllowedAfterDeath) {
			addAnomaly(p, "Event after death", fmt.Sprintf("%s dated %s is more than a year after %s dated %s",
				text.UpperFirst(ev.Type()), ev.GetDate().String(), dev.Type(), dev.GetDate().String()))
		}
	}
}

// canFollowDeath reports whether an event can be dated after the death of the person it
// involves. An occupation is recorded after death when a child's marriage register gives the
// father's occupation.
func canFollowDeath(ev model.TimelineEvent) bool {
	switch ev.(type) {
	case *model.DeathEvent, *model.PossibleDeathEvent, *model.BurialEvent, *model.CremationEvent,
		*model.MemorialEvent, *model.InquestEvent, *model.ProbateEvent, *model.WillEvent,
		*model.OccupationEvent:
		return true
	}
	return false
}

func checkBurialBeforeDeath(p *model.Person) {
	var deaths, burials []model.TimelineEvent
	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) {
			continue
		}
		switch ev.(type) {
		case *model.DeathEvent:
			deaths = append(deaths, ev)
		case *model.BurialEvent:
			burials = append(burials, ev)
		}
	}

	for _, bev := range burials {
		for _, dev := range deaths {
			if bev.GetDate().DefinitelyBefore(dev.GetDate()) {
				addAnomaly(p, "Burial before death", "Burial dated "+bev.GetDate().String()+" is before death dated "+dev.GetDate().String())
				break
			}
		}
	}
}

// checkMarriageAge flags a union event dated when the person was too young to marry. Family
// start dates inferred from other events are not checked.
func checkMarriageAge(p *model.Person) {
	bev := birthForAge(p)
	if bev == nil || !isFirm(bev) {
		return
	}

	for _, f := range uniqueFamilies(p) {
		uev, ok := f.BestStartEvent.(model.UnionTimelineEvent)
		if !ok || !isFirm(uev) {
			continue
		}
		years, ok := yearsBetween(bev, uev)
		if !ok || years >= minMarriageAge {
			continue
		}
		what := text.UpperFirst(uev.Type())
		if other := f.OtherParent(p); !other.IsUnknown() {
			what += " to " + other.PreferredFullName
		}
		addAnomaly(p, "Marriage at a young age", fmt.Sprintf("%s dated %s is %s %s dated %s",
			what, uev.GetDate().String(), relativeYears(years), bev.Type(), bev.GetDate().String()))
	}
}

// checkChildren flags children born when the person was too young or too old to be their
// parent, or after the person's death. The anomaly is recorded on the parent and the child.
func checkChildren(p *model.Person) {
	seen := make(map[string]bool)
	for _, f := range uniqueFamilies(p) {
		for _, c := range f.Children {
			if seen[c.ID] {
				continue
			}
			seen[c.ID] = true
			checkParentAge(p, c)
			checkBornAfterParentsDeath(p, c)
		}
	}
}

func checkParentAge(p, c *model.Person) {
	pev, cev := birthForAge(p), birthForAge(c)
	if pev == nil || cev == nil || !isFirm(pev) || !isFirm(cev) {
		return
	}
	years, ok := yearsBetween(pev, cev)
	if !ok {
		return
	}

	maxAge := maxFatherAge
	if p.Gender.IsFemale() {
		maxAge = maxMotherAge
	}

	var context string
	switch {
	case years < minParentAge:
		context = "Parent too young at child's birth"
	case years > maxAge:
		context = "Parent too old at child's birth"
	default:
		return
	}

	txt := fmt.Sprintf("%s of %s dated %s is %s the %s of %s dated %s",
		text.UpperFirst(cev.Type()), c.PreferredFullName, cev.GetDate().String(), relativeYears(years),
		pev.Type(), p.PreferredFullName, pev.GetDate().String())
	addAnomaly(p, context, txt)
	addAnomaly(c, context, txt)
}

// checkBornAfterParentsDeath flags a child's birth after the mother's death or more than a year
// after the father's death. A baptism is not checked since it can follow a parent's death.
func checkBornAfterParentsDeath(p, c *model.Person) {
	cev, ok := c.BestBirthlikeEvent.(*model.BirthEvent)
	if !ok || !isCheckable(cev) {
		return
	}
	dev := p.BestDeathlikeEvent
	if dev == nil || !isCheckable(dev) {
		return
	}

	var txt string
	if p.Gender.IsFemale() {
		if !dev.GetDate().DefinitelyBefore(cev.GetDate()) {
			return
		}
		txt = "is after the"
	} else {
		if !dev.GetDate().DefinitelyBeforeBy(cev.GetDate(), daysAllowedAfterFather) {
			return
		}
		txt = "is more than a year after the"
	}

	txt = fmt.Sprintf("Birth of %s dated %s %s %s of %s dated %s",
		c.PreferredFullName, cev.GetDate().String(), txt, dev.Type(), p.PreferredFullName, dev.GetDate().String())
	addAnomaly(p, "Child born after parent's death", txt)
	addAnomaly(c, "Child born after parent's death", txt)
}

// birthForAge returns the event that ages are measured from: a firm birth event when the person
// has one, since a baptism chosen as the best birth-like event for its precision may be the
// date that is wrong, otherwise the best birth-like event.
func birthForAge(p *model.Person) model.IndividualTimelineEvent {
	for _, ev := range p.Timeline {
		if bev, ok := ev.(*model.BirthEvent); ok && bev.DirectlyInvolves(p) && isFirm(bev) {
			return bev
		}
	}
	return p.BestBirthlikeEvent
}

func addAnomaly(p *model.Person, context, txt string) {
	p.Anomalies = append(p.Anomalies, &model.Anomaly{
		Category: model.AnomalyCategoryEvent,
		Text:     txt,
		Context:  context,
	})
}

// isCheckable reports whether an event has a date from a source, rather than one that was
// inferred, estimated or calculated.
func isCheckable(ev model.TimelineEvent) bool {
	return !ev.IsInferred() && !ev.GetDate().IsUnknown() && ev.GetDate().Derivation == model.DateDerivationStandard
}

// isFirm reports whether an event has a firm date from a source, precise enough to compare
// years.
func isFirm(ev model.TimelineEvent) bool {
	return !ev.IsInferred() && ev.GetDate().IsFirm()
}

// yearsBetween returns the number of calendar years from the year of a to the year of b.
func yearsBetween(a, b model.TimelineEvent) (int, bool) {
	ay, ok := a.GetDate().Year()
	if !ok {
		return 0, false
	}
	by, ok := b.GetDate().Year()
	if !ok {
		return 0, false
	}
	return by - ay, true
}

// relativeYears describes a number of years as "10 years after", "3 years before" or "in the
// same year as".
func relativeYears(years int) string {
	switch {
	case years == 0:
		return "in the same year as"
	case years == 1:
		return "1 year after"
	case years == -1:
		return "1 year before"
	case years < 0:
		return fmt.Sprintf("%d years before", -years)
	default:
		return fmt.Sprintf("%d years after", years)
	}
}

// uniqueFamilies returns the families in which the person is a parent, without the duplicates
// the loaders can produce.
func uniqueFamilies(p *model.Person) []*model.Family {
	seen := make(map[string]bool)
	var fams []*model.Family
	for _, f := range p.Families {
		if f == nil || seen[f.ID] {
			continue
		}
		seen[f.ID] = true
		fams = append(fams, f)
	}
	return fams
}

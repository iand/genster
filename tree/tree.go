package tree

import (
	"container/heap"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/iand/genster/identifier"
	"github.com/iand/genster/infer"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/place"
	"golang.org/x/exp/slog"
)

var _ = slog.Debug

type Tree struct {
	IdentityMap *IdentityMap
	Gazeteer    *Gazeteer
	Overrides   *Annotations
	People      map[string]*model.Person
	Sources     map[string]*model.Source
	Places      map[string]*model.Place
	Families    map[string]*model.Family
	KeyPerson   *model.Person
}

func NewTree(m *IdentityMap, g *Gazeteer, o *Annotations) *Tree {
	return &Tree{
		IdentityMap: m,
		Gazeteer:    g,
		Overrides:   o,
		People:      make(map[string]*model.Person),
		Sources:     make(map[string]*model.Source),
		Places:      make(map[string]*model.Place),
		Families:    make(map[string]*model.Family),
	}
}

func (t *Tree) GetPerson(id string) (*model.Person, bool) {
	p, ok := t.People[id]
	return p, ok
}

func (t *Tree) FindPerson(scope string, sid string) *model.Person {
	id := t.IdentityMap.ID(scope, sid)
	p, ok := t.People[id]
	if !ok {
		p = &model.Person{
			ID: id,
		}
		p.BestBirthlikeEvent = &model.BirthEvent{
			GeneralEvent: model.GeneralEvent{Date: model.UnknownDate()},
			GeneralIndividualEvent: model.GeneralIndividualEvent{
				Principal: p,
			},
		}

		t.People[id] = p
	}
	return p
}

func (t *Tree) GetSource(id string) (*model.Source, bool) {
	so, ok := t.Sources[id]
	return so, ok
}

func (t *Tree) FindSource(scope string, sid string) *model.Source {
	id := t.IdentityMap.ID(scope, sid)
	so, ok := t.Sources[id]
	if !ok {
		so = &model.Source{
			ID: id,
		}
		t.Sources[id] = so
	}
	return so
}

func (t *Tree) GetPlace(id string) (*model.Place, bool) {
	p, ok := t.Places[id]
	return p, ok
}

func (t *Tree) FindPlaceUnstructured(name string, hints ...place.Hint) *model.Place {
	id := t.Gazeteer.ID(name, hints...)
	p, ok := t.Places[id]
	if !ok {

		pn := place.ClassifyName(name, hints...)
		cleanName := pn.Name

		p = &model.Place{
			ID:                  id,
			OriginalText:        name,
			Hints:               hints,
			PreferredName:       cleanName,
			PreferredUniqueName: cleanName,
			PreferredFullName:   cleanName,
			PreferredSortName:   cleanName,
			PlaceType:           model.PlaceTypeUnknown,
			Kind:                place.PlaceKindUnknown,
			CountryName:         place.UnknownPlaceName(),
			UKNationName:        place.UnknownPlaceName(),
		}

		if c, ok := pn.FindContainerKind(place.PlaceKindCountry); ok {
			p.CountryName = c
		}

		if c, ok := pn.FindContainerKind(place.PlaceKindUKNation); ok {
			p.UKNationName = c
		}

		logging.Debug("adding place", "name", name, "id", id, "country", p.CountryName.Name)

		t.Places[id] = p
	}

	return p
}

func (t *Tree) findPlaceFromGazeteer(name string, gp GazeteerPlace) *model.Place {
	p, ok := t.Places[gp.id]
	if !ok {
		p = &model.Place{
			ID: gp.id,
			// Page:                fmt.Sprintf(s.PlacePagePattern, gp.id),
			OriginalText:        name,
			PreferredName:       gp.name,
			PreferredUniqueName: gp.name,
			PreferredFullName:   gp.name,
			PreferredSortName:   gp.name,
			PlaceType:           model.PlaceTypeUnknown,
			Kind:                gp.kind,
		}

		if gp.parentID != "" {
			parentgp, ok := t.Gazeteer.LookupPlace(gp.parentID)
			if ok {
				parent := t.findPlaceFromGazeteer(parentgp.name, parentgp)
				if !parent.IsUnknown() {
					p.Parent = parent
					p.PreferredFullName = gp.name + ", " + parent.PreferredFullName
					p.PreferredUniqueName = gp.name + ", " + parent.PreferredUniqueName
				}
			} else {
				logging.Warn("parent id not found in gazeteer", "parent_id", gp.parentID, "id", gp.id)
			}
		}

		t.Places[gp.id] = p
	}
	return p
}

func (t *Tree) FindFamily(father *model.Person, mother *model.Person) *model.Family {
	fatherID := father.ID
	if father.IsUnknown() {
		fatherID = "unknown"
	}
	motherID := mother.ID
	if mother.IsUnknown() {
		motherID = "unknown"
	}

	id := identifier.New("family", fatherID, motherID)
	f, ok := t.Families[id]
	if !ok {
		f = t.newFamily(id)
		f.Father = father
		f.Mother = mother

		father.Families = append(father.Families, f)
		mother.Families = append(mother.Families, f)

	}
	return f
}

func (t *Tree) FindFamilyOneParent(parent *model.Person, child *model.Person) *model.Family {
	id := identifier.New("familyoneparent", parent.ID, child.ID)
	f, ok := t.Families[id]
	if !ok {
		f = t.newFamily(id)
		if parent.Gender.IsMale() {
			f.Father = parent
			f.Mother = model.UnknownPerson()
		} else {
			// unknown gender treated as mother for now
			f.Father = model.UnknownPerson()
			f.Mother = parent
		}
		parent.Families = append(parent.Families, f)
	}
	return f
}

func (s *Tree) newFamily(id string) *model.Family {
	f := &model.Family{
		ID: id,
		// Page:      fmt.Sprintf(s.FamilyPagePattern, id),
		Bond:      model.FamilyBondUnknown,
		EndReason: model.FamilyEndReasonUnknown,
	}
	s.Families[id] = f
	return f
}

func (t *Tree) Generate(redactLiving bool) error {
	// Apply any overrides first, they may be redacted after
	if t.Overrides != nil {
		for _, p := range t.People {
			t.Overrides.ApplyPerson(p)
		}
		for _, p := range t.Places {
			t.Overrides.ApplyPlace(p)
		}
	}

	// Add data to each person
	for _, p := range t.People {
		t.PropagateParents(p)
		t.AddFamilies(p)
	}

	// Need to make sure all parents and children are linked first
	// since expand timeline will add events to parents timelines
	for _, f := range t.Families {
		t.InferFamilyRelationships(f)
	}

	// Add data to each person
	for _, p := range t.People {
		t.SelectPersonBestBirthDeathEvents(p)
		t.RefinePersonNames(p)
		t.RefinePersonOccupations(p)
		t.ExpandPersonTimeline(p)
	}

	// Fill in gaps with inferences
	for _, p := range t.People {
		infer.InferPersonBirthEventDate(p)
		infer.InferPersonAliveOrDead(p, time.Now().Year())
		infer.InferPersonDeathEventDate(p)
		infer.InferPersonCauseOfDeath(p)
		infer.InferPersonGeneralFacts(p)
	}

	for _, f := range t.Families {
		t.InferFamilyStartEndDates(f)
	}

	t.BuildRelationsToKeyPerson()

	// Redact any personal information
	t.Redact(redactLiving)

	for _, p := range t.People {
		t.TrimPersonTimeline(p)
		p.RemoveDuplicateFamilies()
		p.RemoveDuplicateChildren()
		p.RemoveDuplicateSpouses()
		t.BuildOlb(p)

		// sort families by date
		sort.Slice(p.Families, func(a, b int) bool {
			return p.Families[a].BestStartDate.SortsBefore(p.Families[b].BestStartDate)
		})

	}

	for _, p := range t.Places {
		t.TrimPlaceTimeline(p)
	}
	for _, s := range t.Sources {
		t.TrimSourceTimeline(s)
	}
	return nil
}

func (t *Tree) PropagateParents(p *model.Person) error {
	if !p.Father.IsUnknown() {
		p.Father.Children = append(p.Father.Children, p)
	}
	if !p.Mother.IsUnknown() {
		p.Mother.Children = append(p.Mother.Children, p)
	}
	return nil
}

func (t *Tree) AddFamilies(p *model.Person) error {
	var parentFamily *model.Family
	if p.Father.IsUnknown() {
		if p.Mother.IsUnknown() {
			// no known mother or father
			return nil
		} else {
			parentFamily = t.FindFamilyOneParent(p.Mother, p)
		}
	} else {
		if p.Mother.IsUnknown() {
			parentFamily = t.FindFamilyOneParent(p.Father, p)
		} else {
			parentFamily = t.FindFamily(p.Father, p.Mother)
		}
	}
	parentFamily.Children = append(parentFamily.Children, p)

	sortMaleFemale := func(p1 *model.Person, p2 *model.Person) (*model.Person, *model.Person, bool) {
		if p1.Gender == model.GenderMale && p2.Gender == model.GenderFemale {
			return p1, p2, true
		}
		if p1.Gender == model.GenderFemale && p2.Gender == model.GenderMale {
			return p2, p1, true
		}

		return p1, p2, false
	}

	addMarriage := func(t *Tree, ev model.PartyTimelineEvent) {
		p1 := ev.GetParty1()
		p2 := ev.GetParty2()
		if p1.IsUnknown() || p2.IsUnknown() {
			return
		}
		father, mother, ok := sortMaleFemale(p1, p2)
		if !ok {
			return
		}

		marriageFamily := t.FindFamily(father, mother)
		marriageFamily.Bond = model.FamilyBondMarried
		marriageFamily.Timeline = append(marriageFamily.Timeline, ev)
		marriageFamily.BestStartEvent = ev
		marriageFamily.BestStartDate = ev.GetDate()
	}

	for _, ev := range p.Timeline {
		switch tev := ev.(type) {
		case *model.MarriageEvent:
			addMarriage(t, tev)
		case *model.MarriageLicenseEvent:
			addMarriage(t, tev)
		case *model.MarriageBannsEvent:
			addMarriage(t, tev)
		case *model.DivorceEvent:
		case *model.AnnulmentEvent:
		}
	}

	return nil
}

func (t *Tree) InferFamilyRelationships(f *model.Family) error {
	// for _, c := range f.Children {
	// 	if c.ID == "LUEW7HLX2PMWC" {
	// 		fmt.Printf("Found LUEW7HLX2PMWC, father is %q, family father is %q, mother is %q, family mother is %q\n", personid(c.Father), personid(f.Father), personid(c.Mother), personid(f.Mother))
	// 	}
	// 	if f.Father != nil {
	// 		if c.Father != nil {
	// 			fmt.Printf("tried to add father %q to child %q but they already had a father %q\n", f.Father.ID, c.ID, c.Father.ID)
	// 		} else {
	// 			c.Father = f.Father
	// 		}
	// 	}
	// 	if f.Mother != nil {
	// 		if c.Mother != nil {
	// 			fmt.Printf("tried to add mother %q to child %q but they already had a mother %q\n", f.Mother.ID, c.ID, c.Mother.ID)
	// 		} else {
	// 			c.Mother = f.Mother
	// 		}
	// 	}
	// 	if c.ID == "LUEW7HLX2PMWC" {
	// 		fmt.Printf("AFTER LUEW7HLX2PMWC, father is %q, family father is %q, mother is %q, family mother is %q\n", personid(c.Father), personid(f.Father), personid(c.Mother), personid(f.Mother))
	// 	}
	// }
	return nil
}

func (t *Tree) SelectPersonBestBirthDeathEvents(p *model.Person) error {
	for _, ev := range p.Timeline {
		if iev, ok := ev.(model.IndividualTimelineEvent); ok && p.SameAs(iev.GetPrincipal()) {
			switch tev := ev.(type) {
			case *model.BirthEvent:
				if bev, ok := p.BestBirthlikeEvent.(*model.BirthEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestBirthlikeEvent = tev
					}
				} else {
					// Birth overrides all
					p.BestBirthlikeEvent = tev
				}

			case *model.BaptismEvent:
				if bev, ok := p.BestBirthlikeEvent.(*model.BaptismEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestBirthlikeEvent = tev
					}
				} else if p.BestBirthlikeEvent == nil {
					// use event only if no better event
					p.BestBirthlikeEvent = tev
				}

			case *model.DeathEvent:
				if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestDeathlikeEvent = tev
					}
				} else {
					// Death overrides all
					p.BestDeathlikeEvent = tev
				}

			case *model.BurialEvent:
				if bev, ok := p.BestDeathlikeEvent.(*model.BurialEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestDeathlikeEvent = ev
					}
				} else if p.BestDeathlikeEvent == nil {
					// use event only if no better event
					p.BestDeathlikeEvent = ev
				}

			case *model.CremationEvent:
				if bev, ok := p.BestDeathlikeEvent.(*model.CremationEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestDeathlikeEvent = ev
					}
				} else if p.BestDeathlikeEvent == nil {
					// use event only if no better event
					p.BestDeathlikeEvent = ev
				}

			case *model.ProbateEvent:
				if bev, ok := p.BestDeathlikeEvent.(*model.ProbateEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestDeathlikeEvent = tev
					}
				} else if p.BestDeathlikeEvent == nil {
					p.BestDeathlikeEvent = tev
				}

			}
		}
	}

	var startYear, endYear int

	if p.BestBirthlikeEvent != nil {
		if year, ok := p.BestBirthlikeEvent.GetDate().Year(); ok {
			startYear = year
		}
	}

	// if startYear+120 > time.Now().Year() {
	// 	logging.Debug("marking person as possibly alive since they were born within 120 years from present", "id", p.ID)
	// 	p.PossiblyAlive = true
	// }

	if p.BestDeathlikeEvent != nil {
		if year, ok := p.BestDeathlikeEvent.GetDate().Year(); ok {
			endYear = year
		}
	}

	if startYear == 0 && endYear == 0 {
		p.VitalYears = "-?-"
	} else if startYear != 0 && endYear == 0 {
		p.VitalYears = fmt.Sprintf("%d-", startYear)
		if !p.PossiblyAlive {
			p.VitalYears += "?"
		}
	} else if startYear == 0 && endYear != 0 {
		p.VitalYears = fmt.Sprintf("?-%d", endYear)
	} else {
		p.VitalYears = fmt.Sprintf("%d-%d", startYear, endYear)
	}

	if p.PossiblyAlive {
		p.BeingTense = "is"
	} else {
		p.BeingTense = "was"
	}

	return nil
}

func (t *Tree) InferFamilyStartEndDates(f *model.Family) error {
	if f.BestEndDate == nil {
		if f.Mother == nil || f.Mother.BestDeathlikeEvent == nil {
			if f.Father != nil && f.Father.BestDeathlikeEvent != nil {
				f.BestEndDate = f.Father.BestDeathlikeEvent.GetDate()
				f.EndDeathPerson = f.Father
				f.EndReason = model.FamilyEndReasonDeath
			}
		} else if f.Father == nil || f.Father.BestDeathlikeEvent == nil {
			if f.Mother != nil && f.Mother.BestDeathlikeEvent != nil {
				f.BestEndDate = f.Mother.BestDeathlikeEvent.GetDate()
				f.EndDeathPerson = f.Mother
				f.EndReason = model.FamilyEndReasonDeath
			}
		} else {
			if f.Mother.BestDeathlikeEvent.GetDate().SortsBefore(f.Father.BestDeathlikeEvent.GetDate()) {
				f.BestEndDate = f.Mother.BestDeathlikeEvent.GetDate()
				f.EndDeathPerson = f.Mother
				f.EndReason = model.FamilyEndReasonDeath
				f.Father.Timeline = append(f.Father.Timeline, f.Mother.BestDeathlikeEvent)
			} else {
				f.BestEndDate = f.Father.BestDeathlikeEvent.GetDate()
				f.EndDeathPerson = f.Father
				f.EndReason = model.FamilyEndReasonDeath
				f.Mother.Timeline = append(f.Mother.Timeline, f.Father.BestDeathlikeEvent)
			}
		}
	}

	inferFromPersonBirth := func(f *model.Family, p *model.Person) {
		if p == nil {
			return
		}

		// Set family start date to the be about person's birth if it is better than the current start date
		if p.BestBirthlikeEvent != nil {
			if f.BestStartDate == nil || p.BestBirthlikeEvent.GetDate().SortsBefore(f.BestStartDate) {
				if yr, ok := p.BestBirthlikeEvent.GetDate().Year(); ok {
					f.BestStartDate = model.AboutYear(yr)
				}
			}
		}
	}

	inferFromPersonDeath := func(f *model.Family, p *model.Person) {
		if p == nil {
			return
		}

		// Set family start date to the be before person's death if it is better than the current start date
		if p.BestDeathlikeEvent != nil {
			if f.BestStartDate == nil || p.BestDeathlikeEvent.GetDate().SortsBefore(f.BestStartDate) {
				if yr, ok := p.BestDeathlikeEvent.GetDate().Year(); ok {
					f.BestStartDate = model.BeforeYear(yr)
				}
			}
		}
	}

	inferFromPersonDeath(f, f.Father)
	inferFromPersonDeath(f, f.Mother)
	for _, c := range f.Children {
		inferFromPersonBirth(f, c)
		inferFromPersonDeath(f, c)
	}

	return nil
}

func (t *Tree) RefinePersonNames(p *model.Person) error {
	if p.NickName != "" {
		p.PreferredFamiliarName = p.NickName
	} else {
		givenParts := strings.Split(p.PreferredGivenName, " ")
		p.PreferredFamiliarName = givenParts[0]
	}
	p.PreferredFamiliarFullName = p.PreferredFamiliarName + " " + p.PreferredFamilyName

	// Adjust names to include vital years
	if p.VitalYears != "" {
		p.PreferredUniqueName = fmt.Sprintf("%s (%s)", p.PreferredFullName, p.VitalYears)
		p.PreferredSortName = fmt.Sprintf("%s (%s)", p.PreferredSortName, p.VitalYears)
	}

	return nil
}

func (t *Tree) RefinePersonOccupations(p *model.Person) error {
	// TODO: move from gedcom import to here

	if len(p.Occupations) == 0 {
		return nil
	}
	if len(p.Occupations) == 1 {
		p.PrimaryOccupation = p.Occupations[0].Detail
		return nil
	}

	return nil
}

func (t *Tree) BuildOlb(p *model.Person) error {
	return nil
}

func (t *Tree) Redact(redactLiving bool) error {
	for _, p := range t.People {
		redact := false
		if redactLiving {
			if p.PossiblyAlive {
				logging.Debug("redacting possibly alive person", "id", p.ID, "name", p.PreferredFullName)
				redact = true
			} else if years, known := model.YearsSinceDeath(p); known && years < 21 {
				logging.Debug("redacting recently deceased person", "id", p.ID, "name", p.PreferredFullName)
				redact = true
			}
		}
		if p.Redacted || redact {
			infer.RedactPersonalDetailsWithDescendants(p)
		}
	}

	return nil
}

func (t *Tree) ExpandPersonTimeline(p *model.Person) error {
	// Add birth event to mother and father's timelines
	if p.BestBirthlikeEvent != nil {
		if p.Father != nil {
			p.Father.Timeline = append(p.Father.Timeline, p.BestBirthlikeEvent)
		}
		if p.Mother != nil {
			p.Mother.Timeline = append(p.Mother.Timeline, p.BestBirthlikeEvent)
		}

	}

	// Add death event to mother and father's timelines if it happened before the parent died
	if p.BestDeathlikeEvent != nil {
		if p.Father != nil {
			if p.Father.BestDeathlikeEvent == nil || p.BestDeathlikeEvent.GetDate().SortsBefore(p.Father.BestDeathlikeEvent.GetDate()) {
				p.Father.Timeline = append(p.Father.Timeline, p.BestDeathlikeEvent)
			}
		}
		if p.Mother != nil {
			if p.Mother.BestDeathlikeEvent == nil || p.BestDeathlikeEvent.GetDate().SortsBefore(p.Mother.BestDeathlikeEvent.GetDate()) {
				p.Mother.Timeline = append(p.Mother.Timeline, p.BestDeathlikeEvent)
			}
		}
	}

	// Add family timeline events
	// for _, f := range p.Families {
	// 	for _, ev := range f.Timeline {
	// 		// p.Timeline = append(p.Timeline, ev)
	// 	}
	// }

	// TODO: add parent deaths to child timelines
	// TODO: add sibling deaths to person timelines

	return nil
}

func (t *Tree) TrimPersonTimeline(p *model.Person) error {
	evs := make([]model.TimelineEvent, 0, len(p.Timeline))

EventLoop:
	for _, ev := range p.Timeline {

		// Keep all events that directly involve this person
		if ev.DirectlyInvolves(p) {
			evs = append(evs, ev)
			continue
		}

		// Drop all events from redacted people
		for _, o := range ev.Participants() {
			if o.Redacted {
				continue EventLoop
			}
		}

		// Skip events before the person's birth
		if p.BestBirthlikeEvent != nil {
			if ev.GetDate().SortsBefore(p.BestBirthlikeEvent.GetDate()) {
				continue
			}
		}

		// Skip events after the person's birth
		if p.BestDeathlikeEvent != nil {
			if p.BestDeathlikeEvent.GetDate().SortsBefore(ev.GetDate()) {
				continue
			}
		}

		evs = append(evs, ev)
	}

	p.Timeline = evs
	return nil
}

func (t *Tree) TrimPlaceTimeline(p *model.Place) error {
	evs := make([]model.TimelineEvent, 0, len(p.Timeline))

EventLoop:
	for _, ev := range p.Timeline {
		// Drop all events from redacted people
		for _, o := range ev.Participants() {
			if o.Redacted {
				continue EventLoop
			}
		}
		evs = append(evs, ev)
	}

	p.Timeline = evs
	return nil
}

func (t *Tree) TrimSourceTimeline(s *model.Source) error {
	evs := make([]model.TimelineEvent, 0, len(s.EventsCiting))

EventLoop:
	for _, ev := range s.EventsCiting {
		// Drop all events from redacted people
		for _, o := range ev.Participants() {
			if o.Redacted {
				continue EventLoop
			}
		}
		evs = append(evs, ev)
	}

	s.EventsCiting = evs
	return nil
}

func (t *Tree) SetKeyPerson(p *model.Person) {
	t.KeyPerson = p
}

func (t *Tree) BuildRelationsToKeyPerson() {
	if t.KeyPerson == nil {
		return
	}
	t.KeyPerson.RelationToKeyPerson = model.Self(t.KeyPerson)
	t.KeyPerson.RedactionKeepsName = true

	descendKeyPersonRelationship(t.KeyPerson)
	roots := ascendKeyPersonRelationship(t.KeyPerson)
	for _, r := range roots {
		descendKeyPersonRelationship(r)
	}
}

func ascendKeyPersonRelationship(p *model.Person) []*model.Person {
	if p.Father == nil && p.Mother == nil {
		// this person is a root of the tree
		return []*model.Person{p}
	}

	var roots []*model.Person
	if p.Father != nil {
		p.Father.RelationToKeyPerson = p.RelationToKeyPerson.ExtendToParent(p.Father)
		froots := ascendKeyPersonRelationship(p.Father)
		roots = append(roots, froots...)
	}

	if p.Mother != nil {
		p.Mother.RelationToKeyPerson = p.RelationToKeyPerson.ExtendToParent(p.Mother)
		mroots := ascendKeyPersonRelationship(p.Mother)
		roots = append(roots, mroots...)
	}

	return roots
}

func descendKeyPersonRelationship(p *model.Person) {
	for _, ch := range p.Children {
		if ch.RelationToKeyPerson != nil {
			continue
		}
		ch.RelationToKeyPerson = p.RelationToKeyPerson.ExtendToChild(ch)
		descendKeyPersonRelationship(ch)
	}

	for _, sp := range p.Spouses {
		if sp.RelationToKeyPerson != nil {
			continue
		}
		sp.RelationToKeyPerson = p.RelationToKeyPerson.ExtendToSpouse(sp)
		// recurseKeyPersonRelationship(sp)
	}
}

func (t *Tree) ListPeopleMatching(m model.PersonMatcher, limit int) []*model.Person {
	matches := make([]*model.Person, 0, limit)
	for _, p := range t.People {
		if m(p) {
			matches = append(matches, p)
		}
		if len(matches) == limit {
			break
		}
	}
	return matches
}

func (t *Tree) ListPlacesMatching(m model.PlaceMatcher, limit int) []*model.Place {
	matches := make([]*model.Place, 0, limit)
	for _, p := range t.Places {
		if m(p) {
			matches = append(matches, p)
		}
		if len(matches) == limit {
			break
		}
	}
	return matches
}

// Metrics

// In general all metrics exclude redacted people

// NumberOfPeople returns the number of people in the tree.
// It excludes redacted people.
func (t *Tree) NumberOfPeople() int {
	num := 0
	for _, p := range t.People {
		if !p.Redacted {
			num++
		}
	}
	return num
}

// AncestorSurnameDistribution returns a map of surnames and the number
// of direct ancestors with that surname
// It excludes redacted people.
func (t *Tree) AncestorSurnameDistribution() map[string]int {
	dist := make(map[string]int)
	for _, p := range t.People {
		if p.Redacted {
			continue
		}
		if p.PreferredFamilyName == model.UnknownNamePlaceholder {
			continue
		}
		if p.RelationToKeyPerson.IsDirectAncestor() {
			dist[p.PreferredFamilyName]++
		}
	}
	return dist
}

// TreeSurnameDistribution returns a map of surnames and the number
// of people in the tree with that surname
// It excludes redacted people.
func (t *Tree) TreeSurnameDistribution() map[string]int {
	dist := make(map[string]int)
	for _, p := range t.People {
		if p.Redacted {
			continue
		}
		if p.PreferredFamilyName == model.UnknownNamePlaceholder {
			continue
		}
		dist[p.PreferredFamilyName]++
	}
	return dist
}

// OldestPeople returns a list of the oldest people in the tree, sorted by descending age
// It excludes redacted people.
func (t *Tree) OldestPeople(limit int) []*model.Person {
	h := new(PersonWithAgeHeap)
	heap.Init(h)

	for _, p := range t.People {
		if p.Redacted {
			continue
		}
		age, ok := p.AgeInYearsAtDeath()
		if !ok {
			continue
		}
		heap.Push(h, &PersonWithAge{Person: p, Age: age})
		if h.Len() > limit {
			heap.Pop(h)
		}
	}

	list := make([]*model.Person, h.Len())
	for i := len(list) - 1; i >= 0; i-- {
		pa := heap.Pop(h).(*PersonWithAge)
		list[i] = pa.Person
	}
	return list
}

type PersonWithAge struct {
	Person *model.Person
	Age    int
}

type PersonWithAgeHeap []*PersonWithAge

func (h PersonWithAgeHeap) Len() int           { return len(h) }
func (h PersonWithAgeHeap) Less(i, j int) bool { return h[i].Age < h[j].Age }
func (h PersonWithAgeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PersonWithAgeHeap) Push(x interface{}) {
	*h = append(*h, x.(*PersonWithAge))
}

func (h *PersonWithAgeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

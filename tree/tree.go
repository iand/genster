package tree

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/iand/genster/identifier"
	"github.com/iand/genster/infer"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/place"
	"github.com/iand/genster/text"
)

type Tree struct {
	ID            string
	Name          string
	Description   string
	IdentityMap   *IdentityMap
	Gazeteer      *Gazeteer
	Annotations   *Annotations
	SurnameGroups *SurnameGroups
	People        map[string]*model.Person
	Citations     map[string]*model.GeneralCitation
	Sources       map[string]*model.Source
	Repositories  map[string]*model.Repository
	Places        map[string]*model.Place
	Families      map[string]*model.Family
	MediaObjects  map[string]*model.MediaObject
	KeyPerson     *model.Person
}

func NewTree(id string, m *IdentityMap, g *Gazeteer, a *Annotations, sg *SurnameGroups) *Tree {
	return &Tree{
		ID:            id,
		IdentityMap:   m,
		Gazeteer:      g,
		Annotations:   a,
		SurnameGroups: sg,
		People:        make(map[string]*model.Person),
		Citations:     make(map[string]*model.GeneralCitation),
		Sources:       make(map[string]*model.Source),
		Repositories:  make(map[string]*model.Repository),
		Places:        make(map[string]*model.Place),
		Families:      make(map[string]*model.Family),
		MediaObjects:  make(map[string]*model.MediaObject),
	}
}

func (t *Tree) GetPerson(id string) (*model.Person, bool) {
	p, ok := t.People[id]
	return p, ok
}

func (t *Tree) FindPerson(scope string, sid string) *model.Person {
	id := t.CanonicalID(scope, sid)
	p, ok := t.People[id]
	if !ok {
		p = &model.Person{
			ID: id,
		}
		// p.BestBirthlikeEvent = &model.BirthEvent{
		// 	GeneralEvent: model.GeneralEvent{Date: model.UnknownDate()},
		// 	GeneralIndividualEvent: model.GeneralIndividualEvent{
		// 		Principal: p,
		// 	},
		// }

		t.People[id] = p
	}
	return p
}

func (t *Tree) GetCitation(id string) (*model.GeneralCitation, bool) {
	c, ok := t.Citations[id]
	return c, ok
}

func (t *Tree) FindCitation(scope string, sid string) (*model.GeneralCitation, bool) {
	id := t.CanonicalID(scope, sid)
	c, ok := t.Citations[id]
	if !ok {
		c = &model.GeneralCitation{
			ID: id,
		}
		t.Citations[id] = c
		return c, false
	}
	return c, true
}

func (t *Tree) GetSource(id string) (*model.Source, bool) {
	so, ok := t.Sources[id]
	return so, ok
}

func (t *Tree) FindSource(scope string, sid string) *model.Source {
	id := t.CanonicalID(scope, sid)
	so, ok := t.Sources[id]
	if !ok {
		so = &model.Source{
			ID: id,
		}
		t.Sources[id] = so
	}
	return so
}

func (t *Tree) FindRepository(scope string, sid string) *model.Repository {
	id := t.CanonicalID(scope, sid)
	re, ok := t.Repositories[id]
	if !ok {
		re = &model.Repository{
			ID: id,
		}
		t.Repositories[id] = re
	}
	return re
}

func (t *Tree) GetPlace(id string) (*model.Place, bool) {
	p, ok := t.Places[id]
	return p, ok
}

func (t *Tree) FindPlace(scope string, sid string) *model.Place {
	id := t.CanonicalID(scope, sid)
	pl, ok := t.Places[id]
	if !ok {
		pl = &model.Place{
			ID: id,
		}
		t.Places[id] = pl
	}
	return pl
}

func (t *Tree) FindPlaceUnstructured(name string, hints ...place.Hint) *model.Place {
	id := t.Gazeteer.ID(name, hints...)
	p, ok := t.Places[id]
	if !ok {

		pn := place.ClassifyName(name, hints...)
		cleanName := pn.Name

		p = &model.Place{
			ID:                id,
			OriginalText:      name,
			Hints:             hints,
			Name:              cleanName,
			FullName:          cleanName,
			PreferredSortName: cleanName,
			PlaceType:         model.PlaceTypeUnknown,
			// Kind:                pn.Kind,
			CountryName:  place.UnknownPlaceName(),
			UKNationName: place.UnknownPlaceName(),
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

func (t *Tree) FindFamily(scope string, sid string) *model.Family {
	id := t.CanonicalID(scope, sid)
	f, ok := t.Families[id]
	if !ok {
		f = t.newFamily(id)
	}
	return f
}

func (t *Tree) FindFamilyByParents(father *model.Person, mother *model.Person) *model.Family {
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

func (t *Tree) FindMediaObject(path string) *model.MediaObject {
	id := t.CanonicalID("mediaobject", path)
	mo, ok := t.MediaObjects[id]
	if !ok {
		mo = &model.MediaObject{
			ID:          id,
			SrcFilePath: path,
		}
		t.MediaObjects[id] = mo
	}
	return mo
}

func (t *Tree) CanonicalID(scope string, sid string) string {
	return t.IdentityMap.ID(scope, sid)
}

func (t *Tree) AddAlias(alias string, canonical string) {
	t.IdentityMap.AddAlias(alias, canonical)
}

func (t *Tree) Generate(redactLiving bool) error {
	// Apply any annotations first, they may be redacted after
	if t.Annotations != nil {
		t.Annotations.ApplyTree(t)

		for _, p := range t.People {
			t.Annotations.ApplyPerson(p)
		}
		for _, p := range t.Places {
			t.Annotations.ApplyPlace(p)
		}
		for _, p := range t.Sources {
			t.Annotations.ApplySource(p)
		}
	}

	// Add data to each person
	for _, p := range t.People {
		t.PropagateParents(p)
	}

	// Need to make sure all parents and children are linked first
	// since expand timeline will add events to parents timelines
	for _, f := range t.Families {
		t.InferFamilyRelationships(f)
	}

	t.BuildRelationsToKeyPerson()

	// Add data to each person
	for _, p := range t.People {
		t.SelectPersonBestBirthDeathEvents(p)
		t.RefinePersonNames(p)
		t.RefinePersonOccupations(p)
		t.ExpandPersonTimeline(p)
	}

	// Fill in gaps with inferences
	for _, p := range t.People {
		// infer.InferPersonBirthEventDate(p)
		infer.InferPersonAliveOrDead(p, time.Now().Year())
		// infer.InferPersonDeathEventDate(p)
		infer.InferPersonCauseOfDeath(p)
		infer.InferPersonGeneralFacts(p)
	}

	for _, f := range t.Families {
		t.AddSpouses(f)
		t.RefineFamilyNames(f)
		t.InferFamilyStartEndDates(f)
	}

	// Redact any personal information
	t.Redact(redactLiving)

	for _, p := range t.People {
		t.TrimPersonTimeline(p)
		t.CrossReferenceCitations(p)
		p.RemoveDuplicateFamilies()
		p.RemoveDuplicateChildren()
		p.RemoveDuplicateSpouses()
		t.BuildOlb(p)

		// sort families by date
		sort.Slice(p.Families, func(a, b int) bool {
			return p.Families[a].BestStartDate.SortsBefore(p.Families[b].BestStartDate)
		})

		// sort children in families by birthlike date
		for _, f := range p.Families {
			sort.Slice(f.Children, func(a, b int) bool {
				return f.Children[a].BestBirthDate().SortsBefore(f.Children[b].BestBirthDate())
			})
		}

		// Sort all children of person
		sort.Slice(p.Children, func(a, b int) bool {
			return p.Children[a].BestBirthDate().SortsBefore(p.Children[b].BestBirthDate())
		})

		// Sort all occupations of person
		sort.Slice(p.Occupations, func(a, b int) bool {
			return p.Occupations[a].Date.SortsBefore(p.Occupations[b].Date)
		})
	}

	for _, p := range t.Places {
		t.RefinePlaceNames(p)
		t.TrimPlaceTimeline(p)
	}
	for _, s := range t.Sources {
		t.TrimSourceTimeline(s)
	}
	for _, c := range t.Citations {
		t.TrimCitationPeopleCited(c)
		t.TrimCitationEventsCited(c)
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

func (t *Tree) InferFamilyRelationships(f *model.Family) error {
	for _, c := range f.Children {
		if c.ParentFamily != nil {
			logging.Warn("person already has a parent family", "id", c.ID)
		}
		c.ParentFamily = f
	}
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
				if p.BestBirthlikeEvent == nil {
					p.BestBirthlikeEvent = tev
				} else if bev, ok := p.BestBirthlikeEvent.(*model.BirthEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestBirthlikeEvent = tev
					}
				} else {
					if tev.GetDate().IsMorePreciseThan(p.BestBirthlikeEvent.GetDate()) {
						p.BestBirthlikeEvent = tev
					}
				}

			case *model.BaptismEvent:
				if p.BestBirthlikeEvent == nil {
					// use event if no better event
					p.BestBirthlikeEvent = tev
				} else if bev, ok := p.BestBirthlikeEvent.(*model.BaptismEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestBirthlikeEvent = tev
					}
				} else if tev.GetDate().IsMorePreciseThan(p.BestBirthlikeEvent.GetDate()) {
					p.BestBirthlikeEvent = tev
				}

			case *model.NamingEvent:
				if p.BestBirthlikeEvent == nil {
					// use event if no better event
					p.BestBirthlikeEvent = tev
				} else if bev, ok := p.BestBirthlikeEvent.(*model.BaptismEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestBirthlikeEvent = tev
					}
				} else if bev, ok := p.BestBirthlikeEvent.(*model.NamingEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestBirthlikeEvent = tev
					}
				} else if tev.GetDate().IsMorePreciseThan(p.BestBirthlikeEvent.GetDate()) {
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
						p.BestDeathlikeEvent = tev
					}
				} else if p.BestDeathlikeEvent == nil {
					// use event only if no better event
					p.BestDeathlikeEvent = tev
				}

			case *model.CremationEvent:
				if bev, ok := p.BestDeathlikeEvent.(*model.CremationEvent); ok {
					if tev.Date.SortsBefore(bev.Date) {
						p.BestDeathlikeEvent = tev
					}
				} else if p.BestDeathlikeEvent == nil {
					// use event only if no better event
					p.BestDeathlikeEvent = tev
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

	if startYear+120 > time.Now().Year() {
		logging.Debug("marking person as possibly alive since they were born within 120 years from present", "id", p.ID)
		p.PossiblyAlive = true
	}

	if p.BestDeathlikeEvent != nil {
		p.PossiblyAlive = false
		if year, ok := p.BestDeathlikeEvent.GetDate().Year(); ok {
			endYear = year
		}
	}

	if startYear == 0 && endYear == 0 {
		p.VitalYears = model.UnknownDateRangePlaceholder
	} else if startYear != 0 && endYear == 0 {
		p.VitalYears = fmt.Sprintf("%d–", startYear)
		if !p.PossiblyAlive {
			p.VitalYears += "?"
		}
	} else if startYear == 0 && endYear != 0 {
		p.VitalYears = fmt.Sprintf("?–%d", endYear)
	} else {
		p.VitalYears = fmt.Sprintf("%d–%d", startYear, endYear)
	}

	if p.PossiblyAlive {
		p.BeingTense = "is"
	} else {
		p.BeingTense = "was"
	}

	return nil
}

func (t *Tree) AddSpouses(f *model.Family) error {
	if f.Mother.IsUnknown() || f.Father.IsUnknown() {
		return nil
	}

	f.Mother.Spouses = append(f.Mother.Spouses, f.Father)
	f.Father.Spouses = append(f.Father.Spouses, f.Mother)
	return nil
}

func (t *Tree) RefineFamilyNames(f *model.Family) error {
	if f.Mother.IsUnknown() {
		if f.Father.IsUnknown() {
			f.PreferredUniqueName = "Unknown parents"
			f.PreferredFullName = "Unknown parents"
			f.PreferredFamiliarName = "Unknown parents"
		} else {
			f.PreferredUniqueName = f.Father.PreferredUniqueName
			f.PreferredFullName = f.Father.PreferredFullName
			f.PreferredFamiliarName = f.Father.PreferredFamiliarName
		}
	} else {
		if f.Father.IsUnknown() {
			f.PreferredUniqueName = f.Mother.PreferredUniqueName
			f.PreferredFullName = f.Mother.PreferredFullName
			f.PreferredFamiliarName = f.Mother.PreferredFamiliarName
		} else {
			f.PreferredUniqueName = f.Father.PreferredUniqueName + " and " + f.Mother.PreferredUniqueName
			f.PreferredFullName = f.Father.PreferredFullName + " and " + f.Mother.PreferredFullName
			f.PreferredFamiliarName = f.Father.PreferredFamiliarName + " and " + f.Mother.PreferredFamiliarName
		}
	}

	return nil
}

func (t *Tree) InferFamilyStartEndDates(f *model.Family) error {
	if f.BestStartEvent != nil && (f.Bond == model.FamilyBondMarried || f.Bond == model.FamilyBondLikelyMarried) {
		f.BestStartDate = f.BestStartEvent.GetDate()
	}

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
			if f.BestStartDate == nil || ((f.Bond == model.FamilyBondUnmarried || f.Bond == model.FamilyBondLikelyUnmarried) && p.BestBirthlikeEvent.GetDate().SortsBefore(f.BestStartDate)) {
				_, isBirth := p.BestBirthlikeEvent.(*model.BirthEvent)

				if isBirth && (f.Bond == model.FamilyBondUnmarried || f.Bond == model.FamilyBondLikelyUnmarried) {
					f.BestStartDate = p.BestBirthlikeEvent.GetDate()
					return
				}

				if yr, ok := p.BestBirthlikeEvent.GetDate().Year(); ok {
					f.BestStartDate = model.BeforeYear(yr)
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
	p.PreferredFullName = text.RemoveRedundantWhitespace(p.PreferredFullName)
	p.PreferredGivenName = text.RemoveRedundantWhitespace(p.PreferredGivenName)
	p.PreferredFamiliarName = text.RemoveRedundantWhitespace(p.PreferredFamiliarName)
	p.PreferredFamiliarFullName = text.RemoveRedundantWhitespace(p.PreferredFamiliarFullName)
	p.PreferredFamilyName = text.RemoveRedundantWhitespace(p.PreferredFamilyName)
	p.PreferredSortName = text.RemoveRedundantWhitespace(p.PreferredSortName)
	p.PreferredUniqueName = text.RemoveRedundantWhitespace(p.PreferredUniqueName)
	p.NickName = text.RemoveRedundantWhitespace(p.NickName)

	if p.PreferredFullName == model.UnknownNamePlaceholder+" "+model.UnknownNamePlaceholder {
		if p.NickName != "" {
			p.PreferredFullName = p.NickName
			p.PreferredGivenName = p.NickName
			p.PreferredFamiliarName = p.NickName
			p.PreferredFamiliarFullName = p.NickName
			p.PreferredFamilyName = p.NickName
			p.PreferredSortName = p.NickName
			p.PreferredUniqueName = p.NickName
		} else {
			p.PreferredFullName = "an unidentified person"
			p.PreferredGivenName = "unidentified"
			p.PreferredFamiliarName = "unidentified"
			p.PreferredFamiliarFullName = "unidentified"
			// p.PreferredFamilyName = "unidentified"
			// p.PreferredSortName = "unidentified person"
			// p.PreferredUniqueName = "an unidentified person"
		}
	} else {
		// if p.NickName != "" {
		// 	p.PreferredFamiliarName = p.NickName
		// } else {
		// 	givenParts := strings.Split(p.PreferredGivenName, " ")
		// 	p.PreferredFamiliarName = givenParts[0]
		// }
		p.PreferredFamiliarFullName = p.PreferredFamiliarName + " " + p.PreferredFamilyName

		// Adjust names to include vital years
		if p.VitalYears != "" {
			if p.VitalYears != model.UnknownDateRangePlaceholder {
				p.PreferredUniqueName = fmt.Sprintf("%s (%s)", p.PreferredFullName, p.VitalYears)
			}
			p.PreferredSortName = fmt.Sprintf("%s (%s)", p.PreferredSortName, p.VitalYears)
		}
	}

	if g, ok := t.SurnameGroups.surnames[p.PreferredFamilyName]; ok {
		p.FamilyNameGrouping = g.String()
	} else {
		p.FamilyNameGrouping = p.PreferredFamilyName
	}

	return nil
}

func (t *Tree) RefinePersonOccupations(p *model.Person) error {
	// TODO: move from gedcom import to here
	// NOTE: PrimaryOccupation is set from Gedcom FACT type Occupation
	// if len(p.Occupations) == 0 {
	// 	return nil
	// }
	// if len(p.Occupations) == 1 {
	// 	logging.Warn("setting primary occupation", "id", p.ID, "primary", p.Occupations[0].Detail)
	// 	p.PrimaryOccupation = p.Occupations[0].Detail

	// 	return nil
	// }
	// if len(p.Occupations) > 1 {
	// 	var occs []string
	// 	for i := range p.Occupations {
	// 		occs = append(occs, p.Occupations[i].Detail)
	// 	}
	// 	logging.Warn("not setting primary occupation", "id", p.ID, "occupations", strings.Join(occs, "|"))

	// }

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

	// add parent deaths to child timelines
	if p.Father != nil && p.Father.BestDeathlikeEvent != nil {
		p.Timeline = append(p.Timeline, p.Father.BestDeathlikeEvent)
	}
	if p.Mother != nil && p.Mother.BestDeathlikeEvent != nil {
		p.Timeline = append(p.Timeline, p.Mother.BestDeathlikeEvent)
	}

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

		// Drop all events from excluded and redacted people
		for _, o := range ev.GetParticipants() {
			if o.Person.Redacted {
				continue EventLoop
			}
		}

		// Skip other events before the person's birth
		if p.BestBirthlikeEvent != nil {
			if ev.GetDate().SortsBefore(p.BestBirthlikeEvent.GetDate()) {
				continue
			}
		}

		// Skip other events after the person's birth
		if p.BestDeathlikeEvent != nil {
			if p.BestDeathlikeEvent.GetDate().SortsBefore(ev.GetDate()) {
				continue
			}
		}

		evs = append(evs, ev)
	}

	p.Timeline = model.CollapseEventList(evs)
	return nil
}

func (t *Tree) TrimPlaceTimeline(p *model.Place) error {
	evs := make([]model.TimelineEvent, 0, len(p.Timeline))

EventLoop:
	for _, ev := range p.Timeline {
		for _, o := range ev.GetParticipants() {
			// Drop all events from redacted people
			if o.Person.Redacted {
				continue EventLoop
			}
		}
		evs = append(evs, ev)
	}

	p.Timeline = model.CollapseEventList(evs)
	return nil
}

func (t *Tree) TrimSourceTimeline(s *model.Source) error {
	evs := make([]model.TimelineEvent, 0, len(s.EventsCiting))

EventLoop:
	for _, ev := range s.EventsCiting {
		// count number of non excluded people involved in event
		for _, o := range ev.GetParticipants() {
			// Drop all events from redacted people
			if o.Person.Redacted {
				continue EventLoop
			}
		}
		evs = append(evs, ev)
	}

	s.EventsCiting = model.CollapseEventList(evs)
	return nil
}

func (t *Tree) CrossReferenceCitations(p *model.Person) {
	if p.Redacted {
		return
	}
EventLoop:
	for _, ev := range p.Timeline {
		// Skip all events from redacted people
		for _, o := range ev.GetParticipants() {
			if o.Person.Redacted {
				continue EventLoop
			}
		}

		for _, cit := range ev.GetCitations() {
			cit.EventsCited = append(cit.EventsCited, ev)
		}

	}

	for _, n := range p.KnownNames {
		for _, cit := range n.Citations {
			cit.PeopleCited = append(cit.PeopleCited, p)
		}
	}
}

func (t *Tree) TrimCitationPeopleCited(c *model.GeneralCitation) error {
	if len(c.PeopleCited) < 2 {
		return nil
	}
	res := make([]*model.Person, 0, len(c.PeopleCited))
	seen := make(map[*model.Person]bool, len(c.PeopleCited))
	for _, p := range c.PeopleCited {
		if !seen[p] {
			res = append(res, p)
			seen[p] = true
		}
	}
	c.PeopleCited = res
	return nil
}

func (t *Tree) TrimCitationEventsCited(c *model.GeneralCitation) error {
	c.EventsCited = model.CollapseEventList(c.EventsCited)
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
	if p.Father.IsUnknown() && p.Mother.IsUnknown() {
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
		if ch.RelationToKeyPerson == nil {
			ch.RelationToKeyPerson = p.RelationToKeyPerson.ExtendToChild(ch)
		}
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

// ApplyPeopleMatching applies fn to each person that matches m until fn returns false or an error
// which is returned if encountered
func (t *Tree) ApplyPeopleMatching(m model.PersonMatcher, fn model.PersonActionFunc) error {
	for _, p := range t.People {
		if m(p) {
			ok, err := fn(p)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}
	}
	return nil
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

func (t *Tree) RefinePlaceNames(pl *model.Place) error {
	if pl.IsUnknown() {
		return nil
	}
	if pl.Name == "" {
		pl.Name = model.UnknownPlaceName
	}

	hierarchy := []*model.Place{pl}
	hierarchyNames := []string{pl.Name}
	par := pl.Parent
	for par != nil {
		hierarchy = append(hierarchy, par)
		hierarchyNames = append(hierarchyNames, par.Name)
		par = par.Parent
	}

	for i := len(hierarchy) - 1; i >= 0; i-- {
		if hierarchy[i].PlaceType == model.PlaceTypeCountry {
			pl.Country = hierarchy[i]
		}
		if hierarchy[i].PlaceType == model.PlaceTypeCounty ||
			hierarchy[i].PlaceType == model.PlaceTypeState ||
			hierarchy[i].PlaceType == model.PlaceTypeArchipelago ||
			hierarchy[i].PlaceType == model.PlaceTypeNation {
			pl.Region = hierarchy[i]
		}
		if hierarchy[i].PlaceType == model.PlaceTypeParish ||
			hierarchy[i].PlaceType == model.PlaceTypeRegistrationDistrict ||
			hierarchy[i].PlaceType == model.PlaceTypeCity ||
			hierarchy[i].PlaceType == model.PlaceTypeTown ||
			hierarchy[i].PlaceType == model.PlaceTypeVillage ||
			hierarchy[i].PlaceType == model.PlaceTypeHamlet {
			pl.District = hierarchy[i]
		}
		if hierarchy[i].PlaceType == model.PlaceTypeUnknown && pl.District == nil {
			pl.District = hierarchy[i]
		}
	}

	pl.FullName = strings.Join(hierarchyNames, ", ")
	pl.FullContext = strings.Join(hierarchyNames[1:], ", ")

	switch pl.PlaceType {
	case model.PlaceTypeParish:
		pl.ProseName = "the parish of " + pl.Name
	case model.PlaceTypeRegistrationDistrict:
		pl.ProseName = "the " + pl.Name + " registration district"
	default:
		pl.ProseName = pl.Name
	}
	for i := 1; i < len(hierarchy); i++ {
		switch hierarchy[i].PlaceType {
		case model.PlaceTypeParish:
			pl.ProseName += " in the parish of " + hierarchy[i].Name
		case model.PlaceTypeRegistrationDistrict:
			pl.ProseName += " in the " + hierarchy[i].Name + " registration district"
		case model.PlaceTypeCity, model.PlaceTypeTown, model.PlaceTypeVillage, model.PlaceTypeHamlet:
			pl.ProseName += " in " + hierarchy[i].Name
		default:
			pl.ProseName += ", " + hierarchy[i].Name
		}
	}

	pl.NameWithDistrict = pl.Name
	if pl.District != nil && !pl.District.SameAs(pl) {
		pl.NameWithDistrict += ", " + pl.District.Name
		pl.DistrictContext = pl.District.Name
	}
	pl.NameWithRegion = pl.NameWithDistrict
	if pl.Region != nil && !pl.Region.SameAs(pl) {
		pl.NameWithRegion += ", " + pl.Region.Name
		pl.RegionContext = pl.DistrictContext + ", " + pl.Region.Name
	}
	pl.NameWithCountry = pl.NameWithRegion
	if pl.Country != nil && !pl.Country.SameAs(pl) {
		pl.NameWithCountry += ", " + pl.Country.Name
		pl.CountryContext = pl.RegionContext + ", " + pl.Country.Name
	}

	return nil
}

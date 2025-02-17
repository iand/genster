package site

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

type PublishSet struct {
	KeyPerson    *model.Person
	People       map[string]*model.Person
	Citations    map[string]*model.GeneralCitation
	Sources      map[string]*model.Source
	Repositories map[string]*model.Repository
	Places       map[string]*model.Place
	Families     map[string]*model.Family
	MediaObjects map[string]*model.MediaObject
	Events       map[model.TimelineEvent]bool
	LastUpdated  time.Time
}

func NewPublishSet(t *tree.Tree, include model.PersonMatcher) (*PublishSet, error) {
	if t.KeyPerson != nil && !include(t.KeyPerson) {
		return nil, fmt.Errorf("key person must be included in subset")
	}

	ps := &PublishSet{
		KeyPerson:    t.KeyPerson,
		People:       make(map[string]*model.Person),
		Citations:    make(map[string]*model.GeneralCitation),
		Sources:      make(map[string]*model.Source),
		Repositories: make(map[string]*model.Repository),
		Places:       make(map[string]*model.Place),
		Families:     make(map[string]*model.Family),
		MediaObjects: make(map[string]*model.MediaObject),
		Events:       make(map[model.TimelineEvent]bool),
	}

	includePlace := func(pl *model.Place) {
		ps.Places[pl.ID] = pl
		for pl.Parent != nil {
			ps.Places[pl.Parent.ID] = pl.Parent
			pl = pl.Parent
		}
	}

	includePlacesInTexts := func(txts ...model.Text) {
		for _, txt := range txts {
			for _, l := range txt.Links {
				switch t := l.Object.(type) {
				case *model.Place:
					includePlace(t)
				}
			}
		}
	}

	maybeIncludeCitation := func(c *model.GeneralCitation) {
		if c.Redacted {
			return
		}
		ps.Citations[c.ID] = c
	}

	maybeIncludeMediaObject := func(mo *model.MediaObject) {
		if mo.Redacted {
			return
		}
		ps.MediaObjects[mo.ID] = mo
	}

	maybeIncludeFamily := func(f *model.Family) {
		if f == nil {
			return
		}
		ps.Families[f.ID] = f
	}

	for _, p := range t.People {
		if !include(p) {
			continue
		}
		ps.People[p.ID] = p
		for _, ev := range p.Timeline {
			ps.Events[ev] = true
		}
		if p.BestBirthlikeEvent != nil {
			ps.Events[p.BestBirthlikeEvent] = true
		}
		if p.BestDeathlikeEvent != nil {
			ps.Events[p.BestDeathlikeEvent] = true
		}

		if p.CauseOfDeath != nil {
			for _, c := range p.CauseOfDeath.Citations {
				maybeIncludeCitation(c)
			}
		}

		for _, n := range p.KnownNames {
			for _, c := range n.Citations {
				maybeIncludeCitation(c)
			}
		}
		for _, f := range p.MiscFacts {
			for _, c := range f.Citations {
				maybeIncludeCitation(c)
			}
		}
		for _, o := range p.Occupations {
			for _, c := range o.Citations {
				maybeIncludeCitation(c)
			}
		}
		for _, a := range p.Associations {
			for _, c := range a.Citations {
				maybeIncludeCitation(c)
			}
		}
		for _, cmo := range p.Gallery {
			maybeIncludeMediaObject(cmo.Object)
		}

		maybeIncludeFamily(p.ParentFamily)
		for _, f := range p.Families {
			maybeIncludeFamily(f)
		}

		// Add events concerning children so citations are brought in
		for _, ch := range p.Children {
			if ch.Redacted {
				continue
			}
			if ch.BestBirthlikeEvent != nil {
				ps.Events[ch.BestBirthlikeEvent] = true
			}
			if ch.BestDeathlikeEvent != nil {
				ps.Events[ch.BestDeathlikeEvent] = true
			}
			for _, f := range p.Families {
				if f.BestStartEvent != nil {
					ps.Events[f.BestStartEvent] = true
				}
				if f.BestEndEvent != nil {
					ps.Events[f.BestEndEvent] = true
				}
			}

		}

		includePlacesInTexts(p.ResearchNotes...)
		includePlacesInTexts(p.Comments...)

	}

	includedEvents := func(ev model.TimelineEvent) bool {
		return ps.Events[ev]
	}

	for ev := range ps.Events {
		pl := ev.GetPlace()
		if pl != nil {
			if _, ok := ps.Places[pl.ID]; !ok {
				pl.Timeline = model.FilterEventList(pl.Timeline, includedEvents)
				includePlace(pl)
			}
		}

		for _, c := range ev.GetCitations() {
			maybeIncludeCitation(c)
		}

		// include any places mentioned in texts
		includePlacesInTexts(ev.GetNarrative())
	}

	for _, c := range ps.Citations {
		includePlacesInTexts(c.ResearchNotes...)
		includePlacesInTexts(c.Comments...)

		for _, cmo := range c.MediaObjects {
			maybeIncludeMediaObject(cmo.Object)
		}
		if c.Source != nil {
			ps.Sources[c.Source.ID] = c.Source
		}

		c.PeopleCited = model.FilterPersonList(c.PeopleCited, include)
		c.EventsCited = model.FilterEventList(c.EventsCited, includedEvents)
	}

	for _, pl := range ps.Places {
		for _, cmo := range pl.Gallery {
			maybeIncludeMediaObject(cmo.Object)
		}
		includePlacesInTexts(pl.ResearchNotes...)
		includePlacesInTexts(pl.Comments...)

	}

	includedCitations := func(c *model.GeneralCitation) bool {
		_, ok := ps.Citations[c.ID]
		return ok
	}

	for _, mo := range ps.MediaObjects {
		mo.Citations = model.FilterCitationList(mo.Citations, includedCitations)
	}

	for _, so := range ps.Sources {
		so.EventsCiting = model.FilterEventList(so.EventsCiting, includedEvents)
		for _, rr := range so.RepositoryRefs {
			if rr.Repository != nil {
				ps.Repositories[rr.Repository.ID] = rr.Repository
			}
		}
	}

	for _, p := range ps.People {
		if p.UpdateTime != nil && p.UpdateTime.After(ps.LastUpdated) {
			ps.LastUpdated = *p.UpdateTime
		}
	}

	for _, p := range ps.Places {
		if p.UpdateTime != nil && p.UpdateTime.After(ps.LastUpdated) {
			ps.LastUpdated = *p.UpdateTime
		}
	}

	for _, c := range ps.Citations {
		if c.UpdateTime != nil && c.UpdateTime.After(ps.LastUpdated) {
			ps.LastUpdated = *c.UpdateTime
		}
	}

	if ps.LastUpdated.IsZero() {
		ps.LastUpdated = time.Now()
	}

	return ps, nil
}

func (ps *PublishSet) Includes(v any) bool {
	switch vt := v.(type) {
	case *model.Person:
		_, ok := ps.People[vt.ID]
		return ok
	case *model.GeneralCitation:
		_, ok := ps.Citations[vt.ID]
		return ok
	case *model.Source:
		_, ok := ps.Sources[vt.ID]
		return ok
	case *model.Family:
		_, ok := ps.Families[vt.ID]
		return ok
	case *model.Place:
		_, ok := ps.Places[vt.ID]
		return ok
	case *model.MediaObject:
		_, ok := ps.MediaObjects[vt.ID]
		return ok
	case *model.Repository:
		_, ok := ps.Repositories[vt.ID]
		return ok
	case model.TimelineEvent:
		_, ok := ps.Events[vt]
		return ok
	default:
		return false
	}
}

// Metrics

// In general all metrics exclude redacted people

// NumberOfPeople returns the number of people in the tree.
// It excludes redacted people.
func (t *PublishSet) NumberOfPeople() int {
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
func (ps *PublishSet) AncestorSurnameDistribution() map[string]int {
	dist := make(map[string]int)
	for _, p := range ps.People {
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

// AncestorSurnameList returns a list of surname groups in generation
// order, starting with the father of the key person, then the mother,
// then the father's father etc..
func (ps *PublishSet) AncestorSurnameGroupList() []string {
	ppl := ps.Ancestors(ps.KeyPerson, 6)
	seen := make(map[string]struct{})
	names := make([]string, 0)

	for i := range ppl {
		if ppl[i] == nil || ppl[i].Redacted {
			continue
		}
		if _, found := seen[ppl[i].FamilyNameGrouping]; !found {
			seen[ppl[i].FamilyNameGrouping] = struct{}{}
			names = append(names, ppl[i].FamilyNameGrouping)
		}
	}
	return names
}

// TreeSurnameDistribution returns a map of surnames and the number
// of people in the tree with that surname
// It excludes redacted people.
func (ps *PublishSet) TreeSurnameDistribution() map[string]int {
	dist := make(map[string]int)
	for _, p := range ps.People {
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
func (ps *PublishSet) OldestPeople(limit int) []*model.Person {
	h := new(PersonWithGreatestNumberHeap)
	heap.Init(h)

	for _, p := range ps.People {
		if p.Redacted {
			continue
		}
		age, ok := p.AgeInYearsAtDeath()
		if !ok {
			continue
		}
		heap.Push(h, &PersonWithNumber{Person: p, Number: age})
		if h.Len() > limit {
			heap.Pop(h)
		}
	}

	list := make([]*model.Person, h.Len())
	for i := len(list) - 1; i >= 0; i-- {
		pa := heap.Pop(h).(*PersonWithNumber)
		list[i] = pa.Person
	}
	return list
}

// GreatestChildren returns a list of the people with the greatest number of children in
// the tree
// It excludes redacted people.
func (ps *PublishSet) GreatestChildren(limit int) []*model.Person {
	h := new(PersonWithGreatestNumberHeap)
	heap.Init(h)

	for _, p := range ps.People {
		if p.Redacted {
			continue
		}
		if len(p.Children) == 0 {
			continue
		}
		heap.Push(h, &PersonWithNumber{Person: p, Number: len(p.Children)})
		if h.Len() > limit {
			heap.Pop(h)
		}
	}

	list := make([]*model.Person, h.Len())
	for i := len(list) - 1; i >= 0; i-- {
		pa := heap.Pop(h).(*PersonWithNumber)
		list[i] = pa.Person
	}
	return list
}

// EarliestBorn returns a list of the earliest born people in the tree, sorted by descending date
// It excludes redacted people.
func (ps *PublishSet) EarliestBorn(limit int) []*model.Person {
	h := new(PersonWithLeastNumberHeap)
	heap.Init(h)

	for _, p := range ps.People {
		if p.Redacted {
			continue
		}

		dt := p.BestBirthDate()
		yr, ok := dt.Year()
		if !ok {
			continue
		}
		heap.Push(h, &PersonWithNumber{Person: p, Number: yr})
		if h.Len() > limit {
			heap.Pop(h)
		}
	}

	list := make([]*model.Person, h.Len())
	for i := len(list) - 1; i >= 0; i-- {
		pa := heap.Pop(h).(*PersonWithNumber)
		list[i] = pa.Person
	}
	return list
}

// Ancestors returns the ancestors of p. The returned list is ordered such that the
// father of entry n is found at (n+2)*2-2, the mother of entry n is found at (n+2)*2-1
// The list will always contain 2^n entries, with unknown ancestors left as nil at the
// appropriate index.
// Odd numbers are female, even numbers are male.
// The child of entry n is found at (n-2)/2 if n is even and (n-3)/2 if n is odd.
// 0: father
// 1: mother
// 2: father's father
// 3: father's mother
// 4: mother's father
// 5: mother's mother
// 6: father's father's father
// 7: father's father's mother
// 8: father's mother's father
// 9: father's mother's mother
// 10: mother's father's father
// 11: mother's father's mother
// 12: mother's mother's father
// 13: mother's mother's mother
// 14: father's father's father's father
// 15: father's father's father's mother
// 16: father's father's mother's father
// 17: father's father's mother's mother
// 18: father's mother's father's father
// 19: father's mother's father's mother
// 20: father's mother's mother's father
// 21: father's mother's mother's mother
// 22: mother's father's father's father
// 23: mother's father's father's mother
// 24: mother's father's mother's father
// 25: mother's father's mother's mother
// 26: mother's mother's father's father
// 27: mother's mother's father's mother
// 28: mother's mother's mother's father
// 29: mother's mother's mother's mother
func (ps *PublishSet) Ancestors(p *model.Person, generations int) []*model.Person {
	n := 0
	f := 2
	for i := 0; i < generations; i++ {
		n += f
		f *= 2
	}
	a := make([]*model.Person, n)

	a[0] = p.Father
	a[1] = p.Mother
	for idx := 0; idx < n; idx++ {
		if a[idx] == nil {
			continue
		}
		if a[idx].Father != nil {
			if (idx+2)*2-2 < n {
				a[(idx+2)*2-2] = a[idx].Father
			}
		}
		if a[idx].Mother != nil {
			if (idx+2)*2-1 < n {
				a[(idx+2)*2-1] = a[idx].Mother
			}
		}
	}

	return a
}

type PersonWithNumber struct {
	Person *model.Person
	Number int // age, year etc...
}

type PersonWithGreatestNumberHeap []*PersonWithNumber

func (h PersonWithGreatestNumberHeap) Len() int           { return len(h) }
func (h PersonWithGreatestNumberHeap) Less(i, j int) bool { return h[i].Number < h[j].Number }
func (h PersonWithGreatestNumberHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PersonWithGreatestNumberHeap) Push(x interface{}) {
	*h = append(*h, x.(*PersonWithNumber))
}

func (h *PersonWithGreatestNumberHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type PersonWithLeastNumberHeap []*PersonWithNumber

func (h PersonWithLeastNumberHeap) Len() int           { return len(h) }
func (h PersonWithLeastNumberHeap) Less(i, j int) bool { return h[j].Number < h[i].Number }
func (h PersonWithLeastNumberHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PersonWithLeastNumberHeap) Push(x interface{}) {
	*h = append(*h, x.(*PersonWithNumber))
}

func (h *PersonWithLeastNumberHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

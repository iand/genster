package chart

import (
	"fmt"
	"slices"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
)

// includeSurname reports whether the surname of p should be shown when
// minimalSurnames is active.
func includeSurname(p *model.Person, minimalSurnames bool) bool {
	if !minimalSurnames || p.IsUnknown() {
		return true
	}
	if p.Father.IsUnknown() {
		if p.Mother.IsUnknown() {
			return true
		}
		return p.PreferredFamilyName != p.Mother.PreferredFamilyName
	}
	return p.PreferredFamilyName != p.Father.PreferredFamilyName
}

func excludeSingleSpouse(exclude *model.Person) model.PersonMatcher {
	return func(p *model.Person) bool {
		return !p.SameAs(exclude)
	}
}

func excludeSpouseList(excludes []*model.Person) model.PersonMatcher {
	return func(p *model.Person) bool {
		return !slices.ContainsFunc(excludes, p.SameAs) // ok, did not match any in exclude list
	}
}

func excludeAllSpouses() model.PersonMatcher {
	return func(p *model.Person) bool {
		return false
	}
}

func includeAllSpouses() model.PersonMatcher {
	return func(p *model.Person) bool {
		return true // ok, never matches a person
	}
}

// formatPersonName formats a person's name as chart headings.
func formatPersonName(p *model.Person, firstUseOfSurname bool, compact bool, minimalSurnames bool, showStars bool) []string {
	if p.IsUnknown() {
		return []string{"Not Known"}
	}
	if compact {
		var names []string
		names = append(names, p.PreferredGivenName)
		if firstUseOfSurname || includeSurname(p, minimalSurnames) {
			names = append(names, p.PreferredFamilyName)
		}
		if showStars && p.IsDirectAncestor() {
			names[len(names)-1] += "★"
		}
		return names
	}
	var name string
	if firstUseOfSurname || includeSurname(p, minimalSurnames) {
		name = p.PreferredFullName
	} else {
		name = p.PreferredGivenName
	}
	if showStars && p.IsDirectAncestor() {
		name += "★"
	}
	return []string{name}
}

// familyWhenDetails returns family event details using abbreviated when format.
func familyWhenDetails(f *model.Family) []string {
	var details []string
	if f.BestStartEvent != nil {
		details = append(details, model.AbbrevWhatWhen(f.BestStartEvent))
	}
	if f.BestEndEvent != nil {
		switch f.BestEndEvent.(type) {
		case *model.DivorceEvent, *model.AnnulmentEvent:
			details = append(details, model.AbbrevWhatWhen(f.BestEndEvent))
		}
	}
	return details
}

// familyWhereDetails returns family event details using abbreviated when+where format.
func familyWhereDetails(f *model.Family) []string {
	var details []string
	if f.BestStartEvent != nil {
		details = appendAbbrevWhatWhenWhere(details, f.BestStartEvent, false)
	}
	if f.BestEndEvent != nil {
		switch f.BestEndEvent.(type) {
		case *model.DivorceEvent, *model.AnnulmentEvent:
			details = appendAbbrevWhatWhenWhere(details, f.BestEndEvent, false)
		}
	}
	return details
}

// familyWhereDetailsCompact returns family event details using abbreviated when+where format
// in a more compact vertical format.
func familyWhereDetailsCompact(f *model.Family) []string {
	var details []string
	if f.BestStartEvent != nil {
		details = appendAbbrevWhatWhenWhere(details, f.BestStartEvent, true)
	}
	if f.BestEndEvent != nil {
		switch f.BestEndEvent.(type) {
		case *model.DivorceEvent, *model.AnnulmentEvent:
			details = appendAbbrevWhatWhenWhere(details, f.BestEndEvent, true)
		}
	}
	return details
}

func appendAbbrevWhatWhenWhere(details []string, ev model.TimelineEvent, compact bool) []string {
	if compact {
		details = append(details, model.AbbrevWhatWhen(ev))
		if !ev.GetPlace().IsUnknown() {
			details = append(details, model.AbbrevWhere(ev))
		}
	} else {
		details = append(details, model.AbbrevWhatWhenWhere(ev))
	}
	return details
}

func BuildDescendantChart(t *tree.Tree, startPerson *model.Person, detail int, depth int, compact bool, children string, parents bool, minimalSurnames bool, showStars bool) (*gtree.DescendantChart, error) {
	var personDetailFn personDetailFunc
	var familyDetailFn familyDetailFunc

	switch detail {
	case 0:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), []string{}
		}
		familyDetailFn = func(f *model.Family) []string {
			return []string{}
		}
	case 1:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			details := []string{p.VitalYears}
			if compact {
				details = appendDescendantPersonSpouses(details, p, false, compact, includeSpouse)
			}
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), details
		}
		familyDetailFn = func(f *model.Family) []string {
			var details []string
			startYear, ok := f.BestStartDate.Year()
			if ok {
				details = append(details, fmt.Sprintf("%d", startYear))
			}
			return details
		}
	case 2:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			var details []string
			if p.BestBirthlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestBirthlikeEvent))
			}
			if p.BestDeathlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestDeathlikeEvent))
			}
			if compact {
				var marriages []*model.Family
				for _, f := range p.Families {
					if f.Bond == model.FamilyBondMarried {
						marriages = append(marriages, f)
					}
				}
				for i, f := range marriages {
					if !includeSpouse(f.OtherParent(p)) {
						continue
					}
					if len(marriages) > 1 {
						details = append(details, fmt.Sprintf("+ (%d) %s", i+1, f.OtherParent(p).PreferredFamiliarFullName))
					} else {
						details = append(details, fmt.Sprintf("+ %s", f.OtherParent(p).PreferredFamiliarFullName))
					}
					details = append(details, model.AbbrevWhatWhen(f.BestStartEvent))
				}
			}
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), details
		}
		familyDetailFn = familyWhenDetails
	case 3:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			var details []string
			if p.NickName != "" {
				details = append(details, "Known as \""+p.NickName+"\"")
			}
			if p.Epithet != "" {
				details = append(details, p.Epithet)
			}
			if p.BestBirthlikeEvent != nil {
				details = appendAbbrevWhatWhenWhere(details, p.BestBirthlikeEvent, compact)
			}
			if p.BestDeathlikeEvent != nil {
				details = appendAbbrevWhatWhenWhere(details, p.BestDeathlikeEvent, compact)
			}
			if compact {
				details = appendDescendantPersonSpouses(details, p, true, compact, includeSpouse)
			}
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), details
		}
		if compact {
			familyDetailFn = familyWhereDetailsCompact
		} else {
			familyDetailFn = familyWhereDetails
		}
	default:
		return nil, fmt.Errorf("unsupported detail level: %d", detail)
	}

	seq := new(sequence)
	ch := new(gtree.DescendantChart)

	if parents && (!startPerson.Father.IsUnknown() || !startPerson.Mother.IsUnknown()) {
		f := &gtree.DescendantPerson{ID: seq.next()}
		tf := &gtree.DescendantFamily{}
		if !startPerson.Father.IsUnknown() {
			headings, details := personDetailFn(startPerson.Father, true, compact, excludeSingleSpouse(startPerson.Mother))
			f.Headings = headings
			f.Details = details

			if !startPerson.Mother.IsUnknown() {
				oh, od := personDetailFn(startPerson.Mother, true, compact, excludeSingleSpouse(startPerson.Father))
				tf.Other = &gtree.DescendantPerson{ID: seq.next(), Headings: oh, Details: od}
			}
		} else {
			headings, details := personDetailFn(startPerson.Mother, true, compact, includeAllSpouses())
			f.Headings = headings
			f.Details = details
		}
		f.Families = append(f.Families, tf)
		ch.Root = f

		child := descendants(startPerson, seq, depth, children, compact, personDetailFn, familyDetailFn)
		tf.Children = []*gtree.DescendantPerson{child}
	} else {
		ch.Root = descendants(startPerson, seq, depth, children, compact, personDetailFn, familyDetailFn)
	}

	return ch, nil
}

func BuildAncestorChart(t *tree.Tree, startPerson *model.Person, detail int, depth int, compact bool) (*gtree.AncestorChart, error) {
	var personDetailFn func(*model.Person, int) []string
	switch detail {
	case 0:
		personDetailFn = func(p *model.Person, generation int) []string {
			return []string{}
		}
	case 1:
		personDetailFn = func(p *model.Person, generation int) []string {
			var details []string
			if p.VitalYears != model.UnknownDateRangePlaceholder {
				details = append(details, p.VitalYears)
			}
			return details
		}
	case 2:
		personDetailFn = func(p *model.Person, generation int) []string {
			var details []string
			if p.Epithet != "" {
				details = append(details, p.Epithet)
			}
			if p.BestBirthlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestBirthlikeEvent))
			}
			if p.BestDeathlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestDeathlikeEvent))
			}

			return details
		}
	case 3:
		personDetailFn = func(p *model.Person, generation int) []string {
			var details []string
			if p.NickName != "" {
				details = append(details, "Known as \""+p.NickName+"\"")
			}
			if p.Epithet != "" {
				details = append(details, p.Epithet)
			}
			if generation <= depth {
				if p.BestBirthlikeEvent != nil {
					details = appendAbbrevWhatWhenWhere(details, p.BestBirthlikeEvent, compact)
				}
				if p.BestDeathlikeEvent != nil {
					details = appendAbbrevWhatWhenWhere(details, p.BestDeathlikeEvent, compact)
				}
			} else {
				if p.VitalYears != model.UnknownDateRangePlaceholder {
					details = append(details, p.VitalYears)
				}
			}

			return details
		}
	default:
		return nil, fmt.Errorf("unsupported detail level: %d", detail)

	}

	ch := new(gtree.AncestorChart)
	ch.Root = ancestors(startPerson, new(sequence), 1, depth+1, personDetailFn)
	return ch, nil
}

func BuildButterflyChart(t *tree.Tree, startPerson *model.Person) (*gtree.ButterflyChart, error) {
	ch := new(gtree.ButterflyChart)
	ch.Root = butterflyAncestors(startPerson, new(sequence), 1, 7)
	return ch, nil
}

func BuildFanChart(t *tree.Tree, startPerson *model.Person, maxGeneration int) (*gtree.FanChart, error) {
	ch := new(gtree.FanChart)
	ch.Root = fanAncestors(startPerson, new(sequence), 1, maxGeneration)
	return ch, nil
}

func BuildFocusChart(t *tree.Tree, startPerson *model.Person, detail int, minimalSurnames bool, showStars bool) (*gtree.DescendantChart, error) {
	var personDetailFn personDetailFunc
	var familyDetailFn familyDetailFunc

	switch detail {
	case 0:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), []string{}
		}
		familyDetailFn = func(f *model.Family) []string {
			return []string{}
		}
	case 1:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			details := []string{p.VitalYears}
			if compact {
				details = appendDescendantPersonSpouses(details, p, false, compact, includeSpouse)
			}
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), details
		}
		familyDetailFn = func(f *model.Family) []string {
			var details []string
			startYear, ok := f.BestStartDate.Year()
			if ok {
				details = append(details, fmt.Sprintf("%d", startYear))
			}
			return details
		}
	case 2:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			var details []string
			if p.IsDirectAncestor() {
				details = append(details, "("+text.UpperFirst(p.RelationToKeyPerson.Name())+")")
			}
			if p.BestBirthlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestBirthlikeEvent))
			}
			if p.BestDeathlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestDeathlikeEvent))
			}
			if compact {
				details = appendDescendantPersonSpouses(details, p, true, compact, includeSpouse)
			}
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), details
		}
		familyDetailFn = familyWhenDetails
	case 3:
		personDetailFn = func(p *model.Person, firstUseOfSurname bool, compact bool, includeSpouse model.PersonMatcher) ([]string, []string) {
			var details []string
			if p.IsDirectAncestor() {
				details = append(details, "("+text.UpperFirst(p.RelationToKeyPerson.Name())+")")
			}
			if p.NickName != "" {
				details = append(details, "Known as \""+p.NickName+"\"")
			}
			if p.Epithet != "" {
				details = append(details, p.Epithet)
			}
			if p.BestBirthlikeEvent != nil {
				details = appendAbbrevWhatWhenWhere(details, p.BestBirthlikeEvent, compact)
			}
			if p.BestDeathlikeEvent != nil {
				details = appendAbbrevWhatWhenWhere(details, p.BestDeathlikeEvent, compact)
			}
			if compact {
				details = appendDescendantPersonSpouses(details, p, true, compact, includeSpouse)
			}
			return formatPersonName(p, firstUseOfSurname, compact, minimalSurnames, showStars), details
		}
		familyDetailFn = familyWhereDetails
	default:
		return nil, fmt.Errorf("unsupported detail level: %d", detail)
	}

	seq := new(sequence)
	ch := new(gtree.DescendantChart)
	var startdp *gtree.DescendantPerson

	if !startPerson.Father.IsUnknown() || !startPerson.Mother.IsUnknown() {
		var parent *gtree.DescendantPerson

		f := &gtree.DescendantPerson{ID: seq.next()}
		tf := &gtree.DescendantFamily{}
		if !startPerson.Father.IsUnknown() {
			parent = newDescendantPerson(startPerson.Father, seq, personDetailFn, true, false, excludeAllSpouses())
			parent.Details = appendDescendantPersonSpouses(parent.Details, startPerson.Father, false, false, excludeSingleSpouse(startPerson.Mother))
			if !startPerson.Mother.IsUnknown() {
				tf.Other = newDescendantPerson(startPerson.Mother, seq, personDetailFn, true, false, excludeAllSpouses())
				tf.Other.Details = appendDescendantPersonSpouses(tf.Other.Details, startPerson.Mother, false, false, excludeSingleSpouse(startPerson.Father))
			}
		} else {
			parent = newDescendantPerson(startPerson.Mother, seq, personDetailFn, true, false, excludeAllSpouses())
			parent.Details = appendDescendantPersonSpouses(parent.Details, startPerson.Mother, false, false, excludeSingleSpouse(startPerson.Father))
		}
		parent.Families = append(f.Families, tf)
		ch.Root = parent

		// add siblings of start person
		if startPerson.ParentFamily != nil {
			for _, sib := range startPerson.ParentFamily.Children {
				sibdp := newDescendantPerson(sib, seq, personDetailFn, false, false, excludeAllSpouses())
				if sib.SameAs(startPerson) {
					startdp = sibdp
				} else {
					sibdp.Details = appendDescendantPersonSpouses(sibdp.Details, sib, false, false, includeAllSpouses())
				}
				tf.Children = append(tf.Children, sibdp)
			}
		}

	} else {
		startdp = newDescendantPerson(startPerson, seq, personDetailFn, true, false, excludeAllSpouses())
		ch.Root = startdp
	}

	for _, f := range startPerson.Families {
		tf := new(gtree.DescendantFamily)
		startdp.Families = append(startdp.Families, tf)

		tf.Details = familyDetailFn(f)
		o := f.OtherParent(startPerson)
		if o != nil {
			tf.Other = newDescendantPerson(o, seq, personDetailFn, true, false, excludeAllSpouses())
			tf.Other.Details = appendDescendantPersonSpouses(tf.Other.Details, o, false, false, excludeSingleSpouse(startPerson))
		}

		for _, c := range f.Children {
			chdp := newDescendantPerson(c, seq, personDetailFn, false, false, excludeAllSpouses())
			chdp.Details = appendDescendantPersonSpouses(chdp.Details, c, false, false, includeAllSpouses())
			tf.Children = append(tf.Children, chdp)
		}
	}

	return ch, nil
}

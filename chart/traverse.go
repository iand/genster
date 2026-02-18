package chart

import (
	"fmt"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/gtree"
)

type (
	personDetailFunc func(p *model.Person, firstUseOfSurname bool, compact bool) ([]string, []string)
	familyDetailFunc func(*model.Family) []string
)

func newDescendantPerson(p *model.Person, seq *sequence, personDetailFn personDetailFunc, firstUseOfSurname bool, compact bool) *gtree.DescendantPerson {
	headings, details := personDetailFn(p, firstUseOfSurname, compact)

	return &gtree.DescendantPerson{ID: seq.next(), Headings: headings, Details: details}
}

func appendDescendantPersonSpouses(details []string, p *model.Person, exclude *model.Person, inclAbbrevDetails bool) []string {
	var marriages []*model.Family
	for _, f := range p.Families {
		if f.Bond == model.FamilyBondMarried {
			marriages = append(marriages, f)
		}
	}
	for i, f := range marriages {
		if f.OtherParent(p).SameAs(exclude) {
			continue
		}
		if len(marriages) > 1 {
			details = append(details, fmt.Sprintf("+ (%d) %s", i+1, f.OtherParent(p).PreferredFamiliarFullName))
		} else {
			details = append(details, fmt.Sprintf("+ %s", f.OtherParent(p).PreferredFamiliarFullName))
		}
		if inclAbbrevDetails {
			details = append(details, model.AbbrevWhatWhenWhere(f.BestStartEvent))
		}
	}

	return details
}

func descendants(p *model.Person, seq *sequence, generations int, directOnly bool, compact bool, personDetailFn personDetailFunc, familyDetailFn familyDetailFunc) *gtree.DescendantPerson {
	tp := newDescendantPerson(p, seq, personDetailFn, seq.n == 0, compact)

	if !directOnly || p.IsDirectAncestor() {
		if generations > 0 {
			for _, f := range p.Families {
				tf := new(gtree.DescendantFamily)
				tp.Families = append(tp.Families, tf)
				// Show spouses separately unless compact has been requested
				if !compact || p.IsDirectAncestor() {
					tf.Details = familyDetailFn(f)
					o := f.OtherParent(p)
					if o != nil {
						oh, od := personDetailFn(o, true, compact)
						tf.Other = &gtree.DescendantPerson{ID: seq.next(), Headings: oh, Details: od}
					}
				}
				// TODO: sort by date
				for _, c := range f.Children {
					tf.Children = append(tf.Children, descendants(c, seq, generations-1, directOnly, compact, personDetailFn, familyDetailFn))
				}
			}
		}
		// Show both parents of direct ancestor at bottom of tree
		if p.IsDirectAncestor() && generations == 0 {
			for _, f := range p.Families {
				if f.OtherParent(p).IsDirectAncestor() {
					tf := new(gtree.DescendantFamily)
					tf.Details = familyDetailFn(f)
					tp.Families = append(tp.Families, tf)
					oh, od := personDetailFn(f.OtherParent(p), true, compact)
					tf.Other = &gtree.DescendantPerson{ID: seq.next(), Headings: oh, Details: od}
					break
				}
			}
		}
	}

	return tp
}

func ancestors(p *model.Person, seq *sequence, generation int, maxGeneration int, personDetailFn func(*model.Person, int) []string) *gtree.AncestorPerson {
	tp := &gtree.AncestorPerson{ID: seq.next(), Headings: []string{p.PreferredFullName}, Details: personDetailFn(p, generation)}
	if generation < maxGeneration {
		if p.Father != nil {
			tp.Father = ancestors(p.Father, seq, generation+1, maxGeneration, personDetailFn)
		}
		if p.Mother != nil {
			tp.Mother = ancestors(p.Mother, seq, generation+1, maxGeneration, personDetailFn)
		}
	}
	return tp
}

func butterflyAncestors(p *model.Person, seq *sequence, generation int, maxGeneration int) *gtree.ButterflyPerson {
	tp := &gtree.ButterflyPerson{
		ID:        seq.next(),
		Forenames: p.PreferredGivenName,
		Surname:   strings.ToUpper(p.PreferredFamilyName),
	}
	if len(tp.Forenames) > 13 {
		tp.Forenames = p.PreferredFamiliarName
	}

	appendPlace := func(name string, ev model.TimelineEvent) string {
		if !ev.GetPlace().IsUnknown() && !ev.GetPlace().Country.IsUnknown() {
			if !ev.GetPlace().Region.IsUnknown() && ev.GetPlace().Country.Name == "England" {
				return name + ", " + ev.GetPlace().Region.Name
			} else {
				return name + ", " + ev.GetPlace().Country.Name
			}
		}
		return name
	}

	if generation < 7 {
		if p.BestBirthlikeEvent != nil {
			tp.DetailLine1 = appendPlace(model.AbbrevWhatWhen(p.BestBirthlikeEvent), p.BestBirthlikeEvent)
			if p.BestDeathlikeEvent != nil {
				tp.DetailLine2 = appendPlace(model.AbbrevWhatWhen(p.BestDeathlikeEvent), p.BestDeathlikeEvent)
			}
		} else {
			if p.BestDeathlikeEvent != nil {
				tp.DetailLine1 = appendPlace(model.AbbrevWhatWhen(p.BestDeathlikeEvent), p.BestDeathlikeEvent)
			}
		}
	} else {
		tp.DetailLine1 = p.VitalYears
	}

	if generation < maxGeneration {
		if p.Father != nil {
			tp.Father = butterflyAncestors(p.Father, seq, generation+1, maxGeneration)
		}
		if p.Mother != nil {
			tp.Mother = butterflyAncestors(p.Mother, seq, generation+1, maxGeneration)
		}
	}
	return tp
}

func fanAncestors(p *model.Person, seq *sequence, generation int, maxGeneration int) *gtree.FanPerson {
	tp := &gtree.FanPerson{
		ID: seq.next(),
		Headings: []string{
			p.PreferredGivenName,
			strings.ToUpper(p.PreferredFamilyName),
		},
	}
	if len(tp.Headings[0]) > 13 {
		tp.Headings[0] = p.PreferredFamiliarName
	}

	appendPlace := func(name string, includeDistrict bool, ev model.TimelineEvent) string {
		if !ev.GetPlace().IsUnknown() {
			if !ev.GetPlace().Country.IsUnknown() {
				if !ev.GetPlace().Region.IsUnknown() && ev.GetPlace().Country.Name == "England" {
					if includeDistrict && !ev.GetPlace().District.IsUnknown() {
						return name + ", " + ev.GetPlace().District.Name + ", " + ev.GetPlace().Region.Name
					}
					return name + ", " + ev.GetPlace().Region.Name
				} else {
					return name + ", " + ev.GetPlace().Country.Name
				}
			}
		}
		return name
	}

	wrapText := func(s string) []string {
		const limit = 30
		var lines []string

		for len(s) > 0 {
			if len(s) <= limit {
				lines = append(lines, s)
				break
			}

			// Look for the last space within the limit
			end := limit
			for end > 0 && s[end] != ' ' {
				end--
			}

			// If no space found, force break at limit
			if end == 0 {
				end = limit
			}

			lines = append(lines, strings.TrimSpace(s[:end]))
			s = strings.TrimLeft(s[end:], " ")
		}

		return lines
	}

	if p.Epithet != "" {
		tp.Details = append(tp.Details, text.UpperFirst(p.Epithet))
	}
	if p.Notable != "" {
		tp.Details = append(tp.Details, wrapText(text.UpperFirst(p.Notable))...)
	}
	if generation < 7 {
		includeDistrict := generation < 5
		if p.BestBirthlikeEvent != nil {
			tp.Details = append(tp.Details, wrapText(appendPlace(model.AbbrevWhatWhen(p.BestBirthlikeEvent), includeDistrict, p.BestBirthlikeEvent))...)
			if p.BestDeathlikeEvent != nil {
				tp.Details = append(tp.Details, wrapText(appendPlace(model.AbbrevWhatWhen(p.BestDeathlikeEvent), includeDistrict, p.BestDeathlikeEvent))...)
			}
		} else {
			if p.BestDeathlikeEvent != nil {
				tp.Details = append(tp.Details, wrapText(appendPlace(model.AbbrevWhatWhen(p.BestDeathlikeEvent), includeDistrict, p.BestDeathlikeEvent))...)
			}
		}
	} else {
		tp.Details = append(tp.Details, p.VitalYears)
	}

	if generation < maxGeneration {
		if !p.Father.IsUnknown() && !p.Father.Unidentified {
			tp.Father = fanAncestors(p.Father, seq, generation+1, maxGeneration)
		}
		if !p.Mother.IsUnknown() && !p.Mother.Unidentified {
			tp.Mother = fanAncestors(p.Mother, seq, generation+1, maxGeneration)
		}
	}
	return tp
}

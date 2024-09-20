package chart

import (
	"fmt"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
)

func BuildDescendantChart(t *tree.Tree, startPerson *model.Person, detail int, depth int, directOnly bool, compact bool) (*gtree.DescendantChart, error) {
	var personDetailFn func(*model.Person) ([]string, []string)
	var familyDetailFn func(*model.Family) []string

	formatName := func(p *model.Person) []string {
		if p.IsUnknown() {
			return []string{"Not Known"}
		}
		if compact {
			var names []string
			names = append(names, p.PreferredGivenName)
			fname := p.PreferredFamilyName
			if p.IsDirectAncestor() {
				fname += "★"
			}
			names = append(names, fname)
			return names
		}

		name := p.PreferredFullName
		if p.IsDirectAncestor() {
			name += "★"
		}

		return []string{name}
	}

	switch detail {
	case 0:
		personDetailFn = func(p *model.Person) ([]string, []string) {
			return formatName(p), []string{}
		}
		familyDetailFn = func(p *model.Family) []string {
			return []string{}
		}
	case 1:
		personDetailFn = func(p *model.Person) ([]string, []string) {
			var details []string
			details = append(details, p.VitalYears)
			if compact {
				for i, f := range p.Families {
					if len(p.Families) > 1 {
						details = append(details, fmt.Sprintf("+ (%d) %s", i+1, f.OtherParent(p).PreferredFullName))
					} else {
						details = append(details, fmt.Sprintf("+ %s", f.OtherParent(p).PreferredFullName))
					}
				}
			}

			return formatName(p), details
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
		personDetailFn = func(p *model.Person) ([]string, []string) {
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
				for i, f := range p.Families {
					if len(p.Families) > 1 {
						details = append(details, fmt.Sprintf("+ (%d) %s", i+1, f.OtherParent(p).PreferredFullName))
					} else {
						details = append(details, fmt.Sprintf("+ %s", f.OtherParent(p).PreferredFullName))
					}
					details = append(details, model.AbbrevWhatWhen(f.BestStartEvent))
				}
			}

			return formatName(p), details
		}
		familyDetailFn = func(f *model.Family) []string {
			var details []string
			if f.BestStartEvent != nil {
				details = append(details, model.AbbrevWhatWhen(f.BestStartEvent))
			}
			if f.BestEndEvent != nil {
				details = append(details, model.AbbrevWhatWhen(f.BestEndEvent))
			}
			return details
		}
	case 3:
		personDetailFn = func(p *model.Person) ([]string, []string) {
			var details []string
			if p.IsDirectAncestor() {
				details = append(details, "("+text.UpperFirst(p.RelationToKeyPerson.Name())+")")
			}
			if p.NickName != "" {
				details = append(details, "Known as \""+p.NickName+"\"")
			}
			if p.Olb != "" {
				details = append(details, p.Olb)
			} else if p.PrimaryOccupation != "" {
				details = append(details, p.PrimaryOccupation)
			}
			if p.BestBirthlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhenWhere(p.BestBirthlikeEvent))
			}
			if p.BestDeathlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhenWhere(p.BestDeathlikeEvent))
			}

			if compact {
				for i, f := range p.Families {
					if len(p.Families) > 1 {
						details = append(details, fmt.Sprintf("+ (%d) %s", i+1, f.OtherParent(p).PreferredFullName))
					} else {
						details = append(details, fmt.Sprintf("+ %s", f.OtherParent(p).PreferredFullName))
					}
					details = append(details, model.AbbrevWhatWhenWhere(f.BestStartEvent))
				}
			}

			return formatName(p), details
		}
		familyDetailFn = func(f *model.Family) []string {
			var details []string
			if f.BestStartEvent != nil {
				details = append(details, model.AbbrevWhatWhenWhere(f.BestStartEvent))
			}
			if f.BestEndEvent != nil {
				details = append(details, model.AbbrevWhatWhenWhere(f.BestEndEvent))
			}
			return details
		}
	default:
		return nil, fmt.Errorf("unsupported detail level: %d", detail)

	}

	ch := new(gtree.DescendantChart)
	ch.Root = descendants(startPerson, new(sequence), depth, directOnly, compact, personDetailFn, familyDetailFn)
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
			if p.Olb != "" {
				details = append(details, p.Olb)
			} else if p.PrimaryOccupation != "" {
				details = append(details, p.PrimaryOccupation)
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
			if p.Olb != "" {
				details = append(details, p.Olb)
			} else if p.PrimaryOccupation != "" {
				details = append(details, p.PrimaryOccupation)
			}
			if generation <= depth {
				if p.BestBirthlikeEvent != nil {
					if p.BestBirthlikeEvent.GetPlace().IsUnknown() {
						details = append(details, model.AbbrevWhatWhen(p.BestBirthlikeEvent))
					} else {
						if compact {
							details = append(details, model.AbbrevWhatWhen(p.BestBirthlikeEvent))
							details = append(details, model.AbbrevWhere(p.BestBirthlikeEvent))
						} else {
							details = append(details, model.AbbrevWhatWhenWhere(p.BestBirthlikeEvent))
						}
					}
				}
				if p.BestDeathlikeEvent != nil {
					if p.BestDeathlikeEvent.GetPlace().IsUnknown() {
						details = append(details, model.AbbrevWhatWhen(p.BestDeathlikeEvent))
					} else {
						if compact {
							details = append(details, model.AbbrevWhatWhen(p.BestDeathlikeEvent))
							details = append(details, model.AbbrevWhere(p.BestDeathlikeEvent))
						} else {
							details = append(details, model.AbbrevWhatWhenWhere(p.BestDeathlikeEvent))
						}
					}
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

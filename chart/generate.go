package chart

import (
	"fmt"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
)

func BuildDescendantChart(t *tree.Tree, startPerson *model.Person, detail int, depth int, directOnly bool) (*gtree.DescendantChart, error) {
	var personDetailFn func(*model.Person) []string
	var familyDetailFn func(*model.Family) []string
	switch detail {
	case 0:
		personDetailFn = func(p *model.Person) []string {
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			return []string{name}
		}
		familyDetailFn = func(p *model.Family) []string {
			return []string{}
		}
	case 1:
		personDetailFn = func(p *model.Person) []string {
			var details []string
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			details = append(details, name)
			details = append(details, p.VitalYears)
			return details
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
		personDetailFn = func(p *model.Person) []string {
			var details []string
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			details = append(details, name)
			if p.IsDirectAncestor() {
				details = append(details, "("+text.UpperFirst(p.RelationToKeyPerson.Name())+")")
			}
			if p.BestBirthlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestBirthlikeEvent))
			}
			if p.BestDeathlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhen(p.BestDeathlikeEvent))
			}

			return details
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
		personDetailFn = func(p *model.Person) []string {
			var details []string
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			details = append(details, name)
			if p.IsDirectAncestor() {
				details = append(details, "("+text.UpperFirst(p.RelationToKeyPerson.Name())+")")
			}
			if p.PrimaryOccupation != "" {
				details = append(details, p.PrimaryOccupation)
			}
			if p.BestBirthlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhenWhere(p.BestBirthlikeEvent))
			}
			if p.BestDeathlikeEvent != nil {
				details = append(details, model.AbbrevWhatWhenWhere(p.BestDeathlikeEvent))
			}

			return details
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
	ch.Root = descendants(startPerson, new(sequence), depth, directOnly, personDetailFn, familyDetailFn)
	return ch, nil
}

func BuildAncestorChart(t *tree.Tree, startPerson *model.Person, detail int, depth int) (*gtree.AncestorChart, error) {
	var personDetailFn func(*model.Person, int) []string
	switch detail {
	case 0:
		personDetailFn = func(p *model.Person, generation int) []string {
			name := p.PreferredFullName
			return []string{name}
		}
	case 1:
		personDetailFn = func(p *model.Person, generation int) []string {
			var details []string
			name := p.PreferredFullName
			details = append(details, name)
			details = append(details, p.VitalYears)
			return details
		}
	case 2:
		personDetailFn = func(p *model.Person, generation int) []string {
			var details []string
			name := p.PreferredFullName
			details = append(details, name)
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
			name := p.PreferredFullName
			details = append(details, name)
			if generation < 6 {
				if p.PrimaryOccupation != "" {
					details = append(details, p.PrimaryOccupation)
				}
				if p.BestBirthlikeEvent != nil {
					if p.BestBirthlikeEvent.GetPlace().IsUnknown() {
						details = append(details, model.AbbrevWhatWhen(p.BestBirthlikeEvent))
					} else {
						details = append(details, model.AbbrevWhatWhenWhere(p.BestBirthlikeEvent))
					}
				}
				if p.BestDeathlikeEvent != nil {
					if p.BestDeathlikeEvent.GetPlace().IsUnknown() {
						details = append(details, model.AbbrevWhatWhen(p.BestDeathlikeEvent))
					} else {
						details = append(details, model.AbbrevWhatWhenWhere(p.BestDeathlikeEvent))
					}
				}
			} else {
				details = append(details, p.VitalYears)
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

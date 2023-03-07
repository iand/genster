package infer

import (
	"fmt"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
)

func RedactPersonalDetails(p *model.Person) {
	if !p.Redacted {
		p.Redacted = true
		// p.Page = ""
		if !p.RedactionKeepsName {
			p.PreferredFullName = "(living or recently deceased person)"
			p.PreferredGivenName = "(living or recently deceased person)"
			p.PreferredFamiliarName = "(living or recently deceased person)"
			p.PreferredFamilyName = "(living or recently deceased person)"
			p.PreferredSortName = "(living or recently deceased person)"
			p.PreferredUniqueName = "(living or recently deceased person)"
			p.NickName = ""
		}
		p.Olb = ""
		p.Gender = model.GenderUnknown
		p.Tags = []string{}
		p.Timeline = []model.TimelineEvent{}
		p.Occupations = []*model.Occupation{}
		p.Links = []model.Link{}
		p.VitalYears = "(?-?)"
		p.BestBirthlikeEvent = &model.BirthEvent{
			GeneralEvent:           model.GeneralEvent{Date: &gdate.Unknown{}},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
		}
		p.BestDeathlikeEvent = nil
		p.Families = []*model.Family{}
	}

	// redact all descendants
	model.RecurseDescendants(p, RedactPersonalDetails)
}

func InferPersonBirthEventDate(p *model.Person) error {
	var inference *model.Inference
	if bev, ok := p.BestBirthlikeEvent.(*model.BirthEvent); ok && gdate.IsUnknown(bev.GetDate()) {
		for _, ev := range p.Timeline {
			if gdate.SortsBefore(bev.GetDate(), ev.GetDate()) {
				continue
			}
			if yearer, ok := gdate.AsYear(ev.GetDate()); ok {
				latestYear := yearer.Year()
				switch tev := ev.(type) {
				case *model.BirthEvent:
					if tev.Principal.SameAs(p) {
						break
					}
					latestYear -= 13
					bev.Date = &gdate.BeforeYear{Y: latestYear}
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", latestYear),
						Reason: fmt.Sprintf("%s had a child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, yearer.Year()),
					}
				case *model.BaptismEvent:
					if tev.Principal.SameAs(p) {
						break
					}
					latestYear -= 13
					bev.Date = &gdate.BeforeYear{Y: latestYear}
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", latestYear),
						Reason: fmt.Sprintf("%s had child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, yearer.Year()),
					}
				case *model.MarriageEvent:
					if !tev.Party1.SameAs(p) && tev.Party2.SameAs(p) {
						break
					}
					latestYear -= 16
					bev.Date = &gdate.BeforeYear{Y: latestYear}
					other := tev.GetOther(p)
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", latestYear),
						Reason: fmt.Sprintf("%s married %s in %d", p.Gender.SubjectPronoun(), other.PreferredUniqueName, yearer.Year()),
					}
				}
			}
		}

		if inference != nil {
			bev.Inferred = true
			bev.Citations = append(bev.Citations, inference.AsCitation())
			p.Inferences = append(p.Inferences, *inference)
		}
	}

	return nil
}

func InferPersonDeathEventDate(p *model.Person) error {
	var inference *model.Inference
	if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && gdate.IsUnknown(bev.GetDate()) {
		for _, ev := range p.Timeline {
			if yearer, ok := gdate.AsYear(ev.GetDate()); ok {
				if gdate.SortsBefore(ev.GetDate(), bev.GetDate()) {
					latestYear := yearer.Year()
					switch tev := ev.(type) {
					case *model.BurialEvent:
						if tev.Principal.SameAs(p) {
							break
						}
						bev.Date = &gdate.Year{Y: latestYear}
						inference = &model.Inference{
							Type:   model.InferenceTypeYearOfDeath,
							Value:  fmt.Sprintf("%d", latestYear),
							Reason: fmt.Sprintf("%s was buried that year", p.Gender.SubjectPronoun()),
						}
					}
				} else {
					earliestYear := yearer.Year()
					switch tev := ev.(type) {
					case *model.BirthEvent:
						if tev.Principal.SameAs(p) {
							break
						}
						bev.Date = &gdate.AfterYear{Y: earliestYear}
						inference = &model.Inference{
							Type:   model.InferenceTypeYearOfDeath,
							Value:  fmt.Sprintf("after %d", earliestYear),
							Reason: fmt.Sprintf("%s had child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, yearer.Year()),
						}
					case *model.MarriageEvent:
						if !tev.Party1.SameAs(p) && tev.Party2.SameAs(p) {
							break
						}
						bev.Date = &gdate.AfterYear{Y: earliestYear}
						other := tev.GetOther(p)
						inference = &model.Inference{
							Type:   model.InferenceTypeYearOfDeath,
							Value:  fmt.Sprintf("after %d", earliestYear),
							Reason: fmt.Sprintf("%s married %s in %d", p.Gender.SubjectPronoun(), other.PreferredUniqueName, yearer.Year()),
						}
					}
				}
			}
		}

		if inference != nil {
			bev.Inferred = true
			bev.Citations = append(bev.Citations, inference.AsCitation())
			p.Inferences = append(p.Inferences, *inference)
		}

	}

	return nil
}

func InferPersonAliveOrDead(p *model.Person, year int) error {
	const maximumLifespan = 120

	if p.BestDeathlikeEvent != nil {
		if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && !gdate.IsUnknown(bev.GetDate()) {
			p.PossiblyAlive = false
			return nil
		}
	}

	if p.BestBirthlikeEvent != nil {
		if startEventYearer, ok := gdate.AsYear(p.BestBirthlikeEvent.GetDate()); ok {
			lastPossibleYearAlive := startEventYearer.Year() + maximumLifespan
			if lastPossibleYearAlive > year {
				p.PossiblyAlive = true
			} else {
				dt := &gdate.BeforeYear{Y: lastPossibleYearAlive}
				inf := model.Inference{
					Type:   model.InferenceTypeYearOfDeath,
					Value:  fmt.Sprintf("before %d", lastPossibleYearAlive),
					Reason: fmt.Sprintf("it is %d years after birth year of %d", maximumLifespan, startEventYearer.Year()),
				}

				if p.BestDeathlikeEvent == nil {
					p.BestDeathlikeEvent = &model.DeathEvent{
						GeneralEvent: model.GeneralEvent{Date: dt, Inferred: true, Citations: []*model.GeneralCitation{inf.AsCitation()}},
						GeneralIndividualEvent: model.GeneralIndividualEvent{
							Principal: p,
						},
					}
				} else if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && gdate.IsUnknown(bev.GetDate()) {
					bev.Inferred = true
					bev.Date = dt
					bev.Citations = append(bev.Citations, inf.AsCitation())
				}
				p.Inferences = append(p.Inferences, inf)
			}
		}
	}
	return nil
}

package infer

import (
	"fmt"
	"regexp"
	"strings"

	// "github.com/iand/genster/logging"
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
			GeneralEvent:           model.GeneralEvent{Date: model.UnknownDate()},
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
	if bev, ok := p.BestBirthlikeEvent.(*model.BirthEvent); ok && bev.GetDate().IsUnknown() {
		for _, ev := range p.Timeline {
			if bev.GetDate().SortsBefore(ev.GetDate()) {
				continue
			}
			if year, ok := ev.GetDate().Year(); ok {
				switch tev := ev.(type) {
				case *model.BirthEvent:
					if tev.Principal.SameAs(p) {
						break
					}
					year -= 13
					bev.Date = model.BeforeYear(year)
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", year),
						Reason: fmt.Sprintf("%s had a child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, year),
					}
				case *model.BaptismEvent:
					if tev.Principal.SameAs(p) {
						break
					}
					year -= 13
					bev.Date = model.BeforeYear(year)
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", year),
						Reason: fmt.Sprintf("%s had child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, year),
					}
				case *model.MarriageEvent:
					if !tev.Party1.SameAs(p) && tev.Party2.SameAs(p) {
						break
					}
					year -= 16
					bev.Date = model.BeforeYear(year)
					other := tev.GetOther(p)
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", year),
						Reason: fmt.Sprintf("%s married %s in %d", p.Gender.SubjectPronoun(), other.PreferredUniqueName, year),
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
	if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && bev.GetDate().IsUnknown() {
		for _, ev := range p.Timeline {
			if year, ok := ev.GetDate().Year(); ok {
				if ev.GetDate().SortsBefore(bev.GetDate()) {
					switch tev := ev.(type) {
					case *model.BurialEvent:
						if tev.Principal.SameAs(p) {
							break
						}
						bev.Date = model.Year(year)
						inference = &model.Inference{
							Type:   model.InferenceTypeYearOfDeath,
							Value:  fmt.Sprintf("%d", year),
							Reason: fmt.Sprintf("%s was buried that year", p.Gender.SubjectPronoun()),
						}
					}
				} else {
					switch tev := ev.(type) {
					case *model.BirthEvent:
						if tev.Principal.SameAs(p) {
							break
						}
						bev.Date = model.AfterYear(year)
						inference = &model.Inference{
							Type:   model.InferenceTypeYearOfDeath,
							Value:  fmt.Sprintf("after %d", year),
							Reason: fmt.Sprintf("%s had child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, year),
						}
					case *model.MarriageEvent:
						if !tev.Party1.SameAs(p) && tev.Party2.SameAs(p) {
							break
						}
						bev.Date = model.AfterYear(year)
						other := tev.GetOther(p)
						inference = &model.Inference{
							Type:   model.InferenceTypeYearOfDeath,
							Value:  fmt.Sprintf("after %d", year),
							Reason: fmt.Sprintf("%s married %s in %d", p.Gender.SubjectPronoun(), other.PreferredUniqueName, year),
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
		if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && !bev.GetDate().IsUnknown() {
			p.PossiblyAlive = false
			return nil
		}
	}

	if p.BestBirthlikeEvent != nil {
		if year, ok := p.BestBirthlikeEvent.GetDate().Year(); ok {
			lastPossibleYearAlive := year + maximumLifespan
			if lastPossibleYearAlive > year {
				p.PossiblyAlive = true
			} else {
				dt := model.BeforeYear(lastPossibleYearAlive)
				inf := model.Inference{
					Type:   model.InferenceTypeYearOfDeath,
					Value:  fmt.Sprintf("before %d", lastPossibleYearAlive),
					Reason: fmt.Sprintf("it is %d years after birth year of %d", maximumLifespan, year),
				}

				if p.BestDeathlikeEvent == nil {
					p.BestDeathlikeEvent = &model.DeathEvent{
						GeneralEvent: model.GeneralEvent{Date: dt, Inferred: true, Citations: []*model.GeneralCitation{inf.AsCitation()}},
						GeneralIndividualEvent: model.GeneralIndividualEvent{
							Principal: p,
						},
					}
				} else if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && bev.GetDate().IsUnknown() {
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

var (
	reWorkhouse = regexp.MustCompile(`(?i)\bwork.?house\b`)
	rePauper    = regexp.MustCompile(`(?i)\bPauper\b`)
)

func InferPersonGeneralFacts(p *model.Person) error {
	if p.BestBirthlikeEvent != nil {
		if !p.BestBirthlikeEvent.GetPlace().IsUnknown() {
			pl := p.BestBirthlikeEvent.GetPlace()
			if reWorkhouse.MatchString(pl.PreferredName) {
				p.BornInWorkhouse = true
				inf := model.Inference{
					Type:   model.InferenceTypeGeneralFact,
					Value:  fmt.Sprintf("born in workhouse"),
					Reason: fmt.Sprintf("place of birth appears to contains the word workhouse"),
				}
				p.Inferences = append(p.Inferences, inf)
			}
		}
	}

	if p.BestDeathlikeEvent != nil {
		if !p.BestDeathlikeEvent.GetPlace().IsUnknown() {
			pl := p.BestDeathlikeEvent.GetPlace()
			if reWorkhouse.MatchString(pl.PreferredName) {
				p.DiedInWorkhouse = true
				inf := model.Inference{
					Type:   model.InferenceTypeGeneralFact,
					Value:  fmt.Sprintf("died in workhouse"),
					Reason: fmt.Sprintf("place of death appears to contains the word workhouse"),
				}
				p.Inferences = append(p.Inferences, inf)
			}
			if rePauper.MatchString(p.BestDeathlikeEvent.GetDetail()) {
				p.Pauper = true
				inf := model.Inference{
					Type:   model.InferenceTypeGeneralFact,
					Value:  fmt.Sprintf("pauper"),
					Reason: fmt.Sprintf("detail of death appears to contains the word pauper"),
				}
				p.Inferences = append(p.Inferences, inf)
			}
		}
	}

	return nil
}

var (
	reSuicide   = regexp.MustCompile(`(?i)\bsuicide\b`)
	reLostAtSea = regexp.MustCompile(`(?i)\blost at sea\b`)
	reDrowning  = regexp.MustCompile(`(?i)\bdrown(ed|ing)\b`)
)

func InferPersonCauseOfDeath(p *model.Person) error {
	if p.BestDeathlikeEvent != nil {
		detail := strings.TrimSpace(p.BestDeathlikeEvent.GetDetail())
		if detail == "" {
			return nil
		}
		if reSuicide.MatchString(p.BestDeathlikeEvent.GetDetail()) {
			p.CauseOfDeath = model.CauseOfDeathSuicide
			inf := model.Inference{
				Type:   model.InferenceTypeCauseOfDeath,
				Value:  string(p.CauseOfDeath),
				Reason: fmt.Sprintf("detail of death event contains the word suicide"),
			}
			p.Inferences = append(p.Inferences, inf)
		} else if reLostAtSea.MatchString(p.BestDeathlikeEvent.GetDetail()) {
			p.CauseOfDeath = model.CauseOfDeathLostAtSea
			inf := model.Inference{
				Type:   model.InferenceTypeCauseOfDeath,
				Value:  string(p.CauseOfDeath),
				Reason: fmt.Sprintf("detail of death event contains the words lost at sea"),
			}
			p.Inferences = append(p.Inferences, inf)
		} else if reDrowning.MatchString(p.BestDeathlikeEvent.GetDetail()) {
			p.CauseOfDeath = model.CauseOfDeathDrowned
			inf := model.Inference{
				Type:   model.InferenceTypeCauseOfDeath,
				Value:  string(p.CauseOfDeath),
				Reason: fmt.Sprintf("detail of death event contains the words drowned or drowning"),
			}
			p.Inferences = append(p.Inferences, inf)
		} else {
			// logging.Dump("death detail: " + p.BestDeathlikeEvent.GetDetail())
		}
	}
	return nil
}

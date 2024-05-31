package infer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
)

func RedactPersonalDetailsWithDescendants(p *model.Person) {
	model.RecurseDescendantsAndApply(p, RedactPersonalDetails)
}

func RedactPersonalDetails(p *model.Person) (bool, error) {
	logging.Debug("redacting person", "id", p.ID, "name", p.PreferredFullName)
	p.Redacted = true
	p.RedactNames("(living or recently deceased person)")

	p.Olb = "information withheld to preserve privacy"
	p.Gender = model.GenderUnknown
	p.Tags = []string{}
	p.Timeline = []model.TimelineEvent{}
	p.Occupations = []*model.Occupation{}
	p.Links = []model.Link{}
	p.VitalYears = "-?-"

	var birthDecade int
	var hasBirthDecade bool
	if p.BestBirthlikeEvent != nil {
		birthDecade, hasBirthDecade = p.BestBirthlikeEvent.GetDate().DecadeStart()
	}

	if hasBirthDecade {
		birthDate := model.WithinDecade(birthDecade)
		p.RedactNames("(person born " + birthDate.When() + ")")

		if p.BestDeathlikeEvent != nil {
			if deathDecade, ok := p.BestDeathlikeEvent.GetDate().DecadeStart(); ok && deathDecade != birthDecade {
				deathDate := model.WithinDecade(deathDecade)
				p.RedactNames("(person lived " + birthDate.String() + " to " + deathDate.String() + ")")
			}
		}
	} else {
		if p.BestDeathlikeEvent != nil {
			if decade, ok := p.BestDeathlikeEvent.GetDate().DecadeStart(); ok {
				deathDate := model.WithinDecade(decade)
				p.RedactNames("(person died " + deathDate.When() + ")")
			}
		}
	}

	p.BestBirthlikeEvent = &model.BirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.UnknownDate()},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}

	if p.BestDeathlikeEvent != nil {
		p.BestDeathlikeEvent = &model.DeathEvent{
			GeneralEvent:           model.GeneralEvent{Date: model.UnknownDate()},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
		}
	}
	p.Families = []*model.Family{}
	p.Children = []*model.Person{}

	return false, nil
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
					inferredYear := year - 13
					bev.Date = model.BeforeYear(inferredYear)
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", inferredYear),
						Reason: fmt.Sprintf("%s had a child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, year),
					}
				case *model.BaptismEvent:
					if tev.Principal.SameAs(p) {
						break
					}
					inferredYear := year - 13
					bev.Date = model.BeforeYear(inferredYear)
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", inferredYear),
						Reason: fmt.Sprintf("%s had child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, year),
					}
				case *model.MarriageEvent:
					if !tev.Party1.Person.SameAs(p) && tev.Party2.Person.SameAs(p) {
						break
					}
					inferredYear := year - 16
					bev.Date = model.BeforeYear(year)
					other := tev.GetOther(p)
					inference = &model.Inference{
						Type:   model.InferenceTypeYearOfBirth,
						Value:  fmt.Sprintf("before %d", inferredYear),
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
	// do we already have a BestDeathlikeEvent?
	if p.BestDeathlikeEvent != nil && !p.BestDeathlikeEvent.GetDate().IsUnknown() {
		logging.Debug("not inferring death event date", "id", p.ID)
		// best deathlike event has a known date, so we're good with that
		return nil
	}

	const longLifespan = 95

	probableLastDeathYear := 0
	if byear, ok := p.BestBirthDate().Year(); ok {
		probableLastDeathYear = byear + longLifespan
	}

	var dt *model.Date
	var inference *model.Inference

	latestYear := 0
eventloop:
	for _, ev := range p.Timeline {
		if year, ok := ev.GetDate().Year(); ok {
			switch tev := ev.(type) {
			case *model.BurialEvent:
				if !tev.Principal.SameAs(p) {
					break
				}
				logging.Debug("inferring death date from burial event", "id", p.ID)
				dt = model.Year(year)
				inference = &model.Inference{
					Type:   model.InferenceTypeYearOfDeath,
					Value:  fmt.Sprintf("%d", year),
					Reason: fmt.Sprintf("%s was buried that year", p.Gender.SubjectPronoun()),
				}
				break eventloop
			case *model.CremationEvent:
				if !tev.Principal.SameAs(p) {
					break
				}
				logging.Debug("inferring death date from cremation event", "id", p.ID)
				dt = model.Year(year)
				inference = &model.Inference{
					Type:   model.InferenceTypeYearOfDeath,
					Value:  fmt.Sprintf("%d", year),
					Reason: fmt.Sprintf("%s was buried that year", p.Gender.SubjectPronoun()),
				}
				break eventloop
			case *model.MarriageEvent:
				if !tev.DirectlyInvolves(p) {
					break
				}
				if year < latestYear {
					break
				}
				latestYear = year
				if probableLastDeathYear != 0 && probableLastDeathYear > year {
					dt = model.YearRange(year, probableLastDeathYear)
				} else {
					dt = model.AfterYear(year)
				}
				other := tev.GetOther(p)
				inference = &model.Inference{
					Type:   model.InferenceTypeYearOfDeath,
					Value:  fmt.Sprintf("after %d", year),
					Reason: fmt.Sprintf("%s married %s in %d", p.Gender.SubjectPronoun(), other.PreferredUniqueName, year),
				}
			case *model.CensusEvent:
				if !tev.DirectlyInvolves(p) {
					break
				}
				if year < latestYear {
					break
				}
				latestYear = year
				if probableLastDeathYear != 0 && probableLastDeathYear > year {
					dt = model.YearRange(year, probableLastDeathYear)
				} else {
					dt = model.AfterYear(year)
				}
				inference = &model.Inference{
					Type:   model.InferenceTypeYearOfDeath,
					Value:  fmt.Sprintf("after %d", year),
					Reason: fmt.Sprintf("%s appeared in the %d census", p.Gender.SubjectPronoun(), year),
				}
			case *model.BirthEvent:
				// ignore own birth
				if tev.DirectlyInvolves(p) {
					break
				}
				if year < latestYear {
					break
				}
				latestYear = year
				if probableLastDeathYear != 0 && probableLastDeathYear > year {
					dt = model.YearRange(year, probableLastDeathYear)
				} else {
					dt = model.AfterYear(year)
				}
				inference = &model.Inference{
					Type:   model.InferenceTypeYearOfDeath,
					Value:  fmt.Sprintf("after %d", year),
					Reason: fmt.Sprintf("%s had child, %s, in %d", p.Gender.SubjectPronoun(), tev.Principal.PreferredUniqueName, year),
				}
			}
		}
	}

	// if dt.IsUnknown() {
	// 	// use children's birth dates
	// 	for _, c := range p.Children {
	// 		if c.BestBirthlikeEvent == nil {
	// 			continue
	// 		}
	// 		if c.BestBirthlikeEvent.GetDate().IsUnknown() {
	// 			continue
	// 		}

	// 		year, ok := c.BestBirthlikeEvent.GetDate().Year()
	// 		if !ok {
	// 			continue
	// 		}
	// 		if year > latestYear {
	// 			year = latestYear
	// 			if probableLastDeathYear != 0 && probableLastDeathYear > year {
	// 				dt = model.YearRange(year, probableLastDeathYear)
	// 			} else {
	// 				dt = model.AfterYear(year)
	// 			}
	// 			inference = &model.Inference{
	// 				Type:   model.InferenceTypeYearOfDeath,
	// 				Value:  fmt.Sprintf("after %d", year),
	// 				Reason: fmt.Sprintf("%s had child, %s, in %d", p.Gender.SubjectPronoun(), c.PreferredUniqueName, year),
	// 			}
	// 		}
	// 	}
	// }

	if inference != nil && !dt.IsUnknown() {
		switch tev := p.BestDeathlikeEvent.(type) {
		case nil:
			// no deathlike event so synthesise one
			logging.Debug("adding inferred best deathlike event", "id", p.ID)
			dev := &model.DeathEvent{
				GeneralEvent: model.GeneralEvent{
					Date:      dt,
					Inferred:  true,
					Citations: []*model.GeneralCitation{inference.AsCitation()},
				},
				GeneralIndividualEvent: model.GeneralIndividualEvent{
					Principal: p,
				},
			}
			p.BestDeathlikeEvent = dev

		case *model.DeathEvent:
			// set the date to the inferred one
			logging.Debug("inferring death year from other events", "id", p.ID)
			tev.Date = dt
			tev.Inferred = true
			tev.Citations = append(tev.Citations, inference.AsCitation())
		case *model.BurialEvent:
			logging.Debug("inferring burial year from other events", "id", p.ID)
			tev.Date = dt
			tev.Inferred = true
			tev.Citations = append(tev.Citations, inference.AsCitation())
		case *model.CremationEvent:
			logging.Debug("inferring cremation year from other events", "id", p.ID)
			tev.Date = dt
			tev.Inferred = true
			tev.Citations = append(tev.Citations, inference.AsCitation())
		default:
			logging.Error("unexpected deathlike event", "type", fmt.Sprintf("%T", p.BestDeathlikeEvent), "id", p.ID, "name", p.PreferredUniqueName)
			return fmt.Errorf("unexpected deathlike event: %T", p.BestDeathlikeEvent)
		}

		p.Inferences = append(p.Inferences, *inference)
	}

	return nil
}

func InferPersonAliveOrDead(p *model.Person, year int) error {
	const maximumLifespan = 120

	lastPossibleYearAlive := year

	// deathInferenceReason := ""
	if byear, ok := p.BestBirthDate().Year(); ok {
		lastPossibleYearAlive = byear + maximumLifespan
		// deathInferenceReason = fmt.Sprintf("it is %d years after birth year of %d", maximumLifespan, byear)
	} else {
		lastEventYear := 0
		for _, ev := range p.Timeline {
			if ev.DirectlyInvolves(p) {
				if eyear, ok := ev.GetDate().Year(); ok {
					if eyear > lastEventYear {
						lastEventYear = eyear
					}
				}
			}
		}

		if lastEventYear > 0 {
			lastPossibleYearAlive = lastEventYear + maximumLifespan
			// deathInferenceReason = fmt.Sprintf("it is %d years after the last event involving this person", maximumLifespan)
		}

	}

	if lastPossibleYearAlive < year {
		// set person and all ancestors as historic
		if err := model.ApplyAndRecurseAncestors(p, func(a *model.Person) (bool, error) {
			if !a.Historic {
				logging.Debug("marking person as historic since they lived more than one lifespan ago", "id", a.ID)
			}
			a.Historic = true
			a.PossiblyAlive = false
			return true, nil
		}); err != nil {
			return fmt.Errorf("mark person as historic: %w", err)
		}

		// dt := model.BeforeYear(lastPossibleYearAlive)
		// inf := model.Inference{
		// 	Type:   model.InferenceTypeYearOfDeath,
		// 	Value:  fmt.Sprintf("before %d", lastPossibleYearAlive),
		// 	Reason: deathInferenceReason,
		// }

		// if p.BestDeathlikeEvent == nil {
		// 	p.BestDeathlikeEvent = &model.DeathEvent{
		// 		GeneralEvent: model.GeneralEvent{Date: dt, Inferred: true, Citations: []*model.GeneralCitation{inf.AsCitation()}},
		// 		GeneralIndividualEvent: model.GeneralIndividualEvent{
		// 			Principal: p,
		// 		},
		// 	}
		// } else if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && bev.GetDate().IsUnknown() {
		// 	bev.Inferred = true
		// 	bev.Date = dt
		// 	bev.Citations = append(bev.Citations, inf.AsCitation())
		// }
		// p.Inferences = append(p.Inferences, inf)
	}

	if p.Historic {
		logging.Debug("marking person as not alive since they are marked as historic", "id", p.ID)
		p.PossiblyAlive = false
		return nil
	}

	if p.BestDeathlikeEvent != nil {
		if bev, ok := p.BestDeathlikeEvent.(*model.DeathEvent); ok && !bev.GetDate().IsUnknown() {
			logging.Debug("marking person as not alive since they have a deathlike event with a known date", "id", p.ID)
			p.PossiblyAlive = false
			return nil
		}
	}

	logging.Debug("marking person as possibly alive", "id", p.ID)
	p.PossiblyAlive = true

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
					Value:  "born in workhouse",
					Reason: "place of birth appears to contains the word workhouse",
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
					Value:  "died in workhouse",
					Reason: "place of death appears to contains the word workhouse",
				}
				p.Inferences = append(p.Inferences, inf)
			}
			if rePauper.MatchString(p.BestDeathlikeEvent.GetDetail()) {
				p.Pauper = true
				inf := model.Inference{
					Type:   model.InferenceTypeGeneralFact,
					Value:  "pauper",
					Reason: "detail of death appears to contains the word pauper",
				}
				p.Inferences = append(p.Inferences, inf)
			}
		}
	}

	if yrs, ok := p.AgeInYearsAtDeath(); ok && yrs < 18 {
		p.DiedYoung = true
	}

	return nil
}

var (
	reSuicide   = regexp.MustCompile(`(?i)\bsuicide\b`)
	reLostAtSea = regexp.MustCompile(`(?i)\blost at sea\b`)
	reDrowning  = regexp.MustCompile(`(?i)\bdrown(ed|ing)\b`)
)

func InferPersonCauseOfDeath(p *model.Person) error {
	if p.CauseOfDeath != nil {
		return nil
	}
	if p.BestDeathlikeEvent != nil {
		detail := strings.TrimSpace(p.BestDeathlikeEvent.GetDetail())
		if detail == "" {
			return nil
		}
		if reSuicide.MatchString(p.BestDeathlikeEvent.GetDetail()) {
			p.ModeOfDeath = model.ModeOfDeathSuicide
			inf := model.Inference{
				Type:   model.InferenceTypeModeOfDeath,
				Value:  string(model.ModeOfDeathSuicide),
				Reason: "detail of death event contains the word suicide",
			}
			p.Inferences = append(p.Inferences, inf)
		} else if reLostAtSea.MatchString(p.BestDeathlikeEvent.GetDetail()) {
			p.ModeOfDeath = model.ModeOfDeathLostAtSea
			inf := model.Inference{
				Type:   model.InferenceTypeModeOfDeath,
				Value:  string(model.ModeOfDeathLostAtSea),
				Reason: "detail of death event contains the words lost at sea",
			}
			p.Inferences = append(p.Inferences, inf)
		} else if reDrowning.MatchString(p.BestDeathlikeEvent.GetDetail()) {
			p.ModeOfDeath = model.ModeOfDeathDrowned
			inf := model.Inference{
				Type:   model.InferenceTypeModeOfDeath,
				Value:  string(model.ModeOfDeathDrowned),
				Reason: "detail of death event contains the words drowned or drowning",
			}
			p.Inferences = append(p.Inferences, inf)
		} else {
			// logging.Dump("death detail: " + p.BestDeathlikeEvent.GetDetail())
		}
	}
	return nil
}

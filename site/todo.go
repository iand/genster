package site

import (
	"fmt"

	"github.com/iand/genster/model"
)

func (s *Site) ScanPersonTodos(p *model.Person) []*model.ToDo {
	var todos []*model.ToDo

	if p.PreferredFamilyName == model.UnknownNamePlaceholder || p.PreferredFamilyName == "" {
		p.ToDos = append(p.ToDos, &model.ToDo{
			Category: model.ToDoCategoryMissing,
			Context:  "surname",
			Goal:     "Find the person's surname",
			Reason:   "Surname is missing",
		})
	}

	if p.PreferredGivenName == model.UnknownNamePlaceholder || p.PreferredGivenName == "" {
		p.ToDos = append(p.ToDos, &model.ToDo{
			Category: model.ToDoCategoryMissing,
			Context:  "forename",
			Goal:     "Find the person's forename",
			Reason:   "Forename is missing",
		})
	}

	if p.BestBirthlikeEvent == nil || p.BestBirthlikeEvent.GetDate().IsUnknown() {
		p.ToDos = append(p.ToDos, &model.ToDo{
			Category: model.ToDoCategoryMissing,
			Context:  "birth",
			Goal:     "Find the person's birth or baptism date",
			Reason:   "No date for the person's birth is known",
		})
	} else if p.BestBirthlikeEvent != nil && p.BestBirthlikeEvent.IsInferred() {
		p.ToDos = append(p.ToDos, &model.ToDo{
			Category: model.ToDoCategoryMissing,
			Context:  "birth",
			Goal:     "Find the person's birth or baptism date",
			Reason:   fmt.Sprintf("No date is known but it is inferred to be %s", p.BestBirthlikeEvent.GetDate().When()),
		})
	} else if p.BestBirthlikeEvent != nil && !p.BestBirthlikeEvent.GetDate().IsFirm() {
		p.ToDos = append(p.ToDos, &model.ToDo{
			Category: model.ToDoCategoryMissing,
			Context:  "birth",
			Goal:     "Find a firm date for the person's birth or baptism",
			Reason:   fmt.Sprintf("Only the approximate date %q is known", p.BestBirthlikeEvent.GetDate().When()),
		})
	}

	if !p.PossiblyAlive {
		if p.BestDeathlikeEvent == nil || p.BestDeathlikeEvent.GetDate().IsUnknown() {
			p.ToDos = append(p.ToDos, &model.ToDo{
				Category: model.ToDoCategoryMissing,
				Context:  "death",
				Goal:     "Find the person's death or burial date",
				Reason:   "No date for the person's death is known",
			})
		} else if p.BestDeathlikeEvent != nil && p.BestDeathlikeEvent.IsInferred() {
			p.ToDos = append(p.ToDos, &model.ToDo{
				Category: model.ToDoCategoryMissing,
				Context:  "death",
				Goal:     "Find the person's death or burial date",
				Reason:   fmt.Sprintf("No date is known but it is inferred to be %s", p.BestDeathlikeEvent.GetDate().When()),
			})
		} else if p.BestDeathlikeEvent != nil && !p.BestDeathlikeEvent.GetDate().IsFirm() {
			p.ToDos = append(p.ToDos, &model.ToDo{
				Category: model.ToDoCategoryMissing,
				Context:  "death",
				Goal:     "Find a firm date for the person's death or burial",
				Reason:   fmt.Sprintf("Only the approximate date %q is known", p.BestDeathlikeEvent.GetDate().When()),
			})
		}
	}

	if p.IsDirectAncestor() {
		if p.Father.IsUnknown() && !p.Illegitimate {
			p.ToDos = append(p.ToDos, &model.ToDo{
				Category: model.ToDoCategoryMissing,
				Context:  "father",
				Goal:     "Find the person's father",
				Reason:   "No father is known and the person is not known to be illegitimate",
			})
		}
		if p.Mother.IsUnknown() {
			p.ToDos = append(p.ToDos, &model.ToDo{
				Category: model.ToDoCategoryMissing,
				Context:  "mother",
				Goal:     "Find the person's mother",
				Reason:   "No mother is known",
			})
		}
	}

	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) {
			continue
		}

		// TODO: will transcription

		hasCitations := len(ev.GetCitations()) > 0
		unreliableCitations := 0
		censusCitations := 0
		civilRegistrationCitations := 0
		transcribedCitations := 0
		for _, c := range ev.GetCitations() {
			if c.Source.IsUnknown() {
				continue
			}
			if len(c.TranscriptionText) > 0 {
				transcribedCitations++
			}
			if c.Source.IsCensus {
				censusCitations++
			}
			if c.Source.IsCivilRegistration {
				civilRegistrationCitations++
				if len(c.TranscriptionText) == 0 {
					p.ToDos = append(p.ToDos, &model.ToDo{
						Category: model.ToDoCategoryCitations,
						Context:  ev.Type() + " event",
						Goal:     "Transcribe the civil registration document.",
						Reason:   "a citation for a civil registration certificate was found but it had no attached transcription",
					})
				}
			}
			if c.Source.IsUnreliable {
				unreliableCitations++
			}
		}

		hasOnlyCensusCitations := hasCitations && censusCitations == len(ev.GetCitations())
		hasOnlyUnreliableCitations := hasCitations && unreliableCitations == len(ev.GetCitations())
		hasCivilRegistrationCitation := civilRegistrationCitations > 0
		hasTranscribedCitation := transcribedCitations > 0

		if hasOnlyUnreliableCitations {
			p.ToDos = append(p.ToDos, &model.ToDo{
				Category: model.ToDoCategoryCitations,
				Context:  ev.Type() + " event",
				Goal:     "Find a more reliable source for this event.",
				Reason:   "all sources for this event are deemed unreliable",
			})
		}

		// births should have more than just census sources
		if _, ok := ev.(*model.BirthEvent); ok {
			if hasOnlyCensusCitations {
				p.ToDos = append(p.ToDos, &model.ToDo{
					Category: model.ToDoCategoryCitations,
					Context:  ev.Type() + " event",
					Goal:     "Find a non-census source for this event.",
					Reason:   "census records appear to be the only source of this event but a direct record of birth or baptism is preferred",
				})
			}
		}

		if ev.GetDate().IsFirm() {
			if !hasCitations {
				p.ToDos = append(p.ToDos, &model.ToDo{
					Category: model.ToDoCategoryCitations,
					Context:  ev.Type() + " event",
					Goal:     "Find a source for this event.",
					Reason:   fmt.Sprintf("event appears to have a firm date %q but no source citation", ev.GetDate().String()),
				})
			}

			// for direct ancestors only
			if p.IsDirectAncestor() {
				// if in UK and date >= 1837 look for a GRO citation
				switch ev.(type) {
				case *model.BirthEvent, *model.DeathEvent, *model.MarriageEvent:
					if ev.GetPlace().IsUnknown() || ev.GetPlace().UKNationName.IsUnknown() {
						break
					}

					yr, ok := ev.GetDate().Year()
					if !ok || yr < 1837 {
						break
					}

					if !hasCivilRegistrationCitation {
						var goal string
						if _, ok := ev.(*model.MarriageEvent); ok {
							goal = fmt.Sprintf("obtain a copy of the certificate for the marriage in %d", yr)
						} else {
							goal = fmt.Sprintf("obtain a copy of the %s certificate", ev.Type())
						}

						p.ToDos = append(p.ToDos, &model.ToDo{
							Category: model.ToDoCategoryRecords,
							Context:  fmt.Sprintf("%s event", ev.Type()),
							Goal:     goal,
							Reason:   "the date and place of the event is known and it is within the period of Civil Registration in the United Kingdom, so a copy of the relevant certificate can be requested",
						})
					}
				}
			}
		}

		switch ev.(type) {
		case *model.BirthEvent, *model.DeathEvent, *model.BaptismEvent, *model.BurialEvent, *model.CremationEvent, *model.MarriageEvent:
			if ev.GetPlace().IsUnknown() && !ev.GetDate().IsUnknown() {
				p.ToDos = append(p.ToDos, &model.ToDo{
					Category: model.ToDoCategoryMissing,
					Context:  ev.Type() + " event",
					Goal:     "Find the place for this event.",
					Reason:   fmt.Sprintf("event has a date %q but the place is unknown", ev.GetDate().String()),
				})
			}
		case *model.WillEvent:
			if !hasCitations {
				p.ToDos = append(p.ToDos, &model.ToDo{
					Category: model.ToDoCategoryCitations,
					Context:  ev.Type() + " event",
					Goal:     "Find a source for this will.",
					Reason:   "there are no source citations for the will",
				})
			}
			if !hasTranscribedCitation {
				p.ToDos = append(p.ToDos, &model.ToDo{
					Category: model.ToDoCategoryMissing,
					Context:  ev.Type() + " event",
					Goal:     "Transcribe the will.",
					Reason:   "a citation for a will was found but it had no attached transcription",
				})
			}
			// case *model.ProbateEvent:
			// 	if !hasCitations {
			// 		p.ToDos = append(p.ToDos, &model.ToDo{
			// 			Category: model.ToDoCategoryCitations,
			// 			Context:  ev.Type() + " event",
			// 			Goal:     "Find a source for this event.",
			// 			Reason:   "there are no source citations for the will",
			// 		})
			// 	}
			// 	if !hasTranscribedCitation {
			// 		p.ToDos = append(p.ToDos, &model.ToDo{
			// 			Category: model.ToDoCategoryMissing,
			// 			Context:  ev.Type() + " event",
			// 			Goal:     "Transcribe the probate document.",
			// 			Reason:   "a citation for a will was found but it had no attached transcription",
			// 		})
			// 	}
		}
	}

	return todos
}

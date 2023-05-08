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
				Reason:   fmt.Sprintf("No date is known but it is inferred to be %s", p.BestBirthlikeEvent.GetDate().When()),
			})
		}
	}

	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) {
			continue
		}

		if len(ev.GetCitations()) == 0 && ev.GetDate().IsFirm() {
			p.ToDos = append(p.ToDos, &model.ToDo{
				Category: model.ToDoCategoryCitations,
				Context:  ev.Type() + " event",
				Goal:     "Find a source for this event.",
				Reason:   fmt.Sprintf("event appears to have a firm date %q but no source citation", ev.GetDate().String()),
			})
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
		}
	}

	return todos
}

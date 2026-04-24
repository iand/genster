package narrative

import (
	"fmt"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

type NarrativeTimelineEntryFormatter[T render.EncodedText] struct {
	pov      *model.POV
	enc      render.TextEncoder[T]
	omitDate bool
	logger   *logging.Logger
	nc       NameChooser
}

func NewNarrativeTimelineEntryFormatter[T render.EncodedText](pov *model.POV, enc render.TextEncoder[T], logger *logging.Logger, nc NameChooser, omitDate bool) *NarrativeTimelineEntryFormatter[T] {
	return &NarrativeTimelineEntryFormatter[T]{
		pov:      pov,
		enc:      enc,
		omitDate: omitDate,
		logger:   logger,
		nc:       nc,
	}
}

func (t *NarrativeTimelineEntryFormatter[T]) Title(seq int, ev model.TimelineEvent) string {
	title := ""
	switch tev := ev.(type) {
	case *model.CensusEvent:
		title = t.censusEventTitle(seq, tev)
	case *model.BirthEvent:
		title = t.vitalEventTitle(seq, tev)
	case *model.BaptismEvent:
		title = t.vitalEventTitle(seq, tev)
	case *model.DeathEvent:
		title = t.vitalEventTitle(seq, tev)
	case *model.BurialEvent:
		title = t.vitalEventTitle(seq, tev)
	case *model.CremationEvent:
		title = t.vitalEventTitle(seq, tev)
	case *model.ProbateEvent:
		title = t.probateEventTitle(seq, tev)
	case *model.WillEvent:
		title = t.willEventTitle(seq, tev)
	case *model.IndividualNarrativeEvent:
		// does not appear in timeline
		return ""
	case *model.ResidenceRecordedEvent:
		title = t.residenceEventTitle(seq, tev)
	case *model.MarriageEvent:
		title = t.marriageEventTitle(seq, tev)
	case *model.MarriageLicenseEvent:
		title = t.marriageEventTitle(seq, tev)
	case *model.MarriageBannsEvent:
		title = t.marriageEventTitle(seq, tev)
	case *model.ArrivalEvent:
		title = t.arrivalEventTitle(seq, tev)
	case *model.DepartureEvent:
		title = t.departureEventTitle(seq, tev)
	case *model.InstitutionEntryEvent:
		title = t.institutionEntryEventTitle(seq, tev)
	case *model.InstitutionDepartureEvent:
		title = t.institutionDepartureEventTitle(seq, tev)
	case *model.DivorceEvent:
		t.logger.Debug("timeline: unhandled divorce event")
	case *model.MusterEvent:
		title = t.musterEventTitle(seq, tev)
		// title = t.generalEventTitle(seq, tev)
	default:
		title = t.generalEventTitle(seq, tev)
		// t.logger.Debug("timeline: unhandled event type", "type", fmt.Sprintf("%T", ev), "title", title)
	}
	if title == "" {
		t.logger.Debug("timeline: ignored event type", "type", fmt.Sprintf("%T", ev))
		return ""
	}

	title = text.FormatSentence(title)
	title = t.enc.EncodeWithCitations(t.enc.EncodeText(title), ev.GetCitations()).String()
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) Detail(seq int, ev model.TimelineEvent) string {
	switch tev := ev.(type) {
	default:
		return tev.GetDetail()
	}
}

func (t *NarrativeTimelineEntryFormatter[T]) whenWhat(ev model.TimelineEvent) string {
	date := ev.GetDate()
	if !date.IsUnknown() {
		switch ev.GetDate().Derivation {
		case model.DateDerivationEstimated, model.DateDerivationCalculated:
			return model.ConditionalWhat(ev, "probably") + " around this time"
		}
	}
	return model.What(ev)
}

func (t *NarrativeTimelineEntryFormatter[T]) vitalEventTitle(seq int, ev model.IndividualTimelineEvent) string {
	var title string
	var trailer string
	what := t.whenWhat(ev)

	eps := ev.GetParticipant(t.pov.Person)
	if len(eps) == 0 {
		// Does not participate so this is someone else's event
		obsContext := t.observerContext(ev, false)
		title = text.JoinSentenceParts(title, obsContext, what)
	} else if len(eps) > 0 {
		ep := eps[0]
		switch ep.Role {
		case model.EventRolePrincipal:
			// This is the POV person's vital event
			title = text.JoinSentenceParts(title, what)

			date := ev.GetDate()

			includeAge := true
			if t.omitDate || date.IsUnknown() {
				includeAge = false
			}
			switch ev.(type) {
			case *model.BirthEvent:
				includeAge = false
			case *model.BurialEvent:
				includeAge = false
			case *model.CremationEvent:
				includeAge = false
			}

			// add their age, if its not their earliest event
			if includeAge {
				if ev != t.pov.Person.BestBirthlikeEvent {
					if age, ok := t.pov.Person.AgeInYearsAt(date); ok {
						title = text.JoinSentenceParts(title, AgeQualifier(age))
					}
				}
			}
			switch ev.(type) {
			case *model.BirthEvent, *model.DeathEvent:
				infs := ev.GetParticipantsByRole(model.EventRoleInformant)
				if len(infs) == 1 {
					name := t.enc.EncodeModelLinkNamed(infs[0].Person, t.nc, t.pov).String()
					rel := infs[0].Person.RelationTo(t.pov.Person, ev.GetDate())
					if rel != "" {
						rel = t.pov.Person.Gender.PossessivePronounSingular() + " " + rel
					}
					trailer = text.JoinSentenceParts("the informant was", rel, name)
				}
			}

		case model.EventRoleInformant:
			obsContext := t.observerContext(ev, true)
			title = text.JoinSentenceParts(title, obsContext, what)
			trailer = text.JoinSentenceParts(t.pov.Person.Gender.SubjectPronounWithLink(), "the informant")
		default:
			t.logger.Warn("unsupported vital event role", "role", ep.Role)
			obsContext := t.observerContext(ev, true)
			title = text.JoinSentenceParts(title, obsContext, what)
		}
	}

	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = WhatWhere(title, pl, t.enc, t.nc)
	}

	if trailer != "" {
		title = text.JoinSentences(title, trailer)
	}

	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) censusEventTitle(seq int, ev *model.CensusEvent) string {
	var title string
	title = text.JoinSentenceParts(title, t.observerContext(ev, false), "recorded in the")
	year, ok := ev.GetDate().Year()
	if ok {
		title = text.JoinSentenceParts(title, fmt.Sprintf("%d ", year))
	}
	title = text.JoinSentenceParts(title, "census")

	entry, _ := ev.Entry(t.pov.Person)
	if entry != nil && entry.RelationToHead.IsImpersonal() {
		title = text.JoinSentenceParts(title, "as", "a"+text.MaybeAn(string(entry.RelationToHead)))
	}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		residing := ChooseFrom(seq, "residing", "", "living")
		title = text.JoinSentenceParts(title, residing)
		title = WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) generalEventTitle(seq int, ev model.TimelineEvent) string {
	if ev.What() == "" {
		return ""
	}
	var title string
	title = text.JoinSentenceParts(title, t.observerContext(ev, false))
	title = text.JoinSentenceParts(title, t.whenWhat(ev))

	pl := ev.GetPlace()
	title = WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) probateEventTitle(seq int, ev model.TimelineEvent) string {
	var title string
	what := t.whenWhat(ev)
	pl := ev.GetPlace()

	eps := ev.GetParticipant(t.pov.Person)
	if len(eps) == 0 {
		// Does not participate so this is someone else's event
		title = what
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentenceParts(title, "here")
		}

		obsContext := t.observerContext(ev, true)
		if obsContext != "" {
			title = text.JoinSentenceParts(title, "for", obsContext)
		}

		if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
			title = text.JoinSentenceParts(title, WhatWhere("", pl, t.enc, t.nc), "probate office")
		}
	} else if len(eps) > 0 {
		ep := eps[0]
		switch ep.Role {
		case model.EventRolePrincipal:
			title = what
			if pl.SameAs(t.pov.Place) {
				title = text.JoinSentenceParts(title, "here")
			}
			if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
				title = text.JoinSentenceParts(title, WhatWhere("", pl, t.enc, t.nc), "probate office")
			}
		case model.EventRoleBeneficiary:
			// James John Simpson
			title = text.JoinSentenceParts("Beneficiary of the probate of", t.observerContext(ev, true))
		case model.EventRoleExecutor:
			title = text.JoinSentenceParts("Executor of the probate of", t.observerContext(ev, true))
		case model.EventRoleAdministrator:
			title = text.JoinSentenceParts("Administrator of the probate of", t.observerContext(ev, true))
		default:
			t.logger.Warn("unsupported role in probate event", "role", ep.Role)
		}
	}

	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) willEventTitle(seq int, ev model.TimelineEvent) string {
	var title string
	what := t.whenWhat(ev)
	pl := ev.GetPlace()

	eps := ev.GetParticipant(t.pov.Person)
	if len(eps) == 0 {
		// Does not participate so this is someone else's event
		title = what
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentenceParts(title, "here")
		}

		obsContext := t.observerContext(ev, true)
		if obsContext != "" {
			title = text.JoinSentenceParts(title, "for", obsContext)
		}

		if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
			title = text.JoinSentenceParts(title, WhatWhere("", pl, t.enc, t.nc))
		}
	} else if len(eps) > 0 {
		ep := eps[0]
		switch ep.Role {
		case model.EventRolePrincipal:
			title = what
			if pl.SameAs(t.pov.Place) {
				title = text.JoinSentenceParts(title, "here")
			}
			if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
				title = text.JoinSentenceParts(title, WhatWhere("", pl, t.enc, t.nc), "probate office")
			}
		case model.EventRoleBeneficiary:
			title = text.JoinSentenceParts("Named as beneficiary in the will of", t.observerContext(ev, true))
		case model.EventRoleExecutor:
			title = text.JoinSentenceParts("Named as executor in the will of", t.observerContext(ev, true))
		case model.EventRoleWitness:
			title = text.JoinSentenceParts("Witnessed the will of", t.observerContext(ev, true))
		default:
			t.logger.Warn("unsupported role in will event", "role", ep.Role)
		}
	}

	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) residenceEventTitle(seq int, ev *model.ResidenceRecordedEvent) string {
	pl := ev.GetPlace()

	if pl.IsUnknown() {
		// TODO: record an anomaly?
		return ""
	}

	title := text.JoinSentenceParts(t.observerContext(ev, false), "recorded", ChooseFrom(seq, "as residing", "as living"))
	title = WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) musterEventTitle(seq int, ev *model.MusterEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev, false), "recorded")

	if regiment, ok := ev.GetAttribute(model.EventAttributeRegiment); ok {
		if battalion, ok := ev.GetAttribute(model.EventAttributeBattalion); ok {
			title = text.JoinSentenceParts(title, "in the", battalion, "battalion,", regiment)
		} else {
			title = text.JoinSentenceParts(title, "in the", regiment, "regiment")
		}
	}
	title = text.JoinSentenceParts(title, "at muster")

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		title = WhatWherePov(title+" taken", pl, t.enc, t.nc, t.pov)
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) arrivalEventTitle(seq int, ev *model.ArrivalEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev, false), "arrived")

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentenceParts(title, "here")
		} else {
			title = text.JoinSentenceParts(title, t.enc.EncodeModelLinkNamed(pl, t.nc, t.pov).String())
		}
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) departureEventTitle(seq int, ev *model.DepartureEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev, false), "departed")

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentenceParts(title, "from here")
		} else {
			title = text.JoinSentenceParts(title, "from", t.enc.EncodeModelLinkNamed(pl, t.nc, t.pov).String())
		}
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) institutionEntryEventTitle(seq int, ev *model.InstitutionEntryEvent) string {
	title := ev.GetDetail()
	if title == "" {
		title = text.JoinSentenceParts(t.observerContext(ev, false), "entered")
	}
	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentenceParts(title, "here")
		} else {
			title = text.JoinSentenceParts(title, t.enc.EncodeModelLinkNamed(pl, t.nc, t.pov).String())
		}
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) institutionDepartureEventTitle(seq int, ev *model.InstitutionDepartureEvent) string {
	title := ev.GetDetail()
	if title == "" {
		title = text.JoinSentenceParts(t.observerContext(ev, false), "left")
	}
	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentenceParts(title, "here")
		} else {
			title = text.JoinSentenceParts(title, t.enc.EncodeModelLinkNamed(pl, t.nc, t.pov).String())
		}
	}

	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) marriageEventTitle(seq int, ev model.UnionTimelineEvent) string {
	title := ""
	if t.pov.Person.IsUnknown() {
		party1 := ev.GetHusband()
		party2 := ev.GetWife()

		party1Link := t.enc.EncodeModelLink(t.enc.EncodeText(t.nc.FirstUse(party1)), party1).String()
		party2Link := t.enc.EncodeModelLink(t.enc.EncodeText(t.nc.FirstUse(party2)), party2).String()

		switch ev.(type) {
		case *model.MarriageEvent:
			title = text.JoinSentenceParts(party1Link, "married", party2Link)
		case *model.MarriageLicenseEvent:
			title = text.JoinSentenceParts(model.What(ev), "for the marriage of ", party1Link, "and", party2Link)
		case *model.MarriageBannsEvent:
			title = text.JoinSentenceParts(model.What(ev), "for the marriage of ", party1Link, "and", party2Link)
		default:
			panic(fmt.Sprintf("unhandled marriage event type: %T", ev))
		}

	} else {
		eps := ev.GetParticipant(t.pov.Person)
		if len(eps) == 0 {
			panic(fmt.Sprintf("person missing from event participants (person.ID=%s, native_id=%s, event type=%T)", t.pov.Person.ID, t.pov.Person.NativeID, ev))
		}

		ep := eps[0]
		switch ep.Role {
		case model.EventRoleHusband, model.EventRoleWife:
			spouse := ev.GetOther(t.pov.Person)
			spouseLink := t.enc.EncodeModelLink(t.enc.EncodeText(t.nc.FirstUse(spouse)), spouse).String()
			switch ev.(type) {
			case *model.MarriageEvent:
				title = text.JoinSentenceParts("married", spouseLink)
			case *model.MarriageLicenseEvent:
				title = text.JoinSentenceParts(model.What(ev), "to marry", spouseLink)
			case *model.MarriageBannsEvent:
				title = text.JoinSentenceParts(model.What(ev), "to marry", spouseLink)
			default:
				panic(fmt.Sprintf("unhandled marriage event type: %T", ev))
			}
		case model.EventRoleWitness:
			party1 := ev.GetHusband()
			party2 := ev.GetWife()

			party1Link := t.enc.EncodeModelLink(t.enc.EncodeText(t.nc.FirstUse(party1)), party1).String()
			party2Link := t.enc.EncodeModelLink(t.enc.EncodeText(t.nc.FirstUse(party2)), party2).String()
			title = text.JoinSentenceParts("witness to the marriage of", party1Link, "and", party2Link)
		}
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) observerContext(ev model.TimelineEvent, prefixRelationWithPronoun bool) string {
	observer := t.pov.Person
	switch tev := ev.(type) {
	case model.IndividualTimelineEvent:
		principal := tev.GetPrincipal()
		if observer.SameAs(principal) {
			return "" // no context needed
		}
		// This is someone else's event
		if observer.IsUnknown() || tev.GetDate().IsUnknown() {
			return t.enc.EncodeModelLink(t.enc.EncodeText(t.nc.FirstUse(principal)), principal).String()
		}
		name := t.enc.EncodeModelLinkNamed(principal, t.nc, t.pov).String()
		rel := principal.RelationTo(t.pov.Person, tev.GetDate())
		if rel != "" && prefixRelationWithPronoun {
			rel = t.pov.Person.Gender.PossessivePronounSingular() + " " + rel
		}
		return text.AppendAside(rel, name)
	case model.MultipartyTimelineEvent:
		if tev.DirectlyInvolves(observer) {
			return "" // no context needed
		}
		var ppl []string
		for _, p := range tev.GetPrincipals() {
			if p.IsUnknown() {
				continue
			}
			ppl = append(ppl, t.enc.EncodeModelLink(t.enc.EncodeText(t.nc.FirstUse(p)), p).String())
		}

		return text.JoinList(ppl)

	case *model.CensusEvent:
		if tev.DirectlyInvolves(observer) {
			return ""
		}

		var ppl []string
		head := tev.Head()
		if head != nil {
			ppl = append(ppl, t.enc.EncodeModelLinkNamed(head, t.nc, t.pov).String())
		}

		for _, en := range tev.Entries {
			if en.Principal.IsUnknown() {
				continue
			}
			if en.Principal.SameAs(head) {
				continue
			}
			ppl = append(ppl, t.enc.EncodeModelLinkNamed(en.Principal, t.nc, t.pov).String())
		}

		return text.JoinList(ppl)

	default:
		return ""
	}
}

func placeIsKnownAndIsNotSameAsPointOfView(pl *model.Place, pov *model.POV) bool {
	return !pl.IsUnknown() && !pl.SameAs(pov.Place)
}

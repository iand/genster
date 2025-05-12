package site

import (
	"fmt"
	"strconv"

	"github.com/iand/gdate"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

func RenderTimeline[T render.EncodedText](t *model.Timeline, pov *model.POV, enc render.ContentBuilder[T], fmtr TimelineEntryFormatter[T]) error {
	enc.EmptyPara()
	if len(t.Events) == 0 {
		return nil
	}
	model.SortTimelineEvents(t.Events)

	logger := logging.Default()
	if !pov.Person.IsUnknown() {
		logger = logger.With("id", pov.Person.ID, "native_id", pov.Person.NativeID)
	}

	monthNames := []string{
		1:  "Jan",
		2:  "Feb",
		3:  "Mar",
		4:  "Apr",
		5:  "May",
		6:  "Jun",
		7:  "Jul",
		8:  "Aug",
		9:  "Sep",
		10: "Oct",
		11: "Nov",
		12: "Dec",
	}

	var row render.TimelineRow[T]
	events := make([]render.TimelineRow[T], 0, len(t.Events))
	for i, ev := range t.Events {
		if !IncludeInTimeline(ev) {
			continue
		}
		title := fmtr.Title(i, ev)
		if title == "" {
			continue
		}

		dt := ev.GetDate()
		if dt.IsUnknown() {
			continue
		}

		var sy, sd string
		if y, m, d, ok := dt.YMD(); ok {
			sy = strconv.Itoa(y)
			sd = fmt.Sprintf("%d %s", d, monthNames[m])
		} else if dt.Span {
			switch d := dt.Date.(type) {
			case *gdate.YearRange:
				sy = d.String()
				sd = ""
			default:
				logger.Warn("timeline: unsupported date span", "type", fmt.Sprintf("%T", d), "value", d.String(), "event", fmt.Sprintf("%T", ev))
			}
		} else {
			switch d := dt.Date.(type) {
			case *gdate.BeforeYear:
				sy = d.Occurrence()
				sd = ""
			case *gdate.AfterYear:
				sy = d.Occurrence()
				sd = ""
			case *gdate.AboutYear:
				sy = d.Occurrence()
				sd = ""
			case *gdate.Year:
				sy = d.String()
				sd = ""
			case *gdate.YearQuarter:
				sy = strconv.Itoa(d.Year())
				sd = d.MonthRange()
			case *gdate.MonthYear:
				sy = strconv.Itoa(d.Year())
				sd = monthNames[d.M]
			default:
				logger.Warn("timeline: unsupported date type", "type", fmt.Sprintf("%T", d), "value", d.String(), "event", fmt.Sprintf("%T", ev))
			}
		}

		if row.Year == sy && row.Date == sd {
			row.Details = append(row.Details, enc.EncodeText(title))
			continue
		} else {
			if row.Year != "" {
				events = append(events, row)
			}
			row = render.TimelineRow[T]{
				Year:    sy,
				Date:    sd,
				Details: []T{enc.EncodeText(title)},
			}
		}
	}
	if row.Year != "" {
		events = append(events, row)
	}
	enc.Timeline(events)
	return nil
}

func IncludeInTimeline(ev model.TimelineEvent) bool {
	switch ev.(type) {
	case *model.IndividualNarrativeEvent:
		return false
	default:
		return true
	}
}

type TimelineEntryFormatter[T render.EncodedText] interface {
	Title(seq int, ev model.TimelineEvent) string
	Detail(seq int, ev model.TimelineEvent) string
}

type NarrativeTimelineEntryFormatter[T render.EncodedText] struct {
	pov      *model.POV
	enc      render.TextEncoder[T]
	omitDate bool
	logger   *logging.Logger
	nc       narrative.NameChooser
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
	} else {
		for _, ep := range eps {
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
							title = text.JoinSentenceParts(title, narrative.AgeQualifier(age))
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

			// TODO: support multiple roles in an event
			break
		}
	}

	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = narrative.WhatWhere(title, pl, t.enc, t.nc)
	}

	if trailer != "" {
		title = text.AppendSentence(text.FormatSentence(title), trailer)
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
		residing := narrative.ChooseFrom(seq, "residing", "", "living")
		title = text.JoinSentenceParts(title, residing)
		title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
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
	title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
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
			title = text.JoinSentenceParts(title, narrative.WhatWhere("", pl, t.enc, t.nc), "probate office")
		}
	} else {
		for _, ep := range eps {
			switch ep.Role {
			case model.EventRolePrincipal:
				title = what
				if pl.SameAs(t.pov.Place) {
					title = text.JoinSentenceParts(title, "here")
				}
				if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
					title = text.JoinSentenceParts(title, narrative.WhatWhere("", pl, t.enc, t.nc), "probate office")
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

			// TODO: support multiple event roles
			break
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
			title = text.JoinSentenceParts(title, narrative.WhatWhere("", pl, t.enc, t.nc))
		}
	} else {
		for _, ep := range eps {
			switch ep.Role {
			case model.EventRolePrincipal:
				title = what
				if pl.SameAs(t.pov.Place) {
					title = text.JoinSentenceParts(title, "here")
				}
				if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
					title = text.JoinSentenceParts(title, narrative.WhatWhere("", pl, t.enc, t.nc), "probate office")
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

			// TODO: support multiple event roles
			break
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

	title := text.JoinSentenceParts(t.observerContext(ev, false), "recorded", narrative.ChooseFrom(seq, "as residing", "as living"))
	title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
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
		title = narrative.WhatWherePov(title+" taken", pl, t.enc, t.nc, t.pov)
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

		party1Link := t.enc.EncodeModelLink(t.enc.EncodeText(party1.PreferredFullName), party1).String()
		party2Link := t.enc.EncodeModelLink(t.enc.EncodeText(party2.PreferredFullName), party2).String()

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

		for _, ep := range eps {
			switch ep.Role {
			case model.EventRoleHusband, model.EventRoleWife:
				spouse := ev.GetOther(t.pov.Person)
				spouseLink := t.enc.EncodeModelLink(t.enc.EncodeText(spouse.PreferredFullName), spouse).String()
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

				party1Link := t.enc.EncodeModelLink(t.enc.EncodeText(party1.PreferredFullName), party1).String()
				party2Link := t.enc.EncodeModelLink(t.enc.EncodeText(party2.PreferredFullName), party2).String()
				title = text.JoinSentenceParts("witness to the marriage of", party1Link, "and", party2Link)
			}

			// TODO: support multiple roles
			break
		}
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
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
			return t.enc.EncodeModelLink(t.enc.EncodeText(principal.PreferredFullName), principal).String()
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
			ppl = append(ppl, t.enc.EncodeModelLink(t.enc.EncodeText(p.PreferredFullName), p).String())
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

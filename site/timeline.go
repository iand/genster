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
	case *model.WitnessToMarriageEvent:
		title = t.witnessToMarriageEventTitle(seq, tev)
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
	title = t.enc.EncodeWithCitations(t.enc.EncodeText(title), ev.GetCitations()).String()
	return text.FormatSentence(title)
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

	principal := ev.GetPrincipal()
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

	what := t.whenWhat(ev)

	// Add person info if known and not the pov person
	if !principal.IsUnknown() {
		if principal.SameAs(t.pov.Person) {
			title = text.JoinSentenceParts(title, what)
			// add their age, if its not their earliest event
			if includeAge {
				if ev != t.pov.Person.BestBirthlikeEvent {
					if age, ok := t.pov.Person.AgeInYearsAt(date); ok {
						title = text.JoinSentenceParts(title, narrative.AgeQualifier(age))
					}
				}
			}
		} else {
			// This is someone else's event
			obsContext := t.observerContext(ev)
			title = text.JoinSentenceParts(title, obsContext, what)
			// add their age, if its not their earliest event
			if includeAge {
				if ev != ev.GetPrincipal().BestBirthlikeEvent {
					if age, ok := ev.GetPrincipal().AgeInYearsAt(date); ok {
						title = text.JoinSentenceParts(title, narrative.AgeQualifier(age))
					}
				}
			}
		}
	}
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = narrative.WhatWhere(title, pl, t.enc, t.nc)
	}

	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) censusEventTitle(seq int, ev *model.CensusEvent) string {
	var title string
	title = text.JoinSentenceParts(title, t.observerContext(ev), "recorded in the")
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
	title = text.JoinSentenceParts(title, t.observerContext(ev))
	title = text.JoinSentenceParts(title, t.whenWhat(ev))

	pl := ev.GetPlace()
	title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) probateEventTitle(seq int, ev model.TimelineEvent) string {
	title := t.whenWhat(ev)
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	obsContext := t.observerContext(ev)
	if obsContext != "" {
		title = text.JoinSentenceParts(title, "for", t.observerContext(ev))
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, narrative.WhatWhere("by the", pl, t.enc, t.nc), "probate office")
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) residenceEventTitle(seq int, ev *model.ResidenceRecordedEvent) string {
	pl := ev.GetPlace()

	if pl.IsUnknown() {
		// TODO: record an anomaly?
		return ""
	}

	title := text.JoinSentenceParts(t.observerContext(ev), "recorded", narrative.ChooseFrom(seq, "as residing", "as living"))
	title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) musterEventTitle(seq int, ev *model.MusterEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev), "recorded")

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
	title := text.JoinSentenceParts(t.observerContext(ev), "arrived")

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
	title := text.JoinSentenceParts(t.observerContext(ev), "departed")

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
		title = text.JoinSentenceParts(t.observerContext(ev), "entered")
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
		title = text.JoinSentenceParts(t.observerContext(ev), "left")
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

	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) witnessToMarriageEventTitle(seq int, ev *model.WitnessToMarriageEvent) string {
	if ev.MarriageEvent == nil {
		return ""
	}

	title := ""
	party1 := ev.MarriageEvent.GetHusband()
	party2 := ev.MarriageEvent.GetWife()

	party1Link := t.enc.EncodeModelLink(t.enc.EncodeText(party1.PreferredFullName), party1).String()
	party2Link := t.enc.EncodeModelLink(t.enc.EncodeText(party2.PreferredFullName), party2).String()

	title = text.JoinSentenceParts("witnessed the marriage of", party1Link, "and", party2Link)

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = narrative.WhatWherePov(title, pl, t.enc, t.nc, t.pov)
	}
	return title
}

func (t *NarrativeTimelineEntryFormatter[T]) observerContext(ev model.TimelineEvent) string {
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
		return text.AppendAside(principal.RelationTo(t.pov.Person, tev.GetDate()), name)
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

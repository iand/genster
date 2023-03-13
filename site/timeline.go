package site

import (
	"fmt"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"golang.org/x/exp/slog"
)

func RenderTimeline(t *model.Timeline, pov *model.POV, enc ExtendedMarkdownBuilder) error {
	enc.EmptyPara()
	if len(t.Events) == 0 {
		return nil
	}
	model.SortTimelineEvents(t.Events)

	fmtr := &TimelineEntryFormatter{
		pov: pov,
		enc: enc,
	}

	events := make([][2]string, 0, len(t.Events))
	for i, ev := range t.Events {
		title := fmtr.Title(i, ev)
		detail := fmtr.Detail(i, ev)
		events = append(events, [2]string{
			title,
			detail,
		})
	}

	enc.DefinitionList(events)
	return nil
}

func WhatWhenWhere(ev model.TimelineEvent, enc ExtendedInlineEncoder) string {
	title := ev.What()
	date := ev.GetDate()
	if !date.IsUnknown() {
		title = text.JoinSentence(title, date.When())
	}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		title = text.JoinSentence(title, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}
	return title
}

func AgeWhenWhere(ev model.IndividualTimelineEvent, enc ExtendedInlineEncoder) string {
	title := ""

	date := ev.GetDate()
	if !date.IsUnknown() {
		if age, ok := ev.GetPrincipal().AgeInYearsAt(ev.GetDate()); ok {
			title = text.JoinSentence(title, AgeQualifier(age))
		}
		title = text.JoinSentence(title, date.When())
	}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		title = text.JoinSentence(title, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}
	return title
}

type TimelineEntryFormatter struct {
	pov *model.POV
	enc ExtendedMarkdownEncoder
}

func (t *TimelineEntryFormatter) Title(seq int, ev model.TimelineEvent) string {
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
		title = t.generalEventTitle(seq, tev)
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
		slog.Debug(fmt.Sprintf("arrival eventfor %s (%s)", tev.GetPrincipal().PreferredUniqueName, tev.GetPrincipal().ID))
	case *model.DepartureEvent:
		title = t.departureEventTitle(seq, tev)
		slog.Debug(fmt.Sprintf("departure eventfor %s (%s)", tev.GetPrincipal().PreferredUniqueName, tev.GetPrincipal().ID))
	case *model.PlaceholderIndividualEvent:
		slog.Debug(fmt.Sprintf("placeholder event %q for %s", tev.ExtraInfo, tev.GetPrincipal().PreferredUniqueName))
		title = tev.ExtraInfo
	default:
		slog.Debug(fmt.Sprintf("unhandled event type: %T", ev))
		title = t.generalEventTitle(seq, tev)
	}
	title = EncodeWithCitations(title, ev.GetCitations(), t.enc)
	return text.FormatSentence(title)
}

func (t *TimelineEntryFormatter) Detail(seq int, ev model.TimelineEvent) string {
	switch tev := ev.(type) {
	default:
		return tev.GetDetail()
	}
}

func (t *TimelineEntryFormatter) vitalEventTitle(seq int, ev model.IndividualTimelineEvent) string {
	var title string

	principal := ev.GetPrincipal()
	date := ev.GetDate()

	inference := ""
	if ev.IsInferred() {
		inference = "inferred to have"
	}

	includeAge := true
	if date.IsUnknown() {
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

	// Add person info if known and not the pov person
	if !principal.IsUnknown() {
		if principal.SameAs(t.pov.Person) {
			title = text.JoinSentence(title, inference, ev.What())
			// add their age, if its not their earliest event
			if includeAge {
				if ev != t.pov.Person.BestBirthlikeEvent {
					if age, ok := t.pov.Person.AgeInYearsAt(date); ok {
						title = text.JoinSentence(title, AgeQualifier(age))
					}
				}
			}
		} else {
			// This is someone else's event
			obsContext := t.observerContext(ev)
			title = text.JoinSentence(title, obsContext, inference, ev.What())
			// add their age, if its not their earliest event
			if includeAge {
				if ev != ev.GetPrincipal().BestBirthlikeEvent {
					if age, ok := ev.GetPrincipal().AgeInYearsAt(date); ok {
						title = text.JoinSentence(title, AgeQualifier(age))
					}
				}
			}
		}
	}

	// Add date if known
	if !date.IsUnknown() {
		title += " " + date.When()
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentence(title, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}

	return title
}

func (t *TimelineEntryFormatter) censusEventTitle(seq int, ev *model.CensusEvent) string {
	var title string
	title = text.JoinSentence(title, t.observerContext(ev), "recorded in the")
	year, ok := ev.GetDate().Year()
	if ok {
		title = text.JoinSentence(title, fmt.Sprintf("%d ", year))
	}
	title = text.JoinSentence(title, "census")

	entry, _ := ev.Entry(t.pov.Person)
	if entry != nil && entry.RelationToHead.IsImpersonal() {
		title = text.JoinSentence(title, "as", "a"+text.MaybeAn(string(entry.RelationToHead)))
	}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		residing := ChooseFrom(seq, "residing", "", "living")
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentence(title, residing, "here")
		} else {
			title = text.JoinSentence(title, residing, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
		}

	}
	return title
}

func (t *TimelineEntryFormatter) generalEventTitle(seq int, ev model.TimelineEvent) string {
	var title string
	title = text.JoinSentence(title, t.observerContext(ev))

	if ev.IsInferred() {
		title = text.JoinSentence(title, "inferred to have")
	}

	title = text.JoinSentence(title, ev.What())
	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentence(title, date.When())
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentence(title, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) probateEventTitle(seq int, ev model.TimelineEvent) string {
	title := "probate was granted"
	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentence(title, date.When())
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentence(title, "by the", t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl), "probate office")
	}
	return title
}

func (t *TimelineEntryFormatter) residenceEventTitle(seq int, ev *model.ResidenceRecordedEvent) string {
	title := "recorded"

	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentence(title, date.When())
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		residing := ChooseFrom(seq, "as residing", "as living")
		title = text.JoinSentence(title, residing, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) arrivalEventTitle(seq int, ev *model.ArrivalEvent) string {
	title := "arrived"

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentence(title, "at", t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}

	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentence(title, date.When())
	}

	return title
}

func (t *TimelineEntryFormatter) departureEventTitle(seq int, ev *model.DepartureEvent) string {
	title := "departed"

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentence(title, "from", t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}

	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentence(title, date.When())
	}

	return title
}

func (t *TimelineEntryFormatter) marriageEventTitle(seq int, ev model.PartyTimelineEvent) string {
	title := ""
	if t.pov.Person == nil {
		party1 := ev.GetParty1()
		party2 := ev.GetParty1()
		switch ev.(type) {
		case *model.MarriageEvent:
			title = text.JoinSentence(party1.PreferredFullName, "married", party2.PreferredFullName)
		case *model.MarriageLicenseEvent:
			title = text.JoinSentence("license was obtained for the marriage of ", party1.PreferredFullName, "and", party2.PreferredFullName)
		case *model.MarriageBannsEvent:
			title = text.JoinSentence("banns were read for the marriage of ", party1.PreferredFullName, "and", party2.PreferredFullName)
		default:
			panic(fmt.Sprintf("unhandled marriage event type: %T", ev))
		}

	} else {
		spouse := ev.GetOther(t.pov.Person)
		switch ev.(type) {
		case *model.MarriageEvent:
			title = text.JoinSentence("married", spouse.PreferredFullName)
		case *model.MarriageLicenseEvent:
			title = text.JoinSentence("obtained license to marry", spouse.PreferredFullName)
		case *model.MarriageBannsEvent:
			title = text.JoinSentence("banns were read to marry", spouse.PreferredFullName)
		default:
			panic(fmt.Sprintf("unhandled marriage event type: %T", ev))
		}

	}

	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentence(title, date.When())
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentence(title, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) observerContext(ev model.TimelineEvent) string {
	observer := t.pov.Person
	switch tev := ev.(type) {
	case model.IndividualTimelineEvent:
		principal := tev.GetPrincipal()
		if observer.SameAs(principal) {
			return "" // no context needed
		}
		// This is someone else's event
		if observer.IsUnknown() || tev.GetDate().IsUnknown() {
			return t.enc.EncodeModelLinkDedupe(principal.PreferredFullName, principal.PreferredFullName, principal)
		}
		name := t.enc.EncodeModelLinkDedupe(principal.PreferredFullName, principal.PreferredGivenName, principal)
		return text.AppendAside(principal.RelationTo(t.pov.Person, tev.GetDate()), name)

	case *model.CensusEvent:
		if tev.DirectlyInvolves(observer) {
			return ""
		}

		var ppl []string
		head := tev.Head()
		if head != nil {
			ppl = append(ppl, t.enc.EncodeModelLinkDedupe(head.PreferredFullName, head.PreferredFullName, head))
		}

		for _, en := range tev.Entries {
			if en.Principal.IsUnknown() {
				continue
			}
			if en.Principal.SameAs(head) {
				continue
			}
			ppl = append(ppl, t.enc.EncodeModelLinkDedupe(en.Principal.PreferredFullName, en.Principal.PreferredFullName, en.Principal))
		}

		return text.JoinList(ppl)

	default:
		return ""
	}
}

func dateIsKnownOrThereIsNoObserver(date *model.Date, pov *model.POV) bool {
	return !date.IsUnknown() && !pov.Person.IsUnknown()
}

func placeIsKnownAndIsNotSameAsPointOfView(pl *model.Place, pov *model.POV) bool {
	return !pl.IsUnknown() && !pl.SameAs(pov.Place)
}

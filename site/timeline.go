package site

import (
	"fmt"
	"log/slog"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
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

type TimelineEntryFormatter struct {
	pov      *model.POV
	enc      ExtendedMarkdownEncoder
	omitDate bool
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
	case *model.DepartureEvent:
		title = t.departureEventTitle(seq, tev)
	case *model.DivorceEvent:
		slog.Debug("timeline: unhandled divorce event")
	case *model.PlaceholderIndividualEvent:
		slog.Debug("timeline: ignored placeholder event", "id", tev.GetPrincipal().ID, "info", tev.ExtraInfo)
		title = tev.ExtraInfo
	default:
		slog.Debug(fmt.Sprintf("timeline: unhandled event type: %T", ev))
		title = t.generalEventTitle(seq, tev)
	}
	title = t.enc.EncodeWithCitations(title, ev.GetCitations())
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

	qual := ""
	if ev.IsInferred() {
		qual = "inferred to have"
	} else if !t.omitDate && !date.IsUnknown() {
		qual = date.Derivation.Qualifier()
		if qual != "" {
			qual += " to have"
		}
	}

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

	// Add person info if known and not the pov person
	if !principal.IsUnknown() {
		if principal.SameAs(t.pov.Person) {
			title = text.JoinSentenceParts(title, qual, ev.What())
			// add their age, if its not their earliest event
			if includeAge {
				if ev != t.pov.Person.BestBirthlikeEvent {
					if age, ok := t.pov.Person.AgeInYearsAt(date); ok {
						title = text.JoinSentenceParts(title, AgeQualifier(age))
					}
				}
			}
		} else {
			// This is someone else's event
			obsContext := t.observerContext(ev)
			title = text.JoinSentenceParts(title, obsContext, qual, ev.What())
			// add their age, if its not their earliest event
			if includeAge {
				if ev != ev.GetPrincipal().BestBirthlikeEvent {
					if age, ok := ev.GetPrincipal().AgeInYearsAt(date); ok {
						title = text.JoinSentenceParts(title, AgeQualifier(age))
					}
				}
			}
		}
	}
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	// Add date if known
	if !t.omitDate && !date.IsUnknown() {
		title = text.JoinSentenceParts(title, date.When())
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}

	return title
}

func (t *TimelineEntryFormatter) censusEventTitle(seq int, ev *model.CensusEvent) string {
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
		residing := ChooseFrom(seq, "residing", "", "living")
		if pl.SameAs(t.pov.Place) {
			title = text.JoinSentenceParts(title, residing, "here")
		} else {
			title = text.JoinSentenceParts(title, residing, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
		}

	}
	return title
}

func (t *TimelineEntryFormatter) generalEventTitle(seq int, ev model.TimelineEvent) string {
	var title string
	title = text.JoinSentenceParts(title, t.observerContext(ev))

	if ev.IsInferred() {
		title = text.JoinSentenceParts(title, "inferred to have")
	}

	title = text.JoinSentenceParts(title, ev.What())
	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) probateEventTitle(seq int, ev model.TimelineEvent) string {
	title := "probate was granted"
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	obsContext := t.observerContext(ev)
	if obsContext != "" {
		title = text.JoinSentenceParts(title, "for", t.observerContext(ev))
	}

	date := ev.GetDate()
	if !t.omitDate && dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, "by the", t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl), "probate office")
	}
	return title
}

func (t *TimelineEntryFormatter) residenceEventTitle(seq int, ev *model.ResidenceRecordedEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev), "recorded")

	pl := ev.GetPlace()
	residing := ChooseFrom(seq, "as residing", "as living")
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, residing, "here")
	}

	date := ev.GetDate()
	if !t.omitDate && dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, residing, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) arrivalEventTitle(seq int, ev *model.ArrivalEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev), "arrived")
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, "at", t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}

	date := ev.GetDate()
	if !t.omitDate && dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	return title
}

func (t *TimelineEntryFormatter) departureEventTitle(seq int, ev *model.DepartureEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev), "departed")
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "from here")
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, "from", t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	}

	date := ev.GetDate()
	if !t.omitDate && dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	return title
}

func (t *TimelineEntryFormatter) marriageEventTitle(seq int, ev model.PartyTimelineEvent) string {
	title := ""
	if t.pov.Person.IsUnknown() {
		party1 := ev.GetParty1()
		party2 := ev.GetParty2()

		party1Link := t.enc.EncodeModelLink(party1.PreferredFullName, party1)
		party2Link := t.enc.EncodeModelLink(party2.PreferredFullName, party2)

		switch ev.(type) {
		case *model.MarriageEvent:
			title = text.JoinSentenceParts(party1Link, "married", party2Link)
		case *model.MarriageLicenseEvent:
			title = text.JoinSentenceParts("license was obtained for the marriage of ", party1Link, "and", party2Link)
		case *model.MarriageBannsEvent:
			title = text.JoinSentenceParts("banns were read for the marriage of ", party1Link, "and", party2Link)
		default:
			panic(fmt.Sprintf("unhandled marriage event type: %T", ev))
		}

	} else {
		spouse := ev.GetOther(t.pov.Person)
		spouseLink := t.enc.EncodeModelLink(spouse.PreferredFullName, spouse)
		switch ev.(type) {
		case *model.MarriageEvent:
			title = text.JoinSentenceParts("married", spouseLink)
		case *model.MarriageLicenseEvent:
			title = text.JoinSentenceParts("obtained license to marry", spouseLink)
		case *model.MarriageBannsEvent:
			title = text.JoinSentenceParts("banns were read to marry", spouseLink)
		default:
			panic(fmt.Sprintf("unhandled marriage event type: %T", ev))
		}

	}

	date := ev.GetDate()
	if !t.omitDate && dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.PlaceType.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
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
	return !date.IsUnknown() || pov.Person.IsUnknown()
}

func placeIsKnownAndIsNotSameAsPointOfView(pl *model.Place, pov *model.POV) bool {
	return !pl.IsUnknown() && !pl.SameAs(pov.Place)
}

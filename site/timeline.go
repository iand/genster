package site

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iand/gdate"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

func RenderTimeline(t *model.Timeline, pov *model.POV, enc render.MarkupBuilder) error {
	enc.EmptyPara()
	if len(t.Events) == 0 {
		return nil
	}
	model.SortTimelineEvents(t.Events)

	logger := logging.Default()
	if !pov.Person.IsUnknown() {
		logger = logger.With("id", pov.Person.ID)
	}

	fmtr := &TimelineEntryFormatter{
		pov:    pov,
		enc:    enc,
		logger: logger,
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

	var row render.TimelineRow
	events := make([]render.TimelineRow, 0, len(t.Events))
	for i, ev := range t.Events {
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
				sd = fmt.Sprintf("%s", monthNames[d.M])
			default:
				logger.Warn("timeline: unsupported date type", "type", fmt.Sprintf("%T", d), "value", d.String(), "event", fmt.Sprintf("%T", ev))
			}
		}

		if row.Year == sy && row.Date == sd {
			row.Details = append(row.Details, render.Markdown(title))
			continue
		} else {
			if row.Year != "" {
				events = append(events, row)
			}
			row = render.TimelineRow{
				Year:    sy,
				Date:    sd,
				Details: []render.Markdown{render.Markdown(title)},
			}
		}
	}
	if row.Year != "" {
		events = append(events, row)
	}
	enc.Timeline(events)
	return nil
}

type TimelineEntryFormatter struct {
	pov      *model.POV
	enc      render.PageMarkdownEncoder
	omitDate bool
	logger   *logging.Logger
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
		t.logger.Debug("timeline: unhandled event type", "type", fmt.Sprintf("%T", ev), "title", title)
	}
	if title == "" {
		t.logger.Debug("timeline: ignored event type", "type", fmt.Sprintf("%T", ev))
		return ""
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

func (t *TimelineEntryFormatter) whenWhat(what string, ev model.TimelineEvent) string {
	date := ev.GetDate()
	if !date.IsUnknown() {
		switch ev.GetDate().Derivation {
		case model.DateDerivationEstimated:
			what = text.MaybeWasVerb(what)
			if strings.HasPrefix(what, "was ") {
				what = what[4:]
			}
			what = "probably " + what + " around this time"
		case model.DateDerivationCalculated:
			what = "calculated to " + text.MaybeHaveBeenVerb(what) + " around this time"
		}
	}
	return what
}

func (t *TimelineEntryFormatter) vitalEventTitle(seq int, ev model.IndividualTimelineEvent) string {
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

	what := t.whenWhat(ev.What(), ev)

	// Add person info if known and not the pov person
	if !principal.IsUnknown() {
		if principal.SameAs(t.pov.Person) {
			title = text.JoinSentenceParts(title, what)
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
			title = text.JoinSentenceParts(title, obsContext, what)
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

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
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
			title = text.JoinSentenceParts(title, residing, pl.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
		}

	}
	return title
}

func (t *TimelineEntryFormatter) generalEventTitle(seq int, ev model.TimelineEvent) string {
	if ev.What() == "" {
		return ""
	}
	var title string
	title = text.JoinSentenceParts(title, t.observerContext(ev))
	title = text.JoinSentenceParts(title, t.whenWhat(ev.What(), ev))

	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	} else if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) probateEventTitle(seq int, ev model.TimelineEvent) string {
	title := text.JoinSentenceParts("probate was granted", t.whenWhat(ev.What(), ev))
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	obsContext := t.observerContext(ev)
	if obsContext != "" {
		title = text.JoinSentenceParts(title, "for", t.observerContext(ev))
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, "by the", t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl), "probate office")
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

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, residing, pl.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) musterEventTitle(seq int, ev *model.MusterEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev), "recorded")

	if regiment, ok := ev.GetAttribute(model.EventAttributeRegiment); ok {
		if battalion, ok := ev.GetAttribute(model.EventAttributeBattalion); ok {
			title = text.JoinSentenceParts(title, "in the", battalion, "battalion,", regiment)
		} else {
			title = text.JoinSentenceParts(title, "in the", regiment, "regiment")
		}
	}
	title = text.JoinSentenceParts(title, "at muster taken ")

	pl := ev.GetPlace()

	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	date := ev.GetDate()
	if !t.omitDate && dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
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
		title = text.JoinSentenceParts(title, "at", t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
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
		title = text.JoinSentenceParts(title, "from", t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}

	return title
}

func (t *TimelineEntryFormatter) institutionEntryEventTitle(seq int, ev *model.InstitutionEntryEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev), "entered")
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}

	return title
}

func (t *TimelineEntryFormatter) institutionDepartureEventTitle(seq int, ev *model.InstitutionDepartureEvent) string {
	title := text.JoinSentenceParts(t.observerContext(ev), "left")
	pl := ev.GetPlace()
	if pl.SameAs(t.pov.Place) {
		title = text.JoinSentenceParts(title, "here")
	}

	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}

	return title
}

func (t *TimelineEntryFormatter) occupationEventTitle(seq int, ev model.TimelineEvent) string {
	var title string
	title = text.JoinSentenceParts(title, t.observerContext(ev))

	if ev.IsInferred() {
		title = text.JoinSentenceParts(title, "inferred to have")
	}

	title = text.JoinSentenceParts(title, ev.What())
	title = text.JoinSentenceParts(title, ev.GetDetail())
	date := ev.GetDate()
	if dateIsKnownOrThereIsNoObserver(date, t.pov) {
		title = text.JoinSentenceParts(title, date.When())
	}

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return title
}

func (t *TimelineEntryFormatter) marriageEventTitle(seq int, ev model.UnionTimelineEvent) string {
	title := ""
	if t.pov.Person.IsUnknown() {
		party1 := ev.GetHusband()
		party2 := ev.GetWife()

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

	pl := ev.GetPlace()
	if placeIsKnownAndIsNotSameAsPointOfView(pl, t.pov) {
		title = text.JoinSentenceParts(title, pl.InAt(), t.enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
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
	case model.MultipartyTimelineEvent:
		if tev.DirectlyInvolves(observer) {
			return "" // no context needed
		}
		var ppl []string
		for _, p := range tev.GetPrincipals() {
			if p.IsUnknown() {
				continue
			}
			ppl = append(ppl, t.enc.EncodeModelLink(p.PreferredFullName, p))
		}

		return text.JoinList(ppl)

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

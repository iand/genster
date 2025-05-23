package narrative

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

func AgeQualifier(age int) string {
	if age == 0 {
		return "as an infant"
	} else if age < 10 {
		return "as a child"
	}
	return fmt.Sprintf("at the age of %s", text.CardinalNoun(age))
}

func WhoWhatWhenWhere[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser) string {
	var title string
	switch tev := ev.(type) {
	case model.IndividualTimelineEvent:
		title = enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(tev.GetPrincipal())), tev.GetPrincipal()).String()
	case model.UnionTimelineEvent:
		title = enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(tev.GetHusband())), tev.GetHusband()).String() + " and " + enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(tev.GetWife())), tev.GetWife()).String()
	case model.MultipartyTimelineEvent:
		var names []string
		for _, p := range tev.GetPrincipals() {
			names = append(names, enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(p)), p).String())
		}
		title = text.JoinList(names)
	}

	title = text.JoinSentenceParts(title, EventWhatWhenWhere(ev, enc, nc))

	return title
}

func WhoWhatWhenWherePov[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	var title string
	switch tev := ev.(type) {
	case model.IndividualTimelineEvent:
		title = enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(tev.GetPrincipal())), tev.GetPrincipal()).String()
	case model.UnionTimelineEvent:
		title = enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(tev.GetHusband())), tev.GetHusband()).String() + " and " + enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(tev.GetWife())), tev.GetWife()).String()
	case model.MultipartyTimelineEvent:
		var names []string
		for _, p := range tev.GetPrincipals() {
			names = append(names, enc.EncodeModelLink(enc.EncodeText(nc.FirstUse(p)), p).String())
		}
		title = text.JoinList(names)
	}

	title = text.JoinSentenceParts(title, EventWhatWhenWherePov(ev, enc, nc, pov))

	return title
}

func EventWhatWhenWhere[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser) string {
	return WhatWhenWhere(InferredWhat(ev, ev), ev.GetDate(), ev.GetPlace(), enc, nc)
}

func EventWhatWhere[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser) string {
	return WhatWhere(InferredWhat(ev, ev), ev.GetPlace(), enc, nc)
}

func EventWhenWhere[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser) string {
	return WhenWhere(ev.GetDate(), ev.GetPlace(), enc, nc)
}

func EventWhatWhenWherePov[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	return WhatWhenWherePov(InferredWhat(ev, ev), ev.GetDate(), ev.GetPlace(), enc, nc, pov)
}

func EventWhatWherePov[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	return WhatWherePov(InferredWhat(ev, ev), ev.GetPlace(), enc, nc, pov)
}

func EventWhenWherePov[T render.EncodedText](ev model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	return WhenWherePov(ev.GetDate(), ev.GetPlace(), enc, nc, pov)
}

func InferredWhat(w model.Whater, ev model.TimelineEvent) string {
	if ev.IsInferred() {
		return "is inferred to " + model.PresentPerfectWhat(w)
	}

	if !ev.GetDate().IsUnknown() {
		switch ev.GetDate().Derivation {
		case model.DateDerivationEstimated:
			return model.PassiveConditionalWhat(w, "probably")
		case model.DateDerivationCalculated:
			return "is calculated to " + model.PresentPerfectWhat(w)
		}
	}

	return model.PassiveWhat(w)
}

func WhatWhenWhere[T render.EncodedText](what string, dt *model.Date, pl *model.Place, enc render.TextEncoder[T], nc NameChooser) string {
	return text.JoinSentenceParts(what, WhenWhere(dt, pl, enc, nc))
}

func WhatWhenWherePov[T render.EncodedText](what string, dt *model.Date, pl *model.Place, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	return text.JoinSentenceParts(what, WhenWherePov(dt, pl, enc, nc, pov))
}

func WhatWhere[T render.EncodedText](what string, pl *model.Place, enc render.TextEncoder[T], nc NameChooser) string {
	if !pl.IsUnknown() {
		what = text.JoinSentenceParts(what, pl.InAt(), enc.EncodeModelLinkNamed(pl, nc, &model.POV{}).String())
	}
	return what
}

func WhatWherePov[T render.EncodedText](what string, pl *model.Place, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	if pl.IsUnknown() {
		return what
	} else if pl.SameAs(pov.Place) {
		what = text.JoinSentenceParts(what, "here")
	} else {
		what = text.JoinSentenceParts(what, pl.InAt(), enc.EncodeModelLinkNamed(pl, nc, pov).String())
	}

	return what
}

func WhatWhen[T render.EncodedText](what string, dt *model.Date, enc render.TextEncoder[T]) string {
	if !dt.IsUnknown() {
		return text.JoinSentenceParts(what, dt.When())
	}
	return what
}

func WhenWhere[T render.EncodedText](dt *model.Date, pl *model.Place, enc render.TextEncoder[T], nc NameChooser) string {
	title := ""
	if !dt.IsUnknown() {
		title = text.JoinSentenceParts(title, dt.When())
	}

	if !pl.IsUnknown() {
		title = text.JoinSentenceParts(title, pl.InAt(), enc.EncodeModelLinkNamed(pl, nc, &model.POV{}).String())
	}
	return title
}

func WhenWherePov[T render.EncodedText](dt *model.Date, pl *model.Place, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	title := ""
	if !dt.IsUnknown() {
		title = text.JoinSentenceParts(title, dt.When())
	}

	title = WhatWherePov(title, pl, enc, nc, pov)
	return title
}

func AgeWhenWhere[T render.EncodedText](ev model.IndividualTimelineEvent, enc render.TextEncoder[T], nc NameChooser) string {
	title := ""

	date := ev.GetDate()
	if !date.IsUnknown() {
		if age, ok := ev.GetPrincipal().AgeInYearsAt(ev.GetDate()); ok {
			title = text.JoinSentenceParts(title, AgeQualifier(age))
		}
		title = text.JoinSentenceParts(title, date.When())
	}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		title = text.JoinSentenceParts(title, pl.InAt(), enc.EncodeModelLinkNamed(pl, nc, &model.POV{}).String())
	}
	return title
}

func AgeWhenWherePov[T render.EncodedText](ev model.IndividualTimelineEvent, enc render.TextEncoder[T], nc NameChooser, pov *model.POV) string {
	title := ""

	date := ev.GetDate()
	if !date.IsUnknown() {
		if age, ok := ev.GetPrincipal().AgeInYearsAt(ev.GetDate()); ok {
			title = text.JoinSentenceParts(title, AgeQualifier(age))
		}
		title = text.JoinSentenceParts(title, date.When())
	}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		title = WhatWherePov(title, pl, enc, nc, pov)
	}
	return title
}

func FollowingWhatWhenWhere[T render.EncodedText](what string, dt *model.Date, pl *model.Place, preceding model.TimelineEvent, enc render.TextEncoder[T], nc NameChooser) string {
	detail := what

	if pl.SameAs(preceding.GetPlace()) {
		detail = text.JoinSentenceParts(detail, "there")
	}

	if !dt.IsUnknown() {
		suppressDate := false
		intervalDesc := ""
		in := preceding.GetDate().IntervalUntil(dt)
		y, m, d, ok := in.YMD()
		if ok {
			if y == 0 {
				if m == 0 {
					if d == 0 {
						intervalDesc = "the same day"
						suppressDate = true
					} else if d == 1 {
						intervalDesc = "the next day"
						suppressDate = true
					} else if d == 2 {
						intervalDesc = "two days later"
						suppressDate = true
					} else if d < 7 {
						intervalDesc = "a few days later"
					} else if d < 10 {
						intervalDesc = "just a week later"
					} else {
						intervalDesc = "a couple of weeks later"
					}
				} else {
					if m == 1 {
						intervalDesc = "the next month"
					} else if m < 4 {
						intervalDesc = "a few months later"
					} else {
						intervalDesc = "less than a year later"
					}
				}
			} else {
				intervalDesc = fmt.Sprintf("%d %s later", y, text.MaybePluralise("year", y))
			}
		} else {
			yrs, ok := in.WholeYears()
			if ok && yrs > 0 {
				intervalDesc = fmt.Sprintf("%d %s later", yrs, text.MaybePluralise("year", yrs))
			}
		}

		if intervalDesc != "" {
			detail = text.JoinSentenceParts(detail, intervalDesc)
		}

		if !suppressDate {
			detail = text.JoinSentenceParts(detail, dt.When())
		}
	}

	if !pl.IsUnknown() && !preceding.GetPlace().SameAs(pl) {
		detail = text.JoinSentenceParts(detail, pl.InAt(), enc.EncodeModelLinkNamed(pl, nc, &model.POV{}).String())
	}

	return detail
}

func DeathWhat(ev model.IndividualTimelineEvent, mode model.ModeOfDeath) string {
	if mode == model.ModeOfDeathNatural {
		return InferredWhat(ev, ev)
	}
	switch ev.(type) {
	case *model.DeathEvent:
		return InferredWhat(mode, ev)
	case *model.BurialEvent:
		return text.JoinSentenceParts(model.PassiveWhat(mode), "and", InferredWhat(ev, ev))
	case *model.CremationEvent:
		return text.JoinSentenceParts(model.PassiveWhat(mode), "and", InferredWhat(ev, ev))
	default:
		panic("unhandled deathlike event in DeathWhat")
	}
}

// WhoDoing returns a persons full or familiar name with their occupation as an aside if known.
func WhoDoing[T render.EncodedText](p *model.Person, dt *model.Date, enc render.TextEncoder[T], nc NameChooser) string {
	detail := enc.EncodeModelLinkNamed(p, nc, &model.POV{}).String()

	occ := p.OccupationAt(dt)
	if !occ.IsUnknown() {
		detail += ", " + occ.String() + ","
	}

	return detail
}

func PositionInFamily(p *model.Person) string {
	if p.ParentFamily == nil {
		return ""
	}
	if p.ParentFamily.Father.IsUnknown() && p.ParentFamily.Mother.IsUnknown() {
		return ""
	}

	if len(p.ParentFamily.Children) == 0 {
		return ""
	}

	var children []*model.Person
	if p.ParentFamily.Father.IsUnknown() {
		children = p.ParentFamily.Mother.Children
	} else if p.ParentFamily.Mother.IsUnknown() {
		children = p.ParentFamily.Father.Children
	} else {
		children = p.ParentFamily.Children
	}
	if len(children) == 0 {
		return text.LowerFirst(p.Gender.RelationToParentNoun())
	}

	if len(children) == 1 {
		if p.ParentFamily.AllChildrenKnown {
			return "only " + text.LowerFirst(p.Gender.RelationToParentNoun())
		}

		return "only known " + text.LowerFirst(p.Gender.RelationToParentNoun())
	}

	if !p.ParentFamily.AllChildrenKnown {
		return ""
	}

	if children[0].SameAs(p) {
		return "first child"
	}

	olderSameGender := 0
	youngerSameGender := 0
	olderSameGenderSurvived := 0
	childOlder := true
	for _, c := range children {
		if c.SameAs(p) {
			childOlder = false
			continue
		}
		if c.Gender == p.Gender {
			if childOlder {
				olderSameGender++
				if !c.DiedYoung {
					olderSameGenderSurvived++
				}
			} else {
				youngerSameGender++
			}
		}
	}

	if olderSameGender == 0 {
		return "eldest " + text.LowerFirst(p.Gender.RelationToParentNoun())
	}

	if youngerSameGender == 0 {
		return "youngest " + text.LowerFirst(p.Gender.RelationToParentNoun())
	}

	if olderSameGenderSurvived != olderSameGender {
		return text.OrdinalNoun(olderSameGenderSurvived+1) + " surviving " + text.LowerFirst(p.Gender.RelationToParentNoun())
	}

	return text.OrdinalNoun(olderSameGender+1) + " " + text.LowerFirst(p.Gender.RelationToParentNoun())
}

func PersonParentage[T render.EncodedText](p *model.Person, enc render.TextEncoder[T]) string {
	rel := PositionInFamily(p)
	if rel == "" {
		rel = text.LowerFirst(p.Gender.RelationToParentNoun())
	}
	intro := "the " + rel + " of "

	if p.Father.IsUnknown() {
		if p.Mother.IsUnknown() {
			return intro + "unknown parents"
		} else {
			return intro + enc.EncodeModelLink(enc.EncodeText(p.Mother.PreferredFullName), p.Mother).String()
		}
	} else {
		if p.Mother.IsUnknown() {
			return intro + enc.EncodeModelLink(enc.EncodeText(p.Father.PreferredFullName), p.Father).String()
		} else {
			return intro + enc.EncodeModelLink(enc.EncodeText(p.Father.PreferredFullName), p.Father).String() + " and " + enc.EncodeModelLink(enc.EncodeText(p.Mother.PreferredFullName), p.Mother).String()
		}
	}
}

func PersonSummary[T render.EncodedText](p *model.Person, enc render.TextEncoder[T], nc NameChooser, name T, includeBirth bool, includeParentage bool, activeTense bool, linkname bool, minimal bool) T {
	enc = &PersonLinkingTextEncoder[T]{enc}

	var empty T
	if !name.IsZero() {
		if p.Redacted {
			return enc.EncodeItalic(name)
		}

		if !linkname || enc.EncodeModelLink(empty, p).String() == "" {
			name = enc.EncodeItalic(name)
		} else {
			name = enc.EncodeModelLink(name, p)
		}

		if p.NickName != "" {
			name = enc.EncodeText(text.JoinSentenceParts(name.String(), fmt.Sprintf("(known as %s)", p.NickName)))
		}

	}

	if p.Epithet != "" {
		name = enc.EncodeText(text.AppendAside(name.String(), p.Epithet))
	}

	includeAgeAtDeathIfKnown := true
	var para text.Para
	if age, ok := p.AgeInYearsAtDeath(); ok && age < 14 {
		if !name.IsZero() {
			if age < 1 {
				para.StartSentence(name.String(), " died in infancy")
			} else {
				para.StartSentence(name.String(), fmt.Sprintf(" died age %s, ", text.CardinalNoun(age)))
			}
			includeAgeAtDeathIfKnown = false
			name = empty
			para.FinishSentence()
		}

		if p.BestBirthlikeEvent != nil && p.BestDeathlikeEvent != nil && p.BestBirthlikeEvent.GetPlace().SameAs(p.BestDeathlikeEvent.GetPlace()) {
			para.StartSentence(YoungPersonOnePlaceSummary(p, enc, nc, name, includeBirth, includeParentage, activeTense, linkname, minimal).String())
			return enc.EncodeText(para.Text())
		}
	}

	if includeBirth {
		birth := PersonBirthSummary(p, enc, nc, name, true, true, includeParentage, activeTense)
		if !birth.IsZero() {
			para.StartSentence(birth.String())
			if activeTense {
				name = empty
			} else {
				name = enc.EncodeText(p.Gender.SubjectPronoun())
			}
		}
	}

	marrs := PersonMarriageSummary(p, enc, nc, name, false, activeTense)
	if !marrs.IsZero() {
		para.StartSentence(marrs.String())
		if activeTense {
			name = empty
		} else {
			name = enc.EncodeText(p.Gender.SubjectPronoun())
		}
	}

	var immPhrases []string

	for _, ev := range p.Timeline {
		switch tev := ev.(type) {
		case *model.ImmigrationEvent:
			if tev.GetPlace().IsUnknown() {
				continue
			}

			when, ok := ev.GetDate().WhenYear()
			if !ok {
				continue
			}
			immPhrases = append(immPhrases, enc.EncodeWithCitations(enc.EncodeText(fmt.Sprintf("%s %s", tev.GetPlace().Name, when)), tev.GetCitations()).String())
		}
	}

	if len(immPhrases) > 0 {
		para.StartSentence("emigrated to", text.JoinList(immPhrases))
	}

	death := PersonDeathSummary(p, enc, nc, name, false, activeTense, minimal, includeAgeAtDeathIfKnown)
	if !death.IsZero() {
		para.StartSentence(death.String())
	}

	// TODO: life events
	// emigration
	// imprisonment

	finalDetail := ""
	if p.Unmarried {
		finalDetail = "never married"
		if p.Childless {
			finalDetail += " and had no children"
		}
	} else {
		if p.Childless {
			finalDetail += "had no children"
		}
	}

	if finalDetail != "" {
		para.StartSentence(p.Gender.SubjectPronoun(), finalDetail)
	}

	if para.IsEmpty() {
		para.StartSentence(name.String())
		para.StartSentence("nothing else is known about", p.Gender.PossessivePronounSingular(), "life")
	}

	return enc.EncodeText(para.Text())
}

func YoungPersonOnePlaceSummary[T render.EncodedText](p *model.Person, enc render.TextEncoder[T], nc NameChooser, name T, includeBirth bool, includeParentage bool, activeTense bool, linkname bool, minimal bool) T {
	var para text.Para
	para.StartSentence(name.String())

	var death *model.DeathEvent

	switch tev := p.BestDeathlikeEvent.(type) {
	case *model.DeathEvent:
		death = tev
	}

	var birthWhat string
	if activeTense {
		birthWhat = model.What(p.BestBirthlikeEvent)
	} else {
		birthWhat = model.PassiveWhat(p.BestBirthlikeEvent)
	}
	para.Continue(enc.EncodeWithCitations(enc.EncodeText(WhatWhenWhere(birthWhat, p.BestBirthlikeEvent.GetDate(), p.BestBirthlikeEvent.GetPlace(), enc, nc)), p.BestBirthlikeEvent.GetCitations()).String())

	var deathWhat string
	deathWhat = model.What(p.BestDeathlikeEvent)

	if death != nil {
		deathWhat = DeathWhat(death, p.ModeOfDeath)

		if !p.BestBirthlikeEvent.GetPlace().IsUnknown() {
			deathWhat += " there"
		}
	}
	para.Continue("and", enc.EncodeWithCitations(enc.EncodeText(WhatWhen(deathWhat, p.BestDeathlikeEvent.GetDate(), enc)), p.BestDeathlikeEvent.GetCitations()).String())

	if len(p.Associations) > 0 {
		for _, as := range p.Associations {
			if as.Kind != model.AssociationKindTwin {
				continue
			}
			twinLink := enc.EncodeModelLink(enc.EncodeText(as.Other.PreferredFamiliarName), as.Other)
			para.StartSentence(p.Gender.SubjectPronoun(), "was the twin to", enc.EncodeWithCitations(twinLink, as.Citations).String())
		}
	}

	return enc.EncodeText(para.Text())
}

func PersonBirthSummary[T render.EncodedText](p *model.Person, enc render.TextEncoder[T], nc NameChooser, name T, allowInferred bool, includeBirthDate bool, includeParentage bool, activeTense bool) T {
	var empty T
	var birth *model.BirthEvent
	var bev model.IndividualTimelineEvent

	if p.BestBirthlikeEvent == nil {
		return empty
	}
	switch tev := p.BestBirthlikeEvent.(type) {
	case *model.BirthEvent:
		if allowInferred || !tev.IsInferred() {
			birth = tev
			bev = tev
		}
		for _, ev := range p.Timeline {
			if !ev.DirectlyInvolves(p) {
				continue
			}
			if bapev, ok := ev.(*model.BaptismEvent); ok {
				if bev == nil {
					bev = bapev
				} else if !bev.GetDate().IsMorePreciseThan(bapev.GetDate()) {
					bev = bapev
				}
			}
		}
	case *model.BaptismEvent:
		bev = tev
	}

	tense := func(st string) T {
		if activeTense {
			return enc.EncodeText(text.StripWasIs(st))
		}
		return enc.EncodeText(st)
	}

	if bev == nil {
		return empty
	}

	var para text.Para
	para.StartSentence(name.String())

	if includeBirthDate {
		if birth != nil {
			if _, ok := bev.(*model.BaptismEvent); ok {
				if yrs, ok := birth.GetDate().WholeYearsUntil(bev.GetDate()); ok && yrs > 1 {
					para.Continue(tense("was born").String(), birth.GetDate().When(), "and")
				}
			}
		}
		para.Continue(enc.EncodeWithCitations(tense(EventWhatWhenWhere(bev, enc, nc)), bev.GetCitations()).String())
	} else {
		para.Continue(enc.EncodeWithCitations(tense(EventWhatWhere(bev, enc, nc)), bev.GetCitations()).String())
	}

	if includeParentage {
		para.AppendClause(PersonParentage(p, enc))
	}

	if len(p.Associations) > 0 {
		for _, as := range p.Associations {
			if as.Kind != model.AssociationKindTwin {
				continue
			}
			twinLink := enc.EncodeModelLink(enc.EncodeText(as.Other.PreferredFamiliarName), as.Other)
			para.Continue(text.UpperFirst(p.Gender.SubjectPronoun()), "was the twin to", enc.EncodeWithCitations(twinLink, as.Citations).String())
		}
	}

	return enc.EncodeText(para.Text())
}

type DeathSummarizer struct {
	DeathEvent        *model.DeathEvent
	BurialEvent       *model.BurialEvent
	CremationEvent    *model.CremationEvent
	ModeOfDeath       model.ModeOfDeath
	CauseOfDeath      *model.Fact
	AgeInYearsAtDeath int
}

func PersonDeathSummary[T render.EncodedText](p *model.Person, enc render.TextEncoder[T], nc NameChooser, name T, allowInferred bool, activeTense bool, minimal bool, includeAge bool) T {
	var empty T
	// var death *model.DeathEvent
	var bev model.IndividualTimelineEvent

	if p.BestDeathlikeEvent == nil {
		return empty
	}

	usingBurial := false
	switch tev := p.BestDeathlikeEvent.(type) {
	case *model.DeathEvent:
		if allowInferred || !tev.IsInferred() {
			// death = tev
			bev = tev
			usingBurial = false
		}
		for _, ev := range p.Timeline {
			if !ev.DirectlyInvolves(p) {
				continue
			}
			if tev, ok := ev.(*model.BurialEvent); ok {
				if bev == nil || tev.GetDate().SortsBefore(bev.GetDate()) {
					bev = tev
					usingBurial = true
				}
			} else if tev, ok := ev.(*model.CremationEvent); ok {
				if bev == nil || tev.GetDate().SortsBefore(bev.GetDate()) {
					bev = tev
					usingBurial = true
				}
			}
		}
	case *model.BurialEvent:
		bev = tev
		usingBurial = true
	}

	if bev == nil {
		return empty
	}

	tense := func(st string) T {
		if activeTense {
			return enc.EncodeText(text.StripWasIs(st))
		}
		return enc.EncodeText(st)
	}

	var para text.Para
	para.StartSentence(name.String())
	// deathWhat := model.PassiveWhat(bev)
	deathWhat := DeathWhat(bev, p.ModeOfDeath)

	para.Continue(enc.EncodeWithCitations(tense(WhatWhenWhere(deathWhat, bev.GetDate(), bev.GetPlace(), enc, nc)), bev.GetCitations()).String())

	if includeAge && !usingBurial {
		if age, ok := p.AgeInYearsAt(bev.GetDate()); ok {
			if age < 1 {
				page, ok := p.PreciseAgeAt(bev.GetDate())
				if !ok {
					para.Continue("in infancy")
				} else {
					para.Continue("aged", page.Rough())
				}
			} else {
				para.Continue(fmt.Sprintf("at the age of %s", text.CardinalNoun(age)))
			}
		}
	}

	if p.CauseOfDeath != nil && !minimal {
		para.StartSentence(p.Gender.PossessivePronounSingular(), "death was attributed to", enc.EncodeWithCitations(enc.EncodeText(p.CauseOfDeath.Detail), p.CauseOfDeath.Citations).String())
	}

	return enc.EncodeText(para.Text())
}

func PersonMarriageSummary[T render.EncodedText](p *model.Person, enc render.TextEncoder[T], nc NameChooser, name T, allowInferred bool, activeTense bool) T {
	var empty T
	tense := func(st string) T {
		if activeTense {
			return enc.EncodeText(text.StripWasIs(st))
		}
		return enc.EncodeText(st)
	}

	var fams []*model.Family

	for _, f := range p.Families {
		if f.Bond != model.FamilyBondMarried && f.Bond != model.FamilyBondLikelyMarried {
			continue
		}
		if f.BestStartEvent == nil {
			continue
		}
		fams = append(fams, f)
	}

	var marrs []string
	if len(fams) == 0 {
		return empty
	} else if len(fams) == 1 {
		// more detail
		f := fams[0]
		other := f.OtherParent(p)
		what := f.BestStartEvent.What() + " " + enc.EncodeModelLink(enc.EncodeText(other.PreferredFamiliarFullName), other).String()
		marrs = append(marrs, enc.EncodeWithCitations(tense(WhatWhenWhere(what, f.BestStartEvent.GetDate(), f.BestStartEvent.GetPlace(), enc, nc)), f.BestStartEvent.GetCitations()).String())
	} else {
		var prev model.TimelineEvent
		for _, f := range fams {
			other := f.OtherParent(p)

			y, _ := f.BestStartEvent.GetDate().AsYear()

			if prev != nil {
				what := enc.EncodeModelLink(enc.EncodeText(other.PreferredFamiliarFullName), other)
				marrs = append(marrs, enc.EncodeWithCitations(tense(WhatWhenWhere(what.String(), y, f.BestStartEvent.GetPlace(), enc, nc)), f.BestStartEvent.GetCitations()).String())
			} else {
				what := f.BestStartEvent.What() + " " + enc.EncodeModelLink(enc.EncodeText(other.PreferredFamiliarFullName), other).String()
				marrs = append(marrs, enc.EncodeWithCitations(tense(WhatWhenWhere(what, y, f.BestStartEvent.GetPlace(), enc, nc)), f.BestStartEvent.GetCitations()).String())
			}

			prev = f.BestStartEvent
		}
	}

	var para text.Para
	para.StartSentence(name.String())
	para.Continue(text.JoinList(marrs))
	return enc.EncodeText(para.Text())
}

func GenerateOlb(p *model.Person) string {
	const (
		Mundane     = 1
		Interesting = 2
		Unusual     = 3
		Unique      = 4
	)

	log := false
	logger := logging.With("id", p.ID, "name", p.PreferredFullName)

	type BioFacts struct {
		BirthYear int
		// BirthYearDesc         string
		BirthPlace     *model.Place
		CountryOfBirth *model.Place
		DeathYear      int
		DeathYearDesc  string
		DeathPlace     *model.Place
		DeathType      string
		// CountryOfDeath        *model.Place
		AgeAtDeath            int
		NumberOfMarriages     int
		AgeAtFirstMarriage    int
		AgeAtFirstSpouseDeath int
		NumberOfDivorces      int
		NumberOfAnnulments    int
		Spouses               []*model.Person
		NumberOfChildren      int
		IllegitimateChildren  int
		NumberOfSiblings      int // TODO
		PositionInFamily      int
		AgeAtDeathOfFather    int
		AgeAtDeathOfMother    int
		OrphanedAtAge         int
		TravelEvents          int
		Suicide               bool
	}

	bf := BioFacts{
		AgeAtDeath:            -1, // unknown
		NumberOfChildren:      len(p.Children),
		IllegitimateChildren:  -1,
		NumberOfMarriages:     0,
		NumberOfDivorces:      0,
		NumberOfAnnulments:    0,
		NumberOfSiblings:      -1,
		PositionInFamily:      -1,
		AgeAtDeathOfFather:    -1,
		AgeAtDeathOfMother:    -1,
		OrphanedAtAge:         -1,
		AgeAtFirstSpouseDeath: -1,
	}

	if p.BestBirthlikeEvent != nil {
		if year, ok := p.BestBirthlikeEvent.GetDate().Year(); ok {
			bf.BirthYear = year

			// 	whenYear, ok := p.BestBirthlikeEvent.GetDate().WhenYear()
			// 	if ok {
			// 		bf.BirthYearDesc = "born " + whenYear
			// 	}

			// 	if p.BestBirthlikeEvent.IsInferred() {
			// 		bf.BirthYearDesc = "likely " + bf.BirthYearDesc
			// 	}
		}
		if !p.BestBirthlikeEvent.GetPlace().IsUnknown() {

			pl := p.BestBirthlikeEvent.GetPlace()
			bf.BirthPlace = pl
			// if !pl.Region.IsUnknown() {
			// 	bf.BirthPlace = pl.Region
			// }
			if !pl.Country.IsUnknown() {
				bf.CountryOfBirth = pl.Country
			}
		}
	}

	if p.BestDeathlikeEvent != nil {
		if year, ok := p.BestDeathlikeEvent.GetDate().Year(); ok {
			bf.DeathYear = year
			if bf.BirthYear != 0 {
				if age, ok := p.AgeInYearsAt(p.BestDeathlikeEvent.GetDate()); ok {
					bf.AgeAtDeath = age
				}
			}
			if p.ModeOfDeath == model.ModeOfDeathSuicide {
				bf.Suicide = true
			}

			bf.DeathType = "died"
			if bf.Suicide {
				bf.DeathType = "killed " + p.Gender.ReflexivePronoun()
			}

			whenYear, ok := p.BestDeathlikeEvent.GetDate().WhenYear()
			if ok {
				bf.DeathYearDesc = whenYear
			}

			if p.BestDeathlikeEvent.IsInferred() {
				bf.DeathType = "likely " + bf.DeathType
			}

		}

		if !p.BestDeathlikeEvent.GetPlace().IsUnknown() {
			pl := p.BestDeathlikeEvent.GetPlace()
			bf.DeathPlace = pl
			// if !pl.Region.IsUnknown() {
			// 	bf.DeathPlace = pl.Region
			// }
			// if !pl.Country.IsUnknown() {
			// 	bf.CountryOfDeath = pl.Country
			// }
		}
	}

	if !p.Mother.IsUnknown() {
		if p.BestDeathlikeEvent != nil && !p.BestDeathlikeEvent.GetDate().IsUnknown() {
			if p.Mother.BestDeathlikeEvent != nil && !p.Mother.BestDeathlikeEvent.GetDate().IsUnknown() && !p.BestDeathlikeEvent.GetDate().SortsBefore(p.Mother.BestDeathlikeEvent.GetDate()) {
				if age, ok := p.AgeInYearsAt(p.Mother.BestDeathlikeEvent.GetDate()); ok {
					bf.AgeAtDeathOfMother = age
				}
			}
		}
	}

	if !p.Father.IsUnknown() && p.BestDeathlikeEvent != nil && !p.BestDeathlikeEvent.GetDate().IsUnknown() {
		if p.Father.BestDeathlikeEvent != nil && !p.Father.BestDeathlikeEvent.GetDate().IsUnknown() && !p.BestDeathlikeEvent.GetDate().SortsBefore(p.Father.BestDeathlikeEvent.GetDate()) {
			if age, ok := p.AgeInYearsAt(p.Father.BestDeathlikeEvent.GetDate()); ok {
				bf.AgeAtDeathOfFather = age
			}
		}
	}

	if bf.AgeAtDeathOfFather != -1 && bf.AgeAtDeathOfFather < 18 && bf.AgeAtDeathOfMother != -1 && bf.AgeAtDeathOfMother < 18 {
		if bf.AgeAtDeathOfFather > bf.AgeAtDeathOfMother {
			bf.OrphanedAtAge = bf.AgeAtDeathOfFather
		} else {
			bf.OrphanedAtAge = bf.AgeAtDeathOfMother
		}
	}

	for _, fam := range p.Families {
		if fam.Bond == model.FamilyBondMarried || fam.Bond == model.FamilyBondLikelyMarried {
			other := fam.OtherParent(p)
			if bf.BirthYear != 0 && bf.NumberOfMarriages == 0 && fam.BestStartDate != nil {
				if age, ok := p.AgeInYearsAt(fam.BestStartDate); ok {
					bf.AgeAtFirstMarriage = age
				}
				if !other.IsUnknown() && other.BestDeathlikeEvent != nil && p.BestDeathlikeEvent != nil && !p.BestDeathlikeEvent.GetDate().SortsBefore(other.BestDeathlikeEvent.GetDate()) {
					if age, ok := p.AgeInYearsAt(other.BestDeathlikeEvent.GetDate()); ok {
						bf.AgeAtFirstSpouseDeath = age
					}
				}

			}

			bf.NumberOfMarriages++
			if !other.IsUnknown() {
				bf.Spouses = append(bf.Spouses, other)
			}
		} else {
			if fam.OtherParent(p).IsUnknown() {
				if bf.IllegitimateChildren == -1 {
					bf.IllegitimateChildren = len(fam.Children)
				} else {
					bf.IllegitimateChildren += len(fam.Children)
				}
			}
		}
	}

	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) {
			continue
		}
		switch ev.(type) {
		case *model.DivorceEvent:
			bf.NumberOfDivorces++
		case *model.AnnulmentEvent:
			bf.NumberOfAnnulments++
		case *model.ArrivalEvent:
			bf.TravelEvents++
		case *model.DepartureEvent:
			bf.TravelEvents++
		}
	}

	var clauses []string

	// a pseudo random sequence for predictably choosing phrase components
	seq := bf.BirthYear + bf.DeathYear + bf.NumberOfChildren + len(p.PreferredFullName)

	// Phrasing examples

	// remained unmarried and died without children in [year] in [place].”
	// a lifelong resident of [region/town],
	// born out of wedlock in [place] in [year],
	// lived most of their life in [region]
	// the unmarried mother of [number] children

	// Two marriages
	//  - married twice, first to [Name] and later to [Name]
	//  - first married [Name] in [place], and remarried in [year]
	//  - was widowed young and later remarried

	// No known marriages
	// - never married
	// - remained single throughout life
	// - had no known spouse

	// died of tuberculosis at 24
	// was lost at sea aged 20
	// killed in a mining accident in 1873
	// succumbed to typhoid as a young woman
	// died in childbirth, along with her infant daughter

	// emigration
	// emigrated to Australia in 1852
	// left England for Canada as a young man
	// settled in the United States after 1871
	// was born in Ireland and later moved to Manchester
	// returned to England after decades abroad

	// Some flags that can be set along the way to control output
	childrenMentioned := false
	_ = childrenMentioned

	hasIllegitimateClause := false
	_ = hasIllegitimateClause

	// -----------------------------------------------------------------------------
	// Origin
	// -----------------------------------------------------------------------------
	originPhrase := ""

	// Statement about birth
	isNotableCountry := func(pl *model.Place) bool {
		if pl.IsUnknown() {
			return false
		}
		if pl.Country.IsUnknown() {
			return false
		}
		switch pl.Country.Name {
		case "England":
			return false
		case "United Kingdom":
			return false
		default:
			return true
		}
	}

	if p.Illegitimate && p.Father.IsUnknown() {
		clause := ChooseFrom(seq, "born out of wedlock", "born to a single mother", "born to an unmarried mother")
		// if !p.Mother.IsUnknown() {
		// 	clause += " " + p.Gender.RelationToParentNoun() + " of " + p.Mother.PreferredFamiliarFullName
		// }

		hasIllegitimateClause = true
		clauses = append(clauses, clause)
	}

	if !bf.BirthPlace.IsUnknown() {
		if bf.BirthPlace.PlaceType == model.PlaceTypeBuilding && bf.BirthPlace.BuildingKind == model.BuildingKindWorkhouse {
			if !bf.BirthPlace.Region.IsUnknown() {
				originPhrase = "born in a " + bf.BirthPlace.Region.Name + " workhouse"
			} else {
				originPhrase = "born in a workhouse"
			}
		} else {
			if !bf.BirthPlace.Region.IsUnknown() && !isNotableCountry(bf.BirthPlace) {
				place := bf.BirthPlace.Region.Adjective
				if place == "" {
					place = bf.BirthPlace.Region.Name
				}
				originPhrase = place + "-born"

			} else if !bf.BirthPlace.Country.IsUnknown() {
				place := bf.BirthPlace.Country.Adjective
				if place == "" {
					place = bf.BirthPlace.Country.Name
				}
				originPhrase = place + "-born"
			}
			if originPhrase != "" {
				if p.Epithet != "" {
					originPhrase += " " + p.Epithet
				} else if !bf.BirthPlace.IsUnknown() &&
					!p.Father.IsUnknown() && p.Father.BestBirthlikeEvent != nil && !p.Father.BestBirthlikeEvent.GetPlace().IsUnknown() && p.Father.BestBirthlikeEvent.GetPlace().Country != nil &&
					!p.Mother.IsUnknown() && p.Mother.BestBirthlikeEvent != nil && !p.Mother.BestBirthlikeEvent.GetPlace().IsUnknown() {
					if p.Father.BestBirthlikeEvent.GetPlace().Country.SameAs(p.Mother.BestBirthlikeEvent.GetPlace().Country) &&
						!p.Father.BestBirthlikeEvent.GetPlace().Country.SameAs(bf.BirthPlace.Country) {
						log = true

						originPhrase = text.JoinSentenceParts(originPhrase, "of", p.Father.BestBirthlikeEvent.GetPlace().Country.Adjective, "descent")
					}
				} else if p.Gender == model.GenderFemale && bf.NumberOfChildren > 0 {
					originPhrase += " mother of " + text.CardinalNoun(bf.NumberOfChildren)
					childrenMentioned = true
				}
			}

		}
	}

	if originPhrase == "" {
		if p.Epithet != "" {
			originPhrase = p.Epithet
			originPhrase = text.JoinSentenceParts(originPhrase, ChooseFrom(seq,
				"of unknown origin",
				"of uncertain origin",
				"of unclear origin",
				"with unknown birthplace",
			))
		} else {
			originPhrase = ChooseFrom(seq,
				"of unknown origin",
				"of uncertain origin",
				"of unclear origin",
				"birthplace not known",
			)
		}
	}

	// Intro statement
	if p.NickName != "" {
		log = true
		originPhrase = text.AppendClause(originPhrase, "known as "+p.NickName)
	}

	if originPhrase != "" {
		clauses = append(clauses, originPhrase)
	}

	// -----------------------------------------------------------------------------
	// Early life
	// -----------------------------------------------------------------------------

	parentDeathDesc := func(seq int, gender model.Gender, age int) string {
		switch {
		case age == 0:
			return ChooseFrom(seq,
				"as an infant",
				"shortly after birth",
				"in the first months of life",
				"before reaching "+gender.PossessivePronounSingular()+" first year",
			)
		case age < 5:
			return ChooseFrom(seq,
				"in early childhood",
				"at a very young age",
				"before the age of five",
			)
		case age < 10:
			return ChooseFrom(seq,
				"while still a child",
				"before reaching ten",
				"when still quite young",
				"during "+gender.PossessivePronounSingular()+" early years",
			)
		case age < 16:
			return ChooseFrom(seq,
				"as a teenager",
				"in adolescence",
				"before adulthood",
				"during formative years",
				"while still growing up",
			)
		default:
			return ChooseFrom(seq,
				"in early life",
				"while growing up",
			)
		}
	}

	earlyLifePhraseParts := []string{}
	if p.Twin {
		earlyLifePhraseParts = append(earlyLifePhraseParts, "one of a pair of twins")
	}

	if p.PhysicalImpairment {
		earlyLifePhraseParts = append(earlyLifePhraseParts, "physically impaired")
	}

	if p.MentalImpairment {
		earlyLifePhraseParts = append(earlyLifePhraseParts, "mentally impaired")
	}

	if p.Deaf {
		earlyLifePhraseParts = append(earlyLifePhraseParts, "deaf")
	}

	if p.Blind {
		earlyLifePhraseParts = append(earlyLifePhraseParts, "blind")
	}

	if bf.OrphanedAtAge > -1 && bf.OrphanedAtAge < 18 {
		prefix := ChooseFrom(seq,
			"orphaned ",
			"lost both parents ",
			"both parents died ",
			"both parents passed away ",
			"left without parents ",
			"became an orphan ",
			"was orphaned ",
			"was left parentless ",
			"had no surviving parents ")
		earlyLifePhraseParts = append(earlyLifePhraseParts, prefix+parentDeathDesc(seq, p.Gender, bf.OrphanedAtAge))
	} else if bf.AgeAtDeathOfMother > -1 && bf.AgeAtDeathOfMother < 18 {
		prefix := ChooseFrom(seq,
			"lost mother ",
			"mother died ",
			"mother passed away ")
		earlyLifePhraseParts = append(earlyLifePhraseParts, prefix+parentDeathDesc(seq, p.Gender, bf.AgeAtDeathOfMother))
	} else if bf.AgeAtDeathOfFather > -1 && bf.AgeAtDeathOfFather < 18 {
		prefix := ChooseFrom(seq,
			"lost father ",
			"father died ",
			"father passed away ")
		earlyLifePhraseParts = append(earlyLifePhraseParts, prefix+parentDeathDesc(seq, p.Gender, bf.AgeAtDeathOfFather))
	}

	earlyLifePhrase := strings.Join(earlyLifePhraseParts, ", ")

	if earlyLifePhrase != "" {
		clauses = append(clauses, earlyLifePhrase)
	}

	// -----------------------------------------------------------------------------
	// Notable fact
	// -----------------------------------------------------------------------------
	notableFactPhrase := ""
	if p.Notable != "" {
		notableFactPhrase = p.Notable
	} else {
		if !bf.DeathPlace.IsUnknown() && !bf.BirthPlace.IsUnknown() {
			if bf.DeathPlace.District.SameAs(bf.BirthPlace.District) {
				samePlaceAllLife := true
				for _, ev := range p.Timeline {
					if _, ok := ev.(*model.ResidenceRecordedEvent); ok {
						pl := ev.GetPlace()
						if pl.IsUnknown() || !bf.DeathPlace.District.SameAs(pl.District) {
							samePlaceAllLife = false
						}
					}
				}

				if samePlaceAllLife {
					notableFactPhrase = ChooseFrom(seq,
						"spent entire life in "+bf.BirthPlace.District.Name,
						"lived entire life in "+bf.BirthPlace.District.Name,
					)
				}
			}
		}
	}

	if notableFactPhrase != "" {
		clauses = append(clauses, notableFactPhrase)
	}
	_ = childrenMentioned

	// Statement about families and children
	legitimateChildren := bf.NumberOfChildren
	if bf.IllegitimateChildren != -1 {
		legitimateChildren -= bf.IllegitimateChildren
	}

	if p.Childless && bf.AgeAtDeath > 18 {
		clauses = append(clauses, "had no children")
	} else if p.Gender.IsFemale() || bf.NumberOfChildren == 0 {
		if bf.IllegitimateChildren == 1 {
			clauses = append(clauses, "had one child with an unknown father")
		} else if bf.IllegitimateChildren > 1 {
			clauses = append(clauses, "had "+text.SmallCardinalNoun(bf.IllegitimateChildren)+" children with unknown fathers")
		}

		if p.Unmarried && bf.AgeAtDeath > 18 {
			clauses = append(clauses, "never married")
		} else if bf.NumberOfMarriages > 0 {
			if bf.AgeAtFirstMarriage > 0 && bf.AgeAtFirstMarriage < 18 {
				if bf.NumberOfMarriages == 1 && len(bf.Spouses) > 0 {
					clauses = append(clauses, "married "+bf.Spouses[0].PreferredFamiliarFullName+" at "+strconv.Itoa(bf.AgeAtFirstMarriage))
				} else if bf.NumberOfMarriages == 2 {
					clauses = append(clauses, "married at "+strconv.Itoa(bf.AgeAtFirstMarriage)+" then later remarried")
				} else {
					clauses = append(clauses, "married at "+strconv.Itoa(bf.AgeAtFirstMarriage)+" then "+text.SmallCardinalNoun(bf.NumberOfMarriages-1)+" more times")
				}
			} else {
				if bf.NumberOfMarriages == 1 && len(bf.Spouses) > 0 {
					clauses = append(clauses, "married "+bf.Spouses[0].PreferredFamiliarFullName)
				} else {
					clauses = append(clauses, "married "+text.MultiplicativeAdverb(bf.NumberOfMarriages))
				}
			}
		}

		if legitimateChildren == 1 {
			clauses = append(clauses, "had one child")
		} else if legitimateChildren > 1 {
			clauses = append(clauses, fmt.Sprintf("had %s children", text.SmallCardinalNoun(legitimateChildren)))
		}
	} else {
		// male or has no children

		clause := ""
		if bf.NumberOfChildren == 1 {
			if bf.IllegitimateChildren == 1 {
				clause += "had one child with an unknown mother"
			} else {
				clause += "had one child"
			}
		} else if bf.NumberOfChildren > 1 {
			clause += fmt.Sprintf("had %s children", text.SmallCardinalNoun(bf.NumberOfChildren))
		}

		if bf.NumberOfMarriages == 1 {
			clause += " with his wife"
			if len(bf.Spouses) > 0 {
				clause += " " + bf.Spouses[0].PreferredFamiliarName
			}
		} else if bf.NumberOfMarriages > 1 {
			clause += " with " + text.SmallCardinalNoun(bf.NumberOfMarriages) + " wives"
		}

		clauses = append(clauses, clause)

		if bf.IllegitimateChildren > 0 {
			if bf.IllegitimateChildren == bf.NumberOfChildren {
				if bf.IllegitimateChildren == 2 {
					clauses = append(clauses, "both with unknown mothers")
				} else if bf.IllegitimateChildren > 2 {
					clauses = append(clauses, "all with unknown mothers")
				}
			} else {
				clauses = append(clauses, text.SmallCardinalNoun(bf.IllegitimateChildren)+" with unknown mothers")
			}
		}
	}

	if bf.NumberOfMarriages == 1 && bf.AgeAtFirstSpouseDeath > 0 && bf.AgeAtFirstSpouseDeath < 40 {
		log = true
		if p.Gender.IsFemale() {
			clauses = append(clauses, "widowed at "+strconv.Itoa(bf.AgeAtFirstSpouseDeath))
		} else {
			clauses = append(clauses, "widower at "+strconv.Itoa(bf.AgeAtFirstSpouseDeath))
		}
	}

	if bf.NumberOfDivorces > 0 {
		if bf.NumberOfDivorces < bf.NumberOfMarriages {
			clauses = append(clauses, "divorced "+text.MultiplicativeAdverb(bf.NumberOfDivorces))
		} else if bf.NumberOfDivorces == bf.NumberOfMarriages && bf.NumberOfDivorces == 1 {
			clauses = append(clauses, "later divorced")
		}
	}

	if bf.NumberOfAnnulments > 0 {
		if bf.NumberOfAnnulments < bf.NumberOfMarriages {
			clauses = append(clauses, "anulled "+text.MultiplicativeAdverb(bf.NumberOfDivorces))
		} else if bf.NumberOfAnnulments == bf.NumberOfMarriages && bf.NumberOfAnnulments == 1 {
			clauses = append(clauses, "later anulled")
		}
	}

	// TODO: suicide
	// TODO: imprisoned
	// TODO: deported

	// if p.Pauper {
	// 	clauses = append(clauses, "pauper")
	// }

	// -----------------------------------------------------------------------------
	// Death phrase
	// -----------------------------------------------------------------------------
	deathPhrase := ""
	switch p.ModeOfDeath {
	case model.ModeOfDeathLostAtSea:
		log = true
		deathPhrase = "lost at sea"
	case model.ModeOfDeathKilledInAction:
		log = true
		deathPhrase = "killed in action"
	case model.ModeOfDeathDrowned:
		log = true
		deathPhrase = "drowned"
	case model.ModeOfDeathExecuted:
		log = true
		deathPhrase = "executed"
	case model.ModeOfDeathChildbirth:
		log = true
		deathPhrase = ChooseFrom(seq,
			"died in childbirth",
			"died while giving birth",
		)
	case model.ModeOfDeathSuicide:
		log = true
		deathPhrase = "killed " + p.Gender.ReflexivePronoun()
	}

	deathPlace := ""

	if deathPlace != "" {
		if deathPhrase == "" {
			deathPhrase = "died"
		}
		deathPhrase += " " + deathPlace
	}

	deathAge := ""
	if bf.AgeAtDeath != -1 {
		switch {
		case bf.AgeAtDeath == 0:
			deathAge = ChooseFrom(seq,
				"while an infant",
				"shortly after birth",
				"in the first months of life",
				"before reaching "+p.Gender.PossessivePronounSingular()+" first year",
			)
		case bf.AgeAtDeath < 5:
			deathAge = ChooseFrom(seq,
				"in early childhood",
				"at a very young age",
				"before the age of five",
			)
		case bf.AgeAtDeath < 10:
			deathAge = ChooseFrom(seq,
				"while still a child",
				"before reaching ten",
				"when still quite young",
				"during "+p.Gender.PossessivePronounSingular()+" early years",
			)
		case bf.AgeAtDeath < 16:
			deathAge = ChooseFrom(seq,
				"as a teenager",
				"in adolescence",
				"before adulthood",
				"during formative years",
				"while still growing up",
			)
		case bf.AgeAtDeath < 30:
			deathAge = ChooseFrom(seq,
				"while a young "+p.Gender.Noun(),
				"before "+p.Gender.SubjectPronounWithLink()+" "+strconv.Itoa(bf.AgeAtDeath+1),
				"before "+p.Gender.SubjectPronoun()+" reached "+strconv.Itoa(bf.AgeAtDeath+1),
				"before the age of "+strconv.Itoa(bf.AgeAtDeath+1),
			)
		default:
			deathAge = "at the age of " + strconv.Itoa(bf.AgeAtDeath)
		}
	} else {
		deathAge = "DEATH AGE " + strconv.Itoa(bf.AgeAtDeath)
	}

	if deathAge != "" {
		if deathPhrase == "" {
			deathPhrase = "died"
		}
		deathPhrase += " " + deathAge
	}

	if deathPhrase != "" {
		clauses = append(clauses, deathPhrase)
	}

	if len(clauses) == 0 {
		return ""
	}

	olb := strings.Join(clauses, ", ")

	if olb != "" {
		olb = text.FinishSentence(text.UpperFirst(olb))
	}
	if log {
		logger.Warn("generated olb: " + olb)
	} else {
		logger.Debug("generated olb: " + olb)
	}
	return olb
}

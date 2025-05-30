package narrative

import (
	"fmt"

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

	if p.Olb != "" {
		name = enc.EncodeText(text.AppendAside(name.String(), p.Olb))
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

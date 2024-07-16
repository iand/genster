package site

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/place"
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

func WhoWhatWhenWhere(ev model.TimelineEvent, enc render.PageMarkdownEncoder) string {
	var title string
	switch tev := ev.(type) {
	case model.IndividualTimelineEvent:
		title = enc.EncodeModelLink(tev.GetPrincipal().PreferredFullName, tev.GetPrincipal())
	case model.UnionTimelineEvent:
		title = enc.EncodeModelLink(tev.GetHusband().PreferredFullName, tev.GetHusband()) + " and " + enc.EncodeModelLink(tev.GetWife().PreferredFullName, tev.GetWife())
	case model.MultipartyTimelineEvent:
		var names []string
		for _, p := range tev.GetPrincipals() {
			names = append(names, enc.EncodeModelLink(p.PreferredFullName, p))
		}
		title = text.JoinList(names)
	}

	title = text.JoinSentenceParts(title, EventWhatWhenWhere(ev, enc))

	return title
}

func EventWhatWhenWhere(ev model.TimelineEvent, enc render.PageMarkdownEncoder) string {
	return WhatWhenWhere(InferredWhat(ev, ev), ev.GetDate(), ev.GetPlace(), enc)
}

func EventWhatWhere(ev model.TimelineEvent, enc render.PageMarkdownEncoder) string {
	return WhatWhere(InferredWhat(ev, ev), ev.GetPlace(), enc)
}

func EventWhenWhere(ev model.TimelineEvent, enc render.PageMarkdownEncoder) string {
	return WhenWhere(ev.GetDate(), ev.GetPlace(), enc)
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

func WhatWhenWhere(what string, dt *model.Date, pl *model.Place, enc render.PageMarkdownEncoder) string {
	return text.JoinSentenceParts(what, WhenWhere(dt, pl, enc))
}

func WhatWhere(what string, pl *model.Place, enc render.PageMarkdownEncoder) string {
	if !pl.IsUnknown() {
		what = text.JoinSentenceParts(what, pl.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return what
}

func WhenWhere(dt *model.Date, pl *model.Place, enc render.PageMarkdownEncoder) string {
	title := ""
	if !dt.IsUnknown() {
		title = text.JoinSentenceParts(title, dt.When())
	}

	if !pl.IsUnknown() {
		title = text.JoinSentenceParts(title, pl.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return title
}

func AgeWhenWhere(ev model.IndividualTimelineEvent, enc render.PageMarkdownEncoder) string {
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
		title = text.JoinSentenceParts(title, pl.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return title
}

func FollowingWhatWhenWhere(what string, dt *model.Date, pl *model.Place, preceding model.TimelineEvent, enc render.PageMarkdownEncoder) string {
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
		detail = text.JoinSentenceParts(detail, pl.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}

	return detail
}

func DeathWhat(ev model.IndividualTimelineEvent, mode model.ModeOfDeath) string {
	switch ev.(type) {
	case *model.DeathEvent:
		if mode == model.ModeOfDeathNatural {
			return InferredWhat(ev, ev)
		}
		return InferredWhat(mode, ev)
	case *model.BurialEvent:
		return text.JoinSentenceParts(model.PassiveWhat(mode), "and", InferredWhat(ev, ev))
	case *model.CremationEvent:
		return text.JoinSentenceParts(model.PassiveWhat(mode), "and", InferredWhat(ev, ev))
	default:
		panic("unhandled deathlike event in DeathWhat")
	}
}

// WhoFormalDoing returns a persons unique or full name with their occupation as an aside if known.
func WhoFormalDoing(p *model.Person, dt *model.Date, enc render.PageMarkdownEncoder) string {
	detail := enc.EncodeModelLinkDedupe(p.PreferredUniqueName, p.PreferredFamiliarFullName, p)

	occ := p.OccupationAt(dt)
	if !occ.IsUnknown() {
		detail += ", " + occ.String() + ","
	}

	return detail
}

// WhoDoing returns a persons full or familiar name with their occupation as an aside if known.
func WhoDoing(p *model.Person, dt *model.Date, enc render.PageMarkdownEncoder) string {
	detail := enc.EncodeModelLinkDedupe(p.PreferredFamiliarFullName, p.PreferredFamiliarName, p)

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
		return ""
	}

	if len(children) == 1 {
		return "only " + text.LowerFirst(p.Gender.RelationToParentNoun())
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

func PersonParentage(p *model.Person, enc render.InlineMarkdownEncoder) string {
	rel := PositionInFamily(p)
	if rel == "" {
		rel = text.LowerFirst(p.Gender.RelationToParentNoun())
	}
	intro := "the " + rel + " of "

	if p.Father.IsUnknown() {
		if p.Mother.IsUnknown() {
			return intro + "unknown parents"
		} else {
			return intro + enc.EncodeModelLink(p.Mother.PreferredFullName, p.Mother)
		}
	} else {
		if p.Mother.IsUnknown() {
			return intro + enc.EncodeModelLink(p.Father.PreferredFullName, p.Father)
		} else {
			return intro + enc.EncodeModelLink(p.Father.PreferredFullName, p.Father) + " and " + enc.EncodeModelLink(p.Mother.PreferredFullName, p.Mother)
		}
	}
}

func PersonSummary(p *model.Person, enc render.PageMarkdownEncoder, name string, includeBirth bool, includeParentage bool, activeTense bool) string {
	if name != "" {
		if p.Redacted {
			return enc.EncodeItalic(name)
		}

		if enc.EncodeModelLink("", p) == "" {
			name = enc.EncodeItalic(name)
		}

		name = enc.EncodeModelLink(name, p)

		if p.NickName != "" {
			name = text.JoinSentenceParts(name, fmt.Sprintf("(known as %s)", p.NickName))
		}
	}

	var para text.Para
	if includeBirth {
		birth := PersonBirthSummary(p, enc, name, true, true, includeParentage, activeTense)
		if birth != "" {
			para.NewSentence(birth)
			if activeTense {
				name = ""
			} else {
				name = p.Gender.SubjectPronoun()
			}
		}
	}

	marrs := PersonMarriageSummary(p, enc, name, false, activeTense)
	if marrs != "" {
		para.NewSentence(marrs)
		if activeTense {
			name = ""
		} else {
			name = p.Gender.SubjectPronoun()
		}
	}

	death := PersonDeathSummary(p, enc, name, false, activeTense)
	if death != "" {
		para.NewSentence(death)
	}

	// TODO: life events
	// marriages
	// emigration
	// imprisonment

	finalDetail := ""
	if p.Unmarried {
		finalDetail = "never married"
	}

	if finalDetail != "" {
		para.NewSentence(p.Gender.SubjectPronoun(), finalDetail)
	}

	return para.Text()
}

func PersonBirthSummary(p *model.Person, enc render.PageMarkdownEncoder, name string, allowInferred bool, includeBirthDate bool, includeParentage bool, activeTense bool) string {
	var birth *model.BirthEvent
	var bev model.IndividualTimelineEvent

	if p.BestBirthlikeEvent == nil {
		return ""
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

	tense := func(st string) string {
		if activeTense {
			return text.StripWasIs(st)
		}
		return st
	}

	if bev == nil {
		return ""
	}

	var para text.Para
	para.NewSentence(name)

	if includeBirthDate {
		if birth != nil {
			if _, ok := bev.(*model.BaptismEvent); ok {
				if yrs, ok := birth.GetDate().WholeYearsUntil(bev.GetDate()); ok && yrs > 1 {
					para.Continue("born", birth.GetDate().When(), "and")
				}
			}
		}
		para.Continue(enc.EncodeWithCitations(tense(EventWhatWhenWhere(bev, enc)), bev.GetCitations()))
	} else {
		para.Continue(enc.EncodeWithCitations(tense(EventWhatWhere(bev, enc)), bev.GetCitations()))
	}

	if includeParentage {
		para.AppendClause(PersonParentage(p, enc))
	}

	if len(p.Associations) > 0 {
		for _, as := range p.Associations {
			if as.Kind != model.AssociationKindTwin {
				continue
			}
			twinLink := enc.EncodeModelLink(as.Other.PreferredFamiliarName, as.Other)
			para.Continue(text.UpperFirst(p.Gender.SubjectPronoun()), "was the twin to", p.Gender.PossessivePronounSingular(), as.Other.Gender.RelationToSiblingNoun(), enc.EncodeWithCitations(twinLink, as.Citations))
		}
	}

	return para.Text()
}

func PersonDeathSummary(p *model.Person, enc render.PageMarkdownEncoder, name string, allowInferred bool, activeTense bool) string {
	var death *model.DeathEvent
	var bev model.IndividualTimelineEvent

	if p.BestDeathlikeEvent == nil {
		return ""
	}
	switch tev := p.BestDeathlikeEvent.(type) {
	case *model.DeathEvent:
		if allowInferred || !tev.IsInferred() {
			death = tev
			bev = tev
		}
		for _, ev := range p.Timeline {
			if !ev.DirectlyInvolves(p) {
				continue
			}
			if tev, ok := ev.(*model.BurialEvent); ok {
				if bev == nil || tev.GetDate().SortsBefore(bev.GetDate()) {
					bev = tev
				}
			} else if tev, ok := ev.(*model.CremationEvent); ok {
				if bev == nil || tev.GetDate().SortsBefore(bev.GetDate()) {
					bev = tev
				}
			}
		}
	case *model.BurialEvent:
		bev = tev
	}

	if bev == nil {
		return ""
	}

	tense := func(st string) string {
		if activeTense {
			return text.StripWasIs(st)
		}
		return st
	}

	var para text.Para
	para.NewSentence(name)
	deathWhat := model.PassiveWhat(bev)
	if death != nil {
		deathWhat = DeathWhat(death, p.ModeOfDeath)
	}
	para.Continue(enc.EncodeWithCitations(tense(WhatWhenWhere(deathWhat, bev.GetDate(), bev.GetPlace(), enc)), bev.GetCitations()))

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

	return para.Text()
}

func PersonMarriageSummary(p *model.Person, enc render.PageMarkdownEncoder, name string, allowInferred bool, activeTense bool) string {
	tense := func(st string) string {
		if activeTense {
			return text.StripWasIs(st)
		}
		return st
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
		return ""
	} else if len(fams) == 1 {
		// more detail
		f := fams[0]
		other := f.OtherParent(p)
		what := f.BestStartEvent.What() + " " + enc.EncodeModelLink(other.PreferredFamiliarFullName, other)
		marrs = append(marrs, enc.EncodeWithCitations(tense(WhatWhenWhere(what, f.BestStartEvent.GetDate(), nil, enc)), f.BestStartEvent.GetCitations()))
	} else {
		var prev model.TimelineEvent
		for _, f := range fams {
			other := f.OtherParent(p)

			y, _ := f.BestStartEvent.GetDate().AsYear()

			if prev != nil {
				what := enc.EncodeModelLink(other.PreferredFamiliarFullName, other)
				marrs = append(marrs, enc.EncodeWithCitations(tense(WhatWhenWhere(what, y, nil, enc)), f.BestStartEvent.GetCitations()))
			} else {
				what := f.BestStartEvent.What() + " " + enc.EncodeModelLink(other.PreferredFamiliarFullName, other)
				marrs = append(marrs, enc.EncodeWithCitations(tense(WhatWhenWhere(what, y, nil, enc)), f.BestStartEvent.GetCitations()))
			}

			prev = f.BestStartEvent
		}
	}

	var para text.Para
	para.NewSentence(name)
	para.Continue(text.JoinList(marrs))
	return para.Text()
}

func GenerateOlb(p *model.Person) error {
	if p.Olb != "" {
		return nil
	}

	const (
		Mundane     = 1
		Interesting = 2
		Unusual     = 3
		Unique      = 4
	)

	log := false
	logger := logging.With("id", p.ID, "name", p.PreferredFullName)

	type BioFacts struct {
		BirthYear             int
		BirthYearDesc         string
		BirthPlace            string
		CountryOfBirth        *place.PlaceName
		DeathYear             int
		DeathYearDesc         string
		DeathPlace            string
		DeathType             string
		CountryOfDeath        *place.PlaceName
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

			whenYear, ok := p.BestBirthlikeEvent.GetDate().WhenYear()
			if ok {
				bf.BirthYearDesc = "born " + whenYear
			}

			if p.BestBirthlikeEvent.IsInferred() {
				bf.BirthYearDesc = "likely " + bf.BirthYearDesc
			}

		}
		if !p.BestBirthlikeEvent.GetPlace().IsUnknown() {
			pl := p.BestBirthlikeEvent.GetPlace()
			bf.BirthPlace = pl.PreferredName

			for pl.Parent != nil {
				pl = pl.Parent
			}

			if !pl.CountryName.IsUnknown() {
				bf.CountryOfBirth = pl.CountryName
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
			bf.DeathPlace = pl.PreferredName

			for pl.Parent != nil {
				pl = pl.Parent
			}

			if !pl.CountryName.IsUnknown() {
				bf.CountryOfDeath = pl.CountryName
			}

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

		bf.NumberOfSiblings = len(p.Mother.Children)
		if bf.NumberOfSiblings > 0 && bf.BirthYear > 0 {
			bf.PositionInFamily = 1
			for _, ch := range p.Mother.Children {
				if ch.BestBirthlikeEvent == nil || ch.BestBirthlikeEvent.GetDate().IsUnknown() {
					bf.PositionInFamily = -1
					break
				}
				if ch.SameAs(p) {
					continue
				}
				if ch.BestBirthlikeEvent.GetDate().SortsBefore(p.BestBirthlikeEvent.GetDate()) {
					bf.PositionInFamily++
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

	type Clause struct {
		Interestingness int
		Text            string
	}

	var clauses []Clause

	if p.PrimaryOccupation != "" {
		clauses = append(clauses, Clause{Text: p.PrimaryOccupation, Interestingness: Unique})
	}

	hasIllegitimateClause := false
	if p.Illegitimate && p.Father.IsUnknown() {
		clause := "illegitimate"
		if !p.Mother.IsUnknown() {
			clause += " " + p.Gender.RelationToParentNoun() + " of " + p.Mother.PreferredFamiliarFullName
		}

		hasIllegitimateClause = true
		clauses = append(clauses, Clause{Text: clause, Interestingness: Interesting})
	}

	// Intro statement
	if p.NickName != "" {
		clauses = append(clauses, Clause{Text: "known as " + p.NickName, Interestingness: Interesting})
	}

	// Statement about birth
	// TODO: ideally use primary occupation if it were clean enough
	nonNotableCountries := map[string]bool{
		"England":        true,
		"United Kingdom": true,
	}

	if p.BornInWorkhouse {
		if p.DiedInWorkhouse {
			clauses = append(clauses, Clause{Text: "born and died in workhouse", Interestingness: Interesting})
		} else {
			clauses = append(clauses, Clause{Text: "born in workhouse", Interestingness: Interesting})
		}
	} else if bf.CountryOfBirth != nil && !nonNotableCountries[bf.CountryOfBirth.Name] {
		if bf.BirthYear%3 == 1 {
			clauses = append(clauses, Clause{Text: bf.CountryOfBirth.Adjective + "-born", Interestingness: Mundane})
		} else {
			clauses = append(clauses, Clause{Text: "born in " + bf.CountryOfBirth.Name, Interestingness: Mundane})
		}
	} else if bf.BirthYearDesc != "" {
		clauses = append(clauses, Clause{Text: bf.BirthYearDesc, Interestingness: 0})
	}

	if !hasIllegitimateClause {

		if bf.NumberOfSiblings > 1 {
			clause := ""
			if bf.PositionInFamily == 1 {
				clause = "eldest"
			} else if bf.PositionInFamily == bf.NumberOfSiblings {
				clause = "youngest"
			} else {
				clause = "one"
			}
			clause += " of " + text.CardinalNoun(bf.NumberOfSiblings)

			// add "children" if we aren't repeating the word later
			if !p.Childless && bf.NumberOfChildren < 2 {
				clause += " children"
			}
			clauses = append(clauses, Clause{Text: clause, Interestingness: 0})
		}

		if p.Mother.IsUnknown() && !p.Father.IsUnknown() {
			clauses = append(clauses, Clause{Text: "mother unknown", Interestingness: Mundane})
		} else if p.Father.IsUnknown() && !p.Mother.IsUnknown() {
			clauses = append(clauses, Clause{Text: "father unknown", Interestingness: Mundane})
		}
	}

	if p.Twin {
		clauses = append(clauses, Clause{Text: "twin", Interestingness: Interesting})
	}

	if p.DiedInChildbirth {
		clauses = append(clauses, Clause{Text: "died in childbirth", Interestingness: Unusual})
	}

	if p.PhysicalImpairment {
		clauses = append(clauses, Clause{Text: "physically impaired", Interestingness: Unusual})
	}

	if p.MentalImpairment {
		clauses = append(clauses, Clause{Text: "mentally impaired", Interestingness: Unusual})
	}

	if p.Deaf {
		clauses = append(clauses, Clause{Text: "deaf", Interestingness: Unusual})
	}

	if p.Blind {
		clauses = append(clauses, Clause{Text: "blind", Interestingness: Unusual})
	}

	parentDeathDesc := func(age int) string {
		if age == 0 {
			return "as a baby"
		} else if age < 5 {
			return "while still a child"
		} else {
			return "at " + strconv.Itoa(age)
		}
	}

	if bf.OrphanedAtAge > -1 && bf.OrphanedAtAge < 18 {
		clauses = append(clauses, Clause{Text: "orphaned " + parentDeathDesc(bf.OrphanedAtAge), Interestingness: Unusual})
	} else if bf.AgeAtDeathOfMother > -1 && bf.AgeAtDeathOfMother < 18 {
		clauses = append(clauses, Clause{Text: "mother died " + parentDeathDesc(bf.AgeAtDeathOfMother), Interestingness: Interesting})
	} else if bf.AgeAtDeathOfFather > -1 && bf.AgeAtDeathOfFather < 18 {
		clauses = append(clauses, Clause{Text: "father died " + parentDeathDesc(bf.AgeAtDeathOfFather), Interestingness: Interesting})
	}

	// Statement about families and children
	legitimateChildren := bf.NumberOfChildren
	if bf.IllegitimateChildren != -1 {
		legitimateChildren -= bf.IllegitimateChildren
	}

	if p.Childless && bf.AgeAtDeath > 18 {
		clauses = append(clauses, Clause{Text: "had no children", Interestingness: Mundane})
	} else if p.Gender.IsFemale() || bf.NumberOfChildren == 0 {
		if bf.IllegitimateChildren == 1 {
			clauses = append(clauses, Clause{Text: "had one child with an unknown father", Interestingness: Mundane})
		} else if bf.IllegitimateChildren > 1 {
			clauses = append(clauses, Clause{Text: "had " + text.SmallCardinalNoun(bf.IllegitimateChildren) + " children with unknown fathers", Interestingness: Mundane})
		}

		if p.Unmarried && bf.AgeAtDeath > 18 {
			clauses = append(clauses, Clause{Text: "never married", Interestingness: Interesting})
		} else if bf.NumberOfMarriages > 0 {
			if bf.AgeAtFirstMarriage > 0 && bf.AgeAtFirstMarriage < 18 {
				if bf.NumberOfMarriages == 1 && len(bf.Spouses) > 0 {
					clauses = append(clauses, Clause{Text: "married " + bf.Spouses[0].PreferredFamiliarFullName + " at " + strconv.Itoa(bf.AgeAtFirstMarriage), Interestingness: Interesting})
				} else if bf.NumberOfMarriages == 2 {
					clauses = append(clauses, Clause{Text: "married at " + strconv.Itoa(bf.AgeAtFirstMarriage) + " then later remarried", Interestingness: Interesting})
				} else {
					clauses = append(clauses, Clause{Text: "married at " + strconv.Itoa(bf.AgeAtFirstMarriage) + " then " + text.SmallCardinalNoun(bf.NumberOfMarriages-1) + " more times", Interestingness: Interesting})
				}
			} else {
				if bf.NumberOfMarriages == 1 && len(bf.Spouses) > 0 {
					clauses = append(clauses, Clause{Text: "married " + bf.Spouses[0].PreferredFamiliarFullName, Interestingness: Mundane})
				} else {
					clauses = append(clauses, Clause{Text: "married " + text.MultiplicativeAdverb(bf.NumberOfMarriages), Interestingness: Interesting})
				}
			}
		}

		if legitimateChildren == 1 {
			clauses = append(clauses, Clause{Text: "had one child", Interestingness: Mundane})
		} else if legitimateChildren > 1 {
			clauses = append(clauses, Clause{Text: fmt.Sprintf("had %s children", text.SmallCardinalNoun(legitimateChildren)), Interestingness: Mundane})
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

		clauses = append(clauses, Clause{Text: clause, Interestingness: Interesting})

		if bf.IllegitimateChildren > 0 {
			if bf.IllegitimateChildren == bf.NumberOfChildren {
				if bf.IllegitimateChildren == 2 {
					clauses = append(clauses, Clause{Text: "both with unknown mothers", Interestingness: Mundane})
				} else if bf.IllegitimateChildren > 2 {
					clauses = append(clauses, Clause{Text: "all with unknown mothers", Interestingness: Interesting})
				}
			} else {
				clauses = append(clauses, Clause{Text: text.SmallCardinalNoun(bf.IllegitimateChildren) + " with unknown mothers", Interestingness: Mundane})
			}
		}
	}

	if bf.NumberOfMarriages == 1 && bf.AgeAtFirstSpouseDeath > 0 && bf.AgeAtFirstSpouseDeath < 40 {
		if p.Gender.IsFemale() {
			clauses = append(clauses, Clause{Text: "widowed at " + strconv.Itoa(bf.AgeAtFirstSpouseDeath), Interestingness: Interesting})
		} else {
			clauses = append(clauses, Clause{Text: "widower at " + strconv.Itoa(bf.AgeAtFirstSpouseDeath), Interestingness: Interesting})
		}
	}

	if bf.NumberOfDivorces > 0 {
		if bf.NumberOfDivorces < bf.NumberOfMarriages {
			clauses = append(clauses, Clause{Text: "divorced " + text.MultiplicativeAdverb(bf.NumberOfDivorces), Interestingness: Mundane})
		} else if bf.NumberOfDivorces == bf.NumberOfMarriages && bf.NumberOfDivorces == 1 {
			clauses = append(clauses, Clause{Text: "later divorced", Interestingness: Mundane})
		}
	}

	if bf.NumberOfAnnulments > 0 {
		log = true
		if bf.NumberOfAnnulments < bf.NumberOfMarriages {
			clauses = append(clauses, Clause{Text: "anulled " + text.MultiplicativeAdverb(bf.NumberOfDivorces), Interestingness: Mundane})
		} else if bf.NumberOfAnnulments == bf.NumberOfMarriages && bf.NumberOfAnnulments == 1 {
			clauses = append(clauses, Clause{Text: "later anulled", Interestingness: Interesting})
		}
	}

	if bf.TravelEvents > 4 {
		clauses = append(clauses, Clause{Text: "travelled widely", Interestingness: Interesting})
	} else if !bf.CountryOfDeath.IsUnknown() && !bf.CountryOfBirth.IsUnknown() && !bf.CountryOfDeath.SameAs(bf.CountryOfBirth) {
		clauses = append(clauses, Clause{Text: "travelled to " + bf.CountryOfDeath.Name, Interestingness: Interesting})
	}

	// TODO: suicide
	// TODO: imprisoned
	// TODO: deported

	if p.Pauper {
		clauses = append(clauses, Clause{Text: "pauper", Interestingness: Mundane})
	}

	// Statement about death
	if bf.AgeAtDeath == 0 {
		clauses = append(clauses, Clause{Text: bf.DeathType + " as an infant", Interestingness: Mundane})
	} else if bf.AgeAtDeath > 0 && bf.AgeAtDeath < 10 {
		clauses = append(clauses, Clause{Text: bf.DeathType + " as a child", Interestingness: Mundane})
	} else if bf.AgeAtDeath >= 10 && bf.AgeAtDeath < 30 {
		clauses = append(clauses, Clause{Text: fmt.Sprintf("%s before %s %s", bf.DeathType, p.Gender.SubjectPronounWithLink(), strconv.Itoa(bf.AgeAtDeath+1)), Interestingness: Interesting})
	} else if bf.AgeAtDeath > 90 && bf.Suicide {
		clauses = append(clauses, Clause{Text: fmt.Sprintf("lived to %s", strconv.Itoa(bf.AgeAtDeath)), Interestingness: Interesting})
	} else if p.DiedInWorkhouse && !p.BornInWorkhouse {
		clause := bf.DeathType + " in poverty"
		if bf.AgeAtDeath > 0 {
			clause += " at the age of " + strconv.Itoa(bf.AgeAtDeath)
		}
		clauses = append(clauses, Clause{Text: clause, Interestingness: Interesting})

	} else if bf.DeathYear != 0 {
		clause := bf.DeathType + " " + bf.DeathYearDesc
		if bf.AgeAtDeath > 0 {
			clause += " at the age of " + strconv.Itoa(bf.AgeAtDeath)
		}
		clauses = append(clauses, Clause{Text: clause, Interestingness: Mundane})
	}

	switch p.ModeOfDeath {
	case model.ModeOfDeathLostAtSea:
		clauses = append(clauses, Clause{Text: "lost at sea", Interestingness: Unusual})
	case model.ModeOfDeathKilledInAction:
		clauses = append(clauses, Clause{Text: "killed in action", Interestingness: Unusual})
	case model.ModeOfDeathDrowned:
		clauses = append(clauses, Clause{Text: "drowned", Interestingness: Unusual})
	case model.ModeOfDeathExecuted:
		clauses = append(clauses, Clause{Text: "executed", Interestingness: Unique})
	}

	if len(clauses) == 0 {
		return nil
	}

	// Only keep 4 interesting clauses
	maxClauses := 4
	threshold := 1
	if len(clauses) > maxClauses {
		remove := len(clauses) - maxClauses
		for remove > 0 {
			for i := range clauses {
				if clauses[i].Text != "" && clauses[i].Interestingness < threshold {
					clauses[i].Text = ""
					remove--
					if remove == 0 {
						break
					}
				}
			}
			threshold++
		}
	}

	var texts []string
	for i := range clauses {
		if clauses[i].Text != "" {
			texts = append(texts, clauses[i].Text)
		}
	}
	p.Olb = strings.Join(texts, ", ")

	if p.Olb != "" {
		p.Olb = text.FinishSentence(text.UpperFirst(p.Olb))
	}
	if log {
		logger.Info("generated olb: " + p.Olb)
	} else {
		logger.Debug("generated olb: " + p.Olb)
	}
	return nil
}

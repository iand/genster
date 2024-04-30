package site

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/place"
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

func WhoWhatWhenWhere(ev model.TimelineEvent, enc ExtendedInlineEncoder) string {
	var title string
	switch tev := ev.(type) {
	case model.IndividualTimelineEvent:
		title = enc.EncodeModelLink(tev.GetPrincipal().PreferredFullName, tev.GetPrincipal())
	case model.PartyTimelineEvent:
		title = enc.EncodeModelLink(tev.GetParty1().PreferredFullName, tev.GetParty1()) + " and " + enc.EncodeModelLink(tev.GetParty2().PreferredFullName, tev.GetParty2())
	}

	title = text.JoinSentenceParts(title, EventWhatWhenWhere(ev, enc))

	// title = text.JoinSentenceParts(title, ev.What())
	// date := ev.GetDate()
	// if !date.IsUnknown() {
	// 	title = text.JoinSentenceParts(title, date.When())
	// }

	// pl := ev.GetPlace()
	// if !pl.IsUnknown() {
	// 	title = text.JoinSentenceParts(title, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
	// }
	return title
}

func EventWhatWhenWhere(ev model.TimelineEvent, enc ExtendedInlineEncoder) string {
	return WhatWhenWhere(InferredWhat(ev.What(), ev), ev.GetDate(), ev.GetPlace(), enc)
}

func InferredWhat(what string, ev model.TimelineEvent) string {
	if ev.IsInferred() {
		return "is inferred to " + text.MaybeHaveBeenVerb(what)
	}

	if !ev.GetDate().IsUnknown() {
		qual := ev.GetDate().Derivation.Qualifier()
		if qual != "" {
			return "is " + qual + " to " + text.MaybeHaveBeenVerb(what)
		}
	}

	return text.MaybeWasVerb(what)
}

func WhatWhenWhere(what string, dt *model.Date, pl *model.Place, enc ExtendedInlineEncoder) string {
	return text.JoinSentenceParts(what, WhenWhere(dt, pl, enc))
}

func WhenWhere(dt *model.Date, pl *model.Place, enc ExtendedInlineEncoder) string {
	title := ""
	if !dt.IsUnknown() {
		title = text.JoinSentenceParts(title, dt.When())
	}

	if !pl.IsUnknown() {
		title = text.JoinSentenceParts(title, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return title
}

func AgeWhenWhere(ev model.IndividualTimelineEvent, enc ExtendedInlineEncoder) string {
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
		title = text.JoinSentenceParts(title, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	return title
}

func FollowingWhatWhenWhere(what string, dt *model.Date, pl *model.Place, preceding model.IndividualTimelineEvent, enc ExtendedMarkdownEncoder) string {
	detail := what

	if pl.SameAs(preceding.GetPlace()) {
		detail = text.JoinSentenceParts(detail, "there")
	}

	if !dt.IsUnknown() {
		intervalDesc := ""
		in := preceding.GetDate().IntervalUntil(dt)
		y, m, d, ok := in.YMD()
		if ok {
			if y == 0 {
				if m == 0 {
					if d == 0 {
						intervalDesc = "shortly after"
					} else if d == 1 {
						intervalDesc = "the next day"
					} else if d == 2 {
						intervalDesc = "two days later"
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

		detail = text.JoinSentenceParts(detail, dt.When())
	}

	if !pl.IsUnknown() && !preceding.GetPlace().SameAs(pl) {
		detail = text.JoinSentenceParts(detail, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}

	return detail
}

// func WhatWhenWhere(ev model.TimelineEvent, enc ExtendedInlineEncoder) string {
// 	title := ""
// 	date := ev.GetDate()
// 	if !date.IsUnknown() {
// 		qual := date.Derivation.Qualifier()
// 		if qual != "" {
// 			title = text.JoinSentenceParts(qual, "to have been")
// 		}
// 	}

// 	return text.JoinSentenceParts(title, ev.What(), WhenWhere(ev, enc))
// }

func PersonSummary(p *model.Person, enc ExtendedMarkdownEncoder) string {
	name := p.PreferredGivenName
	if enc.EncodeModelLink("", p) == "" {
		name = enc.EncodeItalic(name)
	}

	summary := enc.EncodeModelLink(name, p)

	// Intro statement
	if p.NickName != "" {
		summary = text.JoinSentenceParts(summary, fmt.Sprintf("(known as %s)", p.NickName))
	}

	var birth *model.BirthEvent
	var bap *model.BaptismEvent
	var death *model.DeathEvent
	var burial *model.BurialEvent // TODO: cremation

	if p.BestBirthlikeEvent != nil {
		switch tev := p.BestBirthlikeEvent.(type) {
		case *model.BirthEvent:
			birth = tev
			for _, ev := range p.Timeline {
				if bev, ok := ev.(*model.BaptismEvent); ok {
					if bap == nil || bev.GetDate().SortsBefore(bap.GetDate()) {
						bap = bev
					}
				}
			}
		case *model.BaptismEvent:
			bap = tev
		}
	}

	if p.BestDeathlikeEvent != nil {
		switch tev := p.BestDeathlikeEvent.(type) {
		case *model.DeathEvent:
			death = tev
			for _, ev := range p.Timeline {
				if bev, ok := ev.(*model.BurialEvent); ok {
					if burial == nil || bev.GetDate().SortsBefore(burial.GetDate()) {
						burial = bev
					}
				}
			}
		case *model.BurialEvent:
			burial = tev
		}
	}

	// TODO: twin
	// TODO: marriages
	// TODO: major life events
	// TODO: age at death
	// TODO: cause of death

	// birthAdditional contains extra information such as whether person was a twin
	birthAdditional := ""

	firstEvent := true
	var precedingEvent model.IndividualTimelineEvent
	if birth != nil {
		summary = text.JoinSentenceParts(summary, enc.EncodeWithCitations(EventWhatWhenWhere(birth, enc), birth.GetCitations()))
		precedingEvent = birth
		// Try to complete the sentence
		if birthAdditional != "" {
			summary = text.JoinSentenceParts(summary, birthAdditional)
			summary = text.FinishSentence(summary)
			precedingEvent = nil
		}
		firstEvent = false
	}

	if bap != nil {
		if precedingEvent != nil {
			summary = text.JoinSentenceParts(summary, "and", FollowingWhatWhenWhere(InferredWhat(bap.What(), bap), bap.GetDate(), bap.GetPlace(), precedingEvent, enc))
			summary = text.FinishSentence(summary)
			precedingEvent = nil
		} else {
			if !firstEvent {
				summary = text.JoinSentenceParts(summary, text.UpperFirst(p.Gender.SubjectPronoun()))
			}
			summary = text.JoinSentenceParts(summary, enc.EncodeWithCitations(EventWhatWhenWhere(bap, enc), bap.GetCitations()))
			precedingEvent = bap
		}

		firstEvent = false
	}

	if len(p.Associations) > 0 {
		for _, as := range p.Associations {
			if as.Kind != model.AssociationKindTwin {
				continue
			}
			twinLink := enc.EncodeModelLink(as.Other.PreferredFamiliarName, as.Other)

			if precedingEvent != nil {
				summary = text.FinishSentence(summary)
				precedingEvent = nil
			}

			summary = text.JoinSentenceParts(summary, text.UpperFirst(p.Gender.SubjectPronoun()), "was the twin to", p.Gender.PossessivePronounSingular(), as.Other.Gender.RelationToSiblingNoun(), enc.EncodeWithCitations(twinLink, as.Citations))
			summary = text.FinishSentence(summary)
			break
		}
	}

	deathAdditional := ""
	if death != nil {
		deathWhat := death.What()
		switch p.CauseOfDeath {
		case model.CauseOfDeathLostAtSea:
			deathWhat = "lost at sea"
		case model.CauseOfDeathKilledInAction:
			deathWhat = "killed in action"
		case model.CauseOfDeathDrowned:
			deathWhat = "drowned"
		case model.CauseOfDeathSuicide:
			deathWhat = "died by suicide"
		}
		deathWhat = InferredWhat(deathWhat, death)

		if age, ok := p.AgeInYearsAt(death.GetDate()); ok {
			if age <= 1 {
				deathAdditional = "while still a young child"
			} else {
				deathAdditional = fmt.Sprintf("at the age of %s", text.CardinalNoun(age))
			}
		}

		if precedingEvent != nil {
			summary = text.JoinSentenceParts(summary, "and", FollowingWhatWhenWhere(deathWhat, death.GetDate(), death.GetPlace(), precedingEvent, enc))
			summary = text.FinishSentence(summary)
			precedingEvent = nil
		} else {
			if !firstEvent {
				summary = text.JoinSentenceParts(summary, text.UpperFirst(p.Gender.SubjectPronoun()))
			}
			summary = text.JoinSentenceParts(summary, enc.EncodeWithCitations(WhatWhenWhere(deathWhat, death.GetDate(), death.GetPlace(), enc), death.GetCitations()))
			if deathAdditional != "" {
				summary = text.JoinSentenceParts(summary, deathAdditional)
				summary = text.FinishSentence(summary)
				precedingEvent = nil
			} else {
				precedingEvent = death
			}
		}

		firstEvent = false
	}

	if burial != nil {
		burialAdditional := ""
		if deathAdditional == "" {
			if age, ok := p.AgeInYearsAt(burial.GetDate()); ok {
				if age <= 1 {
					burialAdditional = "while still a young child"
				} else {
					burialAdditional = fmt.Sprintf("at the age of %s", text.CardinalNoun(age))
				}
			}
		}

		if precedingEvent != nil {
			summary = text.JoinSentenceParts(summary, "and", FollowingWhatWhenWhere(InferredWhat(burial.What(), burial), burial.GetDate(), burial.GetPlace(), precedingEvent, enc))
			summary = text.FinishSentence(summary)
			precedingEvent = nil
		} else {
			if !firstEvent {
				summary = text.JoinSentenceParts(summary, text.UpperFirst(p.Gender.SubjectPronoun()))
			}
			summary = text.JoinSentenceParts(summary, enc.EncodeWithCitations(EventWhatWhenWhere(burial, enc), burial.GetCitations()))
			if burialAdditional != "" {
				summary = text.JoinSentenceParts(summary, burialAdditional)
				summary = text.FinishSentence(summary)
				precedingEvent = nil
			} else {
				precedingEvent = burial
			}
		}
		firstEvent = false
	}

	// for _, f := range p.Families {
	// 	n.Statements = append(n.Statements, &FamilyStatement{
	// 		Principal: p,
	// 		Family:    f,
	// 	})
	// }

	summary = text.FinishSentence(summary)
	return summary
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
	logger := slog.With("id", p.ID, "name", p.PreferredFullName)

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
			if p.CauseOfDeath == model.CauseOfDeathSuicide {
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

	if p.CauseOfDeath == model.CauseOfDeathLostAtSea {
		clauses = append(clauses, Clause{Text: "lost at sea", Interestingness: Unusual})
	} else if p.CauseOfDeath == model.CauseOfDeathKilledInAction {
		clauses = append(clauses, Clause{Text: "killed in action", Interestingness: Unusual})
	} else if p.CauseOfDeath == model.CauseOfDeathDrowned {
		clauses = append(clauses, Clause{Text: "drowned", Interestingness: Unusual})
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

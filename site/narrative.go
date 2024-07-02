package site

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

type Narrative struct {
	Statements []Statement
}

type IntroGenerator struct {
	POV              *model.POV
	NameMinSeq       int                 // the minimum sequence that the person's name may be used in an intro
	AgeMinSeq        int                 // the minimum sequence that the person's age may be used in an intro
	LastIntroDate    *model.Date         //  the date that the last intro was requested
	PeopleIntroduced map[string][]string // a lookup of occupations for people who have been introduced
}

func (n *IntroGenerator) Default(seq int, dt *model.Date) string {
	part1 := n.RelativeTime(seq, dt, true)
	part2 := n.Pronoun(seq, dt)

	if part1 == "" {
		return part2
	}

	if part2 == "" {
		return part1
	}

	return part1 + ", " + part2
}

func (n *IntroGenerator) Pronoun(seq int, dt *model.Date) string {
	defer func() {
		n.LastIntroDate = dt
	}()
	if n.POV.Person == nil {
		return "they"
	}

	if seq >= n.NameMinSeq {
		n.NameMinSeq = seq + 3
		return n.POV.Person.PreferredFamiliarName
	}

	return n.POV.Person.Gender.SubjectPronoun()
}

func (n *IntroGenerator) RelativeTime(seq int, dt *model.Date, includeFullDate bool) string {
	defer func() {
		n.LastIntroDate = dt
	}()
	if n.POV.Person != nil && seq >= n.AgeMinSeq {
		if age, ok := n.POV.Person.AgeInYearsAt(dt); ok && age > 0 {
			n.AgeMinSeq = seq + 2
			if includeFullDate {
				return fmt.Sprintf("%s, at the age of %s", dt.When(), text.CardinalNoun(age))
			}
			return fmt.Sprintf("at the age of %s", text.CardinalNoun(age))
		}
	}

	if n.LastIntroDate != nil {
		sincePrev := n.LastIntroDate.IntervalUntil(dt)
		if years, ok := sincePrev.WholeYears(); ok {
			dateInYear, ok := dt.DateInYear(true)
			if ok {
				dateInYear = "on " + dateInYear
			}

			if years < 0 && dt.SortsBefore(n.LastIntroDate) {
				return ""
			} else if years == 0 {
				days, isPreciseInterval := sincePrev.ApproxDays()
				if isPreciseInterval && days < 5 {
					return ChooseFrom(seq,
						dateInYear,
						dateInYear+", just a few days later",
						"very shortly after "+dateInYear,
						"just a few days later "+dateInYear,
					)
				} else if isPreciseInterval && days < 20 {
					return ChooseFrom(seq,
						"shortly after"+dateInYear,
						"several days later"+dateInYear,
					)
				} else if n.LastIntroDate.SameYear(dt) {
					return ChooseFrom(seq,
						"later that year"+dateInYear,
						"the same year"+dateInYear,
						"later that same year"+dateInYear,
						"that same year"+dateInYear,
					)
				} else {
					return ChooseFrom(seq,
						"shortly after"+dateInYear,
						"some time later"+dateInYear,
						"a short while later"+dateInYear,
					)
				}

			} else if years == 1 {
				return ChooseFrom(seq,
					"the following year, "+dateInYear,
					"the next year, "+dateInYear,
					"",
				)
				// } else if years < 5 {
				// 	return ChooseFrom(seq,
				// 		text.AppendClause("a few years later", dt.When()),
				// 		text.AppendClause("some years later", dt.When()),
				// 		"",
				// 	)
			} else {
				if includeFullDate {
					return ChooseFrom(seq,
						dt.When(),
						text.CardinalNoun(years)+" years later, "+dt.When(),
					)
				}
				return ChooseFrom(seq,
					"",
					text.CardinalNoun(years)+" years later",
				)
			}
		}

	}

	if includeFullDate {
		return dt.When()
	}
	return ""
}

func (n *IntroGenerator) IntroducePerson(seq int, p *model.Person, dt *model.Date, suppressSameSurname bool, enc render.PageMarkdownEncoder) string {
	if n.PeopleIntroduced == nil {
		n.PeopleIntroduced = make(map[string][]string)
	}

	if p.IsUnknown() {
		return "an unknown person"
	}

	occ := p.OccupationAt(dt)
	occDetail := ""
	if !occ.IsUnknown() {
		occDetail = occ.String()
	}

	occs := n.PeopleIntroduced[p.ID]
	if len(occs) == 0 {
		n.PeopleIntroduced[p.ID] = append(n.PeopleIntroduced[p.ID], occDetail)
		detail := ""
		if suppressSameSurname && p.PreferredFamilyName == n.POV.Person.PreferredFamilyName {
			detail = enc.EncodeModelLinkDedupe(p.PreferredGivenName, p.PreferredGivenName, p)
		} else {
			detail = enc.EncodeModelLinkDedupe(p.PreferredUniqueName, p.PreferredFullName, p)
		}
		if occDetail != "" {
			detail += ", " + occDetail + ","
		}
		return detail
	}

	name := p.PreferredFullName
	if suppressSameSurname && p.PreferredFamilyName == n.POV.Person.PreferredFamilyName {
		name = p.PreferredGivenName
	}

	hadPreviousOccupation := false
	for _, od := range n.PeopleIntroduced[p.ID] {
		if od != "" {
			hadPreviousOccupation = true
		}
		if occDetail == od {
			return enc.EncodeModelLinkDedupe(name, p.PreferredGivenName, p)
		}
	}
	n.PeopleIntroduced[p.ID] = append(n.PeopleIntroduced[p.ID], occDetail)
	detail := enc.EncodeModelLinkDedupe(name, p.PreferredGivenName, p)
	if occDetail != "" {
		if hadPreviousOccupation {
			detail += ", now " + occDetail + ","
		} else {
			detail += ", " + occDetail + ","
		}
	}
	return detail
}

const (
	NarrativeSequenceIntro     = 0
	NarrativeSequenceEarlyLife = 1
	NarrativeSequenceLifeStory = 2
	NarrativeSequenceDeath     = 3
	NarrativeSequencePostDeath = 4
)

func (n *Narrative) Render(pov *model.POV, b render.MarkupBuilder) {
	sort.Slice(n.Statements, func(i, j int) bool {
		if n.Statements[i].NarrativeSequence() == n.Statements[j].NarrativeSequence() {
			if n.Statements[i].Start().SameDate(n.Statements[j].Start()) {
				return n.Statements[i].Priority() > n.Statements[j].Priority()
			}
			return n.Statements[i].Start().SortsBefore(n.Statements[j].Start())
		}
		return n.Statements[i].NarrativeSequence() < n.Statements[j].NarrativeSequence()
	})

	currentNarrativeSequence := NarrativeSequenceIntro
	sequenceInNarrative := 0
	nintro := IntroGenerator{
		POV: pov,
	}
	for _, s := range n.Statements {
		if currentNarrativeSequence != s.NarrativeSequence() {
			currentNarrativeSequence = s.NarrativeSequence()
			switch currentNarrativeSequence {

			case NarrativeSequenceEarlyLife:
				// run on from intro, no separate heading
			case NarrativeSequenceLifeStory:
				// b.Heading3("Life Story")
				// reset sequence at start of new section
				sequenceInNarrative = 0
			case NarrativeSequenceDeath:
				// b.Heading3("Death")
				// reset sequence at start of new section
				sequenceInNarrative = 0
			}
		}

		s.RenderDetail(sequenceInNarrative, &nintro, b, nil)
		b.EmptyPara()

		sequenceInNarrative++

	}
}

type GrammarHints struct {
	DateInferred bool
}

type Statement interface {
	RenderDetail(int, *IntroGenerator, render.MarkupBuilder, *GrammarHints)
	Start() *model.Date
	End() *model.Date
	NarrativeSequence() int
	Priority() int // priority within a narrative against another statement with same date, higher will be rendered first
}

type IntroStatement struct {
	Principal        *model.Person
	Baptisms         []*model.BaptismEvent
	SuppressRelation bool
}

var _ Statement = (*IntroStatement)(nil)

func (s *IntroStatement) RenderDetail(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	var birth string
	// Prose birth
	if s.Principal.BestBirthlikeEvent != nil {
		// birth = text.LowerFirst(EventTitle(s.Principal.BestBirthlikeEvent, enc, &model.POV{Person: s.Principal}))
		birth = enc.EncodeWithCitations(text.LowerFirst(EventWhatWhenWhere(s.Principal.BestBirthlikeEvent, enc)), s.Principal.BestBirthlikeEvent.GetCitations())
	}
	// TODO: position in family

	// Prose parentage
	parentUnknownDetail := ""
	parentDetail := ""
	parentageDetailPrefix := "the " + PositionInFamily(s.Principal) + " of "
	if s.Principal.Father.IsUnknown() {
		if s.Principal.Mother.IsUnknown() {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " parents are not known"
		} else {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " father is not known"
			parentDetail = parentageDetailPrefix + intro.IntroducePerson(seq, s.Principal.Mother, s.Start(), false, enc)
			// parentDetail = parentageDetailPrefix + enc.EncodeModelLinkDedupe(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother.PreferredFamiliarName, s.Principal.Mother)
		}
	} else {
		if s.Principal.Mother.IsUnknown() {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " mother is not known"
			parentDetail = parentageDetailPrefix + intro.IntroducePerson(seq, s.Principal.Father, s.Start(), false, enc)
			// parentDetail = parentageDetailPrefix + enc.EncodeModelLinkDedupe(s.Principal.Father.PreferredUniqueName, s.Principal.Father.PreferredFamiliarName, s.Principal.Father)
		} else {
			parentDetail = parentageDetailPrefix + intro.IntroducePerson(seq, s.Principal.Father, s.Start(), false, enc) + " and " + intro.IntroducePerson(seq, s.Principal.Mother, s.Start(), false, enc)
			// parentDetail = parentageDetailPrefix + enc.EncodeModelLinkDedupe(s.Principal.Father.PreferredUniqueName, s.Principal.Father.PreferredFamiliarName, s.Principal.Father) + " and " + enc.EncodeModelLinkDedupe(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother.PreferredFamiliarName, s.Principal.Mother)
		}
	}

	// ---------------------------------------
	// Build detail
	// ---------------------------------------
	detail := ""

	if s.Principal.NickName != "" {
		detail = text.JoinSentenceParts(detail, "(known as ", s.Principal.NickName, ")")
	}

	// detail += " "
	if birth != "" {
		detail = text.JoinSentenceParts(detail, birth)
		if parentDetail != "" {
			detail = text.AppendClause(detail, parentDetail)
		}
	} else {
		if parentDetail != "" {
			detail = text.JoinSentenceParts(detail, parentDetail)
		}
	}
	if detail == "" {
		detail = text.FormatSentence(text.JoinSentenceParts("nothing is known about the early life of", s.Principal.PreferredGivenName))
	} else {
		detail = text.FormatSentence(text.JoinSentenceParts(s.Principal.PreferredGivenName, detail))

		if parentUnknownDetail != "" {
			detail = text.JoinSentences(detail, parentUnknownDetail)
			detail = text.FinishSentence(detail)
		}
	}

	// Twin association?
	twinClause := false
	if len(s.Principal.Associations) > 0 {
		for _, as := range s.Principal.Associations {
			if as.Kind != model.AssociationKindTwin {
				continue
			}
			twinLink := enc.EncodeModelLink(as.Other.PreferredFamiliarName, as.Other)

			detail = text.JoinSentenceParts(detail, text.UpperFirst(s.Principal.Gender.SubjectPronoun()), "was the twin to", s.Principal.Gender.PossessivePronounSingular(), as.Other.Gender.RelationToSiblingNoun(), enc.EncodeWithCitations(twinLink, as.Citations))
			twinClause = true
			break
		}
	}

	// Insert baptism here if there is only one, otherwise leave for a new para
	if len(s.Baptisms) == 1 && s.Baptisms[0] != s.Principal.BestBirthlikeEvent {
		bapDetail := AgeWhenWhere(s.Baptisms[0], enc)
		if bapDetail != "" {

			if twinClause {
				detail = text.JoinSentenceParts(detail, "and")
			} else {
				detail = text.JoinSentenceParts(text.FinishSentence(detail), text.UpperFirst(s.Principal.Gender.SubjectPronoun()))
			}

			detail = text.JoinSentenceParts(detail, "was baptised", enc.EncodeWithCitations(bapDetail, s.Baptisms[0].GetCitations()))
			detail = text.FinishSentence(detail)
		}

	}

	// ---------------------------------------
	// Prose relation to key person
	// ---------------------------------------
	if !s.SuppressRelation {
		if s.Principal.RelationToKeyPerson != nil && !s.Principal.RelationToKeyPerson.IsSelf() {
			detail += " " + text.UpperFirst(s.Principal.Gender.SubjectPronoun()) + " is " + enc.EncodeModelLink(text.MaybePossessiveSuffix(s.Principal.RelationToKeyPerson.From.PreferredFamiliarName), s.Principal.RelationToKeyPerson.From) + " " + s.Principal.RelationToKeyPerson.Name()
		}
	}

	detail = text.FinishSentence(detail)
	enc.Para(render.Markdown(detail))

	if len(s.Baptisms) > 1 {

		var bapDetail string
		for i, bev := range s.Baptisms {
			logging.Debug("adding baptism event to narrative intro statement", "id", s.Principal.ID, "bev", bev, "BestBirthlikeEvent", s.Principal.BestBirthlikeEvent)
			if s.Baptisms[i] == s.Principal.BestBirthlikeEvent {
				continue
			}
			evDetail := ""
			if i == 0 {
				evDetail += "was baptised"
			} else {
				evDetail += "and again"
			}
			aww := AgeWhenWhere(bev, enc)
			if aww != "" {
				bapDetail = text.JoinSentenceParts(bapDetail, evDetail, enc.EncodeWithCitations(bapDetail, s.Baptisms[0].GetCitations()))
			}
		}
		bapDetail = text.FinishSentence(text.JoinSentenceParts(intro.Pronoun(seq, s.Start()), bapDetail))
		enc.Para(render.Markdown(bapDetail))
	}
}

func (s *IntroStatement) Start() *model.Date {
	return s.Principal.BestBirthDate()
}

func (s *IntroStatement) End() *model.Date {
	return s.Principal.BestBirthDate()
}

func (s *IntroStatement) NarrativeSequence() int {
	return NarrativeSequenceIntro
}

func (s *IntroStatement) Priority() int {
	return 10
}

type FamilyStatement struct {
	Principal *model.Person
	Family    *model.Family
}

var _ Statement = (*FamilyStatement)(nil)

func (s *FamilyStatement) RenderDetail(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	// TODO: note for example VFA3VQS22ZHBO George Henry Chambers (1903-1985) who
	// had a child with Dorothy Youngs in 1944 but didn't marry until 1985
	other := s.Family.OtherParent(s.Principal)

	// Special cases for single parents
	if s.Family.Bond == model.FamilyBondUnmarried || s.Family.Bond == model.FamilyBondLikelyUnmarried {
		if other.IsUnknown() {
			s.renderIllegitimate(seq, intro, enc, hints)
			return
		}
		s.renderUnmarried(seq, intro, enc, hints)
		return
	} else if other.IsUnknown() {
		s.renderUnknownPartner(seq, intro, enc, hints)
		return
	}

	// from here, both partners are known

	var detail text.Para
	detail.NewSentence(text.UpperFirst(intro.Default(seq, s.Start())))

	action := ""
	switch s.Family.Bond {
	case model.FamilyBondMarried:
		action += "married"
	case model.FamilyBondLikelyMarried:
		action += ChooseFrom(seq, "likely married", "probably married")
	default:
		action += "met"
	}

	otherName := ""
	if other.IsUnknown() {
		otherName = "an unknown " + s.Principal.Gender.Opposite().Noun()
	} else {
		otherName = intro.IntroducePerson(seq, other, s.Start(), false, enc)
	}

	singleParent := false
	if s.Family.Bond == model.FamilyBondUnmarried || s.Family.Bond == model.FamilyBondLikelyUnmarried ||
		(s.Family.Bond == model.FamilyBondUnknown && other.IsUnknown()) {
		singleParent = true
	}
	if singleParent && len(s.Family.Children) == 0 {
		// nothing to say
		return
	}

	if !singleParent {
		startDate := s.Family.BestStartDate
		var event string
		if !startDate.IsUnknown() {
			detail.ReplaceSentence(intro.Pronoun(seq, s.Start()))
			event += " " + action
			event += " " + otherName
			event += " " + startDate.When()
			if age, ok := s.Principal.AgeInYearsAt(startDate); ok && age < 18 || age > 45 {
				event += " " + AgeQualifier(age)
			}
		} else {
			event += " " + action
			event += " " + otherName
		}
		if s.Family.BestStartEvent != nil && !s.Family.BestStartEvent.GetPlace().IsUnknown() {
			event = WhatWhere(event, s.Family.BestStartEvent.GetPlace(), enc)
		}
		if s.Family.BestStartEvent != nil {
			detail.Continue(enc.EncodeWithCitations(event, s.Family.BestStartEvent.GetCitations()))
		} else {
			detail.Continue(event)
		}
	}

	if len(s.Family.Children) == 0 {
		// single parents already dealt with
		if s.Principal.Childless {
			detail.Continue("they had no children")
		}
	} else {

		childCardinal := s.childCardinal(s.Family.Children)
		if singleParent {
			switch len(s.Family.Children) {
			case 1:
				detail.Continue(ChooseFrom(seq,
					" had "+childCardinal+" with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
					" had "+childCardinal+" by an unknown "+s.Principal.Gender.Opposite().Noun(),
					" had "+childCardinal+"",
				))
			default:
				detail.Continue("had", childCardinal)
				detail.Continue(ChooseFrom(seq,
					"",
					", by an unknown "+s.Principal.Gender.Opposite().Noun(),
					", by an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
				))
			}
		} else {
			switch len(s.Family.Children) {
			case 1:
				detail.Continue(ChooseFrom(seq,
					" and had "+childCardinal+":",
					". They had just one child together:",
					". They had "+childCardinal+":",
				))
			case 2:
				detail.Continue(ChooseFrom(seq,
					" and had "+childCardinal+":",
					". They had "+childCardinal+": ",
				))
			default:
				detail.Continue(ChooseFrom(seq,
					". They had "+childCardinal+": ",
					" and went on to have "+childCardinal+" with "+s.Principal.Gender.Opposite().ObjectPronoun()+": ",
					". They went on to have "+childCardinal+": ",
				))
			}
		}
	}

	if len(s.Family.Children) == 1 {
		c := s.Family.Children[0]
		if !c.Redacted {
			detail.NewSentence(PersonSummary(c, enc, c.PreferredFamiliarName, true, false, false))
		}
		enc.Para(render.Markdown(detail.Text()))
	} else {
		enc.Para(render.Markdown(detail.Text()))
		childList := s.childList(s.Family.Children, enc)
		enc.UnorderedList(childList)
	}
}

func (s *FamilyStatement) Start() *model.Date {
	return s.Family.BestStartDate
}

func (s *FamilyStatement) End() *model.Date {
	return s.Family.BestEndDate
}

func (s *FamilyStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *FamilyStatement) Priority() int {
	return 5
}

func (s *FamilyStatement) childCardinal(clist []*model.Person) string {
	// TODO: note how many children survived if some died
	allSameGender := true
	if s.Family.Children[0].Redacted {
		allSameGender = false
	} else if s.Family.Children[0].Gender != model.GenderUnknown {
		for i := 1; i < len(s.Family.Children); i++ {
			if s.Family.Children[i].Redacted {
				allSameGender = false
				break
			}
			if s.Family.Children[i].Gender != s.Family.Children[0].Gender {
				allSameGender = false
				break
			}
		}
	}

	if allSameGender {
		if s.Family.Children[0].Gender == model.GenderMale {
			return text.CardinalWithUnit(len(s.Family.Children), "son", "sons")
		} else {
			return text.CardinalWithUnit(len(s.Family.Children), "daughter", "daughters")
		}
	}
	return text.CardinalWithUnit(len(s.Family.Children), "child", "children")
}

func (s *FamilyStatement) childList(clist []*model.Person, enc render.MarkupBuilder) []render.Markdown {
	sort.Slice(clist, func(i, j int) bool {
		var d1, d2 *model.Date
		if clist[i].BestBirthlikeEvent != nil {
			d1 = clist[i].BestBirthlikeEvent.GetDate()
		}
		if clist[j].BestBirthlikeEvent != nil {
			d2 = clist[j].BestBirthlikeEvent.GetDate()
		}

		return d1.SortsBefore(d2)
	})

	redactedCount := 0
	childList := make([]render.Markdown, 0, len(clist))
	for _, c := range clist {
		if c.Redacted {
			redactedCount++
			continue
		}
		childList = append(childList, render.Markdown(PersonSummary(c, enc, c.PreferredGivenName, true, false, true)))
	}
	if len(childList) == 0 {
		return childList
	}
	if redactedCount > 0 {
		childList = append(childList, render.Markdown(text.CardinalWithUnit(redactedCount, "other child", "other children")+" living or recently died"))
	}

	return childList
}

func (s *FamilyStatement) renderIllegitimate(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	// unmarried and the other parent is not known
	if len(s.Family.Children) == 0 {
		// no children so nothing to say
		return
	}

	isFirmBirthdate := func(ev model.IndividualTimelineEvent) bool {
		if _, isBirth := ev.(*model.BirthEvent); !isBirth {
			return false
		}

		return ev.GetDate().IsFirm()
	}

	oneChild := len(s.Family.Children) == 1
	isMother := s.Principal.Gender == model.GenderFemale
	childFirmBirthdate := isFirmBirthdate(s.Family.Children[0].BestBirthlikeEvent)

	var detail text.Para

	if oneChild && isMother {
		c := s.Family.Children[0]
		if childFirmBirthdate {
			// this form: "At the age of thirty-four, Annie gave birth to a"
			detail.AppendAsAside(intro.RelativeTime(seq, c.BestBirthlikeEvent.GetDate(), false))
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate()))
			detail.Continue("gave birth to a", c.Gender.RelationToParentNoun())
			detail.AppendAsAside(enc.EncodeModelLink(c.PreferredFullName, c))
			detail.Continue(enc.EncodeWithCitations(EventWhenWhere(c.BestBirthlikeEvent, enc), c.BestBirthlikeEvent.GetCitations()))

		} else {
			// this form: "At the age of thirty-four, Annie had a"
			detail.AppendAsAside(intro.RelativeTime(seq, c.BestBirthlikeEvent.GetDate(), false))
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate()))
			detail.Continue("had a", c.Gender.RelationToParentNoun())
			detail.AppendAsAside(enc.EncodeModelLink(c.PreferredFullName, c))
			detail.Continue("who")
			detail.Continue(enc.EncodeWithCitations(EventWhatWhenWhere(c.BestBirthlikeEvent, enc), c.BestBirthlikeEvent.GetCitations()))

		}

		// pad the sentence to be longer if needed
		if detail.CurrentSentenceLength() < 60 {
			detail.Continue(ChooseFrom(seq,
				"with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
				"by an unknown "+s.Principal.Gender.Opposite().Noun(),
			))
		}

		if !c.Redacted {
			detail.NewSentence(PersonSummary(c, enc, c.PreferredFamiliarName, false, false, false))
		}
	} else {
		panic("Not implemented: renderIllegitimate where person has more than one child or is the father")
	}
	enc.Para(render.Markdown(detail.Text()))
}

func (s *FamilyStatement) renderUnmarried(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	// unmarried but the other parent is known
	if len(s.Family.Children) == 0 {
		// no children so nothing to say
		return
	}

	isFirmBirthdate := func(ev model.IndividualTimelineEvent) bool {
		if _, isBirth := ev.(*model.BirthEvent); !isBirth {
			return false
		}

		return ev.GetDate().IsFirm()
	}

	oneChild := len(s.Family.Children) == 1
	isMother := s.Principal.Gender == model.GenderFemale
	childFirmBirthdate := isFirmBirthdate(s.Family.Children[0].BestBirthlikeEvent)
	useBirthDateInIntro := childFirmBirthdate

	other := s.Family.OtherParent(s.Principal)
	var otherName string
	if other.IsUnknown() {
		otherName = "an unknown " + s.Principal.Gender.Opposite().Noun()
	} else {
		otherName = intro.IntroducePerson(seq, other, s.Start(), false, enc)
	}

	var detail text.Para
	if oneChild {
		c := s.Family.Children[0]
		detail.AppendAsAside(intro.RelativeTime(seq, s.Family.Children[0].BestBirthlikeEvent.GetDate(), useBirthDateInIntro))
		detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate()))
		if isMother && childFirmBirthdate {
			detail.Continue("gave birth to a")
			detail.Continue(c.Gender.RelationToParentNoun())
			if !c.Redacted {
				detail.AppendAsAside(enc.EncodeModelLink(c.PreferredFamiliarName, c))
			}
			detail.AppendClause("the child of")
			detail.Continue(otherName)
		} else {
			detail.Continue(ChooseFrom(seq,
				"had a child",
				"had a "+c.Gender.RelationToParentNoun(),
			))
			if !c.Redacted {
				detail.AppendAsAside(enc.EncodeModelLink(c.PreferredFamiliarName, c))
			}
			detail.Continue("with", otherName)
		}
		detail.FinishSentence()
		if !c.Redacted {
			detail.NewSentence(PersonSummary(c, enc, c.PreferredFullName, !useBirthDateInIntro, false, false))
		}

	} else {
		panic("Not implemented: renderUnmarried where person has more than one child")
	}

	enc.Para(render.Markdown(detail.Text()))
}

func (s *FamilyStatement) renderUnknownPartner(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	// married or unknown relationship but the other parent is unknown
}

type FamilyEndStatement struct {
	Principal *model.Person
	Family    *model.Family
}

var _ Statement = (*FamilyEndStatement)(nil)

func (s *FamilyEndStatement) RenderDetail(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	endDate := s.Family.BestEndDate
	if endDate.IsUnknown() {
		return
	}
	if s.endedWithDeathOf(s.Principal) {
		return
	}

	other := s.Family.OtherParent(s.Principal)
	if other.IsUnknown() {
		return
	}

	var detail text.Para
	end := ""
	switch s.Family.EndReason {
	case model.FamilyEndReasonDivorce:
		detail.NewSentence(s.Principal.PreferredFamiliarName, "and", other.PreferredFamiliarName, "divorced", endDate.When())
	case model.FamilyEndReasonDeath:
		name := s.Principal.Gender.PossessivePronounSingular() + " " + other.Gender.RelationToSpouseNoun()
		if !other.IsUnknown() {
			name = other.PreferredFamiliarName + ", " + name + ", "
		}
		detail.NewSentence(PersonDeathSummary(other, enc, name, true, false))
		if (other.BestDeathlikeEvent != nil && !other.BestDeathlikeEvent.IsInferred()) && (s.Family.Bond == model.FamilyBondMarried || s.Family.Bond == model.FamilyBondLikelyMarried) {
			detail.NewSentence(s.Principal.PreferredFamiliarName, "was left a", s.Principal.Gender.WidowWidower())
		}
	case model.FamilyEndReasonUnknown:
		// TODO: format FamilyEndReasonUnknown
		end += "the marriage ended in " + endDate.When()
	}

	if end != "" {
		detail.Continue(end)
	}

	enc.Para(render.Markdown(detail.Text()))
}

func (s *FamilyEndStatement) Start() *model.Date {
	return s.Family.BestEndDate
}

func (s *FamilyEndStatement) End() *model.Date {
	return s.Family.BestEndDate
}

func (s *FamilyEndStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *FamilyEndStatement) endedWithDeathOf(p *model.Person) bool {
	return p.SameAs(s.Family.EndDeathPerson)
}

func (s *FamilyEndStatement) Priority() int {
	return 4
}

type DeathStatement struct {
	Principal *model.Person
}

var _ Statement = (*DeathStatement)(nil)

func (s *DeathStatement) RenderDetail(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	var detail string

	bev := s.Principal.BestDeathlikeEvent

	evDetail := DeathWhat(bev, s.Principal.ModeOfDeath)

	// switch bev.(type) {
	// case *model.DeathEvent:
	// 	if bev.IsInferred() {
	// 		evDetail = text.JoinSentenceParts(evDetail, "is inferred to have died")
	// 	} else {
	// 		evDetail = text.JoinSentenceParts(evDetail, "died")
	// 	}
	// case *model.BurialEvent:
	// 	if bev.IsInferred() {
	// 		evDetail = text.JoinSentenceParts(evDetail, "is inferred to have been buried")
	// 	} else {
	// 		evDetail = text.JoinSentenceParts(evDetail, "was buried")
	// 	}
	// case *model.CremationEvent:
	// 	if bev.IsInferred() {
	// 		evDetail = text.JoinSentenceParts(evDetail, "is inferred to have been cremated")
	// 	} else {
	// 		evDetail = text.JoinSentenceParts(evDetail, "was cremated")
	// 	}
	// default:
	// 	panic("unhandled deathlike event in DeathStatement")
	// }

	if !bev.GetDate().IsUnknown() {
		if age, ok := s.Principal.AgeInYearsAt(bev.GetDate()); ok {
			ageDetail := ""
			if age == 0 {
				if pi, ok := s.Principal.PreciseAgeAt(bev.GetDate()); ok {
					if pi.Y == 0 {
						if pi.M == 0 {
							if pi.D == 0 {
								ageDetail = "shortly"
							} else if pi.D < 7 {
								ageDetail = "less than a week"
							} else if pi.D < 10 {
								ageDetail = "just a week"
							} else {
								ageDetail = "a couple of weeks"
							}
						} else {
							if pi.M == 1 {
								ageDetail = "just a month"
							} else if pi.M < 4 {
								ageDetail = "a few months"
							} else {
								ageDetail = "less than a year"
							}
						}
						ageDetail += " after " + s.Principal.Gender.SubjectPronoun() + " was born"
					}
				}
			}

			if ageDetail == "" {
				ageDetail = AgeQualifier(age)
			}
			evDetail += " " + ageDetail
		}
		evDetail += " " + bev.GetDate().When()
	} else {
		evDetail += " on an unknown date"
	}
	if !bev.GetPlace().IsUnknown() {
		pl := bev.GetPlace()
		evDetail = text.JoinSentenceParts(evDetail, pl.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}

	detail += s.Principal.PreferredFamiliarName + " " + evDetail

	burialRunOnSentence := true

	if s.Principal.CauseOfDeath != nil {
		detail = text.FinishSentence(detail)
		detail += " " + text.FormatSentence(text.JoinSentenceParts(s.Principal.Gender.PossessivePronounSingular(), "death was attributed to", enc.EncodeWithCitations(s.Principal.CauseOfDeath.Detail, s.Principal.CauseOfDeath.Citations)))
		burialRunOnSentence = false
	}

	additionalDetailFromDeathEvent := EventNarrativeDetail(bev, enc)
	if additionalDetailFromDeathEvent != "" {
		burialRunOnSentence = false
		detail = text.FinishSentence(detail)
		detail = text.JoinSentences(detail, additionalDetailFromDeathEvent)
	}

	detail = enc.EncodeWithCitations(detail, bev.GetCitations())

	if !burialRunOnSentence {
		enc.Para(render.Markdown(detail))
		detail = ""
	}

	funerals := []model.IndividualTimelineEvent{}
	for _, ev := range s.Principal.Timeline {
		switch tev := ev.(type) {
		case *model.BurialEvent:
			if !tev.Principal.SameAs(s.Principal) {
				continue
			}
			if tev != s.Principal.BestDeathlikeEvent {
				funerals = append(funerals, tev)
			}
		case *model.CremationEvent:
			if !tev.Principal.SameAs(s.Principal) {
				continue
			}
			if tev != s.Principal.BestDeathlikeEvent {
				funerals = append(funerals, tev)
			}
		}
	}
	if len(funerals) > 0 {
		// if len(funerals) > 1 {
		// TODO: record an anomaly
		// }

		evDetail := ""
		funeralEvent := funerals[0]
		switch funeralEvent.(type) {
		case *model.BurialEvent:
			evDetail += "buried"
		case *model.CremationEvent:
			evDetail += "cremated"
		default:
			panic("unhandled funeral event")
		}

		interval := bev.GetDate().IntervalUntil(funeralEvent.GetDate())
		if days, ok := interval.ApproxDays(); ok && days < 15 {
			if days == 0 {
				evDetail += " the same day"
			} else if days == 1 {
				evDetail += " the next day"
			} else {
				evDetail += fmt.Sprintf(" %s days later", text.CardinalNoun(days))
			}
		} else {
			evDetail += " " + funeralEvent.GetDate().When()
		}
		if !funeralEvent.GetPlace().IsUnknown() {
			pl := funeralEvent.GetPlace()
			evDetail = text.JoinSentenceParts(evDetail, pl.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
		}

		if detail == "" {
			detail = text.UpperFirst(s.Principal.Gender.SubjectPronounWithLink()) + " "
		} else {
			if burialRunOnSentence {
				detail += " and was "
			} else {
				detail = text.FinishSentence(detail) + " " + text.UpperFirst(s.Principal.Gender.SubjectPronounWithLink()) + " "
			}
		}

		detail = text.JoinSentenceParts(detail, enc.EncodeWithCitations(evDetail, funeralEvent.GetCitations()))

	}

	// if death is not inferred then perhaps make a statement about surviving partner
	if !bev.IsInferred() {
		if len(s.Principal.Families) > 0 {
			sort.Slice(s.Principal.Families, func(i, j int) bool {
				if s.Principal.Families[i].BestStartDate != nil || s.Principal.Families[j].BestStartDate == nil {
					return false
				}
				return s.Principal.Families[i].BestStartDate.SortsBefore(s.Principal.Families[j].BestStartDate)
			})

			lastFamily := s.Principal.Families[len(s.Principal.Families)-1]
			possibleSurvivor := lastFamily.OtherParent(s.Principal)
			if possibleSurvivor != nil && possibleSurvivor.BestDeathlikeEvent != nil && !possibleSurvivor.BestDeathlikeEvent.GetDate().IsUnknown() {
				if s.Principal.BestDeathlikeEvent.GetDate().SortsBefore(possibleSurvivor.BestDeathlikeEvent.GetDate()) {
					detail += " " + text.UpperFirst(s.Principal.Gender.SubjectPronounWithLink()) + " survived by "
					if lastFamily.Bond == model.FamilyBondMarried {
						detail += text.LowerFirst(s.Principal.Gender.PossessivePronounSingular()) + " " + text.LowerFirst(possibleSurvivor.Gender.RelationToSpouseNoun()) + " "
					}

					detail += intro.IntroducePerson(seq, possibleSurvivor, s.Start(), false, enc)
					detail = text.FinishSentence(detail)
				}
			}
		}
	}
	enc.Para(render.Markdown(detail))
}

func (s *DeathStatement) Start() *model.Date {
	return s.Principal.BestDeathlikeEvent.GetDate()
}

func (s *DeathStatement) End() *model.Date {
	return s.Principal.BestDeathlikeEvent.GetDate()
}

func (s *DeathStatement) NarrativeSequence() int {
	return NarrativeSequenceDeath
}

func (s *DeathStatement) Priority() int {
	return 5
}

type CensusStatement struct {
	Principal *model.Person
	Event     *model.CensusEvent
}

var _ Statement = (*CensusStatement)(nil)

func (s *CensusStatement) RenderDetail(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	ce, found := s.Event.Entry(s.Principal)
	if !found {
		return
	}

	year, _ := s.Event.GetDate().Year()

	narrative := ce.Narrative
	if narrative == "" {
		narrative = s.Event.Narrative.Text
	}

	if narrative != "" {
		var detail text.Para
		detail.NewSentence(intro.Pronoun(seq, s.Start()))
		detail.Continue(enc.EncodeWithCitations(WhatWhere(fmt.Sprintf("was recorded in the %d census", year), s.Event.GetPlace(), enc), s.Event.GetCitations())) // fmt.Sprintf("in the %d census", year)
		detail.NewSentence(narrative)
		detail.FinishSentence()
		enc.Para(render.Markdown(detail.Text()))
		return
	}
	// TODO: construct narrative of census

	// detail := fmt.Sprintf("in the %d census", year)
	// detail = text.JoinSentences(detail, intro.NameBased)
	// detail = enc.EncodeWithCitations(detail, s.Event.GetCitations())
	// detail = text.FormatSentence(detail)
	// enc.Para(detail)

	var detail text.Para
	what := ChooseFrom(seq,
		fmt.Sprintf("%s was recorded in the %d census", intro.Pronoun(seq, s.Start()), year),
		fmt.Sprintf("by the time of the %d census %s was living", year, intro.Pronoun(seq, s.Start())),
		fmt.Sprintf("in the %d census %s was living", year, intro.Pronoun(seq, s.Start())),
	)

	detail.NewSentence(enc.EncodeWithCitations(WhatWhere(what, s.Event.GetPlace(), enc), s.Event.GetCitations())) // fmt.Sprintf("in the %d census", year)

	var spouse *model.CensusEntry
	var father *model.CensusEntry
	var mother *model.CensusEntry
	var children []*model.CensusEntry
	var siblings []*model.CensusEntry
	var relations []*model.CensusEntry

	for _, en := range s.Event.Entries {
		if en.Principal.SameAs(s.Principal) {
			continue
		}
		if s.Principal.Father.SameAs(en.Principal) {
			father = en
			continue
		}
		if s.Principal.Mother.SameAs(en.Principal) {
			mother = en
			continue
		}
		if en.Principal.Father.SameAs(s.Principal) || en.Principal.Mother.SameAs(s.Principal) {
			children = append(children, en)
			continue
		}
		rel := strings.ToLower(en.Principal.RelationTo(s.Principal, s.Event.GetDate()))
		if rel != "" {
			if rel == "wife" || rel == "husband" {
				spouse = en
				continue
			}
			if rel == "brother" || rel == "sister" || rel == "half-sister" || rel == "half-brother" || rel == "stepsister" || rel == "stepbrother" {
				siblings = append(siblings, en)
				continue
			}
			relations = append(relations, en)
			continue
		}
	}

	if spouse != nil || father != nil || mother != nil || len(children) != 0 || len(relations) != 0 || len(siblings) != 0 {
		detail.Continue("with")
		detail.Continue(s.Principal.Gender.PossessivePronounSingular())

		peopleList := make([]string, 0, len(children)+len(relations)+len(siblings)+3)

		if spouse != nil {
			rel := strings.ToLower(spouse.Principal.RelationTo(s.Principal, s.Event.GetDate()))
			peopleList = append(peopleList, rel+" "+intro.IntroducePerson(seq, spouse.Principal, s.Start(), false, enc))
		}

		if len(children) > 0 {
			if len(children) == 1 {
				rel := strings.ToLower(children[0].Principal.RelationTo(s.Principal, s.Event.GetDate()))
				peopleList = append(peopleList, rel+" "+intro.IntroducePerson(seq, children[0].Principal, s.Start(), true, enc))
				// peopleList = append(peopleList, rel+" "+enc.EncodeModelLink(children[0].Principal.PreferredGivenName, children[0].Principal))
			} else {
				ens := make([]string, 0, len(children))
				for _, en := range children {
					ens = append(ens, intro.IntroducePerson(seq, en.Principal, s.Start(), true, enc))
					// ens = append(ens, enc.EncodeModelLink(en.Principal.PreferredGivenName, en.Principal))
				}
				peopleList = append(peopleList, text.CardinalNoun(len(ens))+" children "+text.JoinList(ens))
			}
		}

		if father != nil {
			peopleList = append(peopleList, "father "+intro.IntroducePerson(seq, father.Principal, s.Start(), false, enc))
		}
		if mother != nil {
			peopleList = append(peopleList, "mother "+intro.IntroducePerson(seq, mother.Principal, s.Start(), false, enc))
		}

		if len(siblings) > 0 {
			if len(siblings) == 1 {
				rel := strings.ToLower(siblings[0].Principal.RelationTo(s.Principal, s.Event.GetDate()))
				peopleList = append(peopleList, rel+" "+enc.EncodeModelLink(siblings[0].Principal.PreferredGivenName, siblings[0].Principal))
			} else {

				ens := make([]string, 0, len(siblings))
				for _, en := range siblings {
					ens = append(ens, intro.IntroducePerson(seq, en.Principal, s.Start(), true, enc))
				}
				peopleList = append(peopleList, text.CardinalNoun(len(ens))+" siblings "+text.JoinList(ens))
			}
		}

		if len(relations) > 0 {
			ens := make([]string, 0, len(relations))
			for _, en := range relations {
				rel := strings.ToLower(en.Principal.RelationTo(s.Principal, s.Event.GetDate()))
				ens = append(ens, rel+" "+intro.IntroducePerson(seq, en.Principal, s.Start(), true, enc))
			}
			peopleList = append(peopleList, text.JoinList(ens))
		}

		detail.Continue(text.JoinList(peopleList))
	}

	enc.Para(render.Markdown(detail.Text()))
}

func (s *CensusStatement) Start() *model.Date {
	return s.Event.GetDate()
}

func (s *CensusStatement) End() *model.Date {
	return s.Event.GetDate()
}

func (s *CensusStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *CensusStatement) Priority() int {
	return 4
}

// A NarrativeStatement is used for any general event that includes a narrative.
// If the Event is an IndividualNarrativeEvent then the narrative field is used in
// place of any generated text. Otherwise an introductory sentence is prepended.
type NarrativeStatement struct {
	Principal *model.Person
	Event     model.TimelineEvent
}

var _ Statement = (*NarrativeStatement)(nil)

func (s *NarrativeStatement) RenderDetail(seq int, intro *IntroGenerator, enc render.MarkupBuilder, hints *GrammarHints) {
	narrative := EventNarrativeDetail(s.Event, enc)
	if narrative == "" {
		return
	}

	var detail text.Para
	switch s.Event.(type) {
	case *model.IndividualNarrativeEvent:
	default:
		// prepend an intro
		detail.NewSentence(intro.Pronoun(seq, s.Start()))
		detail.Continue(enc.EncodeWithCitations(EventWhatWhenWhere(s.Event, enc), s.Event.GetCitations()))
	}

	detail.NewSentence(narrative)
	detail.FinishSentence()

	// enc.ParaWithFigure(enc.EncodeWithCitations(detail.Text(), s.Event.GetCitations()), "/trees/cg/media/6V7KWAJR2LCVK.png", "alt text", "this is a caption")
	enc.Para(render.Markdown(enc.EncodeWithCitations(detail.Text(), s.Event.GetCitations())))
}

func (s *NarrativeStatement) Start() *model.Date {
	return s.Event.GetDate()
}

func (s *NarrativeStatement) End() *model.Date {
	return s.Event.GetDate()
}

func (s *NarrativeStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *NarrativeStatement) Priority() int {
	return 0
}

// UTILITY

func ChooseFrom(n int, alternatives ...string) string {
	return alternatives[n%len(alternatives)]
}

func EventNarrativeDetail(ev model.TimelineEvent, enc render.MarkupBuilder) string {
	narr := ev.GetNarrative()
	if narr.Text == "" {
		detail := strings.ToLower(ev.GetDetail())
		if strings.HasPrefix(detail, "she was recorded as") || strings.HasPrefix(detail, "he was recorded as") || strings.HasPrefix(detail, "it was recorded that") {
			return ev.GetDetail()
		}
		return ""
	}
	return EncodeText(narr, enc)
}

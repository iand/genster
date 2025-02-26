package narrative

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"sort"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/text"
)

type Statement[T render.EncodedText] interface {
	RenderDetail(int, *IntroGenerator[T], render.ContentBuilder[T], NameChooser)
	Start() *model.Date
	End() *model.Date
	NarrativeSequence() int
	Priority() int // priority within a narrative against another statement with same date, higher will be rendered first
}

type IntroStatement[T render.EncodedText] struct {
	Principal           *model.Person
	Baptisms            []*model.BaptismEvent
	SuppressRelation    bool
	NameChooser         NameChooser
	IncludeMedia        bool
	CropMediaHighlights bool
}

var _ Statement[md.Text] = (*IntroStatement[md.Text])(nil)

func (s *IntroStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	var birth string
	if s.NameChooser == nil {
		s.NameChooser = nc
	}

	birthNarrative := ""
	// Prose birth
	if s.Principal.BestBirthlikeEvent != nil {
		// birth = text.LowerFirst(EventTitle(s.Principal.BestBirthlikeEvent, enc, &model.POV{Person: s.Principal}))
		birth = enc.EncodeWithCitations(enc.EncodeText(text.LowerFirst(EventWhatWhenWherePov(s.Principal.BestBirthlikeEvent, enc, s.NameChooser, intro.POV))), s.Principal.BestBirthlikeEvent.GetCitations()).String()

		birthNarrative = EventNarrativeDetail(s.Principal.BestBirthlikeEvent, enc)
	}

	// Prose parentage
	parentUnknownDetail := ""
	parentDetail := ""
	parentageDetailPrefix := "the " + PositionInFamily(s.Principal) + " of "
	if s.Principal.Father.IsUnknown() {
		if s.Principal.Mother.IsUnknown() {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " parents are not known"
		} else {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " father is not known"
			parentDetail = parentageDetailPrefix + intro.IntroducePerson(seq, s.Principal.Mother, s.Start(), false, enc, nc)
		}
	} else {
		if s.Principal.Mother.IsUnknown() {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " mother is not known"
			parentDetail = parentageDetailPrefix + intro.IntroducePerson(seq, s.Principal.Father, s.Start(), false, enc, nc)
		} else {
			parentDetail = parentageDetailPrefix + intro.IntroducePerson(seq, s.Principal.Father, s.Start(), false, enc, nc) + " and " + intro.IntroducePerson(seq, s.Principal.Mother, s.Start(), false, enc, nc)
		}
	}

	// ---------------------------------------
	// Build detail
	// ---------------------------------------
	detail := ""

	if s.Principal.NickName != "" {
		detail = text.JoinSentenceParts(detail, "(known as "+s.Principal.NickName+")")
	}

	// detail += " "
	if birth != "" {
		detail = text.JoinSentenceParts(detail, birth)
		if birthNarrative != "" {
			// TODO: this is capitalizing the "was" between the name and the birth
			// detail = text.JoinSentences(detail, birthNarrative)
			// detail = text.FinishSentence(detail)
		}
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
			twinLink := enc.EncodeModelLink(enc.EncodeText(as.Other.PreferredFamiliarName), as.Other)

			detail = text.JoinSentenceParts(detail, text.UpperFirst(s.Principal.Gender.SubjectPronoun()), "was the twin to", s.Principal.Gender.PossessivePronounSingular(), as.Other.Gender.RelationToSiblingNoun(), enc.EncodeWithCitations(twinLink, as.Citations).String())
			twinClause = true
			break
		}
	}

	// Insert baptism here if there is only one, otherwise leave for a new para
	if len(s.Baptisms) == 1 && s.Baptisms[0] != s.Principal.BestBirthlikeEvent {
		bapDetail := AgeWhenWherePov(s.Baptisms[0], enc, nc, intro.POV)
		if bapDetail != "" {

			if twinClause {
				detail = text.JoinSentenceParts(detail, "and")
			} else {
				detail = text.JoinSentenceParts(text.FinishSentence(detail), text.UpperFirst(s.Principal.Gender.SubjectPronoun()))
			}

			detail = text.JoinSentenceParts(detail, "was baptised", enc.EncodeWithCitations(enc.EncodeText(bapDetail), s.Baptisms[0].GetCitations()).String())
			detail = text.FinishSentence(detail)
		}

	}

	// ---------------------------------------
	// Prose relation to key person
	// ---------------------------------------
	if !s.SuppressRelation {
		if s.Principal.RelationToKeyPerson != nil && !s.Principal.RelationToKeyPerson.IsSelf() {
			detail += " " + text.UpperFirst(s.Principal.Gender.SubjectPronoun()) + " is " + enc.EncodeModelLink(enc.EncodeText(text.MaybePossessiveSuffix(s.Principal.RelationToKeyPerson.From.PreferredFamiliarName)), s.Principal.RelationToKeyPerson.From).String() + " " + s.Principal.RelationToKeyPerson.Name()
		}
	}

	detail = text.FinishSentence(detail)
	enc.Para(enc.EncodeText(detail))

	if s.IncludeMedia {
		if s.Principal.BestBirthlikeEvent != nil && len(s.Principal.BestBirthlikeEvent.GetMediaObjects()) > 0 {
			// TODO: handle error
			MediaObjectsAsFigures(s.Principal.BestBirthlikeEvent.GetMediaObjects(), enc, s.CropMediaHighlights)
		}
		if len(s.Baptisms) == 1 && s.Baptisms[0] != s.Principal.BestBirthlikeEvent && len(s.Baptisms[0].GetMediaObjects()) > 0 {
			// TODO: handle error
			MediaObjectsAsFigures(s.Baptisms[0].GetMediaObjects(), enc, s.CropMediaHighlights)
		}
	}

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
			aww := AgeWhenWherePov(bev, enc, nc, intro.POV)
			if aww != "" {
				bapDetail = text.JoinSentenceParts(bapDetail, evDetail, enc.EncodeWithCitations(enc.EncodeText(bapDetail), s.Baptisms[0].GetCitations()).String())
			}
		}
		bapDetail = text.FinishSentence(text.JoinSentenceParts(intro.Pronoun(seq, s.Start(), s.Principal), bapDetail))
		enc.Para(enc.EncodeText(bapDetail))
	}
}

func (s *IntroStatement[T]) Start() *model.Date {
	return s.Principal.BestBirthDate()
}

func (s *IntroStatement[T]) End() *model.Date {
	return s.Principal.BestBirthDate()
}

func (s *IntroStatement[T]) NarrativeSequence() int {
	return NarrativeSequenceIntro
}

func (s *IntroStatement[T]) Priority() int {
	return 10
}

type FamilyStatement[T render.EncodedText] struct {
	Principal *model.Person
	Family    *model.Family
}

var _ Statement[md.Text] = (*FamilyStatement[md.Text])(nil)

func (s *FamilyStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	// TODO: note for example VFA3VQS22ZHBO George Henry Chambers (1903-1985) who
	// had a child with Dorothy Youngs in 1944 but didn't marry until 1985
	other := s.Family.OtherParent(s.Principal)

	// Special cases for single parents
	if s.Family.Bond == model.FamilyBondUnmarried || s.Family.Bond == model.FamilyBondLikelyUnmarried {
		if other.IsUnknown() {
			s.renderIllegitimate(seq, intro, enc, nc)
			return
		}
		s.renderUnmarried(seq, intro, enc, nc)
		return
	} else if other.IsUnknown() {
		s.renderUnknownPartner(seq, intro, enc, nc)
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
		otherName = intro.IntroducePerson(seq, other, s.Start(), false, enc, nc)
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
			detail.ReplaceSentence(intro.Pronoun(seq, s.Start(), s.Principal))
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
			event = WhatWherePov(event, s.Family.BestStartEvent.GetPlace(), enc, nc, intro.POV)
		}
		if s.Family.BestStartEvent != nil {
			detail.Continue(enc.EncodeWithCitations(enc.EncodeText(event), s.Family.BestStartEvent.GetCitations()).String())
		} else {
			detail.Continue(event)
		}
	}

	if len(s.Family.Children) == 0 {
		// single parents already dealt with
		if s.Principal.Childless {
			detail.AddCompleteSentence("they had no children")
		}
	} else {

		childCardinal := ChildCardinal(s.Family.Children)
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

	childList := ChildList(s.Family.Children, enc, nc)
	if len(childList) == 0 {
		enc.Para(enc.EncodeText(detail.Text()))
		return
	}

	detail.FinishSentenceWithTerminator(":–")
	enc.Para(enc.EncodeText(detail.Text()))
	enc.UnorderedList(childList)
}

func (s *FamilyStatement[T]) Start() *model.Date {
	return s.Family.BestStartDate
}

func (s *FamilyStatement[T]) End() *model.Date {
	return s.Family.BestEndDate
}

func (s *FamilyStatement[T]) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *FamilyStatement[T]) Priority() int {
	return 5
}

func ChildCardinal(clist []*model.Person) string {
	// TODO: note how many children survived if some died
	allSameGender := true
	if clist[0].Redacted {
		allSameGender = false
	} else if clist[0].Gender != model.GenderUnknown {
		for i := 1; i < len(clist); i++ {
			if clist[i].Redacted {
				allSameGender = false
				break
			}
			if clist[i].Gender != clist[0].Gender {
				allSameGender = false
				break
			}
		}
	}

	if allSameGender {
		if clist[0].Gender == model.GenderMale {
			return text.CardinalWithUnit(len(clist), "son", "sons")
		} else {
			return text.CardinalWithUnit(len(clist), "daughter", "daughters")
		}
	}
	return text.CardinalWithUnit(len(clist), "child", "children")
}

func ChildList[T render.EncodedText](clist []*model.Person, enc render.ContentBuilder[T], nc NameChooser) []T {
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
	childList := make([]T, 0, len(clist))
	for _, c := range clist {
		if c.Redacted {
			redactedCount++
			continue
		}
		childList = append(childList, PersonSummary(c, enc, nc, enc.EncodeText(c.PreferredGivenName), true, false, true, true, true))
	}
	if len(childList) == 0 {
		return childList
	}
	if redactedCount > 0 {
		childList = append(childList, enc.EncodeText(text.CardinalWithUnit(redactedCount, "other child", "other children")+" living or recently died"))
	}

	return childList
}

func (s *FamilyStatement[T]) renderIllegitimate(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
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
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))
			detail.Continue("gave birth to a", c.Gender.RelationToParentNoun())
			detail.AppendAsAside(enc.EncodeModelLink(enc.EncodeText(c.PreferredFullName), c).String())
			detail.Continue(enc.EncodeWithCitations(enc.EncodeText(EventWhenWherePov(c.BestBirthlikeEvent, enc, nc, intro.POV)), c.BestBirthlikeEvent.GetCitations()).String())

		} else {
			// this form: "At the age of thirty-four, Annie had a"
			detail.AppendAsAside(intro.RelativeTime(seq, c.BestBirthlikeEvent.GetDate(), false))
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))
			detail.Continue("had a", c.Gender.RelationToParentNoun())
			detail.AppendAsAside(enc.EncodeModelLink(enc.EncodeText(c.PreferredFullName), c).String())
			detail.Continue("who")
			detail.Continue(enc.EncodeWithCitations(enc.EncodeText(EventWhatWhenWherePov(c.BestBirthlikeEvent, enc, nc, intro.POV)), c.BestBirthlikeEvent.GetCitations()).String())

		}

		// pad the sentence to be longer if needed
		if detail.CurrentSentenceLength() < 60 {
			detail.Continue(ChooseFrom(seq,
				"with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
				"by an unknown "+s.Principal.Gender.Opposite().Noun(),
			))
		}

		if c.Redacted {
			enc.Para(enc.EncodeText(detail.Text()))
		} else {
			detail.FinishSentenceWithTerminator(":–")
			enc.Para(enc.EncodeText(detail.Text()))
			enc.UnorderedList([]T{PersonSummary(c, enc, nc, enc.EncodeText(c.PreferredFamiliarName), false, false, false, true, true)})
		}
	} else {
		panic(fmt.Sprintf("Not implemented: renderIllegitimate where person has more than one child or is the father (id=%s, name=%s)", s.Principal.ID, s.Principal.PreferredUniqueName))
	}
}

func (s *FamilyStatement[T]) renderUnmarried(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
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
	otherName := intro.IntroducePerson(seq, other, s.Start(), false, enc, nc)

	var detail text.Para
	if oneChild {
		c := s.Family.Children[0]
		detail.AppendAsAside(intro.RelativeTime(seq, s.Family.Children[0].BestBirthlikeEvent.GetDate(), useBirthDateInIntro))
		detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))
		if isMother && childFirmBirthdate {
			detail.Continue("gave birth to a")
			detail.Continue(c.Gender.RelationToParentNoun())
			if !c.Redacted {
				detail.AppendAsAside(enc.EncodeModelLink(enc.EncodeText(c.PreferredFamiliarName), c).String())
			}
			detail.AppendClause("the child of")
			detail.Continue(otherName)
		} else {
			detail.Continue(ChooseFrom(seq,
				"had a child",
				"had a "+c.Gender.RelationToParentNoun(),
			))
			if !c.Redacted {
				detail.AppendAsAside(enc.EncodeModelLink(enc.EncodeText(c.PreferredFamiliarName), c).String())
			}
			detail.Continue("with", otherName)
		}
		detail.FinishSentence()

		if c.Redacted {
			enc.Para(enc.EncodeText(detail.Text()))
		} else {
			detail.FinishSentenceWithTerminator(":–")
			enc.Para(enc.EncodeText(detail.Text()))
			enc.UnorderedList([]T{PersonSummary(c, enc, nc, enc.EncodeText(c.PreferredFamiliarName), false, false, false, true, true)})
		}

	} else {
		c := s.Family.Children[0]
		detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))

		childCardinal := ChildCardinal(s.Family.Children)
		detail.Continue("had", childCardinal)

		otherName = intro.IntroducePerson(seq, other, s.Start(), false, enc, nc)
		detail.Continue("with", otherName)

		childList := ChildList(s.Family.Children, enc, nc)
		if len(childList) == 0 {
			enc.Para(enc.EncodeText(detail.Text()))
			return
		}

		detail.FinishSentenceWithTerminator(":–")
		enc.Para(enc.EncodeText(detail.Text()))
		enc.UnorderedList(childList)

	}
}

func (s *FamilyStatement[T]) renderUnknownPartner(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	// married or unknown relationship but the other parent is unknown
	// panic(fmt.Sprintf("Not implemented: renderUnknownPartner (id=%s, name=%s)", s.Principal.ID, s.Principal.PreferredUniqueName))
}

type FamilyEndStatement[T render.EncodedText] struct {
	Principal *model.Person
	Family    *model.Family
}

var _ Statement[md.Text] = (*FamilyEndStatement[md.Text])(nil)

func (s *FamilyEndStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
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
		detail.NewSentence(PersonDeathSummary(other, enc, nc, enc.EncodeText(name), true, false, true, true).String())
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

	enc.Para(enc.EncodeText(detail.Text()))
}

func (s *FamilyEndStatement[T]) Start() *model.Date {
	return s.Family.BestEndDate
}

func (s *FamilyEndStatement[T]) End() *model.Date {
	return s.Family.BestEndDate
}

func (s *FamilyEndStatement[T]) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *FamilyEndStatement[T]) endedWithDeathOf(p *model.Person) bool {
	return p.SameAs(s.Family.EndDeathPerson)
}

func (s *FamilyEndStatement[T]) Priority() int {
	return 4
}

type DeathStatement[T render.EncodedText] struct {
	Principal               *model.Person
	ExcludeSurvivingPartner bool
	IncludeMedia            bool
	CropMediaHighlights     bool
}

var _ Statement[md.Text] = (*DeathStatement[md.Text])(nil)

func (s *DeathStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	var detail string

	bev := s.Principal.BestDeathlikeEvent

	evDetail := DeathWhat(bev, s.Principal.ModeOfDeath)

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
		evDetail = WhatWherePov(evDetail, pl, enc, nc, intro.POV)
		// evDetail = text.JoinSentenceParts(evDetail, pl.InAt(), enc.EncodeModelLinkDedupe(enc.EncodeText(pl.ProseName), enc.EncodeText(pl.NameWithDistrict), pl).String())
	}

	burialRunOnSentence := true

	detail += s.Principal.PreferredFamiliarName + " " + evDetail

	if s.Principal.CauseOfDeath != nil {
		detail = enc.EncodeWithCitations(enc.EncodeText(detail), bev.GetCitations()).String()
		detail = text.FinishSentence(detail)
		detail += " " + text.FormatSentence(text.JoinSentenceParts(s.Principal.Gender.PossessivePronounSingular(), "death was attributed to", enc.EncodeWithCitations(enc.EncodeText(s.Principal.CauseOfDeath.Detail), s.Principal.CauseOfDeath.Citations).String()))
		burialRunOnSentence = false
	}

	additionalDetailFromDeathEvent := EventNarrativeDetail(bev, enc)
	if additionalDetailFromDeathEvent != "" {
		burialRunOnSentence = false
		detail = enc.EncodeWithCitations(enc.EncodeText(detail), bev.GetCitations()).String()
		detail = text.FinishSentence(detail)
		detail = text.JoinSentences(detail, additionalDetailFromDeathEvent)
	}

	if !burialRunOnSentence {
		enc.Para(enc.EncodeText(detail))
		detail = ""
	} else {
		detail = enc.EncodeWithCitations(enc.EncodeText(detail), bev.GetCitations()).String()
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
			evDetail = WhatWherePov(evDetail, pl, enc, nc, intro.POV)
			// evDetail = text.JoinSentenceParts(evDetail, pl.InAt(), enc.EncodeModelLinkDedupe(enc.EncodeText(pl.ProseName), enc.EncodeText(pl.NameWithDistrict), pl).String())
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

		detail = text.JoinSentenceParts(detail, enc.EncodeWithCitations(enc.EncodeText(evDetail), funeralEvent.GetCitations()).String())

	}

	// if death is not inferred then perhaps make a statement about surviving partner
	if !s.ExcludeSurvivingPartner && !bev.IsInferred() {
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

					detail += intro.IntroducePerson(seq, possibleSurvivor, s.Start(), false, enc, nc)
					detail = text.FinishSentence(detail)
				}
			}
		}
	}
	enc.Para(enc.EncodeText(detail))

	if s.IncludeMedia && bev != nil {
		// TODO: handle error
		MediaObjectsAsFigures(bev.GetMediaObjects(), enc, s.CropMediaHighlights)
	}
}

func (s *DeathStatement[T]) Start() *model.Date {
	return s.Principal.BestDeathlikeEvent.GetDate()
}

func (s *DeathStatement[T]) End() *model.Date {
	return s.Principal.BestDeathlikeEvent.GetDate()
}

func (s *DeathStatement[T]) NarrativeSequence() int {
	return NarrativeSequenceDeath
}

func (s *DeathStatement[T]) Priority() int {
	return 5
}

type CensusStatement[T render.EncodedText] struct {
	Principal           *model.Person
	Event               *model.CensusEvent
	IncludeMedia        bool
	CropMediaHighlights bool
}

var _ Statement[md.Text] = (*CensusStatement[md.Text])(nil)

func (s *CensusStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
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
		detail.NewSentence(intro.Pronoun(seq, s.Start(), s.Principal))
		detail.Continue(enc.EncodeWithCitations(enc.EncodeText(WhatWherePov(fmt.Sprintf("%s recorded in the %d census", intro.WasWere(s.Principal), year), s.Event.GetPlace(), enc, nc, intro.POV)), s.Event.GetCitations()).String()) // fmt.Sprintf("in the %d census", year)
		detail.NewSentence(narrative)
		detail.FinishSentence()
		enc.Para(enc.EncodeText(detail.Text()))
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
		fmt.Sprintf("%s %s recorded in the %d census", intro.Pronoun(seq, s.Start(), s.Principal), intro.WasWere(s.Principal), year),
		fmt.Sprintf("by the time of the %d census %s %s living", year, intro.Pronoun(seq, s.Start(), s.Principal), intro.WasWere(s.Principal)),
		fmt.Sprintf("in the %d census %s %s living", year, intro.Pronoun(seq, s.Start(), s.Principal), intro.WasWere(s.Principal)),
	)

	detail.NewSentence(enc.EncodeWithCitations(enc.EncodeText(WhatWherePov(what, s.Event.GetPlace(), enc, nc, intro.POV)), s.Event.GetCitations()).String()) // fmt.Sprintf("in the %d census", year)

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
			peopleList = append(peopleList, rel+" "+intro.IntroducePerson(seq, spouse.Principal, s.Start(), false, enc, nc))
		}

		if len(children) > 0 {
			if len(children) == 1 {
				rel := strings.ToLower(children[0].Principal.RelationTo(s.Principal, s.Event.GetDate()))
				peopleList = append(peopleList, rel+" "+intro.IntroducePerson(seq, children[0].Principal, s.Start(), true, enc, nc))
			} else {
				ens := make([]string, 0, len(children))
				for _, en := range children {
					ens = append(ens, intro.IntroducePerson(seq, en.Principal, s.Start(), true, enc, nc))
				}
				peopleList = append(peopleList, text.CardinalNoun(len(ens))+" children "+text.JoinList(ens))
			}
		}

		if father != nil {
			peopleList = append(peopleList, "father "+intro.IntroducePerson(seq, father.Principal, s.Start(), false, enc, nc))
		}
		if mother != nil {
			peopleList = append(peopleList, "mother "+intro.IntroducePerson(seq, mother.Principal, s.Start(), false, enc, nc))
		}

		if len(siblings) > 0 {
			if len(siblings) == 1 {
				rel := strings.ToLower(siblings[0].Principal.RelationTo(s.Principal, s.Event.GetDate()))
				peopleList = append(peopleList, rel+" "+enc.EncodeModelLink(enc.EncodeText(siblings[0].Principal.PreferredGivenName), siblings[0].Principal).String())
			} else {

				ens := make([]string, 0, len(siblings))
				for _, en := range siblings {
					ens = append(ens, intro.IntroducePerson(seq, en.Principal, s.Start(), true, enc, nc))
				}
				peopleList = append(peopleList, text.CardinalNoun(len(ens))+" siblings "+text.JoinList(ens))
			}
		}

		if len(relations) > 0 {
			ens := make([]string, 0, len(relations))
			for _, en := range relations {
				rel := strings.ToLower(en.Principal.RelationTo(s.Principal, s.Event.GetDate()))
				ens = append(ens, rel+" "+intro.IntroducePerson(seq, en.Principal, s.Start(), true, enc, nc))
			}
			peopleList = append(peopleList, text.JoinList(ens))
		}

		detail.Continue(text.JoinList(peopleList))
	}

	enc.Para(enc.EncodeText(detail.Text()))

	if s.IncludeMedia {
		// TODO: handle error
		MediaObjectsAsFigures(s.Event.GetMediaObjects(), enc, s.CropMediaHighlights)
	}
}

func (s *CensusStatement[T]) Start() *model.Date {
	return s.Event.GetDate()
}

func (s *CensusStatement[T]) End() *model.Date {
	return s.Event.GetDate()
}

func (s *CensusStatement[T]) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *CensusStatement[T]) Priority() int {
	return 4
}

// A NarrativeStatement is used for any general event that includes a narrative.
// If the Event is an IndividualNarrativeEvent then the narrative field is used in
// place of any generated text. Otherwise an introductory sentence is prepended.
type NarrativeStatement[T render.EncodedText] struct {
	Principal *model.Person
	Event     model.TimelineEvent
}

var _ Statement[md.Text] = (*NarrativeStatement[md.Text])(nil)

func (s *NarrativeStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	narrative := EventNarrativeDetail(s.Event, enc)
	if narrative == "" {
		return
	}

	var detail text.Para
	switch s.Event.(type) {
	case *model.IndividualNarrativeEvent:
	default:
		// prepend an intro
		detail.NewSentence(intro.Pronoun(seq, s.Start(), s.Principal))
		detail.Continue(enc.EncodeWithCitations(enc.EncodeText(EventWhatWhenWherePov(s.Event, enc, nc, intro.POV)), s.Event.GetCitations()).String())
	}

	detail.NewSentence(narrative)
	detail.FinishSentence()

	// enc.ParaWithFigure(enc.EncodeWithCitations(detail.Text(), s.Event.GetCitations()), "/trees/cg/media/6V7KWAJR2LCVK.png", "alt text", "this is a caption")
	enc.Para(enc.EncodeWithCitations(enc.EncodeText(detail.Text()), s.Event.GetCitations()))
}

func (s *NarrativeStatement[T]) Start() *model.Date {
	return s.Event.GetDate()
}

func (s *NarrativeStatement[T]) End() *model.Date {
	return s.Event.GetDate()
}

func (s *NarrativeStatement[T]) NarrativeSequence() int {
	if !s.Event.GetDate().SortsBefore(s.Principal.BestDeathDate()) {
		return NarrativeSequenceDeath
	}
	return NarrativeSequenceLifeStory
}

func (s *NarrativeStatement[T]) Priority() int {
	return 0
}

type ChildrenStatement[T render.EncodedText] struct {
	Family *model.Family
}

var _ Statement[md.Text] = (*ChildrenStatement[md.Text])(nil)

func (s *ChildrenStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	var detail text.Para

	singleParent := false
	if s.Family.Bond == model.FamilyBondUnmarried || s.Family.Bond == model.FamilyBondLikelyUnmarried ||
		(s.Family.Bond == model.FamilyBondUnknown && (s.Family.Father.IsUnknown() || s.Family.Mother.IsUnknown())) {
		singleParent = true
	}

	if s.Family.Father.IsUnknown() {
		if s.Family.Mother.IsUnknown() {
			detail.NewSentence("They")
		} else {
			detail.NewSentence(s.Family.Mother.PreferredGivenName)
		}
	} else {
		if s.Family.Mother.IsUnknown() {
			detail.NewSentence(s.Family.Father.PreferredGivenName)
		} else {
			detail.NewSentence("They")
		}
	}

	if len(s.Family.Children) == 0 {
		// single parents already dealt with
		if (!s.Family.Father.IsUnknown() && s.Family.Father.Childless) || (!s.Family.Mother.IsUnknown() && s.Family.Mother.Childless) {
			detail.AddCompleteSentence("they had no children")
		}
	} else {

		childCardinal := ChildCardinal(s.Family.Children)
		if singleParent {
			detail.Continue("had " + childCardinal)
		} else {
			switch len(s.Family.Children) {
			case 1:
				detail.Continue(ChooseFrom(seq,
					"had "+childCardinal+":",
					"had just one child together:",
					"had "+childCardinal+":",
				))
			default:
				detail.Continue("had " + childCardinal + ":")
			}
		}
	}

	childList := ChildList(s.Family.Children, enc, nc)
	if len(childList) == 0 {
		enc.Para(enc.EncodeText(detail.Text()))
		return
	}

	detail.FinishSentenceWithTerminator(":–")
	enc.Para(enc.EncodeText(detail.Text()))
	enc.UnorderedList(childList)
}

func (s *ChildrenStatement[T]) Start() *model.Date {
	return s.Family.BestStartDate
}

func (s *ChildrenStatement[T]) End() *model.Date {
	return s.Family.BestEndDate
}

func (s *ChildrenStatement[T]) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *ChildrenStatement[T]) Priority() int {
	return 5
}

type MediaStatement[T render.EncodedText] struct {
	Date         *model.Date
	Sequence     int
	MediaObjects []*model.CitedMediaObject
}

var _ Statement[md.Text] = (*MediaStatement[md.Text])(nil)

func (s *MediaStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	for _, mo := range s.MediaObjects {
		enc.Figure(mo.Object.SrcFilePath, mo.Object.Title, enc.EncodeText(mo.Object.Title), mo.Highlight)
	}
}

func (s *MediaStatement[T]) Start() *model.Date {
	return s.Date
}

func (s *MediaStatement[T]) End() *model.Date {
	return s.Date
}

func (s *MediaStatement[T]) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

func (s *MediaStatement[T]) Priority() int {
	return 7
}

func cropMedia(mo *model.MediaObject, region *model.Region) (string, error) {
	f, err := os.Open(mo.SrcFilePath)
	if err != nil {
		return "", fmt.Errorf("open image: %w", err)
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	if region != nil {

		type subImager interface {
			SubImage(r image.Rectangle) image.Image
			Bounds() image.Rectangle
		}
		sub, ok := img.(subImager)
		if !ok {
			return "", fmt.Errorf("image does not support cropping")
		}
		bounds := sub.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		crop := image.Rectangle{
			Min: image.Point{
				X: int(float64(width) * float64(region.Left) / 100.0),
				Y: int(float64(height) * (1 - float64(region.Bottom+region.Height)/100.0)),
			},
			Max: image.Point{
				X: int(float64(width) * float64(region.Left+region.Width) / 100.0),
				Y: int(float64(height) * (1 - float64(region.Bottom)/100.0)),
			},
		}
		img = sub.SubImage(crop)
	}

	out := new(bytes.Buffer)
	switch format {
	case "jpeg":
		err := jpeg.Encode(out, img, nil)
		if err != nil {
			return "", fmt.Errorf("encode image: %w", err)
		}
	case "png":
		err := png.Encode(out, img)
		if err != nil {
			return "", fmt.Errorf("encode image: %w", err)
		}
	default:
		return "", fmt.Errorf("unknown image format %q", format)
	}

	base64Img := base64.StdEncoding.EncodeToString(out.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", base64Img), nil
}

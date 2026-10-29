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
	PossibleBirths      []*model.PossibleBirthEvent
	SuppressRelation    bool
	NameChooser         NameChooser
	IncludeMedia        bool
	CropMediaHighlights bool
}

var _ Statement[md.Text] = (*IntroStatement[md.Text])(nil)

func (s *IntroStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	if s.NameChooser == nil {
		s.NameChooser = nc
	}

	var para text.Para

	if s.Principal.BestBirthlikeEvent == nil && s.Principal.Father.IsUnknown() && s.Principal.Mother.IsUnknown() {
		if len(s.PossibleBirths) > 0 {
			s.addPossibleBirths(&para, possessiveName(s.Principal), enc, nc, intro)
		} else if s.Principal.Unidentified {
			para.StartSentence("nothing is known about", s.Principal.Gender.PossessivePronounSingular(), "early life")
		} else {
			para.StartSentence("nothing is known about the early life of", s.Principal.PreferredGivenName)
			s.maybeAppendNickname(&para)
		}
	} else {
		if s.Principal.BestBirthlikeEvent != nil {
			para.StartSentence(s.Principal.PreferredGivenName)
			s.maybeAppendNickname(&para)
			birth := enc.EncodeWithCitations(enc.EncodeText(text.LowerFirst(EventWhatWhenWherePov(s.Principal.BestBirthlikeEvent, enc, s.NameChooser, intro.POV))), s.Principal.BestBirthlikeEvent.GetCitations()).String()
			para.Continue(birth)
			s.addParentage(&para, enc, nc, intro, seq, true)
			s.maybeAddBaptismAfterBirth(&para, enc, nc, intro)
		} else {
			para.StartSentence(s.Principal.PreferredGivenName)
			s.maybeAppendNickname(&para)
			para.Continue("was")
			s.addParentage(&para, enc, nc, intro, seq, false)
			if len(s.PossibleBirths) > 0 {
				s.addPossibleBirths(&para, s.Principal.Gender.PossessivePronounSingular(), enc, nc, intro)
			} else {
				para.StartSentence("nothing is known about", s.Principal.Gender.PossessivePronounSingular(), "date and place of birth")
			}
		}
		s.maybeAddTwinAssociation(&para, enc)
	}

	s.maybeAddRelationToKeyPerson(&para, enc)
	enc.Para(enc.EncodeText(para.Text()))

	if s.Principal.BestBirthlikeEvent != nil {
		if s.IncludeMedia {
			if len(s.Principal.BestBirthlikeEvent.GetMediaObjects()) > 0 {
				// TODO: handle error
				MediaObjectsAsFigures(s.Principal.BestBirthlikeEvent.GetMediaObjects(), enc, s.CropMediaHighlights)
			}
			if len(s.Baptisms) == 1 && s.Baptisms[0] != s.Principal.BestBirthlikeEvent && len(s.Baptisms[0].GetMediaObjects()) > 0 {
				// TODO: handle error
				MediaObjectsAsFigures(s.Baptisms[0].GetMediaObjects(), enc, s.CropMediaHighlights)
			}
		}
		if len(s.Baptisms) > 1 {
			s.renderMultipleBaptisms(enc, nc, intro, seq)
		}
	}
}

func (s *IntroStatement[T]) maybeAppendNickname(para *text.Para) {
	if s.Principal.NickName != "" {
		para.Continue("(known as " + s.Principal.NickName + ")")
	}
}

// addPossibleBirths adds a sentence giving the unconfirmed candidates for the person's birth.
// owner is the possessive used for the birth, such as "Henry's" or "his".
func (s *IntroStatement[T]) addPossibleBirths(para *text.Para, owner string, enc render.ContentBuilder[T], nc NameChooser, intro *IntroGenerator[T]) {
	list := []string{}
	hasCitation := false
	for _, ev := range s.PossibleBirths {
		list = append(list, enc.EncodeWithCitations(enc.EncodeText(EventWhenWherePov(ev, enc, nc, intro.POV)), ev.GetCitations()).String())
		if len(ev.GetCitations()) > 0 {
			hasCitation = true
		}
	}
	para.StartSentence("no record of", owner, "birth has been confirmed, but")
	if hasCitation {
		para.Continue("there is some evidence that", s.Principal.Gender.SubjectPronounWithLink(), "born")
	} else {
		para.Continue(s.Principal.Gender.SubjectPronoun(), "may have been born")
	}
	para.FinishSentence(text.JoinListOr(list))
}

// addParentage adds the person's parents to the current sentence. When asClause is true the
// sentence already describes the birth, so the parentage is set off with a comma.
func (s *IntroStatement[T]) addParentage(para *text.Para, enc render.ContentBuilder[T], nc NameChooser, intro *IntroGenerator[T], seq int, asClause bool) {
	rel := PositionInFamily(s.Principal)
	if rel == "" {
		rel = text.LowerFirst(s.Principal.Gender.RelationToParentNoun())
	}
	prefix := "the " + rel + " of "

	add := para.Continue
	if asClause {
		add = para.AppendClause
	}

	switch {
	case s.Principal.Father.IsUnknown() && s.Principal.Mother.IsUnknown():
		para.StartSentence(s.Principal.Gender.PossessivePronounSingular(), "parents are not known")
	case s.Principal.Father.IsUnknown():
		add(prefix + intro.IntroducePerson(seq, s.Principal.Mother, s.Start(), false, enc, nc))
		para.StartSentence(s.Principal.Gender.PossessivePronounSingular(), "father is not known")
	case s.Principal.Mother.IsUnknown():
		add(prefix + intro.IntroducePerson(seq, s.Principal.Father, s.Start(), false, enc, nc))
		para.StartSentence(s.Principal.Gender.PossessivePronounSingular(), "mother is not known")
	default:
		add(prefix + intro.IntroducePerson(seq, s.Principal.Father, s.Start(), false, enc, nc) + " and " + intro.IntroducePerson(seq, s.Principal.Mother, s.Start(), false, enc, nc))
	}
}

func (s *IntroStatement[T]) maybeAddTwinAssociation(para *text.Para, enc render.ContentBuilder[T]) {
	for _, as := range s.Principal.Associations {
		if as.Kind != model.AssociationKindTwin {
			continue
		}
		addTwinSentence(para, s.Principal, as, enc)
		return
	}
}

// addTwinSentence starts a sentence saying p was a twin, naming the twin when they can be named.
func addTwinSentence[T render.EncodedText](para *text.Para, p *model.Person, as model.Association, enc render.TextEncoder[T]) {
	if !as.Other.CanBeNamed() {
		para.StartSentence(p.Gender.SubjectPronoun(), enc.EncodeWithCitations(enc.EncodeText("was a twin"), as.Citations).String())
		return
	}
	twinLink := enc.EncodeModelLink(enc.EncodeText(as.Other.PreferredFamiliarName), as.Other)
	para.StartSentence(p.Gender.SubjectPronoun(), "was the twin to", enc.EncodeWithCitations(twinLink, as.Citations).String())
}

func (s *IntroStatement[T]) maybeAddBaptismAfterBirth(para *text.Para, enc render.ContentBuilder[T], nc NameChooser, intro *IntroGenerator[T]) {
	if len(s.Baptisms) != 1 || s.Baptisms[0] == s.Principal.BestBirthlikeEvent {
		return
	}
	bapDetail := AgeWhenWherePov(s.Baptisms[0], enc, nc, intro.POV, 2)
	if bapDetail == "" {
		return
	}
	bapCited := enc.EncodeWithCitations(enc.EncodeText(bapDetail), s.Baptisms[0].GetCitations()).String()
	para.StartSentence(s.Principal.Gender.SubjectPronoun(), "was baptised", bapCited)
}

func (s *IntroStatement[T]) renderMultipleBaptisms(enc render.ContentBuilder[T], nc NameChooser, intro *IntroGenerator[T], seq int) {
	var para text.Para
	para.StartSentence(intro.Pronoun(seq, s.Start(), s.Principal))
	for i, bev := range s.Baptisms {
		logging.Debug("adding baptism event to narrative intro statement", "id", s.Principal.ID, "bev", bev, "BestBirthlikeEvent", s.Principal.BestBirthlikeEvent)
		if s.Baptisms[i] == s.Principal.BestBirthlikeEvent {
			continue
		}
		aww := AgeWhenWherePov(bev, enc, nc, intro.POV, 2)
		if aww == "" {
			continue
		}
		cited := enc.EncodeWithCitations(enc.EncodeText(aww), bev.GetCitations()).String()
		if i == 0 {
			para.Continue("was baptised", cited)
		} else {
			para.Continue("and again", cited)
		}
	}
	enc.Para(enc.EncodeText(para.Text()))
}

func (s *IntroStatement[T]) maybeAddRelationToKeyPerson(para *text.Para, enc render.ContentBuilder[T]) {
	// ---------------------------------------
	// Prose relation to key person
	// ---------------------------------------
	if !s.SuppressRelation {
		if s.Principal.RelationToKeyPerson != nil && !s.Principal.RelationToKeyPerson.IsSelf() {
			para.AddCompleteSentence(
				s.Principal.Gender.SubjectPronoun(),
				"is",
				enc.EncodeModelLink(enc.EncodeText(text.MaybePossessiveSuffix(s.Principal.RelationToKeyPerson.From.PreferredFamiliarName)), s.Principal.RelationToKeyPerson.From).String(),
				s.Principal.RelationToKeyPerson.Name(),
			)
		}
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
	if other.IsUnknown() || s.Family.Bond == model.FamilyBondUnmarried || s.Family.Bond == model.FamilyBondLikelyUnmarried {
		s.renderSingleParent(seq, intro, enc, nc)
		return
	}
	s.renderMarried(seq, intro, enc, nc)
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

func (s *FamilyStatement[T]) renderMarried(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	other := s.Family.OtherParent(s.Principal)

	// a redacted partner is not named: "He married in 1903." or "She had another partner before 1935."
	var otherName string
	if !other.Redacted {
		otherName = intro.IntroducePerson(seq, other, s.Start(), false, enc, nc)
	}
	hasOtherFamilies := len(s.Principal.Families) > 1

	var detail text.Para
	detail.StartSentence(text.UpperFirst(intro.Default(seq, s.Start())))

	var action string
	switch s.Family.Bond {
	case model.FamilyBondMarried:
		action = "married"
	case model.FamilyBondLikelyMarried:
		action = ChooseFrom(seq, "likely married", "probably married")
	default:
		action = "met"
		if other.Redacted {
			action = "had a partner"
			if hasOtherFamilies {
				action = "had another partner"
			}
		}
	}

	// a relationship with no start event and no children can only be dated by a death, which
	// says nothing about when it began: "He had a relationship with Hannah Gibbard."
	startDate := s.Family.BestStartDate
	undated := action == "met" && s.Family.BestStartEvent == nil && len(s.Family.Children) == 0
	if undated {
		startDate = nil
		action = "had a relationship with"
		if hasOtherFamilies {
			action = "also had a relationship with"
		}
	}

	var event string
	if undated {
		detail.ReplaceSentence(intro.Pronoun(seq, s.Start(), s.Principal))
		event = action + " " + otherName
	} else if !startDate.IsUnknown() {
		detail.ReplaceSentence(intro.Pronoun(seq, s.Start(), s.Principal))
		event = text.JoinSentenceParts(action, otherName, NamedWhen(startDate))
		if age, ok := s.Principal.AgeInYearsAt(startDate); ok && (age < 18 || age > 45) {
			event += " " + AgeQualifier(age)
		}
	} else if other.Redacted {
		detail.ReplaceSentence(intro.Pronoun(seq, s.Start(), s.Principal))
		event = action
		if hasOtherFamilies && s.Family.Bond == model.FamilyBondMarried {
			event = "also " + action
		}
	} else {
		event = action + " " + otherName
	}
	if s.Family.BestStartEvent != nil && !s.Family.BestStartEvent.GetPlace().IsUnknown() {
		event = WhatWherePov(event, s.Family.BestStartEvent.GetPlace(), enc, nc, intro.POV)
	}
	if s.Family.BestStartEvent != nil {
		detail.Continue(enc.EncodeWithCitations(enc.EncodeText(event), s.Family.BestStartEvent.GetCitations()).String())
	} else {
		detail.Continue(event)
	}

	if len(s.Family.Children) == 0 {
		if s.Principal.Childless {
			detail.AddCompleteSentence("they had no children")
		}
	} else {
		childCardinal := ChildCardinal(s.Family.Children)
		switch len(s.Family.Children) {
		case 1:
			continueOrStartSentence(&detail, ChooseFrom(seq,
				" and had "+childCardinal+":",
				". They had just one child together:",
				". They had "+childCardinal+":",
			))
		case 2:
			continueOrStartSentence(&detail, ChooseFrom(seq,
				" and had "+childCardinal+":",
				". They had "+childCardinal+": ",
			))
		default:
			continueOrStartSentence(&detail, ChooseFrom(seq,
				". They had "+childCardinal+": ",
				" and went on to have "+childCardinal+" with "+s.Principal.Gender.Opposite().ObjectPronoun()+": ",
				". They went on to have "+childCardinal+": ",
			))
		}
	}

	s.renderChildListParagraph(&detail, enc, nc, s.Family.Children)
}

func (s *FamilyStatement[T]) renderSingleParent(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	if len(s.Family.Children) == 0 {
		return
	}

	other := s.Family.OtherParent(s.Principal)
	otherUnknown := other.IsUnknown()
	otherRedacted := !otherUnknown && other.Redacted
	otherKnown := !otherUnknown && !otherRedacted
	var otherName string
	if otherKnown {
		otherName = intro.IntroducePerson(seq, other, s.Start(), false, enc, nc)
	}

	// a redacted partner is not named; the children are said to be by another relationship
	// when the principal had other families
	var byOther string
	if otherRedacted && len(s.Principal.Families) > 1 {
		byOther = "by another relationship"
	}

	c := s.Family.Children[0]
	oneChild := len(s.Family.Children) == 1
	isMother := s.Principal.Gender == model.GenderFemale
	childFirmBirthdate := isFirmBirthdateEvent(c.BestBirthlikeEvent)

	var detail text.Para

	switch {
	case oneChild && c.Redacted:
		// a redacted child is not named or dated: "had a daughter with an unknown father"
		detail.Continue(intro.Pronoun(seq, s.Start(), s.Principal))
		detail.Continue("had a", c.Gender.RelationToParentNoun())
		switch {
		case otherKnown:
			detail.Continue("with", otherName)
		case otherUnknown:
			detail.Continue("with an unknown " + s.Principal.Gender.Opposite().RelationToChildrenNoun())
		default:
			detail.Continue(byOther)
		}

	case oneChild && otherUnknown && c.BestBirthlikeEvent == nil:
		// only the parent-child link is known: state the relationship and leave the child to the list
		detail.Continue(intro.PossessivePronoun(seq, s.Start(), s.Principal), "only known child was a", c.Gender.RelationToParentNoun())

	case oneChild && isMother && childFirmBirthdate && !otherRedacted:
		// mother with firm birthdate: "gave birth to a [gender]..."
		if c.BestBirthlikeEvent != nil {
			detail.AppendAsAside(intro.RelativeTime(seq, c.BestBirthlikeEvent.GetDate(), otherKnown))
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))
		}
		detail.Continue("gave birth to a", c.Gender.RelationToParentNoun())
		if otherKnown {
			// the child is named in the list that follows
			detail.AppendClause("the child of")
			detail.Continue(otherName)
		} else {
			detail.AppendAsAside(enc.EncodeModelLink(enc.EncodeText(c.PreferredFullName), c).String())
			detail.Continue(enc.EncodeWithCitations(enc.EncodeText(EventWhenWherePov(c.BestBirthlikeEvent, enc, nc, intro.POV)), c.BestBirthlikeEvent.GetCitations()).String())
			if detail.CurrentSentenceLength() < 60 {
				detail.Continue(ChooseFrom(seq,
					"with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
					"by an unknown "+s.Principal.Gender.Opposite().Noun(),
				))
			}
		}

	case oneChild && isMother && otherUnknown:
		// unknown partner, non-firm birthdate: "had a [gender] [name] who [what/when/where]"
		if c.BestBirthlikeEvent != nil {
			detail.AppendAsAside(intro.RelativeTime(seq, c.BestBirthlikeEvent.GetDate(), false))
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))
		}
		detail.Continue("had a", c.Gender.RelationToParentNoun())
		detail.AppendAsAside(enc.EncodeModelLink(enc.EncodeText(c.PreferredFullName), c).String())
		if !c.BestBirthlikeEvent.GetDate().IsUnknown() || !c.BestBirthlikeEvent.GetPlace().IsUnknown() {
			detail.Continue("who")
			detail.Continue(enc.EncodeWithCitations(enc.EncodeText(EventWhatWhenWherePov(c.BestBirthlikeEvent, enc, nc, intro.POV)), c.BestBirthlikeEvent.GetCitations()).String())
		}
		if detail.CurrentSentenceLength() < 60 {
			detail.Continue(ChooseFrom(seq,
				"with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
				"by an unknown "+s.Principal.Gender.Opposite().Noun(),
			))
		}

	case oneChild:
		// general single child: "had a child/[gender]... [with otherName if known]"
		if c.BestBirthlikeEvent != nil {
			detail.AppendAsAside(intro.RelativeTime(seq, c.BestBirthlikeEvent.GetDate(), childFirmBirthdate))
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))
		} else {
			detail.Continue(intro.Pronoun(seq, s.Start(), s.Principal))
		}
		detail.Continue(ChooseFrom(seq,
			"had a child",
			"had a "+c.Gender.RelationToParentNoun(),
		))
		// the child is named in the list that follows
		if otherKnown {
			detail.Continue("with", otherName)
		} else {
			detail.Continue(byOther)
		}

	default:
		// multiple children
		if c.BestBirthlikeEvent != nil {
			detail.Continue(intro.Pronoun(seq, c.BestBirthlikeEvent.GetDate(), s.Principal))
		} else {
			detail.Continue(intro.Pronoun(seq, s.Start(), s.Principal))
		}
		detail.Continue("had", ChildCardinal(s.Family.Children))
		if otherKnown {
			detail.Continue("with", otherName)
		} else {
			detail.Continue(byOther)
		}
	}

	if oneChild {
		s.renderSingleChildParagraph(&detail, c, enc, nc)
	} else {
		s.renderChildListParagraph(&detail, enc, nc, s.Family.Children)
	}
}

func (s *FamilyStatement[T]) renderSingleChildParagraph(detail *text.Para, c *model.Person, enc render.ContentBuilder[T], nc NameChooser) {
	if c.Redacted {
		enc.Para(enc.EncodeText(detail.Text()))
		return
	}
	detail.FinishSentenceWithTerminator(":–")
	enc.Para(enc.EncodeText(detail.Text()))
	enc.UnorderedList([]T{PersonSummary(c, enc, nc, enc.EncodeText(c.PreferredFamiliarName), false, false, false, true, true)})
}

func (s *FamilyStatement[T]) renderChildListParagraph(detail *text.Para, enc render.ContentBuilder[T], nc NameChooser, children []*model.Person) {
	childList, shared := ChildList(children, enc, nc)
	if len(childList) == 0 {
		enc.Para(enc.EncodeText(detail.Text()))
		return
	}
	addSharedClause(detail, shared)
	detail.FinishSentenceWithTerminator(":–")
	enc.Para(enc.EncodeText(detail.Text()))
	enc.UnorderedList(childList)
	if other := OtherChildrenSentence(s.Family); other != "" {
		enc.Para(enc.EncodeText(other))
	}
}

// possessiveName is the possessive used for a person on their own page, such as "Henry's", or a
// pronoun such as "her" when neither their forename nor surname is known.
func possessiveName(p *model.Person) string {
	if p.Unidentified {
		return p.Gender.PossessivePronounSingular()
	}
	return text.MaybePossessiveSuffix(p.PreferredGivenName)
}

// continueOrStartSentence starts a new sentence when s begins with ". ", otherwise it continues
// the current sentence.
func continueOrStartSentence(para *text.Para, s string) {
	if rest, ok := strings.CutPrefix(strings.TrimSpace(s), ". "); ok {
		para.StartSentence(rest)
		return
	}
	para.Continue(s)
}

// isFirmBirthdateEvent reports whether ev is a BirthEvent with a firm date.
func isFirmBirthdateEvent(ev model.IndividualTimelineEvent) bool {
	if _, ok := ev.(*model.BirthEvent); !ok {
		return false
	}
	return ev.GetDate().IsFirm()
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

// ChildList returns a summary of each child that is not redacted, in order of birth, and a
// clause for the sentence introducing the list that states what most of them share, such as
// "all baptised at X (unless noted)". The summaries leave out what the clause states.
func ChildList[T render.EncodedText](clist []*model.Person, enc render.ContentBuilder[T], nc NameChooser) ([]T, string) {
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

	var listed []*model.Person
	for _, c := range clist {
		if !c.Redacted {
			listed = append(listed, c)
		}
	}

	shared, opts := sharedBirthDetails(listed, enc, nc)

	childList := make([]T, 0, len(listed))
	for _, c := range listed {
		childSummary := PersonSummary(c, enc, nc, enc.EncodeText(c.PreferredGivenName), true, false, true, true, true, opts...)
		if !childSummary.IsZero() {
			childList = append(childList, childSummary)
		}
	}

	return childList, shared
}

// minSharedChildren is the fewest children that must share a birth place or estimated birth
// year, and they must be more than half of the list, for it to be stated once for the list.
const minSharedChildren = 3

// sharedBirthDetails finds the place of birth or baptism and the estimated birth dates shared by
// most of the children. It returns a clause stating them, such as "all baptised at X, with
// birth years estimated (unless noted)", and the options that leave them out of each child's
// summary.
func sharedBirthDetails[T render.EncodedText](children []*model.Person, enc render.ContentBuilder[T], nc NameChooser) (string, []SummaryOption) {
	isShared := func(n int) bool { return n >= minSharedChildren && n*2 > len(children) }

	placeCounts := make(map[string]int)
	places := make(map[string]*model.Place)
	estimated := 0
	for _, c := range children {
		if c.BestBirthlikeEvent == nil {
			continue
		}
		bev, _ := SummaryBirthEvent(c, true)
		if bev == nil {
			continue
		}
		if pl := bev.GetPlace(); !pl.IsUnknown() {
			placeCounts[pl.ID]++
			places[pl.ID] = pl
		}
		if !bev.IsInferred() && bev.GetDate().Derivation == model.DateDerivationEstimated {
			estimated++
		}
	}

	var sharedPlace *model.Place
	best := 0
	for id, n := range placeCounts {
		if n > best || (n == best && id < sharedPlace.ID) {
			best = n
			sharedPlace = places[id]
		}
	}
	if !isShared(best) {
		sharedPlace = nil
	}

	var clauses []string
	var opts []SummaryOption
	sharedByAll := true // whether every child shares what the clauses state
	if sharedPlace != nil {
		clauses = append(clauses, WhatWhere("all "+sharedBirthVerb(children, sharedPlace), sharedPlace, enc, nc))
		opts = append(opts, WithKnownBirthPlace(sharedPlace))
		sharedByAll = best == len(children)
	}
	if isShared(estimated) {
		clauses = append(clauses, "with birth years estimated")
		opts = append(opts, WithKnownEstimatedBirths())
		sharedByAll = sharedByAll && estimated == len(children)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	clause := strings.Join(clauses, ", ")
	if !sharedByAll {
		clause += " (unless noted)"
	}
	return clause, opts
}

// sharedBirthVerb is "baptised" when every child whose birth is summarised at pl was baptised
// there, "born" when every one was born there, and "born or baptised" otherwise.
func sharedBirthVerb(children []*model.Person, pl *model.Place) string {
	var born, baptised bool
	for _, c := range children {
		if c.BestBirthlikeEvent == nil {
			continue
		}
		bev, _ := SummaryBirthEvent(c, true)
		if bev == nil || !bev.GetPlace().SameAs(pl) {
			continue
		}
		if _, ok := bev.(*model.BaptismEvent); ok {
			baptised = true
		} else {
			born = true
		}
	}
	switch {
	case baptised && !born:
		return "baptised"
	case born && !baptised:
		return "born"
	default:
		return "born or baptised"
	}
}

// addSharedClause adds the clause stating what most children in a list share to the sentence
// introducing the list, before its colon: "They had nine children, all baptised at X (unless
// noted):".
func addSharedClause(detail *text.Para, shared string) {
	if shared == "" {
		return
	}
	detail.TrimCurrentRight(":")
	detail.AppendClause(shared)
}

// OtherChildrenSentence counts the children of f that are left out of a child list because they
// are redacted, such as "They also had two other children." It is empty when there are none.
func OtherChildrenSentence(f *model.Family) string {
	n := 0
	for _, c := range f.Children {
		if c.Redacted {
			n++
		}
	}
	if n == 0 {
		return ""
	}

	subject := "they"
	switch {
	case f.Father.IsUnknown() && !f.Mother.IsUnknown():
		subject = f.Mother.Gender.SubjectPronoun()
	case f.Mother.IsUnknown() && !f.Father.IsUnknown():
		subject = f.Father.Gender.SubjectPronoun()
	}

	var para text.Para
	para.AddCompleteSentence(subject, "also had", text.CardinalWithUnit(n, "other child", "other children"))
	return para.Text()
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
		detail.StartSentence(s.Principal.PreferredFamiliarName, "and", other.PreferredFamiliarName, "divorced", endDate.When())
	case model.FamilyEndReasonDeath:
		detail.StartSentence(PersonDeathSummary(other, enc, nc, enc.EncodeText(s.otherPartyName(other, enc)), true, false, true, true).String())
		// if (other.BestDeathlikeEvent != nil && !other.BestDeathlikeEvent.IsInferred()) && (s.Family.Bond == model.FamilyBondMarried || s.Family.Bond == model.FamilyBondLikelyMarried) {
		// 	detail.StartSentence(s.Principal.PreferredFamiliarName, "was left a", s.Principal.Gender.WidowWidower())
		// }
	case model.FamilyEndReasonUnknown:
		// TODO: format FamilyEndReasonUnknown
		end += "the marriage ended " + endDate.When()
	}

	if end != "" {
		detail.Continue(end)
	}

	enc.Para(enc.EncodeText(detail.Text()))
}

// otherPartyName names the other party of the family for a sentence about their death: "Frances,
// his wife," for a married or likely married couple, otherwise by the children they had, as in
// "Annie Louisa Tew, the mother of his son Bertie,", or by name alone when they had none.
func (s *FamilyEndStatement[T]) otherPartyName(other *model.Person, enc render.ContentBuilder[T]) string {
	if s.Family.Bond == model.FamilyBondMarried || s.Family.Bond == model.FamilyBondLikelyMarried {
		label := s.Principal.Gender.PossessivePronounSingular() + " " + other.Gender.RelationToSpouseNoun()
		if other.Unidentified {
			return label
		}
		return other.PreferredFamiliarName + ", " + label + ", "
	}

	var label string
	switch len(s.Family.Children) {
	case 0:
	case 1:
		c := s.Family.Children[0]
		label = "the " + other.Gender.RelationToChildrenNoun() + " of " + s.Principal.Gender.PossessivePronounSingular() + " " + c.Gender.RelationToParentNoun()
		if c.CanBeNamed() {
			label += " " + enc.EncodeModelLink(enc.EncodeText(c.PreferredFamiliarName), c).String()
		}
	default:
		label = "the " + other.Gender.RelationToChildrenNoun() + " of " + s.Principal.Gender.PossessivePronounSingular() + " children"
	}

	switch {
	case other.Unidentified && label == "":
		return other.UnknownNameDescription()
	case other.Unidentified:
		return label
	case label == "":
		return other.PreferredFullName
	default:
		return other.PreferredFullName + ", " + label + ", "
	}
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
	PossibleDeaths          []*model.PossibleDeathEvent
	ExcludeSurvivingPartner bool
	IncludeMedia            bool
	CropMediaHighlights     bool
}

var _ Statement[md.Text] = (*DeathStatement[md.Text])(nil)

func (s *DeathStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	if s.Principal.BestDeathlikeEvent == nil {
		_ = s.RenderDetailUnknown(seq, intro, enc, nc)
		return
	}

	bev := s.Principal.BestDeathlikeEvent

	evDetail := DeathWhat(bev, s.Principal.ModeOfDeath)

	if !bev.GetDate().IsUnknown() {
		afterBirth := ""
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
						afterBirth = ageDetail + " after " + s.Principal.Gender.SubjectPronoun() + " was born"
						ageDetail = ""
					}
				}
			}

			if ageDetail == "" && afterBirth == "" {
				ageDetail = AgeQualifier(age)
			}
			if ageDetail != "" {
				evDetail += " " + ageDetail
			}
		}
		evDetail += " " + bev.GetDate().When()
		if afterBirth != "" {
			// set off after the date so the date is not read as the date of birth
			evDetail += ", " + afterBirth
			if !bev.GetPlace().IsUnknown() {
				evDetail += ","
			}
		}
	} else {
		evDetail += " on an unknown date"
	}
	if !bev.GetPlace().IsUnknown() {
		evDetail = WhatWherePov(evDetail, bev.GetPlace(), enc, nc, intro.POV)
	}

	burialRunOnSentence := true

	var para text.Para
	para.StartSentence(s.Principal.PreferredFamiliarName, enc.EncodeWithCitations(enc.EncodeText(evDetail), bev.GetCitations()).String())

	if s.Principal.CauseOfDeath != nil {
		para.StartSentence(
			s.Principal.Gender.PossessivePronounSingular(),
			"death was attributed to",
			enc.EncodeWithCitations(enc.EncodeText(s.Principal.CauseOfDeath.Detail), s.Principal.CauseOfDeath.Citations).String(),
		)
		burialRunOnSentence = false
	}

	additionalDetailFromDeathEvent := EventNarrativeDetail(bev, enc)
	if additionalDetailFromDeathEvent != "" {
		para.StartSentence(additionalDetailFromDeathEvent)
		burialRunOnSentence = false
	}

	if !burialRunOnSentence {
		enc.Para(enc.EncodeText(para.Text()))
		para = text.Para{}
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

		funeralEvent := funerals[0]
		funeralEvDetail := ""
		switch funeralEvent.(type) {
		case *model.BurialEvent:
			funeralEvDetail = "buried"
		case *model.CremationEvent:
			funeralEvDetail = "cremated"
		default:
			panic("unhandled funeral event")
		}

		interval := bev.GetDate().IntervalUntil(funeralEvent.GetDate())
		if days, ok := interval.ApproxDays(); ok && days < 15 {
			switch days {
			case 0:
				funeralEvDetail += " the same day"
			case 1:
				funeralEvDetail += " the next day"
			default:
				funeralEvDetail += fmt.Sprintf(" %s days later", text.CardinalNoun(days))
			}
		} else if !funeralEvent.GetDate().IsUnknown() {
			funeralEvDetail += " " + funeralEvent.GetDate().When()
		}
		if !funeralEvent.GetPlace().IsUnknown() {
			funeralEvDetail = WhatWherePov(funeralEvDetail, funeralEvent.GetPlace(), enc, nc, intro.POV)
		}

		funeralCited := enc.EncodeWithCitations(enc.EncodeText(funeralEvDetail), funeralEvent.GetCitations()).String()
		if para.IsEmpty() {
			para.StartSentence(text.UpperFirst(s.Principal.Gender.SubjectPronounWithLink()), funeralCited)
		} else {
			para.Continue("and was", funeralCited)
		}
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
					para.Continue(text.UpperFirst(s.Principal.Gender.SubjectPronounWithLink()), "survived by")
					if lastFamily.Bond == model.FamilyBondMarried {
						para.Continue(text.LowerFirst(s.Principal.Gender.PossessivePronounSingular()), text.LowerFirst(possibleSurvivor.Gender.RelationToSpouseNoun()))
					}
					para.Continue(intro.IntroducePerson(seq, possibleSurvivor, s.Start(), false, enc, nc))
				}
			}
		}
	}

	if !para.IsEmpty() {
		enc.Para(enc.EncodeText(para.Text()))
	}

	if s.IncludeMedia && bev != nil {
		// TODO: handle error
		_ = MediaObjectsAsFigures(bev.GetMediaObjects(), enc, s.CropMediaHighlights)
	}
}

func (s *DeathStatement[T]) RenderDetailUnknown(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) error {
	var para text.Para
	if len(s.PossibleDeaths) > 0 {
		list := []string{}
		hasCitation := false
		for _, ev := range s.PossibleDeaths {
			list = append(list, enc.EncodeWithCitations(enc.EncodeText(EventWhenWherePov(ev, enc, nc, intro.POV)), ev.GetCitations()).String())
			if len(ev.GetCitations()) > 0 {
				hasCitation = true
			}
		}
		para.StartSentence("no record of", possessiveName(s.Principal), "death has been confirmed, but")
		if hasCitation {
			para.Continue("there is some evidence that", s.Principal.Gender.SubjectPronoun(), "died")
		} else {
			para.Continue(s.Principal.Gender.SubjectPronoun(), "may have died")
		}
		para.FinishSentence(text.JoinListOr(list))
	} else {
		para.StartSentence("the time and place of", possessiveName(s.Principal), "death is not known")
	}
	enc.Para(enc.EncodeText(para.Text()))
	return nil
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
		detail.StartSentence(intro.Pronoun(seq, s.Start(), s.Principal))
		detail.Continue(enc.EncodeWithCitations(enc.EncodeText(WhatWherePov(fmt.Sprintf("%s recorded in the %d census", intro.WasWere(s.Principal), year), s.Event.GetPlace(), enc, nc, intro.POV)), s.Event.GetCitations()).String()) // fmt.Sprintf("in the %d census", year)
		detail.StartSentence(narrative)
		detail.FinishSentence()
		enc.Para(enc.EncodeText(detail.Text()))
		return
	}
	// TODO: construct narrative of census

	var detail text.Para
	what := ChooseFrom(seq,
		fmt.Sprintf("%s %s recorded in the %d census", intro.Pronoun(seq, s.Start(), s.Principal), intro.WasWere(s.Principal), year),
		fmt.Sprintf("by the time of the %d census %s %s living", year, intro.Pronoun(seq, s.Start(), s.Principal), intro.WasWere(s.Principal)),
		fmt.Sprintf("in the %d census %s %s living", year, intro.Pronoun(seq, s.Start(), s.Principal), intro.WasWere(s.Principal)),
	)

	detail.StartSentence(enc.EncodeWithCitations(enc.EncodeText(WhatWherePov(what, s.Event.GetPlace(), enc, nc, intro.POV)), s.Event.GetCitations()).String()) // fmt.Sprintf("in the %d census", year)

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

// -------------------------------------------------------------------------------------

// A GeneralEventStatement is used for any general event that may include a narrative.
// If the Event is an IndividualNarrativeEvent then the narrative field is used in
// place of any generated text. Otherwise an introductory sentence is prepended followed
// by the event detail. Any additional narrative text is appended as a new sentence.
type GeneralEventStatement[T render.EncodedText] struct {
	Principal *model.Person
	Event     model.TimelineEvent
}

var _ Statement[md.Text] = (*GeneralEventStatement[md.Text])(nil)

func (s *GeneralEventStatement[T]) RenderDetail(seq int, intro *IntroGenerator[T], enc render.ContentBuilder[T], nc NameChooser) {
	var detail text.Para
	switch s.Event.(type) {
	case *model.IndividualNarrativeEvent:
	default:
		// prepend an intro
		detail.StartSentence(intro.Pronoun(seq, s.Start(), s.Principal))
		detail.Continue(enc.EncodeWithCitations(enc.EncodeText(EventWhatWhenWherePov(s.Event, enc, nc, intro.POV)), s.Event.GetCitations()).String())
	}

	narrative := EventNarrativeDetail(s.Event, enc)
	if narrative != "" {
		detail.StartSentence(narrative)
	}
	detail.FinishSentence()

	// enc.ParaWithFigure(enc.EncodeWithCitations(detail.Text(), s.Event.GetCitations()), "/trees/cg/media/6V7KWAJR2LCVK.png", "alt text", "this is a caption")
	enc.Para(enc.EncodeWithCitations(enc.EncodeText(detail.Text()), s.Event.GetCitations()))
}

func (s *GeneralEventStatement[T]) Start() *model.Date {
	return s.Event.GetDate()
}

func (s *GeneralEventStatement[T]) End() *model.Date {
	return s.Event.GetDate()
}

func (s *GeneralEventStatement[T]) NarrativeSequence() int {
	if !s.Event.GetDate().SortsBefore(s.Principal.BestDeathDate()) {
		return NarrativeSequenceDeath
	}
	return NarrativeSequenceLifeStory
}

func (s *GeneralEventStatement[T]) Priority() int {
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
			detail.StartSentence("They")
		} else {
			detail.StartSentence(s.Family.Mother.PreferredGivenName)
		}
	} else {
		if s.Family.Mother.IsUnknown() {
			detail.StartSentence(s.Family.Father.PreferredGivenName)
		} else {
			detail.StartSentence("They")
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

	childList, shared := ChildList(s.Family.Children, enc, nc)
	if len(childList) == 0 {
		enc.Para(enc.EncodeText(detail.Text()))
		return
	}

	addSharedClause(&detail, shared)
	detail.FinishSentenceWithTerminator(":–")
	enc.Para(enc.EncodeText(detail.Text()))
	enc.UnorderedList(childList)
	if other := OtherChildrenSentence(s.Family); other != "" {
		enc.Para(enc.EncodeText(other))
	}
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
		enc.Figure(mo.Object.SrcFilePath, mo.Object.Title, enc.EncodeText(mo.Object.Title), mo.Highlight, "")
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

package site

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

type Narrative struct {
	Statements []Statement
}

type NarrativeIntro struct {
	Text         string // Default text to use as an intro, composed of a combination of other fields
	NameBased    string // name or he/she
	TimeBased    string // During this time, later that year, some years later
	DateInferred bool   // whether or not the TimeBased or Text intro already includes a version of the date
}

func (n *NarrativeIntro) Default() string {
	return n.Text
	// return text.MaybeAddWithSpace(n.TimeInterval, n.Noun)
}

const (
	NarrativeSequenceIntro     = 0
	NarrativeSequenceEarlyLife = 1
	NarrativeSequenceLifeStory = 2
	NarrativeSequenceDeath     = 3
	NarrativeSequencePostDeath = 4
)

func (n *Narrative) Render(pov *model.POV, b ExtendedMarkdownBuilder) {
	sort.Slice(n.Statements, func(i, j int) bool {
		if n.Statements[i].NarrativeSequence() == n.Statements[j].NarrativeSequence() {
			return n.Statements[i].Start().SortsBefore(n.Statements[j].Start())
		}
		return n.Statements[i].NarrativeSequence() < n.Statements[j].NarrativeSequence()
	})

	currentNarrativeSequence := NarrativeSequenceIntro
	sequenceInNarrative := 0
	for i, s := range n.Statements {
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

		var nintro NarrativeIntro

		if sequenceInNarrative == 0 || pov.Person.BestDeathDate().SortsBefore(s.Start()) {
			nintro.NameBased = pov.Person.PreferredFamiliarName

			if s.NarrativeSequence() == NarrativeSequenceLifeStory {
				nintro.TimeBased = text.UpperFirst(s.Start().When()) + ", "
			}

		} else {
			name := pov.Person.PreferredFamiliarName
			if sequenceInNarrative%4 != 0 {
				name = pov.Person.Gender.SubjectPronoun()
			}
			nintro.NameBased = name

			if i > 0 {
				sincePrev := n.Statements[i-1].End().IntervalUntil(s.Start())
				if years, ok := sincePrev.WholeYears(); ok {
					dateInYear, ok := s.Start().DateInYear(true)
					if ok {
						dateInYear = "on " + dateInYear
					}

					if years < 0 && s.Start().SortsBefore(n.Statements[i-1].End()) {
						nintro.TimeBased = ""
					} else if years == 0 {
						days, isPreciseInterval := sincePrev.ApproxDays()
						if isPreciseInterval && days < 5 {
							nintro.TimeBased = ChooseFrom(sequenceInNarrative,
								dateInYear,
								text.AppendAside(dateInYear, "just a few days later"),
								text.AppendAside("Very shortly after", dateInYear),
								text.AppendAside("Just a few days later", dateInYear),
							)
						} else if isPreciseInterval && days < 20 {
							nintro.TimeBased = ChooseFrom(sequenceInNarrative,
								text.AppendAside("Shortly after", dateInYear),
								text.AppendAside("Several days later", dateInYear),
							)
						} else if n.Statements[i-1].End().SameYear(s.Start()) {
							nintro.TimeBased = ChooseFrom(sequenceInNarrative,
								text.AppendAside("Later that year", dateInYear),
								text.AppendAside("The same year", dateInYear),
								text.AppendAside("Later that same year", dateInYear),
								text.AppendAside("That same year", dateInYear),
							)
						} else {
							nintro.TimeBased = ChooseFrom(sequenceInNarrative,
								text.AppendAside("Shortly after", dateInYear),
								text.AppendAside("Some time later", dateInYear),
								text.AppendAside("A short while later", dateInYear),
							)
						}

					} else if years == 1 {
						nintro.TimeBased = ChooseFrom(sequenceInNarrative,
							text.AppendAside("The next year", dateInYear),
							text.AppendAside("The following year", dateInYear),
							"",
						)
					} else if years < 5 {
						nintro.TimeBased = ChooseFrom(sequenceInNarrative,
							s.Start().When(),
							"",
							text.AppendClause("A few years later", s.Start().When()),
							text.AppendClause("Some years later", s.Start().When()),
							"",
						)
					} else {
						nintro.TimeBased = ChooseFrom(sequenceInNarrative,
							"",
							text.AppendClause("Several years later", s.Start().When()),
						)
						if nintro.TimeBased != "" {
							nintro.TimeBased += ", "
						}
					}
				}
			}
		}

		nintro.Text = nintro.TimeBased
		nintro.DateInferred = true
		if nintro.Text == "" {
			// nintro.Text = ChooseFrom(sequenceInNarrative, "", "", "then ")
			nintro.Text += nintro.NameBased
			nintro.DateInferred = false
		} else if nintro.NameBased != "" {
			nintro.Text = text.AppendClause(nintro.Text, nintro.NameBased)
		}
		nintro.Text = text.UpperFirst(nintro.Text)
		nintro.NameBased = nintro.NameBased
		nintro.TimeBased = nintro.TimeBased

		s.RenderDetail(sequenceInNarrative, &nintro, b, nil)
		b.EmptyPara()

		sequenceInNarrative++

	}
}

type GrammarHints struct {
	DateInferred bool
}

type Statement interface {
	RenderDetail(int, *NarrativeIntro, ExtendedMarkdownBuilder, *GrammarHints)
	Start() *model.Date
	End() *model.Date
	NarrativeSequence() int
}

type IntroStatement struct {
	Principal        *model.Person
	Baptisms         []*model.BaptismEvent
	SuppressRelation bool
}

var _ Statement = (*IntroStatement)(nil)

func (s *IntroStatement) RenderDetail(seq int, intro *NarrativeIntro, enc ExtendedMarkdownBuilder, hints *GrammarHints) {
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
	parentageDetailPrefix := "the " + text.LowerFirst(s.Principal.Gender.RelationToParentNoun()) + " of "
	if s.Principal.Father.IsUnknown() {
		if s.Principal.Mother.IsUnknown() {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " parents are not known"
		} else {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " father is not known"
			parentDetail = parentageDetailPrefix + enc.EncodeModelLinkDedupe(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother.PreferredFamiliarName, s.Principal.Mother)
		}
	} else {
		if s.Principal.Mother.IsUnknown() {
			parentUnknownDetail = s.Principal.Gender.PossessivePronounSingular() + " mother is not known"
			parentDetail = parentageDetailPrefix + enc.EncodeModelLinkDedupe(s.Principal.Father.PreferredUniqueName, s.Principal.Father.PreferredFamiliarName, s.Principal.Father)
		} else {
			parentDetail = parentageDetailPrefix + enc.EncodeModelLinkDedupe(s.Principal.Father.PreferredUniqueName, s.Principal.Father.PreferredFamiliarName, s.Principal.Father) + " and " + enc.EncodeModelLinkDedupe(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother.PreferredFamiliarName, s.Principal.Mother)
		}
	}

	// ---------------------------------------
	// Build detail
	// ---------------------------------------
	detail := s.Principal.PreferredGivenName

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
	detail = text.FinishSentence(detail)

	if parentUnknownDetail != "" {
		detail = text.JoinSentences(detail, parentUnknownDetail)
		detail = text.FinishSentence(detail)
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
			detail += " " + text.UpperFirst(s.Principal.Gender.SubjectPronoun()) + " is the " + s.Principal.RelationToKeyPerson.Name() + " of " + enc.EncodeModelLinkDedupe(s.Principal.RelationToKeyPerson.From.PreferredFamiliarFullName, s.Principal.RelationToKeyPerson.From.PreferredFamiliarName, s.Principal.RelationToKeyPerson.From)
		}
	}

	detail = text.FinishSentence(detail)
	enc.Para(detail)

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
		bapDetail = text.FinishSentence(text.JoinSentenceParts(intro.NameBased, bapDetail))
		enc.Para(bapDetail)
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

type FamilyStatement struct {
	Principal *model.Person
	Family    *model.Family
}

var _ Statement = (*FamilyStatement)(nil)

func (s *FamilyStatement) RenderDetail(seq int, intro *NarrativeIntro, enc ExtendedMarkdownBuilder, hints *GrammarHints) {
	// TODO: note for example VFA3VQS22ZHBO George Henry Chambers (1903-1985) who
	// had a child with Dorothy Youngs in 1944 but didn't marry until 1985

	detail := intro.Text

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
	other := s.Family.OtherParent(s.Principal)
	if other.IsUnknown() {
		otherName = "an unknown " + s.Principal.Gender.Opposite().Noun()
	} else {
		otherName = enc.EncodeModelLinkDedupe(other.PreferredUniqueName, other.PreferredFamiliarFullName, other)
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
			event += " " + action
			event += " " + otherName
			if !intro.DateInferred {
				event += " " + startDate.When()
			}
			if age, ok := s.Principal.AgeInYearsAt(startDate); ok && age < 18 || age > 45 {
				event += " " + AgeQualifier(age)
			}
		} else {
			event += " " + action
			event += " " + otherName
		}
		if s.Family.BestStartEvent != nil {
			detail += enc.EncodeWithCitations(event, s.Family.BestStartEvent.GetCitations())
		} else {
			detail += " " + event
		}
	}

	childList := make([]string, 0, len(s.Family.Children))

	if len(s.Family.Children) == 0 {
		// single parents already dealt with
		if s.Principal.Childless {
			detail = text.JoinSentences(detail, "they had no children")
		}
	} else {
		sort.Slice(s.Family.Children, func(i, j int) bool {
			var d1, d2 *model.Date
			if s.Family.Children[i].BestBirthlikeEvent != nil {
				d1 = s.Family.Children[i].BestBirthlikeEvent.GetDate()
			}
			if s.Family.Children[j].BestBirthlikeEvent != nil {
				d2 = s.Family.Children[j].BestBirthlikeEvent.GetDate()
			}

			return d1.SortsBefore(d2)
		})

		for _, c := range s.Family.Children {
			childList = append(childList, PersonSummary(c, enc))
		}

		// children := make([]string, len(s.Family.Children))
		// for j := range s.Family.Children {
		// 	children[j] = enc.EncodeModelLink(s.Family.Children[j].PreferredGivenName, s.Family.Children[j])
		// 	if s.Family.Children[j].BestBirthlikeEvent != nil && !s.Family.Children[j].BestBirthlikeEvent.GetDate().IsUnknown() {
		// 		children[j] += fmt.Sprintf(" (%s)", s.Family.Children[j].BestBirthlikeEvent.ShortDescription())
		// 	}
		// }

		allSameGender := true
		if s.Family.Children[0].Gender != model.GenderUnknown {
			for i := 1; i < len(s.Family.Children); i++ {
				if s.Family.Children[i].Gender != s.Family.Children[0].Gender {
					allSameGender = false
					break
				}
			}
		}

		var childCardinal string
		if allSameGender {
			if s.Family.Children[0].Gender == model.GenderMale {
				childCardinal = text.CardinalWithUnit(len(s.Family.Children), "son", "sons")
			} else {
				childCardinal = text.CardinalWithUnit(len(s.Family.Children), "daughter", "daughters")
			}
		} else {
			childCardinal = text.CardinalWithUnit(len(s.Family.Children), "child", "children")
		}

		if singleParent {
			switch len(s.Family.Children) {
			case 1:
				detail += ChooseFrom(seq,
					" had one "+childCardinal+" with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
					" had one "+childCardinal+" by an unknown "+s.Principal.Gender.Opposite().Noun(),
					" had one "+childCardinal+"",
				)
			default:
				detail += " had " + childCardinal
				detail += ChooseFrom(seq,
					"",
					", by an unknown "+s.Principal.Gender.Opposite().Noun(),
					", by an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
				)
			}
		} else {
			switch len(s.Family.Children) {
			case 1:
				detail += ChooseFrom(seq,
					" and had "+childCardinal+":",
					". They had just one child together:",
					". They had "+childCardinal+":",
				)
			case 2:
				detail += ChooseFrom(seq,
					" and had "+childCardinal+":",
					". They had "+childCardinal+": ",
				)
			default:
				detail += ChooseFrom(seq,
					". They had "+childCardinal+": ",
					" and went on to have "+childCardinal+" with "+s.Principal.Gender.Opposite().PossessivePronounSingular()+": ",
					". They went on to have "+childCardinal+": ",
				)
			}
		}

	}

	endDate := s.Family.BestEndDate
	end := ""
	if !endDate.IsUnknown() {
		switch s.Family.EndReason {
		case model.FamilyEndReasonDivorce:
			end += "they divorced " + endDate.When()
		case model.FamilyEndReasonDeath:
			if s.Family.Bond == model.FamilyBondMarried || s.Family.Bond == model.FamilyBondLikelyMarried {
				if !s.EndedWithDeathOf(s.Principal) {
					leavingWidow := ""
					if s.Principal.Gender == model.GenderMale {
						leavingWidow = " leaving him a widower"
					} else if s.Principal.Gender == model.GenderFemale {
						leavingWidow = " leaving her a widow"
					}
					end += ChooseFrom(seq,
						other.PreferredFamiliarName+" died "+endDate.When(),
						other.PreferredFamiliarName+" died "+endDate.When()+leavingWidow,
						"however, "+endDate.When()+", "+other.PreferredFamiliarName+" died",
						"however, "+endDate.When()+", "+other.PreferredFamiliarName+" died "+leavingWidow,
					)
				}
			}
		case model.FamilyEndReasonUnknown:
			// TODO: format FamilyEndReasonUnknown
			end += "the marriage ended in " + endDate.When()
		}
	}
	if end != "" {
		detail = text.JoinSentences(detail, end)
	}

	enc.Para(detail)
	enc.UnorderedList(childList)

	// TODO: note how many children survived if some died
}

func (s *FamilyStatement) Start() *model.Date {
	return s.Family.BestStartDate
}

func (s *FamilyStatement) End() *model.Date {
	return s.Family.BestEndDate
}

func (s *FamilyStatement) EndedWithDeathOf(p *model.Person) bool {
	return p.SameAs(s.Family.EndDeathPerson)
}

func (s *FamilyStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

type DeathStatement struct {
	Principal *model.Person
}

var _ Statement = (*DeathStatement)(nil)

func (s *DeathStatement) RenderDetail(seq int, intro *NarrativeIntro, enc ExtendedMarkdownBuilder, hints *GrammarHints) {
	var detail string

	evDetail := ""
	bev := s.Principal.BestDeathlikeEvent
	switch bev.(type) {
	case *model.DeathEvent:
		if bev.IsInferred() {
			evDetail = text.JoinSentenceParts(evDetail, "is inferred to have died")
		} else {
			evDetail = text.JoinSentenceParts(evDetail, "died")
		}
	case *model.BurialEvent:
		if bev.IsInferred() {
			evDetail = text.JoinSentenceParts(evDetail, "is inferred to have been buried")
		} else {
			evDetail = text.JoinSentenceParts(evDetail, "was buried")
		}
	case *model.CremationEvent:
		if bev.IsInferred() {
			evDetail = text.JoinSentenceParts(evDetail, "is inferred to have been cremated")
		} else {
			evDetail = text.JoinSentenceParts(evDetail, "was cremated")
		}
	default:
		panic("unhandled deathlike event in DeathStatement")
	}

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
		evDetail = text.JoinSentenceParts(evDetail, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
	}
	detail += enc.EncodeWithCitations(evDetail, bev.GetCitations())

	additionalDetailFromDeathEvent := EventNarrativeDetail(bev)

	if additionalDetailFromDeathEvent != "" {
		detail += ". " + additionalDetailFromDeathEvent
	}

	funerals := []model.TimelineEvent{}
	for _, ev := range s.Principal.Timeline {
		switch tev := ev.(type) {
		case *model.BurialEvent:
			if tev != s.Principal.BestDeathlikeEvent {
				funerals = append(funerals, tev)
			}
		case *model.CremationEvent:
			if tev != s.Principal.BestDeathlikeEvent {
				funerals = append(funerals, tev)
			}
		}
	}
	if len(funerals) > 0 {
		if len(funerals) > 1 {
			// TODO: record an anomaly
		}

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
			evDetail = text.JoinSentenceParts(evDetail, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredUniqueName, pl.PreferredName, pl))
		}

		if additionalDetailFromDeathEvent != "" {
			detail = text.FinishSentence(detail) + " " + text.UpperFirst(s.Principal.Gender.SubjectPronounWithLink()) + " "
		} else {
			detail += " and was "
		}

		detail += enc.EncodeWithCitations(evDetail, funeralEvent.GetCitations())

	}

	detail += "."

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

				detail += enc.EncodeModelLinkDedupe(possibleSurvivor.PreferredFullName, possibleSurvivor.PreferredFamiliarName, possibleSurvivor)
				detail += "."
			}
		}
	}
	enc.Para(s.Principal.PreferredFamiliarName + " " + detail)
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

type CensusStatement struct {
	Principal *model.Person
	Event     *model.CensusEvent
}

var _ Statement = (*CensusStatement)(nil)

func (s *CensusStatement) RenderDetail(seq int, intro *NarrativeIntro, enc ExtendedMarkdownBuilder, hints *GrammarHints) {
	ce, found := s.Event.Entry(s.Principal)
	if !found {
		return
	}
	year, _ := s.Event.GetDate().Year()
	if ce.Narrative != "" {
		detail := fmt.Sprintf("in the %d census", year)
		detail = text.JoinSentences(detail, ce.Narrative)
		detail = enc.EncodeWithCitations(detail, s.Event.GetCitations())
		detail = text.FormatSentence(detail)
		enc.Para(detail)
		return
	}
	// TODO: construct narrative of census

	// detail := fmt.Sprintf("in the %d census", year)
	// detail = text.JoinSentences(detail, intro.NameBased)
	// detail = enc.EncodeWithCitations(detail, s.Event.GetCitations())
	// detail = text.FormatSentence(detail)
	// enc.Para(detail)

	// year, _ := s.Event.GetDate().Year()
	// enc.Para(fmt.Sprintf("census %d", year))
	// for _, ce := range s.Event.Entries {
	// 	enc.Para(fmt.Sprintf("%+v\n", ce))
	// }
	// if s.Event.Detail != "" {
	// 	enc.Para(s.Event.Detail)
	// }
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

// UTILITY

func ChooseFrom(n int, alternatives ...string) string {
	return alternatives[n%len(alternatives)]
}

func EventNarrativeDetail(ev model.TimelineEvent) string {
	detail := strings.ToLower(ev.GetDetail())
	if strings.HasPrefix(detail, "she was recorded as") || strings.HasPrefix(detail, "he was recorded as") || strings.HasPrefix(detail, "it was recorded that") {
		return ev.GetDetail()
	}
	return ""
}

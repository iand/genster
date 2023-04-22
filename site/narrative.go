package site

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/place"
	"github.com/iand/genster/text"
	"golang.org/x/exp/slog"
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
			nintro.Text = ChooseFrom(sequenceInNarrative, "", "", "then ")
			nintro.Text += nintro.NameBased
			nintro.DateInferred = false
		} else if nintro.NameBased != "" {
			nintro.Text = text.AppendClause(nintro.Text, nintro.NameBased)
		}
		nintro.Text = text.UpperFirst(nintro.Text)
		nintro.NameBased = text.UpperFirst(nintro.NameBased)
		nintro.TimeBased = text.UpperFirst(nintro.TimeBased)

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
		birth = enc.EncodeWithCitations(text.LowerFirst(WhatWhenWhere(s.Principal.BestBirthlikeEvent, enc)), s.Principal.BestBirthlikeEvent.GetCitations())
	}
	// TODO: position in family

	// Prose parentage
	parentage := text.LowerFirst(s.Principal.Gender.RelationToParentNoun()) + " of "
	if s.Principal.Father.IsUnknown() {
		if s.Principal.Mother.IsUnknown() {
			parentage += "unknown parents"
		} else {
			parentage += enc.EncodeModelLinkDedupe(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother.PreferredFamiliarName, s.Principal.Mother)
		}
	} else {
		if s.Principal.Mother.IsUnknown() {
			parentage += enc.EncodeModelLinkDedupe(s.Principal.Father.PreferredUniqueName, s.Principal.Father.PreferredFamiliarName, s.Principal.Father)
		} else {
			parentage += enc.EncodeModelLinkDedupe(s.Principal.Father.PreferredUniqueName, s.Principal.Father.PreferredFamiliarName, s.Principal.Father) + " and " + enc.EncodeModelLinkDedupe(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother.PreferredFamiliarName, s.Principal.Mother)
		}
	}

	// ---------------------------------------
	// Build detail
	// ---------------------------------------
	detail := s.Principal.PreferredGivenName

	if s.Principal.NickName != "" {
		detail += " (known as " + s.Principal.NickName + ")"
	}

	detail += " was "
	if birth != "" {
		detail += birth + ", the " + parentage
	} else {
		detail += "the " + parentage
	}
	detail = text.FinishSentence(detail)

	// Insert baptism here if there is only one, otherwise leave for a new para
	if len(s.Baptisms) == 1 {
		bapDetail := AgeWhenWhere(s.Baptisms[0], enc)
		if bapDetail != "" {
			detail = text.JoinSentence(detail, text.UpperFirst(s.Principal.Gender.SubjectPronoun()), "was baptised", enc.EncodeWithCitations(bapDetail, s.Baptisms[0].GetCitations()))
			detail = text.FinishSentence(detail)
		}

	}

	// ---------------------------------------
	// Prose relation to key person
	// ---------------------------------------
	if !s.SuppressRelation {
		if s.Principal.RelationToKeyPerson != nil && !s.Principal.RelationToKeyPerson.IsSelf() {
			detail += " " + text.UpperFirst(s.Principal.Gender.SubjectPronoun()) + " is the " + s.Principal.RelationToKeyPerson.Name() + " of " + enc.EncodeModelLinkDedupe(s.Principal.RelationToKeyPerson.From.PreferredFullName, s.Principal.RelationToKeyPerson.From.PreferredFamiliarName, s.Principal.RelationToKeyPerson.From)
		}
	}

	detail = text.FinishSentence(detail)
	enc.Para(detail)

	if len(s.Baptisms) > 1 {

		var bapDetail string
		for i, bev := range s.Baptisms {
			evDetail := ""
			if i == 0 {
				evDetail += "was baptised"
			} else {
				evDetail += "and again"
			}
			aww := AgeWhenWhere(bev, enc)
			if aww != "" {
				bapDetail = text.JoinSentence(bapDetail, evDetail, enc.EncodeWithCitations(bapDetail, s.Baptisms[0].GetCitations()))
			}
		}
		bapDetail = text.FinishSentence(text.JoinSentence(intro.NameBased, bapDetail))
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

	var detail string

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
			detail += event
		}
	}

	if len(s.Family.Children) == 0 {
		// single parents already dealt with
		if s.Principal.Childless {
			detail += ". They had no children."
		} else {
			detail += ". They had no known children."
		}
	} else {
		sort.Slice(s.Family.Children, func(i, j int) bool {
			return s.Family.Children[i].BestBirthlikeEvent.GetDate().SortsBefore(s.Family.Children[j].BestBirthlikeEvent.GetDate())
		})

		children := make([]string, len(s.Family.Children))
		for j := range s.Family.Children {
			children[j] = enc.EncodeModelLink(s.Family.Children[j].PreferredGivenName, s.Family.Children[j])
			if !s.Family.Children[j].BestBirthlikeEvent.GetDate().IsUnknown() {
				children[j] += fmt.Sprintf(" (%s)", s.Family.Children[j].BestBirthlikeEvent.ShortDescription())
			}
		}

		if singleParent {
			switch len(s.Family.Children) {
			case 1:
				detail += ChooseFrom(seq,
					" had one child with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun()+": "+children[0],
					" had one child, "+children[0]+", by an unknown "+s.Principal.Gender.Opposite().Noun(),
					" had one child "+children[0],
					" had one child, "+children[0]+", with an unknown "+s.Principal.Gender.Opposite().RelationToChildrenNoun(),
				)
			default:
				detail += " had " + text.CardinalWithUnit(len(s.Family.Children), "child", "children") + ": "
				detail += text.JoinList(children)
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
					" and had one child "+children[0],
					". They had just one child together "+children[0],
					". They had one child "+children[0],
					". "+children[0]+" was their only child",
				)
			case 2:
				list := text.JoinList(children)
				detail += ChooseFrom(seq,
					" and had two children: "+list,
					". They had two children "+list,
					". "+list+" were their only children",
				)
			default:
				number := text.CardinalWithUnit(len(s.Family.Children), "child", "children")
				detail += ChooseFrom(seq,
					". They had "+number+": ",
					" and went on to have "+number+" together: ",
					". They went on to have "+number+" ",
				)
				detail += text.JoinList(children)
			}
		}
		detail += "."

		// TODO: note how many children survived if some died
	}

	endDate := s.Family.BestEndDate
	end := ""
	if !endDate.IsUnknown() {
		switch s.Family.EndReason {
		case model.FamilyEndReasonDivorce:
			end += "They divorced " + endDate.When() + "."
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
						other.PreferredFamiliarName+" died "+endDate.When()+".",
						other.PreferredFamiliarName+" died "+endDate.When()+leavingWidow+".",
						"However, "+endDate.When()+", "+other.PreferredFamiliarName+" died.",
						"However, "+endDate.When()+", "+other.PreferredFamiliarName+" died "+leavingWidow+".",
					)
				}
			}
		case model.FamilyEndReasonUnknown:
			// TODO: format FamilyEndReasonUnknown
			end += " unknown " + endDate.When()
		}
	}
	if end != "" {
		detail += " " + end
	}

	enc.Para(intro.Text + detail)
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
			evDetail = text.JoinSentence(evDetail, "is inferred to have died")
		} else {
			evDetail = text.JoinSentence(evDetail, "died")
		}
	case *model.BurialEvent:
		if bev.IsInferred() {
			evDetail = text.JoinSentence(evDetail, "is inferred to have been buried")
		} else {
			evDetail = text.JoinSentence(evDetail, "was buried")
		}
	case *model.CremationEvent:
		if bev.IsInferred() {
			evDetail = text.JoinSentence(evDetail, "is inferred to have been cremated")
		} else {
			evDetail = text.JoinSentence(evDetail, "was cremated")
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
		evDetail = text.JoinSentence(evDetail, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
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
			evDetail = text.JoinSentence(evDetail, pl.PlaceType.InAt(), enc.EncodeModelLinkDedupe(pl.PreferredFullName, pl.PreferredName, pl))
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

// UTILITY

func ChooseFrom(n int, alternatives ...string) string {
	return alternatives[n%len(alternatives)]
}

func AgeQualifier(age int) string {
	if age == 0 {
		return "as an infant"
	} else if age < 10 {
		return "as a child"
	}
	return fmt.Sprintf("at the age of %s", text.CardinalNoun(age))
}

func EventNarrativeDetail(ev model.TimelineEvent) string {
	detail := strings.ToLower(ev.GetDetail())
	if strings.HasPrefix(detail, "she was recorded as") || strings.HasPrefix(detail, "he was recorded as") || strings.HasPrefix(detail, "it was recorded that") {
		return ev.GetDetail()
	}
	return ""
}

func GenerateOlb(p *model.Person) error {
	if p.Olb != "" {
		return nil
	}
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
				bf.CountryOfBirth = pl.CountryName
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
				if ch.BestBirthlikeEvent.GetDate().IsUnknown() {
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

	hasIllegitimateClause := false
	if p.Illegitimate && p.Father.IsUnknown() {
		clause := "illegitimate"
		if !p.Mother.IsUnknown() {
			clause += " " + p.Gender.RelationToParentNoun() + " of " + p.Mother.PreferredFamiliarFullName
		}

		hasIllegitimateClause = true
		clauses = append(clauses, Clause{Text: clause, Interestingness: 2})
	}

	// Intro statement
	if p.NickName != "" {
		clauses = append(clauses, Clause{Text: "known as " + p.NickName, Interestingness: 2})
	}

	// Statement about birth
	// TODO: ideally use primary occupation if it were clean enough
	nonNotableCountries := map[string]bool{
		"England":        true,
		"United Kingdom": true,
	}

	if p.BornInWorkhouse {
		if p.DiedInWorkhouse {
			clauses = append(clauses, Clause{Text: "born and died in workhouse", Interestingness: 2})
		} else {
			clauses = append(clauses, Clause{Text: "born in workhouse", Interestingness: 2})
		}
	} else if bf.CountryOfBirth != nil && !nonNotableCountries[bf.CountryOfBirth.Name] {
		if bf.BirthYear%3 == 1 {
			clauses = append(clauses, Clause{Text: bf.CountryOfBirth.Adjective + "-born", Interestingness: 1})
		} else {
			clauses = append(clauses, Clause{Text: "born in " + bf.CountryOfBirth.Name, Interestingness: 1})
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
			clauses = append(clauses, Clause{Text: "mother unknown", Interestingness: 1})
		} else if p.Father.IsUnknown() && !p.Mother.IsUnknown() {
			clauses = append(clauses, Clause{Text: "father unknown", Interestingness: 1})
		}
	}

	if p.Twin {
		clauses = append(clauses, Clause{Text: "twin", Interestingness: 2})
	}

	if p.PhysicalImpairment {
		clauses = append(clauses, Clause{Text: "physically impaired", Interestingness: 2})
	}

	if p.MentalImpairment {
		clauses = append(clauses, Clause{Text: "mentally impaired", Interestingness: 2})
	}

	if p.Deaf {
		clauses = append(clauses, Clause{Text: "deaf", Interestingness: 2})
	}

	if p.Blind {
		clauses = append(clauses, Clause{Text: "blind", Interestingness: 2})
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
		clauses = append(clauses, Clause{Text: "orphaned " + parentDeathDesc(bf.OrphanedAtAge), Interestingness: 3})
	} else if bf.AgeAtDeathOfMother > -1 && bf.AgeAtDeathOfMother < 18 {
		clauses = append(clauses, Clause{Text: "mother died " + parentDeathDesc(bf.AgeAtDeathOfMother), Interestingness: 2})
	} else if bf.AgeAtDeathOfFather > -1 && bf.AgeAtDeathOfFather < 18 {
		clauses = append(clauses, Clause{Text: "father died " + parentDeathDesc(bf.AgeAtDeathOfFather), Interestingness: 2})
	}

	// Statement about families and children
	legitimateChildren := bf.NumberOfChildren
	if bf.IllegitimateChildren != -1 {
		legitimateChildren -= bf.IllegitimateChildren
	}

	if p.Childless && bf.AgeAtDeath > 18 {
		clauses = append(clauses, Clause{Text: "had no children", Interestingness: 1})
	} else if p.Gender.IsFemale() || bf.NumberOfChildren == 0 {
		if bf.IllegitimateChildren == 1 {
			clauses = append(clauses, Clause{Text: "had one child with an unknown father", Interestingness: 1})
		} else if bf.IllegitimateChildren > 1 {
			clauses = append(clauses, Clause{Text: "had " + text.SmallCardinalNoun(bf.IllegitimateChildren) + " children with unknown fathers", Interestingness: 1})
		}

		if p.Unmarried && bf.AgeAtDeath > 18 {
			clauses = append(clauses, Clause{Text: "never married", Interestingness: 2})
		} else if bf.NumberOfMarriages > 0 {
			if bf.AgeAtFirstMarriage > 0 && bf.AgeAtFirstMarriage < 18 {
				if bf.NumberOfMarriages == 1 && len(bf.Spouses) > 0 {
					clauses = append(clauses, Clause{Text: "married " + bf.Spouses[0].PreferredFamiliarFullName + " at " + strconv.Itoa(bf.AgeAtFirstMarriage), Interestingness: 2})
				} else if bf.NumberOfMarriages == 2 {
					clauses = append(clauses, Clause{Text: "married at " + strconv.Itoa(bf.AgeAtFirstMarriage) + " then later remarried", Interestingness: 2})
				} else {
					clauses = append(clauses, Clause{Text: "married at " + strconv.Itoa(bf.AgeAtFirstMarriage) + " then " + text.SmallCardinalNoun(bf.NumberOfMarriages-1) + " more times", Interestingness: 2})
				}
			} else {
				if bf.NumberOfMarriages == 1 && len(bf.Spouses) > 0 {
					clauses = append(clauses, Clause{Text: "married " + bf.Spouses[0].PreferredFamiliarFullName, Interestingness: 1})
				} else {
					clauses = append(clauses, Clause{Text: "married " + text.MultiplicativeAdverb(bf.NumberOfMarriages), Interestingness: 2})
				}
			}
		}

		if legitimateChildren == 1 {
			clauses = append(clauses, Clause{Text: "had one child", Interestingness: 1})
		} else if legitimateChildren > 1 {
			clauses = append(clauses, Clause{Text: fmt.Sprintf("had %s children", text.SmallCardinalNoun(legitimateChildren)), Interestingness: 1})
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

		clauses = append(clauses, Clause{Text: clause, Interestingness: 2})

		if bf.IllegitimateChildren > 0 {
			if bf.IllegitimateChildren == bf.NumberOfChildren {
				if bf.IllegitimateChildren == 2 {
					clauses = append(clauses, Clause{Text: "both with unknown mothers", Interestingness: 1})
				} else if bf.IllegitimateChildren > 2 {
					clauses = append(clauses, Clause{Text: "all with unknown mothers", Interestingness: 2})
				}
			} else {
				clauses = append(clauses, Clause{Text: text.SmallCardinalNoun(bf.IllegitimateChildren) + " with unknown mothers", Interestingness: 1})
			}
		}
	}

	if bf.NumberOfMarriages == 1 && bf.AgeAtFirstSpouseDeath > 0 && bf.AgeAtFirstSpouseDeath < 40 {
		if p.Gender.IsFemale() {
			clauses = append(clauses, Clause{Text: "widowed at " + strconv.Itoa(bf.AgeAtFirstSpouseDeath), Interestingness: 2})
		} else {
			clauses = append(clauses, Clause{Text: "widower at " + strconv.Itoa(bf.AgeAtFirstSpouseDeath), Interestingness: 2})
		}
	}

	if bf.NumberOfDivorces > 0 {
		if bf.NumberOfDivorces < bf.NumberOfMarriages {
			clauses = append(clauses, Clause{Text: "divorced " + text.MultiplicativeAdverb(bf.NumberOfDivorces), Interestingness: 1})
		} else if bf.NumberOfDivorces == bf.NumberOfMarriages && bf.NumberOfDivorces == 1 {
			clauses = append(clauses, Clause{Text: "later divorced", Interestingness: 1})
		}
	}

	if bf.NumberOfAnnulments > 0 {
		log = true
		if bf.NumberOfAnnulments < bf.NumberOfMarriages {
			clauses = append(clauses, Clause{Text: "anulled " + text.MultiplicativeAdverb(bf.NumberOfDivorces), Interestingness: 1})
		} else if bf.NumberOfAnnulments == bf.NumberOfMarriages && bf.NumberOfAnnulments == 1 {
			clauses = append(clauses, Clause{Text: "later anulled", Interestingness: 2})
		}
	}

	if bf.TravelEvents > 4 {
		clauses = append(clauses, Clause{Text: "travelled widely", Interestingness: 2})
	}

	// TODO: occupation
	// TODO: suicide
	// TODO: imprisoned
	// TODO: deported

	if p.Pauper {
		clauses = append(clauses, Clause{Text: "pauper", Interestingness: 1})
	}

	// Statement about death
	if bf.AgeAtDeath == 0 {
		clauses = append(clauses, Clause{Text: bf.DeathType + " as an infant", Interestingness: 1})
	} else if bf.AgeAtDeath > 0 && bf.AgeAtDeath < 10 {
		clauses = append(clauses, Clause{Text: bf.DeathType + " as a child", Interestingness: 1})
	} else if bf.AgeAtDeath >= 10 && bf.AgeAtDeath < 30 {
		clauses = append(clauses, Clause{Text: fmt.Sprintf("%s before %s %s", bf.DeathType, p.Gender.SubjectPronounWithLink(), strconv.Itoa(bf.AgeAtDeath+1)), Interestingness: 2})
	} else if bf.AgeAtDeath > 90 && bf.Suicide {
		clauses = append(clauses, Clause{Text: fmt.Sprintf("lived to %s", strconv.Itoa(bf.AgeAtDeath)), Interestingness: 2})
	} else if p.DiedInWorkhouse && !p.BornInWorkhouse {
		clause := bf.DeathType + " in poverty"
		if bf.AgeAtDeath > 0 {
			clause += " at the age of " + strconv.Itoa(bf.AgeAtDeath)
		}
		clauses = append(clauses, Clause{Text: clause, Interestingness: 2})

	} else if bf.DeathYear != 0 {
		clause := bf.DeathType + " " + bf.DeathYearDesc
		if bf.AgeAtDeath > 0 {
			clause += " at the age of " + strconv.Itoa(bf.AgeAtDeath)
		}
		clauses = append(clauses, Clause{Text: clause, Interestingness: 1})
	}

	if p.CauseOfDeath == model.CauseOfDeathLostAtSea {
		clauses = append(clauses, Clause{Text: "lost at sea", Interestingness: 3})
	} else if p.CauseOfDeath == model.CauseOfDeathKilledInAction {
		clauses = append(clauses, Clause{Text: "killed in action", Interestingness: 3})
	} else if p.CauseOfDeath == model.CauseOfDeathDrowned {
		clauses = append(clauses, Clause{Text: "drowned", Interestingness: 3})
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

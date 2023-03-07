package site

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iand/gdate"
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

func (n *Narrative) Render(pov *model.POV, b StructuredMarkdownEncoder) {
	sort.Slice(n.Statements, func(i, j int) bool {
		if n.Statements[i].NarrativeSequence() == n.Statements[j].NarrativeSequence() {
			return gdate.SortsBefore(n.Statements[i].Start(), n.Statements[j].Start())
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
				b.Heading3("Life Story")
				// reset sequence at start of new section
				sequenceInNarrative = 0
			case NarrativeSequenceDeath:
				b.Heading3("Death")
				// reset sequence at start of new section
				sequenceInNarrative = 0
			}
		}

		var nintro NarrativeIntro

		if sequenceInNarrative == 0 || gdate.SortsBefore(pov.Person.BestDeathDate(), s.Start()) {
			nintro.NameBased = pov.Person.PreferredFamiliarName

			if s.NarrativeSequence() == NarrativeSequenceLifeStory {
				nintro.TimeBased = text.UpperFirst(s.Start().Occurrence()) + ", "
			}

		} else {
			name := pov.Person.PreferredFamiliarName
			if sequenceInNarrative%4 != 0 {
				name = pov.Person.Gender.SubjectPronoun()
			}
			nintro.NameBased = name

			if i > 0 {
				sincePrev := gdate.IntervalBetween(n.Statements[i-1].End(), s.Start())
				if yrs, ok := gdate.AsYearsInterval(sincePrev); ok {
					years := yrs.Years()
					preciseStart, isPreciseStart := gdate.AsPrecise(s.Start())
					dateInYear := ""
					if isPreciseStart {
						dateInYear = "on " + preciseStart.DateInYear(true)
					}

					if years < 0 && gdate.SortsBefore(s.Start(), n.Statements[i-1].End()) {
						nintro.TimeBased = ""
					} else if years == 0 {
						preciseInterval, isPreciseInterval := gdate.AsPreciseInterval(sincePrev)
						if isPreciseInterval && preciseInterval.ApproxDays() < 5 {
							nintro.TimeBased = ChooseFrom(sequenceInNarrative,
								dateInYear,
								text.AppendAside(dateInYear, "just a few days later"),
								text.AppendAside("Very shortly after", dateInYear),
								text.AppendAside("Just a few days later", dateInYear),
							)
						} else if isPreciseInterval && preciseInterval.ApproxDays() < 20 {
							nintro.TimeBased = ChooseFrom(sequenceInNarrative,
								text.AppendAside("Shortly after", dateInYear),
								text.AppendAside("Several days later", dateInYear),
							)
						} else if SameYear(n.Statements[i-1].End(), s.Start()) {
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
								text.AppendAside(text.UpperFirst(text.ReplaceFirstNumberWithCardinalNoun(sincePrev.Rough()))+" later", dateInYear),
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
							s.Start().Occurrence(),
							text.AppendClause(text.UpperFirst(text.ReplaceFirstNumberWithCardinalNoun(sincePrev.Rough()))+" later,", s.Start().Occurrence()),
							"",
							text.AppendClause("A few years later", s.Start().Occurrence()),
							text.AppendClause("Some years later", s.Start().Occurrence()),
							"",
						)
					} else {
						nintro.TimeBased = ChooseFrom(sequenceInNarrative,
							"",
							text.AppendClause(text.UpperFirst(text.ReplaceFirstNumberWithCardinalNoun(sincePrev.Rough()))+" later,", s.Start().Occurrence()),
							text.AppendClause("Several years later", s.Start().Occurrence()),
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
	RenderDetail(int, *NarrativeIntro, StructuredMarkdownEncoder, *GrammarHints)
	Start() gdate.Date
	End() gdate.Date
	NarrativeSequence() int
}

type IntroStatement struct {
	Principal *model.Person
}

var _ Statement = (*IntroStatement)(nil)

func (s *IntroStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
	var birth string
	// Prose birth
	if s.Principal.BestBirthlikeEvent != nil {
		birth = text.LowerFirst(EventTitle(s.Principal.BestBirthlikeEvent, enc, &model.POV{Person: s.Principal}))
	}

	// Prose parentage
	parentage := text.LowerFirst(s.Principal.Gender.RelationToParentNoun()) + " of "
	if s.Principal.Father.IsUnknown() {
		if s.Principal.Mother.IsUnknown() {
			parentage += "unknown parents"
		} else {
			parentage += enc.EncodeModelLink(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother, true)
		}
	} else {
		if s.Principal.Mother.IsUnknown() {
			parentage += enc.EncodeModelLink(s.Principal.Father.PreferredUniqueName, s.Principal.Father, true)
		} else {
			parentage += enc.EncodeModelLink(s.Principal.Father.PreferredUniqueName, s.Principal.Father, true) + " and " + enc.EncodeModelLink(s.Principal.Mother.PreferredUniqueName, s.Principal.Mother, true)
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
		detail += birth + ", the " + parentage + "."
	} else {
		detail += "the " + parentage + "."
	}

	// TODO: position in family
	// TODO: known as

	// ---------------------------------------
	// Prose relation to key person
	// ---------------------------------------
	if s.Principal.RelationToKeyPerson != nil && !s.Principal.RelationToKeyPerson.IsSelf() {
		detail += " " + text.UpperFirst(s.Principal.Gender.SubjectPronoun()) + " is the " + s.Principal.RelationToKeyPerson.Name() + " of " + enc.EncodeModelLink(s.Principal.RelationToKeyPerson.From.PreferredFullName, s.Principal.RelationToKeyPerson.From, true) + "."
	}
	enc.Para(detail)
}

func (s *IntroStatement) Start() gdate.Date {
	return s.Principal.BestBirthDate()
}

func (s *IntroStatement) End() gdate.Date {
	return s.Principal.BestBirthDate()
}

func (s *IntroStatement) NarrativeSequence() int {
	return NarrativeSequenceIntro
}

type BaptismsStatement struct {
	Principal *model.Person
	Events    []*model.BaptismEvent
}

var _ Statement = (*BaptismsStatement)(nil)

func (s *BaptismsStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
	var detail string
	for i, bev := range s.Events {
		evDetail := ""
		if i == 0 {
			evDetail += "was baptised"
		} else {
			evDetail += "and again"
		}
		if !gdate.IsUnknown(bev.Date) {
			if age, ok := s.Principal.AgeInYearsAt(bev.Date); ok {
				evDetail += " " + AgeQualifier(age)
			}
			evDetail += " " + bev.Date.Occurrence()
		}
		if !bev.Place.IsUnknown() {
			evDetail += " in " + bev.Place.PreferredName
		}
		detail += EncodeWithCitations(evDetail, bev.GetCitations(), enc)
	}
	detail += "."

	enc.Para(intro.NameBased + " " + detail)
}

func (s *BaptismsStatement) Start() gdate.Date {
	if len(s.Events) > 0 {
		return s.Events[0].GetDate()
	}
	return s.Principal.BestBirthlikeEvent.GetDate()
}

func (s *BaptismsStatement) End() gdate.Date {
	if len(s.Events) > 0 {
		return s.Events[len(s.Events)-1].GetDate()
	}
	return s.Principal.BestBirthlikeEvent.GetDate()
}

func (s *BaptismsStatement) EndedWithDeathOf(p *model.Person) bool {
	return true
}

func (s *BaptismsStatement) NarrativeSequence() int {
	return NarrativeSequenceEarlyLife
}

type FamilyStatement struct {
	Principal *model.Person
	Family    *model.Family
}

var _ Statement = (*FamilyStatement)(nil)

func (s *FamilyStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
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
		otherName = enc.EncodeModelLink(other.PreferredUniqueName, other, true)
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
		if !gdate.IsUnknown(startDate) {
			event += " " + action
			event += " " + otherName
			if !intro.DateInferred {
				event += " " + startDate.Occurrence()
			}
			if age, ok := s.Principal.AgeInYearsAt(startDate); ok && age < 18 || age > 45 {
				event += " " + AgeQualifier(age)
			}
		} else {
			event += " " + action
			event += " " + otherName
		}
		if s.Family.BestStartEvent != nil {
			detail += EncodeWithCitations(event, s.Family.BestStartEvent.GetCitations(), enc)
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
			return gdate.SortsBefore(s.Family.Children[i].BestBirthlikeEvent.GetDate(), s.Family.Children[j].BestBirthlikeEvent.GetDate())
		})

		children := make([]string, len(s.Family.Children))
		for j := range s.Family.Children {
			children[j] = enc.EncodeModelLink(s.Family.Children[j].PreferredGivenName, s.Family.Children[j], true)
			if !gdate.IsUnknown(s.Family.Children[j].BestBirthlikeEvent.GetDate()) {
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
	if !gdate.IsUnknown(endDate) {
		switch s.Family.EndReason {
		case model.FamilyEndReasonDivorce:
			end += "They divorced " + endDate.Occurrence() + "."
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
						other.PreferredFamiliarName+" died "+endDate.Occurrence()+".",
						other.PreferredFamiliarName+" died "+endDate.Occurrence()+leavingWidow+".",
						"However, "+endDate.Occurrence()+", "+other.PreferredFamiliarName+" died.",
						"However, "+endDate.Occurrence()+", "+other.PreferredFamiliarName+" died "+leavingWidow+".",
					)
				}
			}
		case model.FamilyEndReasonUnknown:
			// TODO: format FamilyEndReasonUnknown
			end += " unknown " + endDate.Occurrence()
		}
	}
	if end != "" {
		detail += " " + end
	}

	enc.Para(intro.Text + detail)
}

func (s *FamilyStatement) Start() gdate.Date {
	return s.Family.BestStartDate
}

func (s *FamilyStatement) End() gdate.Date {
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

func (s *DeathStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
	var detail string

	evDetail := ""
	bev := s.Principal.BestDeathlikeEvent
	switch bev.(type) {
	case *model.DeathEvent:
		evDetail += "died"
	case *model.BurialEvent:
		evDetail += "was buried"
	case *model.CremationEvent:
		evDetail += "was cremated"
	default:
		panic("unhandled deathlike event in DeathStatement")
	}

	if !gdate.IsUnknown(bev.GetDate()) {
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
		evDetail += " " + bev.GetDate().Occurrence()
	} else {
		evDetail += " on an unknown date"
	}
	if !bev.GetPlace().IsUnknown() {
		pl := bev.GetPlace()
		evDetail += " " + pl.PlaceType.InAt() + " " + bev.GetPlace().PreferredFullName
	}
	detail += EncodeWithCitations(evDetail, bev.GetCitations(), enc)

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

		interval := gdate.IntervalBetween(bev.GetDate(), funeralEvent.GetDate())
		if precise, ok := gdate.AsPreciseInterval(interval); ok && precise.D < 15 {
			if precise.D == 0 {
				evDetail += " the same day"
			} else if precise.D == 1 {
				evDetail += " the next day"
			} else {
				evDetail += fmt.Sprintf(" %s days later", text.CardinalNounUnderTwenty(precise.D))
			}
		} else {
			evDetail += " " + funeralEvent.GetDate().Occurrence()
		}
		if !funeralEvent.GetPlace().IsUnknown() {
			pl := funeralEvent.GetPlace()
			evDetail += " " + pl.PlaceType.InAt() + " " + funeralEvent.GetPlace().PreferredFullName
		}

		if additionalDetailFromDeathEvent != "" {
			detail = text.FinishSentence(detail) + " " + text.UpperFirst(s.Principal.Gender.SubjectPronounWithLink()) + " "
		} else {
			detail += " and was "
		}

		detail += EncodeWithCitations(evDetail, funeralEvent.GetCitations(), enc)

	}

	detail += "."

	if len(s.Principal.Families) > 0 {
		sort.Slice(s.Principal.Families, func(i, j int) bool {
			if s.Principal.Families[i].BestStartDate != nil || s.Principal.Families[j].BestStartDate == nil {
				return false
			}
			return gdate.SortsBefore(s.Principal.Families[i].BestStartDate, s.Principal.Families[j].BestStartDate)
		})

		lastFamily := s.Principal.Families[len(s.Principal.Families)-1]
		possibleSurvivor := lastFamily.OtherParent(s.Principal)
		if possibleSurvivor != nil && possibleSurvivor.BestDeathlikeEvent != nil && !gdate.IsUnknown(possibleSurvivor.BestDeathlikeEvent.GetDate()) {
			if gdate.SortsBefore(s.Principal.BestDeathlikeEvent.GetDate(), possibleSurvivor.BestDeathlikeEvent.GetDate()) {
				detail += " " + text.UpperFirst(s.Principal.Gender.SubjectPronoun()) + " was survived by "
				if lastFamily.Bond == model.FamilyBondMarried {
					detail += text.LowerFirst(s.Principal.Gender.PossessivePronounSingular()) + " " + text.LowerFirst(possibleSurvivor.Gender.RelationToSpouseNoun()) + " "
				}

				detail += enc.EncodeModelLink(possibleSurvivor.PreferredFamiliarName, possibleSurvivor, true)
				detail += "."
			}
		}
	}
	enc.Para(s.Principal.PreferredFamiliarName + " " + detail)
}

func (s *DeathStatement) Start() gdate.Date {
	return s.Principal.BestDeathlikeEvent.GetDate()
}

func (s *DeathStatement) End() gdate.Date {
	return s.Principal.BestDeathlikeEvent.GetDate()
}

func (s *DeathStatement) NarrativeSequence() int {
	return NarrativeSequenceDeath
}

type WillAndProbateStatement struct {
	Principal *model.Person
	Event     model.TimelineEvent
}

var _ Statement = (*WillAndProbateStatement)(nil)

func (s *WillAndProbateStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
	detail := ""

	switch tev := s.Event.(type) {
	case *model.ProbateEvent:
		detail = "probate was granted "
		if !intro.DateInferred {
			detail += tev.GetDate().Occurrence()
		}
		if !tev.GetPlace().IsUnknown() {
			pl := tev.GetPlace()
			detail += " " + pl.PlaceType.InAt() + " " + pl.PreferredFullName
			detail = EncodeWithCitations(detail, tev.GetCitations(), enc)
		}
	}

	if detail != "" {
		enc.Para(text.FormatSentence(text.JoinSentence(intro.TimeBased, detail)))
	}
}

func (s *WillAndProbateStatement) Start() gdate.Date {
	return s.Event.GetDate()
}

func (s *WillAndProbateStatement) End() gdate.Date {
	return s.Event.GetDate()
}

func (s *WillAndProbateStatement) NarrativeSequence() int {
	return NarrativeSequencePostDeath
}

type CensusStatement struct {
	Principal *model.Person
	Event     *model.CensusEvent
}

var _ Statement = (*CensusStatement)(nil)

func (s *CensusStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
	detail := "was recorded in the "

	if !intro.DateInferred {
		yearer, ok := gdate.AsYear(s.Event.GetDate())
		if ok {
			detail = text.JoinSentence(detail, fmt.Sprintf("%d ", yearer.Year()))
		}
	}
	detail = text.JoinSentence(detail, "census")

	entry, _ := s.Event.Entry(s.Principal)
	if entry != nil && entry.RelationToHead.IsImpersonal() {
		detail = text.JoinSentence(detail, "as", "a"+text.MaybeAn(string(entry.RelationToHead)))
	}

	if !s.Event.GetPlace().IsUnknown() {
		pl := s.Event.GetPlace()

		residing := ChooseFrom(seq, "residing", "", "living")

		detail = text.JoinSentence(detail, residing, pl.PlaceType.InAt(), pl.PreferredFullName)
		detail = EncodeWithCitations(detail, s.Event.GetCitations(), enc)
	}

	if detail != "" {
		enc.Para(text.FormatSentence(text.JoinSentence(intro.Text, detail)))
	}
}

func (s *CensusStatement) Start() gdate.Date {
	return s.Event.GetDate()
}

func (s *CensusStatement) End() gdate.Date {
	return s.Event.GetDate()
}

func (s *CensusStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

// UTILITY

func SameYear(a, b gdate.Date) bool {
	ay, ok := gdate.AsYear(a)
	if !ok {
		return false
	}
	by, ok := gdate.AsYear(b)
	if !ok {
		return false
	}

	return ay.Year() == by.Year()
}

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

type VerbatimStatement struct {
	Principal *model.Person
	Detail    string
	Date      gdate.Date
	Citations []*model.GeneralCitation
}

var _ Statement = (*VerbatimStatement)(nil)

func (s *VerbatimStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
	var detail string
	if intro.TimeBased != "" {
		detail = text.JoinSentence(intro.TimeBased, s.Detail)
	} else if !gdate.IsUnknown(s.Date) {
		detail = text.JoinSentence(s.Date.Occurrence(), s.Detail)
	} else {
		detail = s.Detail
	}

	if detail != "" {
		enc.Para(EncodeWithCitations(text.UpperFirst(text.FinishSentence(detail)), s.Citations, enc))
	}
}

func (s *VerbatimStatement) Start() gdate.Date {
	return s.Date
}

func (s *VerbatimStatement) End() gdate.Date {
	return s.Date
}

func (s *VerbatimStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

type ArrivalDepartureStatement struct {
	Principal *model.Person
	Event     model.TimelineEvent
}

var _ Statement = (*ArrivalDepartureStatement)(nil)

func (s *ArrivalDepartureStatement) RenderDetail(seq int, intro *NarrativeIntro, enc StructuredMarkdownEncoder, hints *GrammarHints) {
	pl := s.Event.GetPlace()
	if pl.IsUnknown() {
		return
	}

	var action string

	switch s.Event.(type) {
	case *model.ArrivalEvent:
		action = "arrived at"
	case *model.DepartureEvent:
		action = "departed"
	default:
		panic("unexpected event type")
	}

	var detail string
	if !gdate.IsUnknown(s.Event.GetDate()) {
		if intro.NameBased == "" {
			detail = text.JoinSentence(s.Event.GetDate().Occurrence(), action, pl.PreferredName)
		} else {
			detail = text.JoinSentence(intro.NameBased, action, pl.PreferredName, s.Event.GetDate().Occurrence())
		}
	} else {
		detail = text.JoinSentence(intro.NameBased, action, pl.PreferredName)
	}

	detail = text.FormatSentence(detail)

	if s.Event.GetDetail() != "" {
		detail += " " + text.FormatSentence(s.Event.GetDetail())
	}

	if detail != "" {
		enc.Para(EncodeWithCitations(detail, s.Event.GetCitations(), enc))
	}
}

func (s *ArrivalDepartureStatement) Start() gdate.Date {
	return s.Event.GetDate()
}

func (s *ArrivalDepartureStatement) End() gdate.Date {
	return s.Event.GetDate()
}

func (s *ArrivalDepartureStatement) NarrativeSequence() int {
	return NarrativeSequenceLifeStory
}

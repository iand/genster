package narrative

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

type PersonNarrative[T render.EncodedText] struct {
	Statements []Statement[T]
}

type IntroGenerator[T render.EncodedText] struct {
	POV              *model.POV
	NameMinSeq       int                 // the minimum sequence that the person's name may be used in an intro
	AgeMinSeq        int                 // the minimum sequence that the person's age may be used in an intro
	LastIntroDate    *model.Date         //  the date that the last intro was requested
	PeopleIntroduced map[string][]string // a lookup of occupations for people who have been introduced
}

func (n *IntroGenerator[T]) Default(seq int, dt *model.Date) string {
	part1 := n.RelativeTime(seq, dt, true)
	part2 := n.Pronoun(seq, dt, nil)

	if part1 == "" {
		return part2
	}

	if part2 == "" {
		return part1
	}

	return part1 + ", " + part2
}

func (n *IntroGenerator[T]) Pronoun(seq int, dt *model.Date, principal *model.Person) string {
	defer func() {
		n.LastIntroDate = dt
	}()
	if n.POV.Person == nil {
		if principal.IsUnknown() {
			return "they"
		}
		return principal.PreferredFamiliarName
	}

	if seq >= n.NameMinSeq {
		n.NameMinSeq = seq + 3
		return n.POV.Person.PreferredFamiliarName
	}

	return n.POV.Person.Gender.SubjectPronoun()
}

func (n *IntroGenerator[T]) WasWere(principal *model.Person) string {
	if n.POV.Person == nil && principal.IsUnknown() {
		return "were"
	}
	return "was"
}

func (n *IntroGenerator[T]) RelativeTime(seq int, dt *model.Date, includeFullDate bool) string {
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

func (n *IntroGenerator[T]) IntroducePerson(seq int, p *model.Person, dt *model.Date, suppressSameSurname bool, enc render.TextEncoder[T], nc NameChooser) string {
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
		if !n.POV.Person.IsUnknown() && suppressSameSurname && p.PreferredFamilyName == n.POV.Person.PreferredFamilyName {
			detail = enc.EncodeModelLinkDedupe(enc.EncodeText(p.PreferredGivenName), enc.EncodeText(p.PreferredGivenName), p).String()
		} else {
			detail = enc.EncodeModelLinkDedupe(enc.EncodeText(p.PreferredUniqueName), enc.EncodeText(p.PreferredFullName), p).String()
		}
		if occDetail != "" {
			detail += ", " + occDetail + ","
		}
		return detail
	}

	name := p.PreferredFullName
	if !n.POV.Person.IsUnknown() && suppressSameSurname && p.PreferredFamilyName == n.POV.Person.PreferredFamilyName {
		name = p.PreferredGivenName
	}

	hadPreviousOccupation := false
	for _, od := range n.PeopleIntroduced[p.ID] {
		if od != "" {
			hadPreviousOccupation = true
		}
		if occDetail == od {
			return enc.EncodeModelLinkDedupe(enc.EncodeText(name), enc.EncodeText(p.PreferredGivenName), p).String()
		}
	}
	n.PeopleIntroduced[p.ID] = append(n.PeopleIntroduced[p.ID], occDetail)
	detail := enc.EncodeModelLinkDedupe(enc.EncodeText(name), enc.EncodeText(p.PreferredGivenName), p).String()
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

func (n *PersonNarrative[T]) Render(pov *model.POV, b render.ContentBuilder[T]) {
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
	nintro := IntroGenerator[T]{
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

		s.RenderDetail(sequenceInNarrative, &nintro, b, &DefaultNameChooser{})
		b.EmptyPara()

		sequenceInNarrative++

	}
}

// UTILITY

func ChooseFrom(n int, alternatives ...string) string {
	return alternatives[n%len(alternatives)]
}

func EventNarrativeDetail[T render.EncodedText](ev model.TimelineEvent, enc render.ContentBuilder[T]) string {
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

package narrative

import (
	"sort"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

type FamilyNarrative[T render.EncodedText] struct {
	Family           *model.Family
	FatherStatements []Statement[T]
	MotherStatements []Statement[T]
	FamilyStatements []Statement[T]
}

func sortStatements[T render.EncodedText](ss []Statement[T]) {
	sort.Slice(ss, func(i, j int) bool {
		if ss[i].NarrativeSequence() == ss[j].NarrativeSequence() {
			if ss[i].Start().SameDate(ss[j].Start()) {
				return ss[i].Priority() > ss[j].Priority()
			}
			return ss[i].Start().SortsBefore(ss[j].Start())
		}
		return ss[i].NarrativeSequence() < ss[j].NarrativeSequence()
	})
}

func (n *FamilyNarrative[T]) Render(b render.ContentBuilder[T], nc NameChooser) {
	sortStatements(n.FatherStatements)
	sortStatements(n.MotherStatements)
	sortStatements(n.FamilyStatements)

	father := n.Family.Father
	mother := n.Family.Mother

	fintro := IntroGenerator[T]{
		POV: &model.POV{
			Person: father,
		},
	}
	mintro := IntroGenerator[T]{
		POV: &model.POV{
			Person: mother,
		},
	}

	if !father.IsUnknown() && !mother.IsUnknown() {
		// Interleave statements from father and mother of family

		var fidx, midx int

		// Father's intro and life before marriage
		headingPrinted := false
		for ; fidx < len(n.FatherStatements); fidx++ {
			s := n.FatherStatements[fidx]
			if s.NarrativeSequence() == NarrativeSequenceIntro || s.Start().SortsBefore(n.Family.BestStartDate) {
				if !headingPrinted {
					b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Father.PreferredFamiliarName)+" background"), "")
					headingPrinted = true
				}
				s.RenderDetail(fidx, &fintro, b, nc)
				b.EmptyPara()
			} else {
				break
			}
		}

		// Mother's intro and life before marriage
		headingPrinted = false
		for ; midx < len(n.MotherStatements); midx++ {
			s := n.MotherStatements[midx]
			if s.NarrativeSequence() == NarrativeSequenceIntro || s.Start().SortsBefore(n.Family.BestStartDate) {
				if !headingPrinted {
					b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Mother.PreferredFamiliarName)+" background"), "")
					headingPrinted = true
				}
				s.RenderDetail(midx, &mintro, b, nc)
				b.EmptyPara()
			} else {
				break
			}
		}

		// Family life
		famIntro := IntroGenerator[T]{
			POV: &model.POV{
				Family: n.Family,
			},
		}
		headingPrinted = false
		for idx, s := range n.FamilyStatements {
			if !headingPrinted {
				b.Heading3(b.EncodeText("Family life"), "")
				headingPrinted = true
			}
			s.RenderDetail(idx, &famIntro, b, nc)
			b.EmptyPara()
		}

		if father.BestDeathlikeEvent != nil && mother.BestDeathlikeEvent != nil && mother.BestDeathlikeEvent.GetDate().SortsBefore(father.BestDeathlikeEvent.GetDate()) {
			// Mother died before father
			headingPrinted = false
			for ; midx < len(n.MotherStatements); midx++ {
				s := n.MotherStatements[midx]
				if !headingPrinted {
					if _, ok := s.(*DeathStatement[T]); ok {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Mother.PreferredFamiliarName)+" death"), "")
					} else {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Mother.PreferredFamiliarName)+" later life"), "")
					}
					headingPrinted = true
				}
				s.RenderDetail(fidx, &mintro, b, nc)
				b.EmptyPara()
			}

			headingPrinted = false
			for ; fidx < len(n.FatherStatements); fidx++ {
				s := n.FatherStatements[fidx]
				if !headingPrinted {
					if _, ok := s.(*DeathStatement[T]); ok {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Father.PreferredFamiliarName)+" death"), "")
					} else {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Father.PreferredFamiliarName)+" later life"), "")
					}
					headingPrinted = true
				}
				s.RenderDetail(fidx, &fintro, b, nc)
				b.EmptyPara()
			}

		} else {

			headingPrinted = false
			for ; fidx < len(n.FatherStatements); fidx++ {
				s := n.FatherStatements[fidx]
				if !headingPrinted {
					if _, ok := s.(*DeathStatement[T]); ok {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Father.PreferredGivenName)+" death"), "")
					} else {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Father.PreferredGivenName)+" later life"), "")
					}
					headingPrinted = true
				}
				s.RenderDetail(fidx, &fintro, b, nc)
				b.EmptyPara()
			}

			headingPrinted = false
			for ; midx < len(n.MotherStatements); midx++ {
				s := n.MotherStatements[midx]
				if !headingPrinted {
					if _, ok := s.(*DeathStatement[T]); ok {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Mother.PreferredGivenName)+" death"), "")
					} else {
						b.Heading3(b.EncodeText(text.MaybePossessiveSuffix(n.Family.Mother.PreferredGivenName)+" later life"), "")
					}
					headingPrinted = true
				}
				s.RenderDetail(fidx, &mintro, b, nc)
				b.EmptyPara()
			}
		}

	} else if !father.IsUnknown() {
		// Just father's narrative
		currentNarrativeSequence := NarrativeSequenceIntro
		sequenceInNarrative := 0
		for _, s := range n.FatherStatements {
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

			s.RenderDetail(sequenceInNarrative, &fintro, b, nc)
			b.EmptyPara()

			sequenceInNarrative++

		}
	} else if !mother.IsUnknown() {
		// Just mother's narrative
		currentNarrativeSequence := NarrativeSequenceIntro
		sequenceInNarrative := 0
		for _, s := range n.MotherStatements {
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

			s.RenderDetail(sequenceInNarrative, &mintro, b, nc)
			b.EmptyPara()

			sequenceInNarrative++

		}
	}
}

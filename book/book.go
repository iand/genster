package book

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/pandoc"
	"github.com/iand/genster/site"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
)

func NewBook(t *tree.Tree) *Book {
	b := &Book{
		Tree: t,
	}
	return b
}

type Book struct {
	Tree             *tree.Tree
	IncludePrivate   bool
	IncludeDebugInfo bool
	PublishSet       *site.PublishSet
	Doc              *pandoc.Document
	Chapters         []*Chapter
	OutputDir        string
}

func (b *Book) BuildPublishSet(m model.PersonMatcher) error {
	subset, err := site.NewPublishSet(b.Tree, m)
	if err != nil {
		return fmt.Errorf("build publish set: %w", err)
	}

	b.PublishSet = subset
	return nil
}

func (s *Book) BuildDocument() error {
	s.Doc = &pandoc.Document{}

	var families []*model.Family

	// Build list of families with a breadth-first traversal

	hasBiographicalDetail := func(p *model.Person) bool {
		if p.IsUnknown() {
			return false
		}
		if p.Redacted {
			return false
		}
		if !p.ParentFamily.IsUnknown() {
			return true
		}
		if p.BestBirthlikeEvent != nil || p.BestDeathlikeEvent != nil || len(p.Children) > 0 {
			return true
		}

		return false
	}

	var ancestors []*model.Person
	ancestors = append(ancestors, s.Tree.KeyPerson)
	for len(ancestors) > 0 {
		p := ancestors[0]
		ancestors = ancestors[1:]
		f := p.ParentFamily
		if !f.IsUnknown() {
			if hasBiographicalDetail(f.Father) || hasBiographicalDetail(f.Mother) {
				families = append(families, f)
			}
			if !f.Father.IsUnknown() {
				ancestors = append(ancestors, f.Father)
			}
			if !f.Mother.IsUnknown() {
				ancestors = append(ancestors, f.Mother)
			}
		}
	}

	s.Doc.Heading1("Families", "")
	for _, f := range families {
		if f.Father.IsUnknown() && f.Mother.IsUnknown() {
			continue
		}
		if !f.Father.IsUnknown() && !f.Father.IsDirectAncestor() {
			continue
		}
		if !f.Mother.IsUnknown() && !f.Mother.IsDirectAncestor() {
			continue
		}
		err := s.AddFamilyChapter(f)
		if err != nil {
			return fmt.Errorf("family page: %w", err)
		}

		// if err := writePage(d, contentDir, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
		// 	return fmt.Errorf("write person page: %w", err)
		// }

	}
	// for _, p := range s.PublishSet.People {
	// 	if s.LinkFor(p) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderPersonPage(s, p)
	// 	if err != nil {
	// 		return fmt.Errorf("render person page: %w", err)
	// 	}

	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
	// 		return fmt.Errorf("write person page: %w", err)
	// 	}

	// }

	// for _, p := range s.PublishSet.Places {
	// 	if s.LinkFor(p) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderPlacePage(s, p)
	// 	if err != nil {
	// 		return fmt.Errorf("render place page: %w", err)
	// 	}

	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.PlaceFilePattern, p.ID)); err != nil {
	// 		return fmt.Errorf("write place page: %w", err)
	// 	}
	// }

	// for _, c := range s.PublishSet.Citations {
	// 	if s.LinkFor(c) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderCitationPage(s, c)
	// 	if err != nil {
	// 		return fmt.Errorf("render citation page: %w", err)
	// 	}
	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.CitationFilePattern, c.ID)); err != nil {
	// 		return fmt.Errorf("write citation page: %w", err)
	// 	}
	// }

	// for _, mo := range s.PublishSet.MediaObjects {
	// 	// TODO: redaction

	// 	// var ext string
	// 	// switch mo.MediaType {
	// 	// case "image/jpeg":
	// 	// 	ext = "jpg"
	// 	// case "image/png":
	// 	// 	ext = "png"
	// 	// case "image/gif":
	// 	// 	ext = "gif"
	// 	// default:
	// 	// 	return fmt.Errorf("unsupported media type: %v", mo.MediaType)
	// 	// }

	// 	fname := filepath.Join(mediaDir, fmt.Sprintf("%s/%s", s.MediaDir, mo.FileName))

	// 	if err := CopyFile(fname, mo.SrcFilePath); err != nil {
	// 		return fmt.Errorf("copy media object: %w", err)
	// 	}
	// }

	// s.BuildCalendar()

	// for month, c := range s.Calendars {
	// 	d, err := c.RenderPage(s)
	// 	if err != nil {
	// 		return fmt.Errorf("generate markdown: %w", err)
	// 	}

	// 	fname := fmt.Sprintf(s.CalendarFilePattern, month)

	// 	f, err := CreateFile(filepath.Join(contentDir, fname))
	// 	if err != nil {
	// 		return fmt.Errorf("create calendar file: %w", err)
	// 	}
	// 	if _, err := d.WriteTo(f); err != nil {
	// 		return fmt.Errorf("write calendar markdown: %w", err)
	// 	}
	// 	f.Close()
	// }

	// if err := s.WritePersonListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write people list pages: %w", err)
	// }

	// if err := s.WritePlaceListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write place list pages: %w", err)
	// }

	// // Not publishing sources at this time
	// // if err := s.WriteSourceListPages(contentDir); err != nil {
	// // 	return fmt.Errorf("write source list pages: %w", err)
	// // }

	// if err := s.WriteSurnameListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write surname list pages: %w", err)
	// }

	// if err := s.WriteInferenceListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write inferences pages: %w", err)
	// }

	// if err := s.WriteAnomalyListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write anomalies pages: %w", err)
	// }

	// if err := s.WriteTodoListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write todo pages: %w", err)
	// }

	// if err := s.WriteTreeOverview(contentDir); err != nil {
	// 	return fmt.Errorf("write tree overview: %w", err)
	// }

	// if err := s.WriteChartAncestors(contentDir); err != nil {
	// 	return fmt.Errorf("write ancestor chart: %w", err)
	// }

	// if err := s.WriteChartTrees(root); err != nil {
	// 	return fmt.Errorf("write chart trees: %w", err)
	// }

	// TODO: order chapters
	for _, c := range s.Chapters {
		s.Doc.AppendText(c.Content.Text())
	}

	return nil
}

func (s *Book) WriteDocument(fname string) error {
	path := filepath.Dir(fname)

	if err := os.MkdirAll(path, 0o777); err != nil {
		return fmt.Errorf("create path: %w", err)
	}

	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if _, err := s.Doc.WriteTo(f); err != nil {
		return fmt.Errorf("write content: %w", err)
	}
	return f.Close()
}

func (b *Book) AddFamilyChapter(f *model.Family) error {
	var indexTerm string
	if !f.Father.IsUnknown() {
		indexTerm = "\\index{" + f.Father.PreferredFamilyName + "!" + f.Father.PreferredGivenName + "}"
	}
	if !f.Mother.IsUnknown() {
		indexTerm += "\\index{" + f.Mother.PreferredFamilyName + "!" + f.Mother.PreferredGivenName + "}"
	}
	ch := NewChapter(indexTerm + f.PreferredFullName + "\\label{" + f.ID + "}")
	b.Chapters = append(b.Chapters, ch)

	nc := &narrative.DefaultNameChooser{}
	enc := ch.Content

	var para text.Para
	_ = para

	// Describe relation to key person
	rel := ""
	dist := 0
	if !f.Father.IsUnknown() {
		dist = f.Father.RelationToKeyPerson.Distance()
		rel = "father"
		if !f.Mother.IsUnknown() {
			rel = "parents"
		}
	} else if !f.Mother.IsUnknown() {
		rel = "mother"
		dist = f.Mother.RelationToKeyPerson.Distance()
	}
	if dist > 4 {
		enc.Para(enc.EncodeText(text.FormatSentence(text.OrdinalNoun(dist-2) + " great grand" + rel)))
	} else if dist == 4 {
		enc.Para(enc.EncodeText(text.FormatSentence("great great grand" + rel)))
	} else if dist == 3 {
		enc.Para(enc.EncodeText(text.FormatSentence("great grand" + rel)))
	} else if dist == 2 {
		enc.Para(enc.EncodeText(text.FormatSentence("grand" + rel)))
	} else if dist == 1 {
		enc.Para(enc.EncodeText(text.FormatSentence(rel)))
	}

	// Describe family relationship
	var what string
	switch f.Bond {
	case model.FamilyBondUnmarried:
		if f.Father.IsUnknown() {
			if len(f.Children) == 1 {
				what = "had a child by an unknown father"
			} else {
				what = "had children by an unknown father"
			}
		} else if f.Father.IsUnknown() {
			if len(f.Children) == 1 {
				what = "had a child by an unknown mother"
			} else {
				what = "had children by an unknown mother"
			}
		} else {
			what = "were unmarried"
		}
	case model.FamilyBondMarried:
		what = "were married"
		if f.BestStartEvent != nil {
			what = enc.EncodeWithCitations(enc.EncodeText(narrative.WhatWhenWhere(what, f.BestStartEvent.GetDate(), f.BestStartEvent.GetPlace(), enc, nc)), f.BestStartEvent.GetCitations()).String()
			// 	para.NewSentence(enc.EncodeWithCitations(enc.EncodeText(narrative.WhatWhenWhere(what, f.BestStartEvent.GetDate(), f.BestStartEvent.GetPlace(), enc, nc)), f.BestStartEvent.GetCitations()).String())
		}
		// } else {
		// 	para.NewSentence(what)
		// }
	case model.FamilyBondLikelyMarried:
		what = "were likely married"
	case model.FamilyBondUnknown:
		if len(f.Children) == 1 {
			what = "had a child together, but we don't have any evidence that they married"
		} else {
			what = "had children together, but we don't have any evidence that they married"
		}
	default:
	}
	if f.Father.IsUnknown() {
		para.NewSentence(enc.EncodeModelLink(enc.EncodeText(f.Mother.PreferredFullName), f.Mother).String(), what)
	} else if f.Father.IsUnknown() {
		para.NewSentence(enc.EncodeModelLink(enc.EncodeText(f.Mother.PreferredFullName), f.Mother).String(), what)
	} else {
		para.NewSentence(enc.EncodeModelLink(enc.EncodeText(f.Father.PreferredFullName), f.Father).String(), "and", enc.EncodeModelLink(enc.EncodeText(f.Mother.PreferredFullName), f.Mother).String(), what)
	}

	enc.Para(enc.EncodeText(para.Text()))

	// Children
	if len(f.Children) > 0 {
		var detail text.Para

		if f.Father.IsUnknown() {
			if f.Mother.IsUnknown() {
				detail.NewSentence("They")
			} else {
				detail.NewSentence(f.Mother.PreferredGivenName)
			}
		} else {
			if f.Mother.IsUnknown() {
				detail.NewSentence(f.Father.PreferredGivenName)
			} else {
				detail.NewSentence("They")
			}
		}

		if len(f.Children) == 0 {
			// single parents already dealt with
			if (!f.Father.IsUnknown() && f.Father.Childless) || (!f.Mother.IsUnknown() && f.Mother.Childless) {
				detail.AddCompleteSentence("they had no children")
			}
		} else {

			childCardinal := narrative.ChildCardinal(f.Children)
			detail.Continue("had " + childCardinal)
		}

		childList := narrative.ChildList(f.Children, enc, nc)
		detail.FinishSentenceWithTerminator(":–")
		enc.Para(enc.EncodeText(detail.Text()))
		enc.UnorderedList(childList)
	}
	// Events

	// End of relationship

	// Later relationships of father

	// Later relationships of mother

	n := b.BuildFamilyNarrative(f)
	n.Render(enc)
	// enc.Image("/home/iand/Documents/genealogy/images/people/chambers-family/James Hall (1917), No 2.jpg", "James Hall (1917)")
	// enc.Figure("/home/iand/Documents/genealogy/images/people/chambers-family/James Hall (1917), No 2.jpg", "James Hall (1917)", "James Hall (1917)", nil)

	return nil
}

func (b *Book) BuildFamilyNarrative(f *model.Family) *narrative.FamilyNarrative[pandoc.Text] {
	n := &narrative.FamilyNarrative[pandoc.Text]{
		Family: f,
	}

	var timeline []model.TimelineEvent

	fintro := &narrative.IntroStatement[pandoc.Text]{
		Principal:           f.Father,
		SuppressRelation:    true,
		IncludeMedia:        true,
		CropMediaHighlights: true,
	}
	if !f.Father.IsUnknown() {
		n.FatherStatements = append(n.FatherStatements, fintro)
		timeline = append(timeline, f.Father.Timeline...)

		// If death is known, add it
		if f.Father.BestDeathlikeEvent != nil {
			n.FatherStatements = append(n.FatherStatements, &narrative.DeathStatement[pandoc.Text]{
				Principal:               f.Father,
				ExcludeSurvivingPartner: true,
				IncludeMedia:            true,
				CropMediaHighlights:     true,
			})
		}
		for _, pf := range f.Father.Families {
			if f.SameAs(pf) {
				continue
			}
			n.FatherStatements = append(n.FatherStatements, &narrative.FamilyStatement[pandoc.Text]{
				Principal: f.Father,
				Family:    pf,
			})
			if !f.BestEndDate.IsUnknown() && f.BestEndEvent != nil && !f.BestEndEvent.IsInferred() {
				n.FatherStatements = append(n.FatherStatements, &narrative.FamilyEndStatement[pandoc.Text]{
					Principal: f.Father,
					Family:    pf,
				})
			}
		}

	}

	mintro := &narrative.IntroStatement[pandoc.Text]{
		Principal:           f.Mother,
		SuppressRelation:    true,
		IncludeMedia:        true,
		CropMediaHighlights: true,
	}
	if !f.Mother.IsUnknown() {
		n.MotherStatements = append(n.MotherStatements, mintro)
		timeline = append(timeline, f.Mother.Timeline...)
		// If death is known, add it
		if f.Mother.BestDeathlikeEvent != nil {
			n.MotherStatements = append(n.MotherStatements, &narrative.DeathStatement[pandoc.Text]{
				Principal:               f.Mother,
				ExcludeSurvivingPartner: true,
				IncludeMedia:            true,
				CropMediaHighlights:     true,
			})
		}
		for _, pf := range f.Mother.Families {
			if f.SameAs(pf) {
				continue
			}
			n.MotherStatements = append(n.MotherStatements, &narrative.FamilyStatement[pandoc.Text]{
				Principal: f.Mother,
				Family:    pf,
			})
			if !f.BestEndDate.IsUnknown() && f.BestEndEvent != nil && !f.BestEndEvent.IsInferred() {
				n.MotherStatements = append(n.MotherStatements, &narrative.FamilyEndStatement[pandoc.Text]{
					Principal: f.Mother,
					Family:    pf,
				})
			}
		}
	}

	isFamilyEvent := func(ev model.TimelineEvent) bool {
		if !ev.DirectlyInvolves(f.Father) && !ev.DirectlyInvolves(f.Mother) {
			return false
		}
		if ev.GetDate().SortsBefore(f.BestStartDate) {
			return false
		}
		if f.BestEndDate.SortsBefore(ev.GetDate()) {
			return false
		}
		if f.Bond == model.FamilyBondUnmarried || f.Bond == model.FamilyBondLikelyUnmarried {
			// events are only family events for unmarried couples if they are both directly involved
			if !ev.DirectlyInvolves(f.Father) || !ev.DirectlyInvolves(f.Mother) {
				return false
			}
		}
		return true
	}

	seenSharedEvents := map[model.TimelineEvent]bool{}

	for _, ev := range timeline {
		switch tev := ev.(type) {
		case *model.BaptismEvent:
			if tev.DirectlyInvolves(f.Father) {
				fintro.Baptisms = append(fintro.Baptisms, tev)
			} else if tev.DirectlyInvolves(f.Mother) {
				mintro.Baptisms = append(mintro.Baptisms, tev)
			}
		case *model.CensusEvent:
			s := &narrative.CensusStatement[pandoc.Text]{
				Principal:           f.Father, // TODO: change
				Event:               tev,
				IncludeMedia:        true,
				CropMediaHighlights: true,
			}
			if isFamilyEvent(ev) {
				if !seenSharedEvents[ev] {
					n.FamilyStatements = append(n.FamilyStatements, s)
					seenSharedEvents[ev] = true
				}
			} else if tev.DirectlyInvolves(f.Father) {
				n.FatherStatements = append(n.FatherStatements, s)
			} else if tev.DirectlyInvolves(f.Mother) {
				n.MotherStatements = append(n.MotherStatements, s)
			}
		case *model.IndividualNarrativeEvent:
			s := &narrative.NarrativeStatement[pandoc.Text]{
				Principal: f.Father, // TODO: change
				Event:     tev,
			}
			if isFamilyEvent(ev) {
				if !seenSharedEvents[ev] {
					n.FamilyStatements = append(n.FamilyStatements, s)
					seenSharedEvents[ev] = true
				}
			} else if tev.DirectlyInvolves(f.Father) {
				n.FatherStatements = append(n.FatherStatements, s)
			} else if tev.DirectlyInvolves(f.Mother) {
				n.MotherStatements = append(n.MotherStatements, s)
			}
		case *model.BirthEvent:
		case *model.DeathEvent:
		case *model.BurialEvent:
		case *model.CremationEvent:
		default:
			if tev.GetNarrative().Text != "" {
				s := &narrative.NarrativeStatement[pandoc.Text]{
					Principal: f.Father, // TODO: change
					Event:     tev,
				}
				if isFamilyEvent(ev) {
					if !seenSharedEvents[ev] {
						n.FamilyStatements = append(n.FamilyStatements, s)
						seenSharedEvents[ev] = true
					}
				} else if tev.DirectlyInvolves(f.Father) {
					n.FatherStatements = append(n.FatherStatements, s)
				} else if tev.DirectlyInvolves(f.Mother) {
					n.MotherStatements = append(n.MotherStatements, s)
				}
			}

		}
	}

	return n
}

func PersonIntro[T render.EncodedText](p *model.Person, enc render.TextEncoder[T], nc narrative.NameChooser, name T) T {
	if !name.IsZero() {
		name = enc.EncodeModelLink(name, p)

		if p.NickName != "" {
			name = enc.EncodeText(text.JoinSentenceParts(name.String(), fmt.Sprintf("(known as %s)", p.NickName)))
		}
	}

	var para text.Para
	birth := narrative.PersonBirthSummary(p, enc, nc, name, true, true, true, false)
	if !birth.IsZero() {
		para.NewSentence(birth.String())
		name = enc.EncodeText(p.Gender.SubjectPronoun())
	}
	return enc.EncodeText(para.Text())
}

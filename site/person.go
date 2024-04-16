package site

import (
	"fmt"
	"sort"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
)

func RenderPersonPage(s *Site, p *model.Person) (*md.Document, error) {
	pov := &model.POV{Person: p}

	doc := s.NewDocument()
	doc.Layout(PageLayoutPerson.String())
	doc.Category(PageCategoryPerson)
	doc.ID(p.ID)
	doc.Title(p.PreferredUniqueName)
	if p.Redacted {
		doc.Summary("information withheld to preserve privacy")
		return doc, nil
	}

	doc.SetFrontMatterField("gender", p.Gender.Noun())
	if s.GenerateWikiTree {
		if l := s.LinkForFormat(p, "wikitree"); l != "" {
			doc.SetFrontMatterField("wikitreeformat", l)
		}
	}

	if p.Olb != "" {
		doc.Summary(p.Olb)
	}
	doc.AddTags(CleanTags(p.Tags))

	// determine the feature image
	if !p.Gender.IsUnknown() {
		birthYear, hasBirthYear := p.BestBirthDate().Year()
		deathYear, hasDeathYear := p.BestDeathDate().Year()
		if hasBirthYear || hasDeathYear {
			eraYear := 0
			age := 0
			if hasBirthYear {
				eraYear = birthYear + 18
				if hasDeathYear {
					age = deathYear - birthYear
					if age < 18 {
						eraYear = birthYear
					}

					if age < 3 {
						doc.SetFrontMatterField("maturity", "infant")
					} else if age < 15 {
						doc.SetFrontMatterField("maturity", "child")
					} else if age < 25 {
						doc.SetFrontMatterField("maturity", "young")
					} else if age < 55 {
						doc.SetFrontMatterField("maturity", "mature")
					} else {
						doc.SetFrontMatterField("maturity", "old")
					}
				}
			} else {
				eraYear = deathYear
			}

			if eraYear != 0 {
				if eraYear < 1700 {
					doc.SetFrontMatterField("era", "1600s")
				} else if eraYear < 1800 {
					doc.SetFrontMatterField("era", "1700s")
				} else if eraYear < 1900 {
					doc.SetFrontMatterField("era", "1800s")
				} else if eraYear < 2000 {
					doc.SetFrontMatterField("era", "1900s")
				} else {
					doc.SetFrontMatterField("era", "modern")
				}
			}

		}
	}

	// Render narrative
	n := &Narrative{
		Statements: make([]Statement, 0),
	}

	// Everyone has an intro
	intro := &IntroStatement{
		Principal: p,
	}
	for _, ev := range p.Timeline {
		switch tev := ev.(type) {
		case *model.BaptismEvent:
			if tev.DirectlyInvolves(p) {
				intro.Baptisms = append(intro.Baptisms, tev)
			}
		case *model.CensusEvent:
			if tev.DirectlyInvolves(p) {
				n.Statements = append(n.Statements, &CensusStatement{
					Principal: p,
					Event:     tev,
				})
			}
		}
	}
	if len(intro.Baptisms) > 0 {
		sort.Slice(intro.Baptisms, func(i, j int) bool {
			return intro.Baptisms[i].GetDate().SortsBefore(intro.Baptisms[j].GetDate())
		})
	}
	n.Statements = append(n.Statements, intro)

	// If death is known, add it
	if p.BestDeathlikeEvent != nil {
		n.Statements = append(n.Statements, &DeathStatement{
			Principal: p,
		})
	}

	for _, f := range p.Families {
		n.Statements = append(n.Statements, &FamilyStatement{
			Principal: p,
			Family:    f,
		})
	}

	n.Render(pov, doc)

	if p.EditLink != nil {
		doc.SetFrontMatterField("editlink", p.EditLink.URL)
		doc.SetFrontMatterField("editlinktitle", p.EditLink.Title)
	}

	if p.WikiTreeID != "" {
		doc.SetFrontMatterField("wikitreeid", p.WikiTreeID)
	}

	if p.GrampsID != "" {
		doc.SetFrontMatterField("grampsid", p.GrampsID)
	}

	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(p.Timeline)),
	}
	for _, ev := range p.Timeline {
		if ev.GetDate().IsUnknown() && ev.GetPlace().IsUnknown() {
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category: ev.GetTitle(),
				Detail:   ev.GetDetail(),
			})
		} else {
			t.Events = append(t.Events, ev)
		}
	}

	if len(p.Timeline) > 0 {
		doc.EmptyPara()
		doc.Heading2("Timeline")

		doc.ResetSeenLinks()
		if err := RenderTimeline(t, pov, doc); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.MiscFacts) > 0 {
		doc.EmptyPara()
		doc.Heading2("Other Information")
		if err := RenderFacts(p.MiscFacts, pov, doc); err != nil {
			return nil, fmt.Errorf("render facts: %w", err)
		}
	}

	links := make([]string, 0, len(p.Links))
	for _, l := range p.Links {
		links = append(links, doc.EncodeLink(l.Title, l.URL))
	}

	if len(links) > 0 {
		doc.Heading2("Links")
		doc.UnorderedList(links)
	}

	if len(p.ResearchNotes) > 0 {
		doc.Heading2("Research Notes")
		mentions := make([]*model.Note, 0)
		for _, n := range p.ResearchNotes {
			if !n.PrimaryPerson.SameAs(p) {
				mentions = append(mentions, n)
				continue
			}
			doc.Heading3(n.Title)

			byline := ""
			if n.Author != "" {
				byline = "by " + n.Author
			}
			if n.Date != "" {
				if byline != "" {
					byline += ", "
				}
				byline += n.Date
			}
			if byline != "" {
				doc.Para(doc.EncodeItalic("Written " + byline))
				doc.EmptyPara()
			}
			doc.RawMarkdown(n.Markdown)
		}

		if len(mentions) > 0 {
			doc.Heading3("Mentioned in the following notes:")
			ss := make([]string, 0, len(mentions))
			for _, n := range mentions {
				ss = append(ss, doc.EncodeModelLink(n.Title, n.PrimaryPerson))
			}
			doc.UnorderedList(ss)
		}

	}

	return doc, nil
}

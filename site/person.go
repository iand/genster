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
	doc.Section(md.PageLayoutPerson)
	doc.ID(p.ID)
	doc.Title(p.PreferredUniqueName)
	doc.SetFrontMatterField("gender", p.Gender.Noun())

	if p.Redacted {
		doc.Summary("information withheld to preserve privacy")
		return doc, nil
	}

	if p.Olb != "" {
		doc.Summary(p.Olb)
	}
	doc.AddTags(CleanTags(p.Tags))

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
			if tev != p.BestBirthlikeEvent {
				intro.Baptisms = append(intro.Baptisms, tev)
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

	return doc, nil
}

package site

import (
	"fmt"
	"sort"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
)

func RenderPersonPage(s *Site, p *model.Person) (*md.Document, error) {
	pov := &model.POV{Person: p}

	d := s.NewDocument()
	d.Section(md.PageLayoutPerson)
	d.ID(p.ID)
	d.Title(p.PreferredUniqueName)
	d.SetFrontMatterField("gender", p.Gender.Noun())

	if p.Redacted {
		d.Summary("information withheld to preserve privacy")
		return d, nil
	}

	if p.Olb != "" {
		d.Summary(p.Olb)
	}
	d.AddTags(CleanTags(p.Tags))

	b := d.Body()

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

	n.Render(pov, b)

	if p.EditLink != nil {
		d.SetFrontMatterField("editlink", p.EditLink.URL)
		d.SetFrontMatterField("editlinktitle", p.EditLink.Title)
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
		b.EmptyPara()
		b.Heading2("Timeline")

		b.ResetSeenLinks()
		if err := RenderTimeline(t, pov, b); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.MiscFacts) > 0 {
		b.EmptyPara()
		b.Heading2("Other Information")
		if err := RenderFacts(p.MiscFacts, pov, b); err != nil {
			return nil, fmt.Errorf("render facts: %w", err)
		}
	}

	links := make([]string, 0, len(p.Links))
	for _, l := range p.Links {
		links = append(links, b.EncodeLink(l.Title, l.URL))
	}

	if len(links) > 0 {
		b.Heading2("Links")
		b.UnorderedList(links)
	}

	return d, nil
}

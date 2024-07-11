package site

import (
	"fmt"
	"sort"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
)

func RenderPersonPage(s *Site, p *model.Person) (render.Page, error) {
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

	if p.FeatureImage != nil {
		doc.Image(s.LinkFor(p.FeatureImage.Object))
	}

	switch p.Gender {
	case model.GenderMale:
		doc.SetFrontMatterField("gender", "male")
	case model.GenderFemale:
		doc.SetFrontMatterField("gender", "female")
	default:
		doc.SetFrontMatterField("gender", "unknown")
	}
	if s.GenerateWikiTree {
		if l := s.LinkForFormat(p, "wikitree"); l != "" {
			doc.SetFrontMatterField("wikitreeformat", l)
		}
	}

	// if p.Olb != "" {
	// 	doc.Summary(p.Olb)
	// }
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

					if age < 14 {
						doc.SetFrontMatterField("maturity", "child")
					} else if age < 30 {
						doc.SetFrontMatterField("maturity", "young")
					} else if age < 70 {
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

			switch p.OccupationGroup {
			case model.OccupationGroupLabouring:
				doc.SetFrontMatterField("trade", "labourer")
			case model.OccupationGroupIndustrial:
				doc.SetFrontMatterField("trade", "miner")
			case model.OccupationGroupMaritime:
				doc.SetFrontMatterField("trade", "nautical")
			case model.OccupationGroupCrafts:
				doc.SetFrontMatterField("trade", "crafts")
			case model.OccupationGroupClerical:
				doc.SetFrontMatterField("trade", "clerical")
			case model.OccupationGroupCommercial:
				doc.SetFrontMatterField("trade", "commerical")
			case model.OccupationGroupMilitary:
				doc.SetFrontMatterField("trade", "military")
			case model.OccupationGroupService:
				doc.SetFrontMatterField("trade", "service")

			}

			for i := len(p.Occupations) - 1; i >= 0; i-- {
			}

		}
	}

	// if p.Olb != "" {
	// 	doc.Para(doc.EncodeBold(doc.EncodeItalic(text.FormatSentence(p.Olb))))
	// }

	ap := p.RelationToKeyPerson.Path()
	if len(ap) > 2 {
		doc.AddDescendant(ap[len(ap)-1].PreferredUniqueName, "", p.Olb)
		for i := len(ap) - 2; i >= 0; i-- {
			if ap[i].Redacted {
				break
			}
			doc.AddDescendant(ap[i].PreferredUniqueName, doc.LinkBuilder.LinkFor(ap[i]), "")
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
		case *model.IndividualNarrativeEvent:
			n.Statements = append(n.Statements, &NarrativeStatement{
				Principal: p,
				Event:     tev,
			})
		case *model.BirthEvent:
		case *model.DeathEvent:
		case *model.BurialEvent:
		case *model.CremationEvent:
		default:
			if tev.DirectlyInvolves(p) && tev.GetNarrative().Text != "" {
				n.Statements = append(n.Statements, &NarrativeStatement{
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
		if !f.BestEndDate.IsUnknown() && f.BestEndEvent != nil && !f.BestEndEvent.IsInferred() {
			n.Statements = append(n.Statements, &FamilyEndStatement{
				Principal: p,
				Family:    f,
			})
		}
	}

	n.Render(pov, doc)

	for _, l := range p.Links {
		doc.AddLink(l.Title, l.URL)
	}

	if p.GrampsID != "" {
		doc.SetFrontMatterField("grampsid", p.GrampsID)
		doc.AddAlias(s.RedirectPath(p.GrampsID))
	}

	if p.Slug != "" {
		doc.AddAlias(s.RedirectPath(p.Slug))
	}

	if len(p.Comments) > 0 {
		doc.Heading3("Comments", "")
		for _, t := range p.Comments {
			RenderText(t, doc)
		}
	}

	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(p.Timeline)),
	}
	for _, ev := range p.Timeline {
		if ev.GetDate().IsUnknown() && ev.GetPlace().IsUnknown() {
			p.MiscFacts = append(p.MiscFacts, model.Fact{
				Category: ev.What(),
				Detail:   ev.GetDetail(),
			})
		} else {
			t.Events = append(t.Events, ev)
		}
	}

	if len(p.Timeline) > 0 {
		doc.EmptyPara()
		doc.Heading2("Timeline", "")

		doc.ResetSeenLinks()
		if err := RenderTimeline(t, pov, doc); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.MiscFacts) > 0 || len(p.KnownNames) > 1 {
		doc.Heading2("Other Information", "")
		if len(p.KnownNames) > 1 {
			doc.EmptyPara()
			doc.Para("Other names and variations")
			if err := RenderNames(p.KnownNames, doc); err != nil {
				return nil, fmt.Errorf("render names: %w", err)
			}
		}
		doc.EmptyPara()
		if err := RenderFacts(p.MiscFacts, pov, doc); err != nil {
			return nil, fmt.Errorf("render facts: %w", err)
		}
	}

	if len(p.ResearchNotes) > 0 {
		doc.Heading2("Research Notes", "")
		for _, t := range p.ResearchNotes {
			RenderText(t, doc)
		}
	}

	return doc, nil
}

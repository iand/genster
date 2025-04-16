package site

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/iand/genster/debug"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
)

func RenderPersonPage(s *Site, p *model.Person) (render.Document[md.Text], error) {
	pov := &model.POV{Person: p}

	if p.BestBirthlikeEvent != nil {
		pl := p.BestBirthlikeEvent.GetPlace()
		if !pl.IsUnknown() {
			pov.Place = pl.Country
		}
	}

	doc := s.NewDocument()
	doc.Layout(PageLayoutPerson.String())
	doc.Category(PageCategoryPerson)
	doc.SetSitemapDisable()
	if p.UpdateTime != nil {
		doc.LastUpdated(*p.UpdateTime)
	}
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
				doc.SetFrontMatterField("trade", "commercial")
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

	if p.Intro != nil {
		narrative.RenderText(*p.Intro, doc)
	}

	// Render narrative
	n := &narrative.PersonNarrative[md.Text]{
		Statements: make([]narrative.Statement[md.Text], 0),
	}

	// Everyone has an intro
	intro := &narrative.IntroStatement[md.Text]{
		Principal: p,
	}
	death := &narrative.DeathStatement[md.Text]{
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
				n.Statements = append(n.Statements, &narrative.CensusStatement[md.Text]{
					Principal: p,
					Event:     tev,
				})
			}
		case *model.IndividualNarrativeEvent:
			n.Statements = append(n.Statements, &narrative.NarrativeStatement[md.Text]{
				Principal: p,
				Event:     tev,
			})
		case *model.BirthEvent:
		case *model.DeathEvent:
		case *model.BurialEvent:
		case *model.CremationEvent:
		case *model.PossibleBirthEvent:
			intro.PossibleBirths = append(intro.PossibleBirths, tev)
		case *model.PossibleDeathEvent:
			death.PossibleDeaths = append(death.PossibleDeaths, tev)
		default:
			if tev.DirectlyInvolves(p) && tev.GetNarrative().Text != "" {
				n.Statements = append(n.Statements, &narrative.NarrativeStatement[md.Text]{
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
	n.Statements = append(n.Statements, death)

	for _, f := range p.Families {
		n.Statements = append(n.Statements, &narrative.FamilyStatement[md.Text]{
			Principal: p,
			Family:    f,
		})
		if !f.BestEndDate.IsUnknown() && f.BestEndEvent != nil && !f.BestEndEvent.IsInferred() {
			n.Statements = append(n.Statements, &narrative.FamilyEndStatement[md.Text]{
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
			narrative.RenderText(t, doc)
		}
	}

	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(p.Timeline)),
	}
	for _, ev := range p.Timeline {
		if !ev.GetDate().IsUnknown() || !ev.GetPlace().IsUnknown() {
			t.Events = append(t.Events, ev)
		}
	}

	if len(p.Timeline) > 0 {
		doc.EmptyPara()
		doc.Heading2("Timeline", "")

		doc.ResetSeenLinks()

		fmtr := &NarrativeTimelineEntryFormatter[md.Text]{
			pov:    pov,
			enc:    doc,
			logger: logging.With("id", p.ID),
			nc:     &narrative.TimelineNameChooser{},
		}

		if err := RenderTimeline(t, pov, doc, fmtr); err != nil {
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
			narrative.RenderText(t, doc)
		}
	}

	if s.IncludeDebugInfo {
		var buf bytes.Buffer
		debug.DumpPerson(p, &buf)
		doc.Comment(buf.String())
	}

	return doc, nil
}

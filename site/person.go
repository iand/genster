package site

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/iand/genster/debug"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/text"
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

	if p.RelationToKeyPerson != nil && p.RelationToKeyPerson.IsDirectAncestor() {
		doc.SetFrontMatterField("ancestor", "true")
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

	olb := p.Olb
	// if olb == "" && s.IncludeDebugInfo {
	// 	olb = narrative.GenerateOlb(p)
	// }
	if olb != "" {
		doc.Para(doc.EncodeBold(doc.EncodeItalic(doc.EncodeText(text.FormatSentence(olb)))))
	}

	ap := p.RelationToKeyPerson.Path()
	if len(ap) > 2 {
		doc.AddDescendant(ap[len(ap)-1].PreferredUniqueName, "", p.Epithet)
		for i := len(ap) - 2; i >= 0; i-- {
			if ap[i].Redacted {
				break
			}
			doc.AddDescendant(ap[i].PreferredUniqueName, doc.LinkBuilder.LinkFor(ap[i]), "")
		}
	}

	var storySubjects, storyMentions, questionSubjects, questionMentions []model.Link
	for _, l := range p.Links {
		switch l.Category {
		case model.LinkCategoryStorySubject:
			storySubjects = append(storySubjects, l)
		case model.LinkCategoryStoryMention:
			storyMentions = append(storyMentions, l)
		case model.LinkCategoryQuestionSubject:
			questionSubjects = append(questionSubjects, l)
		case model.LinkCategoryQuestionMention:
			questionMentions = append(questionMentions, l)
		default:
			doc.AddLink(l.Title, l.URL, l.Category)
		}
	}
	writeContentLinksPara(doc, p.PreferredGivenName, storySubjects, storyMentions, questionSubjects, questionMentions)

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
			if tev.IsParticipant(p) {
				intro.Baptisms = append(intro.Baptisms, tev)
			}
		case *model.CensusEvent:
			if tev.IsParticipant(p) {
				n.Statements = append(n.Statements, &narrative.CensusStatement[md.Text]{
					Principal: p,
					Event:     tev,
				})
			}
		case *model.IndividualNarrativeEvent:
			n.Statements = append(n.Statements, &narrative.GeneralEventStatement[md.Text]{
				Principal: p,
				Event:     tev,
			})
		case *model.BirthEvent:
		case *model.DeathEvent:
		case *model.BurialEvent:
		case *model.CremationEvent:
		case *model.PhysicalDescriptionEvent:
			n.Statements = append(n.Statements, &narrative.GeneralEventStatement[md.Text]{
				Principal: p,
				Event:     tev,
			})
		case *model.PossibleBirthEvent:
			intro.PossibleBirths = append(intro.PossibleBirths, tev)
		case *model.PossibleDeathEvent:
			death.PossibleDeaths = append(death.PossibleDeaths, tev)
		default:
			if tev.DirectlyInvolves(p) && tev.GetNarrative().Text != "" {
				n.Statements = append(n.Statements, &narrative.GeneralEventStatement[md.Text]{
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
		if s.IncludeDebugInfo {
			doc.Para(doc.EncodeModelLink("family", f))
		}
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

	if p.GrampsID != "" {
		doc.SetFrontMatterField("grampsid", p.GrampsID)
		doc.AddAlias(s.RedirectPath(p.GrampsID))
	}

	if p.Slug != "" {
		doc.SetFrontMatterField("slug", p.Slug)
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

		fmtr := narrative.NewNarrativeTimelineEntryFormatter[md.Text](pov, doc, logging.With("id", p.ID), &narrative.TimelineNameChooser{}, false)

		if err := RenderTimeline(t, pov, doc, fmtr); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.MiscFacts) > 0 || len(p.KnownNames) > 1 {
		doc.Heading2("Facts", "")
		if err := RenderFacts(p, pov, doc); err != nil {
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

// linkTitles converts a slice of model.Link into a slice of encoded markdown
// hyperlink strings for use in prose sentences.
func linkTitles(doc *md.Document, links []model.Link) []string {
	titles := make([]string, len(links))
	for i, l := range links {
		titles[i] = string(doc.EncodeLink(md.Text(l.Title), l.URL))
	}
	return titles
}

// writeContentLinksPara adds a paragraph to doc describing which stories and
// questions the person is the subject of, and which they are merely mentioned
// in. Links are rendered as markdown hyperlinks inline in the sentence.
//
// The general form is:
//
//	"{name} is the subject of [questions] and [stories] and is also mentioned in [questions and stories]."
func writeContentLinksPara(doc *md.Document, firstname string, storySubjects, storyMentions, questionSubjects, questionMentions []model.Link) {
	hasSubjects := len(storySubjects) > 0 || len(questionSubjects) > 0
	hasMentions := len(storyMentions) > 0 || len(questionMentions) > 0
	if !hasSubjects && !hasMentions {
		return
	}

	var parts []string

	if hasSubjects {
		var subjectParts []string
		if len(storySubjects) > 0 {
			titles := linkTitles(doc, storySubjects)
			noun := "stories"
			if len(storySubjects) == 1 {
				noun = "story"
			}
			subjectParts = append(subjectParts, text.JoinList(titles)+" "+noun)
		}
		if len(questionSubjects) > 0 {
			titles := linkTitles(doc, questionSubjects)
			noun := "open questions"
			if len(questionSubjects) == 1 {
				noun = "open question"
			}
			subjectParts = append(subjectParts, text.JoinList(titles)+" "+noun)
		}
		parts = append(parts, "is the subject of the "+text.JoinList(subjectParts))
	}

	if hasMentions {
		var mentionParts []string
		prefix := "is mentioned in the"
		if hasSubjects {
			prefix = "is also mentioned in the"
		}

		if len(storyMentions) > 0 {
			titles := linkTitles(doc, storyMentions)
			noun := "stories"
			if len(storyMentions) == 1 {
				noun = "story"
			}
			mentionParts = append(mentionParts, text.JoinList(titles)+" "+noun)
		}

		if len(questionMentions) > 0 {
			titles := linkTitles(doc, questionMentions)
			noun := "stories"
			if len(questionMentions) == 1 {
				noun = "story"
			}
			mentionParts = append(mentionParts, text.JoinList(titles)+" "+noun)
		}

		parts = append(parts, prefix+" "+text.JoinList(mentionParts))
	}

	sentence := firstname + " " + strings.Join(parts, " and ") + "."
	doc.Preface(md.Text(sentence))
}

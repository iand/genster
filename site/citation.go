package site

import (
	"slices"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/text"
)

func RenderCitationPage(s *Site, c *model.GeneralCitation) (render.Page[md.Text], error) {
	doc := s.NewDocument()
	doc.SuppressCitations = true

	doc.Layout(PageLayoutCitation.String())
	doc.Category(PageCategoryCitation)
	if c.LastUpdated != nil {
		doc.LastUpdated(*c.LastUpdated)
	}
	doc.ID(c.ID)

	if c.GrampsID != "" {
		doc.SetFrontMatterField("grampsid", c.GrampsID)
		doc.AddAlias(s.RedirectPath(c.GrampsID))
	}

	title := c.Detail
	if title == "" {
		if c.Source != nil && c.Source.Title != "" {
			title = c.Source.Title
		} else {
			title = "Citation"
		}
	}
	doc.Title(title)

	sourceDesc := "unknown source"

	if c.Detail != "" && c.Source != nil && c.Source.Title != "" {
		sourceDesc = c.Source.Title
		if c.Source.Author != "" {
			sourceDesc += " (" + c.Source.Author + ")"
		}
		doc.Para(doc.EncodeText("Cited from " + sourceDesc))
	}

	for _, cmo := range c.MediaObjects {
		link := s.LinkFor(cmo.Object)
		if link != "" {
			doc.EmptyPara()
			doc.Figure(link, cmo.Object.ID, doc.EncodeText(cmo.Object.Title), cmo.Highlight)
		}
	}

	if len(c.TranscriptionText) > 0 {
		if len(c.TranscriptionText) == 1 {
			doc.Heading3("Transcription", "")
		} else {
			doc.Heading3("Transcriptions", "")
		}
		for _, t := range c.TranscriptionText {
			if t.Formatted {
				doc.Pre(t.Text)
				// doc.Pre("")
			} else if t.Markdown {
				doc.Markdown(t.Text)
				// doc.Pre("")
			} else {
				doc.BlockQuote(doc.EncodeText(t.Text))
				// doc.BlockQuote("")
			}
		}
		if !c.TranscriptionDate.IsUnknown() {
			doc.BlockQuote(doc.EncodeText("-- transcribed " + c.TranscriptionDate.When()))
		}
	}

	if len(c.Comments) > 0 {
		doc.Heading3("Comments", "")
		for _, t := range c.Comments {
			RenderText(t, doc)
		}
	}

	peopleInCitations := make(map[*model.Person]bool)

	var cites string

	events := make([]md.Text, 0, len(c.EventsCited))
	for _, ev := range c.EventsCited {
		if !IncludeInTimeline(ev) {
			continue
		}

		events = append(events, md.Text(WhoWhatWhenWhere(ev, doc, FullNameChooser{})))
		for _, p := range ev.GetParticipants() {
			peopleInCitations[p.Person] = true
		}
	}

	people := make([]md.Text, 0, len(c.PeopleCited))
	for _, p := range c.PeopleCited {
		if peopleInCitations[p] {
			continue
		}
		people = append(people, doc.EncodeModelLink(doc.EncodeText(p.PreferredFullName), p))
	}

	if len(events) > 0 || len(people) > 0 {
		doc.Heading3("Other Information", "other")

		if len(events) > 0 {
			if len(c.EventsCited) == 1 {
				cites = "one event has been derived from this evidence:"
			} else {
				cites = text.JoinSentenceParts(text.SmallCardinalNoun(len(c.EventsCited)), "events have been derived from this evidence:")
			}
			doc.EmptyPara()
			doc.Para(doc.EncodeText(text.FormatSentence(cites)))

			doc.UnorderedList(events)
		}

		if len(people) > 0 {
			var peopleIntro string
			var otherClause string
			if len(peopleInCitations) > 0 {
				otherClause = "other"
			}
			var evidenceClause string
			if len(cites) == 0 {
				evidenceClause = "in this evidence"
			}

			// peopleIntro = text.JoinSentenceParts("no", otherClause, "people are mentioned", evidenceClause)
			if len(people) == 1 {
				peopleIntro = text.JoinSentenceParts("one", otherClause, "person is mentioned", evidenceClause)
			} else {
				peopleIntro = text.JoinSentenceParts(text.SmallCardinalNoun(len(people)), otherClause, "people are mentioned", evidenceClause, ":")
			}

			doc.EmptyPara()
			doc.Para(doc.EncodeText(text.UpperFirst(peopleIntro)))
			slices.Sort(people)
			doc.UnorderedList(people)
		}

	}

	doc.Heading3("Full Citation", "")
	doc.Para(doc.EncodeText(text.FinishSentence(c.String())))

	repos := make([]md.Text, 0)
	if c.URL != nil {
		repos = append(repos, doc.EncodeLink(doc.EncodeText(c.URL.Title), c.URL.URL))
	}

	if c.Source != nil && len(c.Source.RepositoryRefs) > 0 {
		for _, rr := range c.Source.RepositoryRefs {
			// 	rr := c.Source.RepositoryRefs[0]

			s := ""
			if rr.Repository.ShortName != "" {
				s += rr.Repository.ShortName
			} else {
				s += rr.Repository.Name
			}
			if rr.CallNo != "" {
				s += ". " + rr.CallNo
			}

			if s != "" {
				repos = append(repos, doc.EncodeText(s))
			}
		}
	}

	if len(repos) > 0 {
		doc.Heading3("Source", "")
		doc.Para(doc.EncodeText(sourceDesc))
		doc.Para("Available at:")
		doc.UnorderedList(repos)
	}

	if len(c.ResearchNotes) > 0 {
		doc.Heading2("Research Notes", "")
		for _, t := range c.ResearchNotes {
			RenderText(t, doc)
		}
	}

	return doc, nil
}

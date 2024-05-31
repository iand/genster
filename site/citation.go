package site

import (
	"slices"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

func RenderCitationPage(s *Site, c *model.GeneralCitation) (*md.Document, error) {
	doc := s.NewDocument()
	doc.SuppressCitations = true

	doc.Layout(PageLayoutCitation.String())
	doc.Category(PageCategoryCitation)
	doc.ID(c.ID)

	title := c.Detail
	if title == "" {
		if c.Source != nil && c.Source.Title != "" {
			title = c.Source.Title
		} else {
			title = "Citation"
		}
	}
	doc.Title(title)

	if c.Detail != "" && c.Source != nil && c.Source.Title != "" {
		doc.Para("Cited from " + c.Source.Title)
	}

	for _, mo := range c.MediaObjects {
		link := s.LinkFor(mo)
		if link != "" {
			doc.EmptyPara()
			doc.Image(mo.ID, link)
		}
	}

	if len(c.TranscriptionText) > 0 {
		if len(c.TranscriptionText) == 1 {
			doc.Heading3("Transcription")
		} else {
			doc.Heading3("Transcriptions")
		}
		for _, t := range c.TranscriptionText {
			if t.Formatted {
				doc.Pre(t.Text)
				doc.Pre("")
			} else {
				doc.BlockQuote(t.Text)
				doc.BlockQuote("")

			}
		}
		if !c.TranscriptionDate.IsUnknown() {
			doc.BlockQuote("-- transcribed " + c.TranscriptionDate.When())
		}
	}

	if len(c.Notes) > 0 {
		if len(c.Notes) == 1 {
			doc.Heading3("Note")
		} else {
			doc.Heading3("Note")
		}
		for _, t := range c.Notes {
			if t.Formatted {
				doc.Pre(t.Text)
				doc.Pre("")
			} else {
				doc.Para(t.Text)
				doc.EmptyPara()
			}
		}
	}

	peopleInCitations := make(map[*model.Person]bool)

	var cites string

	events := make([]string, 0, len(c.EventsCited))
	for _, ev := range c.EventsCited {
		events = append(events, WhoWhatWhenWhere(ev, doc))
		for _, p := range ev.Participants() {
			peopleInCitations[p.Person] = true
		}
	}

	people := make([]string, 0, len(c.PeopleCited))
	for _, p := range c.PeopleCited {
		if peopleInCitations[p] {
			continue
		}
		people = append(people, doc.EncodeModelLink(p.PreferredFullName, p))
	}

	if len(events) > 0 || len(people) > 0 {
		doc.Heading3("Other Information")

		if len(events) > 0 {
			if len(c.EventsCited) == 1 {
				cites = "one event has been derived from this evidence:"
			} else {
				cites = text.JoinSentenceParts(text.SmallCardinalNoun(len(c.EventsCited)), "events have been derived from this evidence:")
			}
			doc.EmptyPara()
			doc.Para(text.FormatSentence(cites))

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
			doc.Para(text.UpperFirst(peopleIntro))
			slices.Sort(people)
			doc.UnorderedList(people)
		}

	}

	doc.Heading3("Full Citation")
	doc.Para(text.FinishSentence(c.String()))

	if c.Source != nil && len(c.Source.RepositoryRefs) > 0 {
		repos := make([]string, 0, len(c.Source.RepositoryRefs))
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
				repos = append(repos, s)
			}
		}

		if len(repos) > 0 {
			doc.Heading3("Source")
			doc.Para(c.Source.Title + " available at:")
			doc.UnorderedList(repos)
		}
	}

	return doc, nil
}

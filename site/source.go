package site

import (
	// "fmt"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

func RenderSourcePage(s *Site, so *model.Source) (*md.Document, error) {
	doc := s.NewDocument()
	doc.SuppressCitations = true

	doc.Title(so.Title)
	doc.Layout(PageLayoutSource.String())
	doc.Category(PageCategorySource)
	doc.ID(so.ID)
	doc.AddTags(CleanTags(so.Tags))

	if so.RepositoryName != "" {
		para := text.JoinSentenceParts("provided by", doc.EncodeLink(so.RepositoryName, so.RepositoryLink))
		doc.Para(text.FormatSentence(para))
	}

	if so.SearchLink != "" {
		para := text.JoinSentenceParts("It can be ", doc.EncodeLink("searched online", so.SearchLink))
		doc.Para(text.FormatSentence(para))
	}

	const maxEvents = 20

	var cites string
	if len(so.EventsCiting) == 0 {
		cites = "no evidence cites this source."
	} else if len(so.EventsCiting) == 1 {
		cites = "one piece of evidence cites this source:"
	} else {
		cites = text.JoinSentenceParts(text.SmallCardinalNoun(len(so.EventsCiting)), "pieces of evidence cite this source")
		if len(so.EventsCiting) > maxEvents {
			cites = text.JoinSentenceParts(cites, ", some of which are:")
		} else {
			cites += ":"
		}
	}
	doc.EmptyPara()
	doc.Para(text.UpperFirst(cites))

	fmtr := &TimelineEntryFormatter{
		pov: &model.POV{},
		enc: doc,
	}

	events := make([][2]string, 0, len(so.EventsCiting))
	for i, ev := range so.EventsCiting {
		if i == maxEvents {
			break
		}
		citation := ""
		for _, c := range ev.GetCitations() {
			if c.Source != so {
				continue
			}
			citation = c.Detail
			break
		}

		title := fmtr.Title(i, ev)
		detail := ""
		if citation != "" {
			detail = "cites " + citation
		}
		events = append(events, [2]string{
			title,
			detail,
		})
	}

	doc.DefinitionList(events)

	return doc, nil
}

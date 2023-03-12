package site

import (
	"net/url"
	"strings"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

type ExtendedMarkdownBuilder interface {
	ExtendedMarkdownEncoder
	MarkdownBuilder
}

type MarkdownBuilder interface {
	MarkdownEncoder
	Markdown() string
	Para(s string)
	EmptyPara()
	Heading2(s string)
	Heading3(s string)
	Heading4(s string)
	UnorderedList(items []string)
	OrderedList(items []string)
	DefinitionList(items [][2]string)
	BlockQuote(s string)
}

type InlineBuilder interface {
	InlineEncoder
	Markdown() string
}

type MarkdownEncoder interface {
	InlineEncoder
	EncodePara(s string) string
	EncodeEmptyPara() string
	EncodeHeading4(s string) string
	EncodeUnorderedList(items []string) string
	EncodeOrderedList(items []string) string
	EncodeDefinitionList(items [][2]string) string
	EncodeBlockQuote(s string) string
}

type InlineEncoder interface {
	EncodeLink(text string, url string) string
	EncodeModelLink(text string, m any) string
	EncodeItalic(s string) string
	EncodeBold(s string) string
}

type ExtendedMarkdownEncoder interface {
	MarkdownEncoder
	EncodeModelLinkDedupe(firstText string, subsequentText string, m any) string
	EncodeCitation(citation string, detail string, citationID string) string
}

type ExtendedInlineEncoder interface {
	InlineEncoder
	EncodeModelLinkDedupe(firstText string, subsequentText string, m any) string
	EncodeCitation(citation string, detail string, citationID string) string
}

func EncodeWithCitations(s string, citations []*model.GeneralCitation, enc ExtendedMarkdownEncoder) string {
	for i, cit := range citations {
		if i > 0 {
			s += "<sup>,</sup>"
		}
		s += EncodeCitationDetail(cit, enc)
	}
	return s
}

func EncodeCitationDetail(c *model.GeneralCitation, enc ExtendedMarkdownEncoder) string {
	var heading string
	var detail string

	if c.Source != nil && c.Source.Title != "" {
		heading = c.Source.Title
		if c.Detail != "" {
			if !strings.HasSuffix(heading, ".") && !strings.HasSuffix(heading, "!") && !strings.HasSuffix(heading, "?") {
				heading += "; "
			}
			heading += enc.EncodePara(cleanCitationDetail(c.Detail))
		}
		heading = text.FinishSentence(heading)
		heading += " (" + enc.EncodeModelLink("source", c.Source) + ")"

		// heading = enc.EncodeModelLink(text.FinishSentence(c.Source.Title), c.Source, false)
		// if c.Detail != "" {
		// 	detail = enc.EncodePara(cleanCitationDetail(c.Detail))
		// }
	} else {
		heading = cleanCitationDetail(c.Detail)
		detail = ""
	}

	if c.URL != nil {
		detail += enc.EncodeEmptyPara()
		detail += enc.EncodePara("View at " + enc.EncodeLink(c.URL.Title, c.URL.URL))
	}

	if len(c.TranscriptionText) > 0 {
		for _, t := range c.TranscriptionText {
			detail += enc.EncodeBlockQuote(t)
			detail += enc.EncodeBlockQuote("")
		}
		if !gdate.IsUnknown(c.TranscriptionDate) {
			detail += enc.EncodeBlockQuote("-- transcribed " + c.TranscriptionDate.Occurrence())
		}
	}

	// for _, m := range c.Media {
	// }

	return enc.EncodeCitation(heading, detail, c.ID)
}

func EncodeRawLink(u string, enc ExtendedInlineEncoder) string {
	text := u

	pu, err := url.Parse(u)
	if err == nil && pu != nil && pu.Host != "" {
		text = pu.Host
	}

	return enc.EncodeLink(text, u)
}

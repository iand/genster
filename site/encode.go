package site

import (
	"net/url"
	"strings"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

type StructuredMarkdownEncoder interface {
	BlockEncoder
	Para(s string)
	EmptyPara()
	Heading3(s string)
	Heading4(s string)
	UnorderedList(items []string)
	OrderedList(items []string)
	DefinitionList(items [][2]string)
	BlockQuote(s string)
}

type BlockEncoder interface {
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
	EncodeLink(text string, url string, suppressDuplicates bool) string
	EncodeModelLink(text string, m any, suppressDuplicates bool) string
	EncodeCitation(citation string, detail string, citationID string) string
	EncodeItalic(s string) string
	EncodeBold(s string) string
}

func EncodeWithCitations(s string, citations []*model.GeneralCitation, enc BlockEncoder) string {
	for i, cit := range citations {
		if i > 0 {
			s += "<sup>,</sup>"
		}
		s += EncodeCitationDetail(cit, enc)
	}
	return s
}

func EncodeCitationDetail(c *model.GeneralCitation, enc BlockEncoder) string {
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
		heading += " (" + enc.EncodeModelLink("source", c.Source, false) + ")"

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
		detail += enc.EncodePara("View at " + enc.EncodeLink(c.URL.Title, c.URL.URL, false))
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

func EncodeRawLink(u string, enc InlineEncoder) string {
	text := u

	pu, err := url.Parse(u)
	if err == nil && pu != nil && pu.Host != "" {
		text = pu.Host
	}

	return enc.EncodeLink(text, u, false)
}

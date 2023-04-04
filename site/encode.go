package site

import (
	"net/url"

	"github.com/iand/genster/model"
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
	// EncodeCitation(citation string, detail string, citationID string) string
	EncodeCitationDetail(c *model.GeneralCitation) string
	EncodeWithCitations(s string, citations []*model.GeneralCitation) string
}

type ExtendedInlineEncoder interface {
	InlineEncoder
	EncodeModelLinkDedupe(firstText string, subsequentText string, m any) string
	// EncodeCitation(citation string, detail string, citationID string) string
}

func EncodeRawLink(u string, enc ExtendedInlineEncoder) string {
	text := u

	pu, err := url.Parse(u)
	if err == nil && pu != nil && pu.Host != "" {
		text = pu.Host
	}

	return enc.EncodeLink(text, u)
}

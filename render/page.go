package render

import (
	"io"

	"github.com/iand/genster/model"
)

type Page interface {
	PageMarkdownEncoder
	MarkupBuilder
	WriteTo(w io.Writer) (int64, error)
	SetFrontMatterField(k, v string)
	Title(s string)
	Summary(s string)
	Layout(s string)
	Category(s string)
	ID(s string)
	AddTag(s string)
	AddTags(ss []string)
	ResetSeenLinks()
}

// A PageMarkdownEncoder provides methods that encode as markdown but require
// or add additional context at the page level.
type PageMarkdownEncoder interface {
	InlineMarkdownEncoder
	EncodeCitationDetail(c *model.GeneralCitation) string
	EncodeWithCitations(s string, citations []*model.GeneralCitation) string
	EncodeModelLinkDedupe(firstText string, subsequentText string, m any) string
}

type MarkupBuilder interface {
	InlineMarkdownEncoder
	PageMarkdownEncoder
	String() string // used by list pages
	RawMarkdown(string)
	Para(string)
	Pre(string)
	EmptyPara()
	Heading2(string)
	Heading3(string)
	Heading4(string)
	UnorderedList([]string)
	OrderedList([]string)
	DefinitionList([][2]string)
	BlockQuote(string)
}

type InlineMarkdownEncoder interface {
	EncodeItalic(s string) string
	EncodeBold(s string) string
	EncodeLink(text string, url string) string
	EncodeModelLink(text string, m any) string
}

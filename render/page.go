package render

import (
	"io"

	"github.com/iand/genster/model"
)

type Page interface {
	WriteTo(w io.Writer) (int64, error)
	MarkupBuilder
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

type MarkupBuilder interface {
	InlineMarkdownEncoder
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
	EncodeModelLinkDedupe(firstText string, subsequentText string, m any) string
	EncodeCitationDetail(c *model.GeneralCitation) string
	EncodeWithCitations(s string, citations []*model.GeneralCitation) string
}

package render

import (
	"io"

	"github.com/iand/genster/model"
)

type EncodedText interface {
	String() string
	IsZero() bool
}

type Page[T EncodedText] interface {
	TextEncoder[T]
	PageBuilder[T]
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

type PageBuilder[T EncodedText] interface {
	TextEncoder[T]
	String() string // used by list pages
	Markdown(string)
	Para(T)
	Pre(string)
	EmptyPara()
	Heading2(m T, id string)
	Heading3(m T, id string)
	Heading4(m T, id string)
	UnorderedList([]T)
	OrderedList([]T)
	DefinitionList([][2]T)
	BlockQuote(T)
	Timeline([]TimelineRow[T])
}

type TimelineRow[T EncodedText] struct {
	Year    string
	Date    string
	Details []T
}

type TextEncoder[T EncodedText] interface {
	EncodeText(ss ...string) T
	EncodeItalic(T) T
	EncodeBold(T) T
	EncodeLink(s T, url string) T
	EncodeModelLink(s T, m any) T
	EncodeWithCitations(s T, citations []*model.GeneralCitation) T
	EncodeModelLinkDedupe(firstText T, subsequentText T, m any) T
}

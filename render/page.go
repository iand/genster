package render

import (
	"io"

	"github.com/iand/genster/model"
)

type EncodedText interface {
	String() string
	IsZero() bool
}

type Document[T EncodedText] interface {
	TextEncoder[T]
	ContentBuilder[T]
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
	PageBreak()
}

type ContentBuilder[T EncodedText] interface {
	TextEncoder[T]
	String() string // used by list pages
	Markdown(string)
	Para(T)
	Pre(string)
	EmptyPara()
	Preface(m T)
	Heading2(m T, id string)
	Heading3(m T, id string)
	Heading4(m T, id string)
	UnorderedList([]T)
	OrderedList([]T)
	DefinitionList([][2]T)
	BlockQuote(T)
	Timeline([]TimelineRow[T])
	Figure(link string, alt string, caption T, highlight *model.Region, downloadName string)
	FactList([]FactEntry[T])
	GroupedList([]Grouping[T])
}

type TimelineRow[T EncodedText] struct {
	Year    string
	Date    string // day and month
	Details []T
}

type FactEntry[T EncodedText] struct {
	Category string
	Details  []T
}

// Grouping is a titled group of related items for rendering as a GroupedList.
type Grouping[T EncodedText] struct {
	Heading T    // the group's heading line
	Notes   []T  // secondary lines shown under the heading, above the items
	Items   []T  // the members of the group
	Ordered bool // true if the items form a numbered sequence
}

type TextEncoder[T EncodedText] interface {
	EncodeText(ss ...string) T
	EncodeItalic(T) T
	EncodeBold(T) T
	EncodeLink(s T, url string) T
	EncodeModelLink(s T, m any) T
	EncodeWithCitations(s T, citations []*model.GeneralCitation) T
	EncodeModelLinkDedupe(firstText T, subsequentText T, m any) T
	EncodeModelLinkNamed(m any, nc NameChooser, pov *model.POV) T
}

type NameChooser interface {
	FirstUse(m any) string                                          // prefix, name and suffix to use for first occurrence
	Subsequent(m any) string                                        // prefix, name and suffix to use for subsequent occurrences
	FirstUseSplit(m any, pov *model.POV) (string, string, string)   // prefix, name and suffix to use for first occurrence
	SubsequentSplit(m any, pov *model.POV) (string, string, string) // prefix, name and suffix to use for subsequent occurrences
}

package render

import (
	"fmt"
	"io"

	"github.com/iand/genster/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

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

type EncodedText interface {
	String() string
	IsZero() bool
	Render(w io.Writer) error
}

type Markdown string

func (m Markdown) String() string { return string(m) }
func (m Markdown) IsZero() bool   { return m == "" }

var md = goldmark.New(
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

type textBlockParser struct{}

func (b *textBlockParser) Trigger() []byte {
	return nil
}

func (b *textBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	_, segment := reader.PeekLine()
	segment = segment.TrimLeftSpace(reader.Source())
	if segment.IsEmpty() {
		return nil, parser.NoChildren
	}
	node := ast.NewTextBlock()

	// node := ast.NewParagraph()
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return node, parser.NoChildren
}

func (b *textBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if util.IsBlank(line) {
		return parser.Close
	}
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return parser.Continue | parser.NoChildren
}

func (b *textBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	lines := node.Lines()
	if lines.Len() != 0 {
		// trim leading spaces
		for i := 0; i < lines.Len(); i++ {
			l := lines.At(i)
			lines.Set(i, l.TrimLeftSpace(reader.Source()))
		}

		// trim trailing spaces
		length := lines.Len()
		lastLine := node.Lines().At(length - 1)
		node.Lines().Set(length-1, lastLine.TrimRightSpace(reader.Source()))
	}
	if lines.Len() == 0 {
		node.Parent().RemoveChild(node.Parent(), node)
		return
	}
}

func (b *textBlockParser) CanInterruptParagraph() bool {
	return false
}

func (b *textBlockParser) CanAcceptIndentedLine() bool {
	return false
}

func init() {
	md.SetParser(parser.NewParser(
		parser.WithBlockParsers(util.Prioritized(&textBlockParser{}, 1000)),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	))
}

func (m Markdown) ToHTML(w io.Writer) error {
	if err := md.Convert([]byte(m), w); err != nil {
		return fmt.Errorf("goldmark: %v", err)
	}

	return nil
}

func (m Markdown) Render(w io.Writer) error {
	if err := md.Convert([]byte(m), w); err != nil {
		return fmt.Errorf("goldmark: %v", err)
	}

	return nil
}

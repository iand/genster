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
	RawMarkdown(Markdown)
	Para(Markdown)
	Pre(string)
	EmptyPara()
	Heading2(Markdown)
	Heading3(Markdown)
	Heading4(Markdown)
	UnorderedList([]Markdown)
	OrderedList([]Markdown)
	DefinitionList([][2]Markdown)
	BlockQuote(Markdown)
}

type InlineMarkdownEncoder interface {
	EncodeItalic(m string) string
	EncodeBold(m string) string
	EncodeLink(text string, url string) string
	EncodeModelLink(text string, m any) string
}

type Markdown string

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

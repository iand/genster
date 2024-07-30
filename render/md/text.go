package md

import (
	"fmt"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Text is a piece of markdown encoded text
type Text string

func (m Text) String() string { return string(m) }
func (m Text) IsZero() bool   { return m == "" }

func (m Text) ToHTML(w io.Writer) error {
	if err := md.Convert([]byte(m), w); err != nil {
		return fmt.Errorf("goldmark: %v", err)
	}

	return nil
}

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

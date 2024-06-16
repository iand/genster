package render

/*
import (
	"bytes"
	"fmt"
	"html"
	"io"
	"strings"
)

type LinkBuilder interface {
	LinkFor(v any) string
}

type HtmlDocument struct {
	HtmlEncoder
	frontMatter map[string][]string
}

type HtmlEncoder struct {
	LinkBuilder       LinkBuilder
	maintext          strings.Builder
	citetext          strings.Builder
	citationidx       int
	citationMap       map[string]int
	seenLinks         map[string]bool
	SuppressCitations bool
}

type Node interface {
	WriteTo(w io.Writer) (int64, error)
}

type Elem struct {
	Name     string
	Children []Node
}

var _ Node = (*Elem)(nil)

func (e *Elem) WriteTo(w io.Writer) (int64, error) {
	bb := new(bytes.Buffer)
	bb.WriteString("<")
	bb.WriteString(e.Name)
	bb.WriteString(">")

	if len(e.Children) > 0 {
		for _, c := range e.Children {
			c.WriteTo(bb)
		}
		bb.WriteString("</")
		bb.WriteString(e.Name)
		bb.WriteString(">")
	}

	return bb.WriteTo(w)
}

type Text struct {
	Content string
}

var _ Node = (*Text)(nil)

func (t *Text) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(html.EscapeString(t.Content)))
	return int64(n), err
}

func (e *HtmlEncoder) Heading1(s string) {
	e.writeHeading1(&e.maintext, s)
}

func (e *HtmlEncoder) EncodeHeading1(s string) string {
	buf := new(strings.Builder)
	e.writeHeading1(buf, s)
	return buf.String()
}

func (e *HtmlEncoder) writeHeading1(buf io.StringWriter, s string) {
	buf.WriteString("<h1>")
	buf.WriteString(html.EscapeString(s))
	buf.WriteString("</h1>\n")
}

func (e *HtmlEncoder) Heading2(s string) {
	e.writeHeading2(&e.maintext, s)
}

func (e *HtmlEncoder) EncodeHeading2(s string) string {
	buf := new(strings.Builder)
	e.writeHeading2(buf, s)
	return buf.String()
}

func (e *HtmlEncoder) writeHeading2(buf io.StringWriter, s string) {
	buf.WriteString("<h2>")
	buf.WriteString(html.EscapeString(s))
	buf.WriteString("</h2>\n")
}

func (e *HtmlEncoder) Heading3(s string) {
	e.writeHeading3(&e.maintext, s)
}

func (e *HtmlEncoder) EncodeHeading3(s string) string {
	buf := new(strings.Builder)
	e.writeHeading3(buf, s)
	return buf.String()
}

func (e *HtmlEncoder) writeHeading3(buf io.StringWriter, s string) {
	buf.WriteString("<h3>")
	buf.WriteString(html.EscapeString(s))
	buf.WriteString("</h3>\n")
}

func (e *HtmlEncoder) Heading4(s string) {
	e.writeHeading4(&e.maintext, s)
}

func (e *HtmlEncoder) EncodeHeading4(s string) string {
	buf := new(strings.Builder)
	e.writeHeading4(buf, s)
	return buf.String()
}

func (e *HtmlEncoder) writeHeading4(buf io.StringWriter, s string) {
	buf.WriteString("<h4>")
	buf.WriteString(html.EscapeString(s))
	buf.WriteString("</h4>\n")
}

func (e *HtmlEncoder) Para(s string) {
	e.writePara(&e.maintext, s)
}

func (e *HtmlEncoder) EncodePara(s string) string {
	buf := new(strings.Builder)
	e.writePara(buf, s)
	return buf.String()
}

func (e *HtmlEncoder) writePara(buf io.StringWriter, s string) {
	buf.WriteString("<p>")
	buf.WriteString(html.EscapeString(s))
	buf.WriteString("</p>\n")
}

func (e *HtmlEncoder) EmptyPara() {
	e.writeEmptyPara(&e.maintext)
}

func (e *HtmlEncoder) EncodeEmptyPara() string {
	buf := new(strings.Builder)
	e.writeEmptyPara(buf)
	return buf.String()
}

func (e *HtmlEncoder) writeEmptyPara(buf io.StringWriter) {
	buf.WriteString("<p></p>")
}

func (e *HtmlEncoder) BlockQuote(s string) {
	e.writeBlockQuote(&e.maintext, s)
}

func (e *HtmlEncoder) EncodeBlockQuote(s string) string {
	buf := new(strings.Builder)
	e.writeBlockQuote(buf, s)
	return buf.String()
}

func (e *HtmlEncoder) writeBlockQuote(buf io.StringWriter, s string) {
	buf.WriteString("<blockquote>")
	buf.WriteString(html.EscapeString(s))
	buf.WriteString("</blockquote>\n")
}

func (e *HtmlEncoder) EncodeItalic(s string) string {
	return "*" + s + "*"
}

func (e *HtmlEncoder) EncodeBold(s string) string {
	return "**" + s + "**"
}

func (e *HtmlEncoder) UnorderedList(items []string) {
	b.writeUnorderedList(&b.maintext, items)
}

func (e *HtmlEncoder) EncodeUnorderedList(items []string) string {
	buf := new(strings.Builder)
	b.writeUnorderedList(buf, items)
	return buf.String()
}

func (e *HtmlEncoder) writeUnorderedList(buf io.StringWriter, items []string) {
	for _, item := range items {
		buf.WriteString(" - " + item + "\n")
	}
}

func (e *HtmlEncoder) OrderedList(items []string) {
	b.writeOrderedList(&b.maintext, items)
}

func (e *HtmlEncoder) EncodeOrderedList(items []string) string {
	buf := new(strings.Builder)
	b.writeOrderedList(buf, items)
	return buf.String()
}

func (e *HtmlEncoder) writeOrderedList(buf io.StringWriter, items []string) {
	for i, item := range items {
		buf.WriteString(fmt.Sprintf(" %d. %s\n", i+1, item))
	}
}

func (e *HtmlEncoder) DefinitionList(items [][2]string) {
	b.writeDefinitionList(&b.maintext, items)
}

func (e *HtmlEncoder) EncodeDefinitionList(items [][2]string) string {
	buf := new(strings.Builder)
	b.writeDefinitionList(buf, items)
	return buf.String()
}
*/

package md

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

const DirectAncestorMarker = "★"

type Encoder struct {
	LinkBuilder       LinkBuilder
	main              strings.Builder
	citations         strings.Builder
	citationidx       int
	citationMap       map[string]int
	seenLinks         map[string]bool
	SuppressCitations bool
}

func (e *Encoder) SetLinkBuilder(l LinkBuilder) {
	e.LinkBuilder = l
}

func (e *Encoder) Markdown() string {
	s := new(strings.Builder)
	e.WriteMarkdown(s)
	return s.String()
}

func (e *Encoder) WriteMarkdown(w io.Writer) error {
	bw := bufio.NewWriter(w)
	bw.WriteString(e.main.String())
	bw.WriteString("\n")

	reftext := e.citations.String()
	if len(reftext) > 0 {
		bw.WriteString("<div class=\"footnotes\">\n\n")
		bw.WriteString("----\n\n")
		bw.WriteString("#### Citations\n")
		bw.WriteString("\n")
		bw.WriteString(reftext)
		bw.WriteString("</div>\n\n")
	}
	return bw.Flush()
}

func (e *Encoder) SetBody(s string) {
	e.main = strings.Builder{}
	e.main.WriteString(s)
}

func (e *Encoder) RawMarkdown(s string) {
	e.main.WriteString(s)
}

func (e *Encoder) Heading1(s string) {
	e.writeHeading1(&e.main, s)
}

func (e *Encoder) EncodeHeading1(s string) string {
	buf := new(strings.Builder)
	e.writeHeading1(buf, s)
	return buf.String()
}

func (e *Encoder) writeHeading1(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("# " + s)
	buf.WriteString("\n")
}

func (b *Encoder) Heading2(s string) {
	b.writeHeading2(&b.main, s)
}

func (b *Encoder) EncodeHeading2(s string) string {
	buf := new(strings.Builder)
	b.writeHeading2(buf, s)
	return buf.String()
}

func (b *Encoder) writeHeading2(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("## " + s)
	buf.WriteString("\n")
}

func (b *Encoder) Heading3(s string) {
	b.writeHeading3(&b.main, s)
}

func (b *Encoder) EncodeHeading3(s string) string {
	buf := new(strings.Builder)
	b.writeHeading3(buf, s)
	return buf.String()
}

func (b *Encoder) writeHeading3(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("### " + s)
	buf.WriteString("\n")
}

func (b *Encoder) Heading4(s string) {
	b.writeHeading3(&b.main, s)
}

func (b *Encoder) EncodeHeading4(s string) string {
	buf := new(strings.Builder)
	b.writeHeading3(buf, s)
	return buf.String()
}

func (b *Encoder) writeHeading4(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("#### " + s)
	buf.WriteString("\n")
}

func (b *Encoder) Para(s string) {
	b.writePara(&b.main, s)
}

func (b *Encoder) EncodePara(s string) string {
	buf := new(strings.Builder)
	b.writePara(buf, s)
	return buf.String()
}

func (b *Encoder) writePara(buf io.StringWriter, s string) {
	buf.WriteString(s)
	buf.WriteString("\n")
}

func (b *Encoder) EmptyPara() {
	b.writeEmptyPara(&b.main)
}

func (b *Encoder) EncodeEmptyPara() string {
	buf := new(strings.Builder)
	b.writeEmptyPara(buf)
	return buf.String()
}

func (b *Encoder) writeEmptyPara(buf io.StringWriter) {
	buf.WriteString("\n\n")
}

func (b *Encoder) BlockQuote(s string) {
	b.writeBlockQuote(&b.main, s)
}

func (b *Encoder) EncodeBlockQuote(s string) string {
	buf := new(strings.Builder)
	b.writeBlockQuote(buf, s)
	return buf.String()
}

func (b *Encoder) writeBlockQuote(buf io.StringWriter, s string) {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if i > 0 {
			buf.WriteString("\n")
			buf.WriteString("> \n")
		}
		buf.WriteString("> ")
		buf.WriteString(l)
	}
	buf.WriteString("\n")
}

func (b *Encoder) EncodeItalic(s string) string {
	return "*" + s + "*"
}

func (b *Encoder) EncodeBold(s string) string {
	return "**" + s + "**"
}

func (b *Encoder) UnorderedList(items []string) {
	b.writeUnorderedList(&b.main, items)
}

func (b *Encoder) EncodeUnorderedList(items []string) string {
	buf := new(strings.Builder)
	b.writeUnorderedList(buf, items)
	return buf.String()
}

func (b *Encoder) writeUnorderedList(buf io.StringWriter, items []string) {
	for _, item := range items {
		buf.WriteString(" - " + item + "\n")
	}
}

func (b *Encoder) OrderedList(items []string) {
	b.writeOrderedList(&b.main, items)
}

func (b *Encoder) EncodeOrderedList(items []string) string {
	buf := new(strings.Builder)
	b.writeOrderedList(buf, items)
	return buf.String()
}

func (b *Encoder) writeOrderedList(buf io.StringWriter, items []string) {
	for i, item := range items {
		buf.WriteString(fmt.Sprintf(" %d. %s\n", i+1, item))
	}
}

func (b *Encoder) DefinitionList(items [][2]string) {
	b.writeDefinitionList(&b.main, items)
}

func (b *Encoder) EncodeDefinitionList(items [][2]string) string {
	buf := new(strings.Builder)
	b.writeDefinitionList(buf, items)
	return buf.String()
}

func (b *Encoder) writeDefinitionList(buf io.StringWriter, items [][2]string) {
	// buf.WriteString("<dl>\n")
	for _, item := range items {
		// buf.WriteString(fmt.Sprintf("#### %s\n", item[0]))
		// buf.WriteString(fmt.Sprintf("%s\n", item[1]))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("%s\n", item[0]))
		lines := strings.Split(item[1], "\n")
		for _, l := range lines {
			buf.WriteString(fmt.Sprintf(": %s\n", l))
		}

		// buf.WriteString(fmt.Sprintf("<dt>%s</dt>\n", item[0]))
		// buf.WriteString(fmt.Sprintf("<dd>%s</dd>\n", item[1]))
	}
	// buf.WriteString("</dl>\n")
}

func (b *Encoder) ResetSeenLinks() {
	b.seenLinks = make(map[string]bool)
}

func (b *Encoder) EncodeLink(text string, url string) string {
	if url == "" {
		return text
	}

	return fmt.Sprintf("[%s](%s)", text, url)
}

func (b *Encoder) EncodeModelLink(firstText string, m any) string {
	if b.seenLinks == nil {
		b.seenLinks = make(map[string]bool)
	}
	url := b.LinkBuilder.LinkFor(m)
	b.seenLinks[url] = true

	if p, ok := m.(*model.Person); ok {
		if p.RelationToKeyPerson.IsDirectAncestor() && !p.RelationToKeyPerson.IsSelf() {
			firstText += DirectAncestorMarker
		}
	}

	return b.EncodeLink(firstText, url)
}

func (b *Encoder) EncodeModelLinkDedupe(firstText string, subsequentText string, m any) string {
	suffix := ""
	if p, ok := m.(*model.Person); ok {
		if p.RelationToKeyPerson.IsDirectAncestor() && !p.RelationToKeyPerson.IsSelf() {
			suffix = DirectAncestorMarker
		}
	}

	if b.LinkBuilder == nil {
		return firstText + suffix
	}

	url := b.LinkBuilder.LinkFor(m)

	// Only encode the first mention of a link
	if b.seenLinks == nil {
		b.seenLinks = make(map[string]bool)
	}

	if b.seenLinks[url] {
		// return subsequentText
		return b.EncodeLink(subsequentText+suffix, url)
	}
	b.seenLinks[url] = true

	return b.EncodeLink(firstText+suffix, url)
}

func (b *Encoder) EncodeCitation(citation string, detail string, citationID string) string {
	if b.SuppressCitations {
		return ""
	}
	if b.citationMap == nil {
		b.citationMap = make(map[string]int)
	}

	var idx int
	if citationID != "" {
		var exists bool
		idx, exists = b.citationMap[citationID]
		if !exists {
			b.citationidx++
			idx = b.citationidx
			b.citationMap[citationID] = idx
			b.citations.WriteString("\n")
			b.citations.WriteString(fmt.Sprintf("##### %d. %s {#cit_%[1]d}\n", idx, citation))
			b.citations.WriteString("\n")
			if detail != "" {
				b.citations.WriteString(detail)
				b.citations.WriteString("\n")
			}
		}

	} else {
		b.citationidx++
		idx = b.citationidx
		b.citations.WriteString(fmt.Sprintf("##### %d. %s {#cit_%[1]d}\n", idx, citation))
	}

	return fmt.Sprintf("<sup>[%d](#cit_%[1]d)</sup>", idx)
}

func (b *Encoder) EncodeWithCitations(s string, citations []*model.GeneralCitation) string {
	sups := ""
	for i, cit := range citations {
		if i > 0 && sups != "" {
			sups += "<sup>,</sup>"
		}
		sups += b.EncodeCitationDetail(cit)
	}
	return s + sups
}

func (b *Encoder) EncodeCitationDetail(c *model.GeneralCitation) string {
	var heading string
	var detail string

	if c.Source != nil && c.Source.Title != "" {
		heading = c.Source.Title
		if c.Detail != "" {
			if !strings.HasSuffix(heading, ".") && !strings.HasSuffix(heading, "!") && !strings.HasSuffix(heading, "?") {
				heading += "; "
			}
			heading += b.EncodePara(c.Detail)
		}
		heading = text.FinishSentence(heading)
		heading += " (" + b.EncodeModelLink("source", c.Source) + ")"

		// heading = enc.EncodeModelLink(text.FinishSentence(c.Source.Title), c.Source, false)
		// if c.Detail != "" {
		// 	detail = enc.EncodePara(cleanCitationDetail(c.Detail))
		// }
	} else {
		heading = c.Detail
		detail = ""
	}

	if c.URL != nil {
		detail += b.EncodeEmptyPara()
		detail += b.EncodePara("View at " + b.EncodeLink(c.URL.Title, c.URL.URL))
	}

	if len(c.TranscriptionText) > 0 {
		for _, t := range c.TranscriptionText {
			detail += b.EncodeBlockQuote(t)
			detail += b.EncodeBlockQuote("")
		}
		if !c.TranscriptionDate.IsUnknown() {
			detail += b.EncodeBlockQuote("-- transcribed " + c.TranscriptionDate.When())
		}
	}

	// for _, m := range c.Media {
	// }

	return b.EncodeCitation(heading, detail, c.ID)
}

func (b *Encoder) Pre(s string) {
	b.writePre(&b.main, s)
}

func (b *Encoder) EncodePre(s string) string {
	buf := new(strings.Builder)
	b.writePre(buf, s)
	return buf.String()
}

func (b *Encoder) writePre(buf io.StringWriter, s string) {
	buf.WriteString("<pre>\n")
	buf.WriteString(html.EscapeString(s))
	buf.WriteString("</pre>\n")
}

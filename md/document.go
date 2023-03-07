package md

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	MarkdownTagTitle    = "title"
	MarkdownTagSummary  = "summary"
	MarkdownTagLayout   = "layout"
	MarkdownTagTags     = "tags"
	MarkdownTagCategory = "category"
	MarkdownTagID       = "id"
)

const (
	PageLayoutPerson     = "person"
	PageLayoutCalendar   = "calendar"
	PageLayoutSource     = "source"
	PageLayoutPlace      = "place"
	PageLayoutInferences = "inferences"
)

type LinkBuilder interface {
	LinkFor(v any) string
}

type Document struct {
	LinkBuilder      LinkBuilder
	frontMatter      map[string][]string
	main             strings.Builder
	citations        strings.Builder
	citationidx      int
	citationMap      map[string]int
	seenLinks        map[string]bool
	lastHeadingLevel int
}

func (b *Document) Markdown() string {
	s := new(strings.Builder)
	b.WriteMarkdown(s)
	return s.String()
}

func (b *Document) WriteMarkdown(w io.Writer) error {
	bw := bufio.NewWriter(w)

	tagRanks := map[string]byte{
		MarkdownTagID:      4,
		MarkdownTagTitle:   3,
		MarkdownTagLayout:  2,
		MarkdownTagSummary: 1,
	}

	if len(b.frontMatter) > 0 {
		bw.WriteString("---\n")

		type rankedKey struct {
			key  string
			rank int
		}

		keys := make([]string, 0, len(b.frontMatter))
		for k := range b.frontMatter {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			ri := tagRanks[keys[i]]
			rj := tagRanks[keys[j]]
			if ri != rj {
				return ri > rj
			}
			return keys[i] < keys[j]
		})

		for _, k := range keys {
			v := b.frontMatter[k]
			bw.WriteString(k)
			bw.WriteString(": ")
			bw.WriteString(strings.Join(v, ","))
			bw.WriteString("\n")
		}
		bw.WriteString("---\n")
	}
	bw.WriteString("\n")
	bw.WriteString(b.main.String())
	bw.WriteString("\n")

	reftext := b.citations.String()
	if len(reftext) > 0 {
		bw.WriteString("## Citations\n")
		bw.WriteString("\n")
		bw.WriteString(reftext)
	}
	return bw.Flush()
}

func (b *Document) SetFrontMatterField(k, v string) {
	if b.frontMatter == nil {
		b.frontMatter = make(map[string][]string)
	}
	b.frontMatter[k] = []string{v}
}

func (b *Document) appendFrontMatterField(k, v string) {
	if b.frontMatter == nil {
		b.frontMatter = make(map[string][]string)
	}
	b.frontMatter[k] = append(b.frontMatter[k], v)
}

func (b *Document) Title(s string) {
	b.SetFrontMatterField(MarkdownTagTitle, s)
}

func (b *Document) Summary(s string) {
	if s == "" {
		return
	}
	b.SetFrontMatterField(MarkdownTagSummary, s)
}

func (b *Document) Layout(s string) {
	b.SetFrontMatterField(MarkdownTagLayout, s)
}

func (b *Document) ID(s string) {
	b.SetFrontMatterField(MarkdownTagID, s)
}

func (b *Document) Category(s string) {
	b.SetFrontMatterField(MarkdownTagCategory, s)
}

func (b *Document) AddTag(s string) {
	if s == "" {
		return
	}
	b.appendFrontMatterField(MarkdownTagTags, s)
}

func (b *Document) AddTags(ss []string) {
	if len(ss) == 0 {
		return
	}
	for _, s := range ss {
		b.AddTag(s)
	}
}

func (b *Document) Heading1(s string) {
	b.writeHeading1(&b.main, s)
}

func (b *Document) EncodeHeading1(s string) string {
	buf := new(strings.Builder)
	b.writeHeading1(buf, s)
	return buf.String()
}

func (b *Document) writeHeading1(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("# " + s)
	buf.WriteString("\n")
}

func (b *Document) Heading2(s string) {
	b.writeHeading2(&b.main, s)
}

func (b *Document) EncodeHeading2(s string) string {
	buf := new(strings.Builder)
	b.writeHeading2(buf, s)
	return buf.String()
}

func (b *Document) writeHeading2(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("## " + s)
	buf.WriteString("\n")
}

func (b *Document) Heading3(s string) {
	b.writeHeading3(&b.main, s)
}

func (b *Document) EncodeHeading3(s string) string {
	buf := new(strings.Builder)
	b.writeHeading3(buf, s)
	return buf.String()
}

func (b *Document) writeHeading3(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("### " + s)
	buf.WriteString("\n")
}

func (b *Document) Heading4(s string) {
	b.writeHeading3(&b.main, s)
}

func (b *Document) EncodeHeading4(s string) string {
	buf := new(strings.Builder)
	b.writeHeading3(buf, s)
	return buf.String()
}

func (b *Document) writeHeading4(buf io.StringWriter, s string) {
	buf.WriteString("\n")
	buf.WriteString("#### " + s)
	buf.WriteString("\n")
}

func (b *Document) Para(s string) {
	b.writePara(&b.main, s)
}

func (b *Document) EncodePara(s string) string {
	buf := new(strings.Builder)
	b.writePara(buf, s)
	return buf.String()
}

func (b *Document) writePara(buf io.StringWriter, s string) {
	buf.WriteString(s)
	buf.WriteString("\n")
}

func (b *Document) EmptyPara() {
	b.writeEmptyPara(&b.main)
}

func (b *Document) EncodeEmptyPara() string {
	buf := new(strings.Builder)
	b.writeEmptyPara(buf)
	return buf.String()
}

func (b *Document) writeEmptyPara(buf io.StringWriter) {
	buf.WriteString("\n\n")
}

func (b *Document) BlockQuote(s string) {
	b.writeBlockQuote(&b.main, s)
}

func (b *Document) EncodeBlockQuote(s string) string {
	buf := new(strings.Builder)
	b.writeBlockQuote(buf, s)
	return buf.String()
}

func (b *Document) writeBlockQuote(buf io.StringWriter, s string) {
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

func (b *Document) EncodeItalic(s string) string {
	return "*" + s + "*"
}

func (b *Document) EncodeBold(s string) string {
	return "**" + s + "**"
}

func (b *Document) UnorderedList(items []string) {
	b.writeUnorderedList(&b.main, items)
}

func (b *Document) EncodeUnorderedList(items []string) string {
	buf := new(strings.Builder)
	b.writeUnorderedList(buf, items)
	return buf.String()
}

func (b *Document) writeUnorderedList(buf io.StringWriter, items []string) {
	for _, item := range items {
		buf.WriteString(" - " + item + "\n")
	}
}

func (b *Document) OrderedList(items []string) {
	b.writeOrderedList(&b.main, items)
}

func (b *Document) EncodeOrderedList(items []string) string {
	buf := new(strings.Builder)
	b.writeOrderedList(buf, items)
	return buf.String()
}

func (b *Document) writeOrderedList(buf io.StringWriter, items []string) {
	for i, item := range items {
		buf.WriteString(fmt.Sprintf(" %d. %s\n", i+1, item))
	}
}

func (b *Document) DefinitionList(items [][2]string) {
	b.writeDefinitionList(&b.main, items)
}

func (b *Document) EncodeDefinitionList(items [][2]string) string {
	buf := new(strings.Builder)
	b.writeDefinitionList(buf, items)
	return buf.String()
}

func (b *Document) writeDefinitionList(buf io.StringWriter, items [][2]string) {
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

func (b *Document) EncodeLink(text string, url string, suppressDuplicate bool) string {
	if url == "" {
		return text
	}

	// Only encode the first mention of a link
	if b.seenLinks == nil {
		b.seenLinks = make(map[string]bool)
	}

	if suppressDuplicate && b.seenLinks[url] {
		return text
	}

	b.seenLinks[url] = true
	return fmt.Sprintf("[%s](%s)", text, url)
}

func (b *Document) EncodeModelLink(text string, m any, suppressDuplicate bool) string {
	if b.LinkBuilder == nil {
		return text
	}
	return b.EncodeLink(text, b.LinkBuilder.LinkFor(m), suppressDuplicate)
}

func (b *Document) EncodeCitation(citation string, detail string, citationID string) string {
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
		b.citations.WriteString(fmt.Sprintf("#### %d. %s {#cit_%[1]d}\n", idx, citation))
	}

	return fmt.Sprintf("<sup>[%d](#cit_%[1]d)</sup>", idx)
}

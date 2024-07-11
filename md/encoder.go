package md

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/yuin/goldmark"
)

const (
	DirectAncestorMarker   = "★"
	GeneralNotesSourceName = "General Notes"
)

type (
	PlainText  string
	MarkupText string
)

type Encoder struct {
	LinkBuilder       LinkBuilder
	maintext          strings.Builder
	citationAnchors   map[string]citationAnchor // map of citation ids to html anchors
	sourceMap         map[string]sourceEntry    // map of source names to source entries
	seenLinks         map[string]bool
	SuppressCitations bool
}

type citationAnchor struct {
	anchor  string
	display string
}

type sourceEntry struct {
	name      string
	index     int
	prefix    string
	citations []citationEntry
}

type citationEntry struct {
	id       string
	anchor   string
	sortkey  string
	markdown render.Markdown
}

func (e *Encoder) SetLinkBuilder(l LinkBuilder) {
	e.LinkBuilder = l
}

func (e *Encoder) String() string {
	s := new(strings.Builder)
	e.WriteTo(s)
	return s.String()
}

func (e *Encoder) WriteTo(w io.Writer) (int64, error) {
	bb := new(bytes.Buffer)

	bb.WriteString(e.maintext.String())
	bb.WriteString("\n")

	if len(e.sourceMap) > 0 {
		sources := make([]sourceEntry, 0, len(e.sourceMap))
		for id := range e.sourceMap {
			sources = append(sources, e.sourceMap[id])
		}
		sort.Slice(sources, func(i, j int) bool {
			return sources[i].index < sources[j].index
		})

		bb.WriteString("<div class=\"footnotes\">\n\n")
		bb.WriteString("<h2>Citations and Notes</h2\n")
		bb.WriteString("<hr>\n")
		for i := range sources {
			bb.WriteString(fmt.Sprintf("<div class=\"source\">%s</div>\n", html.EscapeString(sources[i].name)))

			for ci := range sources[i].citations {
				bb.WriteString(fmt.Sprintf("<div class=\"citation\"><span class=\"anchor\" id=\"%s\">%s:</span> ", sources[i].citations[ci].anchor, sources[i].citations[ci].anchor))
				sources[i].citations[ci].markdown.ToHTML(bb)
				bb.WriteString("</div>\n")
			}

		}
		bb.WriteString("</div>\n\n")
	}

	return bb.WriteTo(w)
}

func (e *Encoder) SetBody(s string) {
	e.maintext = strings.Builder{}
	e.maintext.WriteString(s)
}

func (e *Encoder) RawMarkdown(s render.Markdown) {
	e.maintext.WriteString("\n")
	e.maintext.WriteString(string(s))
}

func (e *Encoder) Heading2(m render.Markdown, id string) {
	e.maintext.WriteString("<h2>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h2>\n")
}

func (e *Encoder) Heading3(m render.Markdown, id string) {
	e.maintext.WriteString("<h3>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h3>\n")
}

func (e *Encoder) Heading4(m render.Markdown, id string) {
	e.maintext.WriteString("<h4>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h4>\n")
}

func (b *Encoder) Para(m render.Markdown) {
	b.maintext.WriteString("<p>")
	m.ToHTML(&b.maintext)
	b.maintext.WriteString("</p>\n")
}

func (e *Encoder) EmptyPara() {
	e.maintext.WriteString("<p></p>\n")
	// e.writeEmptyPara(&b.maintext)
}

func (b *Encoder) BlockQuote(m render.Markdown) {
	b.maintext.WriteString("<blockquote>")
	b.maintext.WriteString(string(m))
	b.maintext.WriteString("</blockquote>\n")
}

func (e *Encoder) UnorderedList(ms []render.Markdown) {
	e.maintext.WriteString("<ul>\n")
	for _, m := range ms {
		e.maintext.WriteString("<li>")
		m.ToHTML(&e.maintext)
		e.maintext.WriteString("</li>\n")
	}
	e.maintext.WriteString("</ul>\n")
}

func (e *Encoder) OrderedList(ms []render.Markdown) {
	e.maintext.WriteString("<ol>\n")
	for _, m := range ms {
		e.maintext.WriteString("<li>")
		m.ToHTML(&e.maintext)
		e.maintext.WriteString("</li>\n")
	}
	e.maintext.WriteString("</ol>\n")
}

func (e *Encoder) DefinitionList(markups [][2]render.Markdown) {
	e.maintext.WriteString("<dl>\n")
	for _, markup := range markups {
		e.maintext.WriteString("<dt>")
		markup[0].ToHTML(&e.maintext)
		e.maintext.WriteString("</dt>\n")

		e.maintext.WriteString("<dd>")
		markup[1].ToHTML(&e.maintext)
		e.maintext.WriteString("</dd>\n")
	}
	e.maintext.WriteString("</dl>\n")
}

func (e *Encoder) ResetSeenLinks() {
	e.seenLinks = make(map[string]bool)
}

func (e *Encoder) EncodeLink(text string, url string) string {
	if url == "" {
		return text
	}

	// return fmt.Sprintf("<a href=\"%s\">%s</a>", url, html.EscapeString(text))
	return fmt.Sprintf("[%s](%s)", text, url)
}

func (e *Encoder) EncodeModelLink(firstText string, m any) string {
	if e.seenLinks == nil {
		e.seenLinks = make(map[string]bool)
	}
	url := e.LinkBuilder.LinkFor(m)
	e.seenLinks[url] = true

	if p, ok := m.(*model.Person); ok {
		if p.RelationToKeyPerson.IsDirectAncestor() && !p.RelationToKeyPerson.IsSelf() {
			firstText += DirectAncestorMarker
		}
	}

	return e.EncodeLink(firstText, url)
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
		return subsequentText
	}
	b.seenLinks[url] = true

	return b.EncodeLink(firstText+suffix, url)
}

func (b *Encoder) EncodeWithCitations(s string, citations []*model.GeneralCitation) string {
	if len(citations) == 0 {
		return s
	}
	sups := "<sup class=\"citref\">"
	for i, cit := range citations {
		if i > 0 && sups != "" {
			sups += ","
		}
		sups += b.EncodeCitationDetail(cit)
	}

	return s + sups + "</sup>"
}

func (e *Encoder) EncodeCitationDetail(c *model.GeneralCitation) string {
	citationText := c.Detail

	if c.ID != "" && e.LinkBuilder.LinkFor(c) != "" && (len(c.TranscriptionText) > 0 || len(c.MediaObjects) > 0) {
		citationText += " (" + e.EncodeModelLink("more details...", c) + ")"
	} else {
		if c.URL != nil {
			citationText = e.EncodeLink(citationText, c.URL.URL)
		}
	}

	return e.EncodeCitationLink(c.SourceTitle(), render.Markdown(citationText), c.ID)
}

func (b *Encoder) EncodeCitationLink(sourceName string, citationText render.Markdown, citationID string) string {
	if b.SuppressCitations {
		return ""
	}

	if b.citationAnchors == nil {
		b.citationAnchors = make(map[string]citationAnchor)
	}

	ret := ""

	var anchor string
	ca, ok := b.citationAnchors[citationID]
	if ok {
		anchor = ca.anchor
	} else {

		if b.sourceMap == nil {
			b.sourceMap = make(map[string]sourceEntry)
		}
		if sourceName == "" {
			sourceName = GeneralNotesSourceName
		}

		se, ok := b.sourceMap[sourceName]
		if !ok {
			if sourceName == GeneralNotesSourceName {
				se = sourceEntry{
					name:   sourceName,
					index:  -1,
					prefix: "n",
				}
			} else {
				se = sourceEntry{
					name:  sourceName,
					index: len(b.sourceMap),
				}

				const alphabet string = "abcdefghkmpqrstuvwxyz"
				idx := se.index
				if idx > len(alphabet)-1 {
					se.prefix = string(alphabet[idx/len(alphabet)])
					idx = idx % len(alphabet)
				}
				if idx < len(alphabet) {
					se.prefix += string(alphabet[idx])
				} else {
					panic(fmt.Sprintf("too many sources se.index=%d, idx=%d", se.index, idx))
				}
			}
		}
		citIdx := len(se.citations) + 1
		anchor = fmt.Sprintf("%s%d", se.prefix, citIdx)
		ce := citationEntry{
			id:       citationID,
			anchor:   anchor,
			sortkey:  string(citationText),
			markdown: citationText,
		}
		se.citations = append(se.citations, ce)

		b.sourceMap[sourceName] = se

		ca := citationAnchor{
			anchor:  anchor,
			display: anchor,
		}

		b.citationAnchors[citationID] = ca
	}

	ret = fmt.Sprintf("<a href=\"#%s\">%s</a>", anchor, anchor)

	return ret
}

func (e *Encoder) Pre(s string) {
	e.maintext.WriteString("<pre>\n")
	e.maintext.WriteString(html.EscapeString(s))
	e.maintext.WriteString("</pre>\n")
}

func (e *Encoder) Image(alt string, link string) {
	e.writeImage(&e.maintext, alt, link)
}

func (b *Encoder) writeImage(buf io.StringWriter, alt string, link string) {
	buf.WriteString(fmt.Sprintf("![%s](%s)\n", alt, link))
}

func (e *Encoder) ParaWithFigure(text render.Markdown, link string, alt string, caption render.Markdown) {
	e.maintext.WriteString("<p>")
	e.maintext.WriteString("<figure class=\"inline-right\">")
	e.maintext.WriteString(fmt.Sprintf("<img src=\"%s\" alt=\"%s\">", html.EscapeString(link), html.EscapeString(alt)))
	e.maintext.WriteString("<figcaption>")
	e.maintext.WriteString("<p>")
	caption.ToHTML(&e.maintext)
	e.maintext.WriteString("</p>")
	e.maintext.WriteString("</figcaption>")
	e.maintext.WriteString("</figure>\n")
	text.ToHTML(&e.maintext)
	e.maintext.WriteString("</p>\n")
}

func (e *Encoder) Figure(link string, alt string, caption render.Markdown, highlight *model.Region) {
	e.maintext.WriteString("<figure>")
	if highlight == nil {
		e.maintext.WriteString(fmt.Sprintf("<a href=\"%s\" data-dimbox=\"figures\"><img src=\"%[1]s\" alt=\"%s\"></a>", html.EscapeString(link), html.EscapeString(alt)))
	} else {
		e.maintext.WriteString("<div class=\"shade\">")
		e.maintext.WriteString(fmt.Sprintf("<a href=\"%s\" data-dimbox=\"figures\">", html.EscapeString(link)))
		e.maintext.WriteString(fmt.Sprintf("<span class=\"shade\" style=\"bottom: %d%%;left: %d%%;width: %d%%;height: %d%%;\"></span>", highlight.Bottom, highlight.Left, highlight.Width, highlight.Height))
		e.maintext.WriteString(fmt.Sprintf("<img src=\"%[1]s\" alt=\"%s\">", html.EscapeString(link), html.EscapeString(alt)))
		e.maintext.WriteString("</a></div>")
	}

	e.maintext.WriteString("<figcaption>")
	e.maintext.WriteString("<p>")
	caption.ToHTML(&e.maintext)
	e.maintext.WriteString("</p>")
	e.maintext.WriteString("</figcaption>")
	e.maintext.WriteString("</figure>\n")
}

func (e *Encoder) Timeline(rows []render.TimelineRow) {
	e.maintext.WriteString("<dl class=\"timeline\">\n")
	yr := ""
	for _, row := range rows {
		if yr != row.Year {
			if yr != "" {
				e.maintext.WriteString("</dl></dd>\n")
			}
			e.maintext.WriteString("<dt>")
			e.maintext.WriteString(html.EscapeString(row.Year))
			e.maintext.WriteString("</dt>\n")
			e.maintext.WriteString("<dd><dl class=\"tlentry\">\n")
			yr = row.Year
		}
		e.maintext.WriteString("<dt>")
		e.maintext.WriteString(html.EscapeString(row.Date))
		e.maintext.WriteString("</dt>\n")
		e.maintext.WriteString("<dd>")
		for i, det := range row.Details {
			if i > 0 {
				e.maintext.WriteString("<br>\n")
			}
			det.ToHTML(&e.maintext)
		}
		e.maintext.WriteString("</dd>\n")
	}
	e.maintext.WriteString("</dl></dd>\n")
	e.maintext.WriteString("</dl>\n")
}

func (e *Encoder) ConvertMarkdown(text string, w io.Writer) error {
	if err := goldmark.Convert([]byte(text), w); err != nil {
		return fmt.Errorf("goldmark: %v", err)
	}

	return nil
}

func (e *Encoder) EncodeItalic(m string) string {
	return "*" + m + "*"
}

func (e *Encoder) EncodeBold(m string) string {
	return "**" + m + "**"
}

func (e *Encoder) EncodeImage(alt string, link string) string {
	buf := new(strings.Builder)
	e.writeImage(buf, alt, link)
	return buf.String()
}

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

var _ render.PageBuilder[Text] = (*Encoder)(nil)

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
	markdown Text
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

func (e *Encoder) Markdown(s string) {
	e.maintext.WriteString("\n")
	e.maintext.WriteString(s)
}

func (e *Encoder) Heading2(m Text, id string) {
	e.maintext.WriteString("<h2>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h2>\n")
}

func (e *Encoder) Heading3(m Text, id string) {
	e.maintext.WriteString("<h3>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h3>\n")
}

func (e *Encoder) Heading4(m Text, id string) {
	e.maintext.WriteString("<h4>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h4>\n")
}

func (b *Encoder) Para(m Text) {
	b.maintext.WriteString("<p>")
	m.ToHTML(&b.maintext)
	b.maintext.WriteString("</p>\n")
}

func (e *Encoder) EmptyPara() {
	e.maintext.WriteString("<p></p>\n")
	// e.writeEmptyPara(&b.maintext)
}

func (b *Encoder) BlockQuote(m Text) {
	b.maintext.WriteString("<blockquote>")
	b.maintext.WriteString(string(m))
	b.maintext.WriteString("</blockquote>\n")
}

func (e *Encoder) UnorderedList(ms []Text) {
	e.maintext.WriteString("<ul>\n")
	for _, m := range ms {
		e.maintext.WriteString("<li>")
		m.ToHTML(&e.maintext)
		e.maintext.WriteString("</li>\n")
	}
	e.maintext.WriteString("</ul>\n")
}

func (e *Encoder) OrderedList(ms []Text) {
	e.maintext.WriteString("<ol>\n")
	for _, m := range ms {
		e.maintext.WriteString("<li>")
		m.ToHTML(&e.maintext)
		e.maintext.WriteString("</li>\n")
	}
	e.maintext.WriteString("</ol>\n")
}

func (e *Encoder) DefinitionList(markups [][2]Text) {
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

func (e *Encoder) EncodeLink(text Text, url string) Text {
	if url == "" {
		return text
	}

	// return fmt.Sprintf("<a href=\"%s\">%s</a>", url, html.EscapeString(text))
	return e.EncodeText(fmt.Sprintf("[%s](%s)", text, url))
}

func (e *Encoder) EncodeModelLink(firstText Text, m any) Text {
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

func (e *Encoder) EncodeModelLinkDedupe(firstText Text, subsequentText Text, m any) Text {
	suffix := Text("")
	if p, ok := m.(*model.Person); ok {
		if p.RelationToKeyPerson.IsDirectAncestor() && !p.RelationToKeyPerson.IsSelf() {
			suffix = Text(DirectAncestorMarker)
		}
	}

	if e.LinkBuilder == nil {
		return firstText + suffix
	}

	url := e.LinkBuilder.LinkFor(m)

	// Only encode the first mention of a link
	if e.seenLinks == nil {
		e.seenLinks = make(map[string]bool)
	}

	if e.seenLinks[url] {
		return subsequentText
	}
	e.seenLinks[url] = true

	return e.EncodeLink(firstText+suffix, url)
}

func (e *Encoder) EncodeWithCitations(s Text, citations []*model.GeneralCitation) Text {
	if len(citations) == 0 {
		return s
	}
	sups := ""
	for i, cit := range citations {
		if cit.Redacted {
			continue
		}
		if i > 0 && sups != "" {
			sups += ","
		}
		sups += e.encodeCitationDetail(cit)
	}

	if sups == "" {
		return s
	}

	return e.EncodeText(s.String() + "<sup class=\"citref\">" + sups + "</sup>")
}

func (e *Encoder) encodeCitationDetail(c *model.GeneralCitation) string {
	citationText := c.Detail

	if c.ID != "" && e.LinkBuilder.LinkFor(c) != "" && (len(c.TranscriptionText) > 0 || len(c.MediaObjects) > 0) {
		citationText += " (" + e.EncodeModelLink("more details...", c).String() + ")"
	} else {
		if c.URL != nil {
			citationText = e.EncodeLink(e.EncodeText(citationText), c.URL.URL).String()
		}
	}

	return e.encodeCitationLink(c.SourceTitle(), Text(citationText), c.ID)
}

func (b *Encoder) encodeCitationLink(sourceName string, citationText Text, citationID string) string {
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

var htmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`'`, "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&#34;", // "&#34;" is shorter than "&quot;".
)

func (e *Encoder) Pre(s string) {
	e.maintext.WriteString("<pre>\n")
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		e.maintext.WriteString(htmlEscaper.Replace(line))
		e.maintext.WriteString("\n")
	}
	e.maintext.WriteString("</pre>\n")
}

func (e *Encoder) Image(alt string, link string) {
	e.writeImage(&e.maintext, alt, link)
}

func (b *Encoder) writeImage(buf io.StringWriter, alt string, link string) {
	buf.WriteString(fmt.Sprintf("![%s](%s)\n", alt, link))
}

func (e *Encoder) ParaWithFigure(text Text, link string, alt string, caption Text) {
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

func (e *Encoder) Figure(link string, alt string, caption Text, highlight *model.Region) {
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

func (e *Encoder) Timeline(rows []render.TimelineRow[Text]) {
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

func (e *Encoder) EncodeItalic(m Text) Text {
	return "*" + m + "*"
}

func (e *Encoder) EncodeBold(m Text) Text {
	return "**" + m + "**"
}

func (e *Encoder) EncodeImage(alt string, link string) string {
	buf := new(strings.Builder)
	e.writeImage(buf, alt, link)
	return buf.String()
}

func (e *Encoder) EncodeText(ss ...string) Text {
	if len(ss) == 0 {
		return ""
	} else if len(ss) == 1 {
		return Text(ss[0])
	}
	return Text(strings.Join(ss, ""))
}

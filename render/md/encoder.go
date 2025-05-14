package md

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/url"
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

type Content struct {
	LinkBuilder       LinkBuilder
	maintext          strings.Builder
	citationAnchors   map[string]citationAnchor // map of citation ids to html anchors
	sourceMap         map[string]sourceEntry    // map of source names to source entries
	seenLinks         map[string]bool
	SuppressCitations bool
}

var _ render.ContentBuilder[Text] = (*Content)(nil)

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

func (e *Content) SetLinkBuilder(l LinkBuilder) {
	e.LinkBuilder = l
}

func (e *Content) String() string {
	s := new(strings.Builder)
	e.WriteTo(s)
	return s.String()
}

func (e *Content) WriteTo(w io.Writer) (int64, error) {
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
		bb.WriteString("<h2>Citations and Notes</h2>\n")
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

func (e *Content) SetBody(s string) {
	e.maintext = strings.Builder{}
	e.maintext.WriteString(s)
}

func (e *Content) Markdown(s string) {
	e.maintext.WriteString("\n")
	e.maintext.WriteString(s)
}

func (e *Content) Heading2(m Text, id string) {
	e.maintext.WriteString("<h2>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h2>\n")
}

func (e *Content) Heading3(m Text, id string) {
	e.maintext.WriteString("<h3>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h3>\n")
}

func (e *Content) Heading4(m Text, id string) {
	e.maintext.WriteString("<h4>")
	m.ToHTML(&e.maintext)
	if id != "" {
		e.maintext.WriteString(fmt.Sprintf(" <a class=\"anchor\" id=\"%[1]s\" href=\"#%[1]s\">#</a>", html.EscapeString(id)))
	}
	e.maintext.WriteString("</h4>\n")
}

func (b *Content) Para(m Text) {
	b.maintext.WriteString("<p>")
	m.ToHTML(&b.maintext)
	b.maintext.WriteString("</p>\n")
}

func (e *Content) EmptyPara() {
	e.maintext.WriteString("<p></p>\n")
	// e.writeEmptyPara(&b.maintext)
}

func (b *Content) BlockQuote(m Text) {
	b.maintext.WriteString("<blockquote>")
	b.maintext.WriteString(string(m))
	b.maintext.WriteString("</blockquote>\n")
}

func (e *Content) UnorderedList(ms []Text) {
	e.maintext.WriteString("<ul>\n")
	for _, m := range ms {
		e.maintext.WriteString("<li>")
		m.ToHTML(&e.maintext)
		e.maintext.WriteString("</li>\n")
	}
	e.maintext.WriteString("</ul>\n")
}

func (e *Content) OrderedList(ms []Text) {
	e.maintext.WriteString("<ol>\n")
	for _, m := range ms {
		e.maintext.WriteString("<li>")
		m.ToHTML(&e.maintext)
		e.maintext.WriteString("</li>\n")
	}
	e.maintext.WriteString("</ol>\n")
}

func (e *Content) DefinitionList(markups [][2]Text) {
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

func (e *Content) ResetSeenLinks() {
	e.seenLinks = make(map[string]bool)
}

func (e *Content) EncodeLink(text Text, url string) Text {
	if url == "" {
		return text
	}

	// return fmt.Sprintf("<a href=\"%s\">%s</a>", url, html.EscapeString(text))
	return e.EncodeText(fmt.Sprintf("[%s](%s)", text, url))
}

func (e *Content) EncodeModelLink(firstText Text, m any) Text {
	decorator := Text("")
	if p, ok := m.(*model.Person); ok {
		if p.RelationToKeyPerson.IsDirectAncestor() && !p.RelationToKeyPerson.IsSelf() {
			decorator = Text(DirectAncestorMarker)
		}
	}
	firstText += decorator

	var link string
	if e.LinkBuilder != nil {
		link = e.LinkBuilder.LinkFor(m)
	}

	buf := new(strings.Builder)
	e.writeModelLink(buf, link, "", firstText.String(), "")
	return e.EncodeText(buf.String())
}

func (e *Content) EncodeModelLinkDedupe(firstText Text, subsequentText Text, m any) Text {
	decorator := Text("")
	if p, ok := m.(*model.Person); ok {
		if p.RelationToKeyPerson.IsDirectAncestor() && !p.RelationToKeyPerson.IsSelf() {
			decorator = Text(DirectAncestorMarker)
		}
	}
	var link string
	var name Text
	if e.LinkBuilder == nil {
		name = firstText
		name += decorator
	} else {
		link = e.LinkBuilder.LinkFor(m)
		if !e.seenLinks[link] {
			name = firstText
			name += decorator
		} else {
			name = subsequentText
		}
	}

	buf := new(strings.Builder)
	e.writeModelLink(buf, link, "", name.String(), "")
	return e.EncodeText(buf.String())
}

func (e *Content) EncodeModelLinkNamed(m any, nc render.NameChooser, pov *model.POV) Text {
	var decorator string
	if p, ok := m.(*model.Person); ok {
		if p.RelationToKeyPerson.IsDirectAncestor() && !p.RelationToKeyPerson.IsSelf() {
			decorator = DirectAncestorMarker
		}
	}

	var link string
	var prefix, name, suffix string
	if e.LinkBuilder == nil {
		prefix, name, suffix = nc.FirstUseSplit(m, pov)
		name += decorator
	} else {
		link = e.LinkBuilder.LinkFor(m)
		if !e.seenLinks[link] {
			prefix, name, suffix = nc.FirstUseSplit(m, pov)
			name += decorator
		} else {
			prefix, name, suffix = nc.SubsequentSplit(m, pov)
		}
	}

	buf := new(strings.Builder)
	e.writeModelLink(buf, link, prefix, name, suffix)
	return e.EncodeText(buf.String())
}

func (e *Content) writeModelLink(buf io.StringWriter, link string, prefix string, text string, suffix string) {
	if e.seenLinks == nil {
		e.seenLinks = make(map[string]bool)
	}
	buf.WriteString(prefix)
	if link == "" {
		buf.WriteString(text)
	} else {
		e.seenLinks[link] = true
		buf.WriteString(fmt.Sprintf("[%s](%s)", text, link))
	}
	buf.WriteString(suffix)
}

func (e *Content) EncodeWithCitations(s Text, citations []*model.GeneralCitation) Text {
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

func (e *Content) encodeCitationDetail(c *model.GeneralCitation) string {
	citationText := c.Detail

	if c.ID != "" && e.LinkBuilder.LinkFor(c) != "" && (len(c.TranscriptionText) > 0 || len(c.MediaObjects) > 0) {
		citationText += " (" + e.EncodeModelLink("more details...", c).String() + ")"
	} else {
		if c.URL != nil {
			if c.URL.Title != "" {
				citationText += " (" + e.EncodeLink(e.EncodeText(c.URL.Title), c.URL.URL).String() + ")"
			} else {
				pu, err := url.Parse(c.URL.URL)
				if err == nil && pu != nil && pu.Host != "" {
					host := pu.Host
					citationText += " (" + e.EncodeLink(e.EncodeText(host), c.URL.URL).String() + ")"
				} else {
					citationText = e.EncodeLink(e.EncodeText(citationText), c.URL.URL).String()
				}
			}
		}
	}

	return e.encodeCitationLink(c.SourceTitle(), Text(citationText), c.ID)
}

func (b *Content) encodeCitationLink(sourceName string, citationText Text, citationID string) string {
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

func (e *Content) Pre(s string) {
	e.maintext.WriteString("<pre>\n")
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		e.maintext.WriteString(htmlEscaper.Replace(line))
		e.maintext.WriteString("\n")
	}
	e.maintext.WriteString("</pre>\n")
}

func (e *Content) Image(link string, alt string) {
	e.writeImage(&e.maintext, link, alt)
}

func (b *Content) writeImage(buf io.StringWriter, link string, alt string) {
	buf.WriteString(fmt.Sprintf("![%s](%s)\n", alt, link))
}

func (e *Content) ParaWithFigure(text Text, link string, alt string, caption Text) {
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

func (e *Content) Figure(link string, alt string, caption Text, highlight *model.Region) {
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

func (e *Content) Timeline(rows []render.TimelineRow[Text]) {
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

func (e *Content) FactList(items []render.FactEntry[Text]) {
	e.maintext.WriteString(`<div class="facts-grid">` + "\n")
	for _, item := range items {
		e.maintext.WriteString("<div>")
		e.maintext.WriteString("<strong>")
		e.maintext.WriteString(item.Category)
		e.maintext.WriteString("</strong>")
		e.maintext.WriteString("<br>\n")
		for i, det := range item.Details {
			if i > 0 {
				e.maintext.WriteString("<br>\n")
			}
			det.ToHTML(&e.maintext)
		}
		e.maintext.WriteString("</div>")
	}
	e.maintext.WriteString("</div>")
}

func (e *Content) ConvertMarkdown(text string, w io.Writer) error {
	if err := goldmark.Convert([]byte(text), w); err != nil {
		return fmt.Errorf("goldmark: %v", err)
	}

	return nil
}

func (e *Content) EncodeItalic(m Text) Text {
	return "*" + m + "*"
}

func (e *Content) EncodeBold(m Text) Text {
	return "**" + m + "**"
}

func (e *Content) EncodeImage(alt string, link string) string {
	buf := new(strings.Builder)
	e.writeImage(buf, alt, link)
	return buf.String()
}

func (e *Content) EncodeText(ss ...string) Text {
	if len(ss) == 0 {
		return ""
	} else if len(ss) == 1 {
		return Text(ss[0])
	}
	return Text(strings.Join(ss, ""))
}

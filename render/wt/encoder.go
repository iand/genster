// Package wt provides types and functions for encoding wikitree markup
package wt

import (
	"fmt"
	"io"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

// Text is a piece of wikitree markup text
type Text string

func (m Text) String() string { return string(m) }
func (m Text) IsZero() bool   { return m == "" }

var _ render.ContentBuilder[Text] = (*Encoder)(nil)

type Encoder struct {
	main        strings.Builder
	seenLinks   map[any]bool
	citationidx int
	citationMap map[string]int
}

func (w *Encoder) String() string {
	s := new(strings.Builder)
	s.WriteString(w.main.String())
	s.WriteString("\n")

	if w.citationidx > 0 {
		s.WriteString("== Sources ==\n")
		s.WriteString("<references />")
	}
	return s.String()
}

func (w *Encoder) Para(m Text) {
	w.main.WriteString("\n")
	w.main.WriteString(string(m))
	w.main.WriteString("\n")
}

func (w *Encoder) EmptyPara() {
	w.main.WriteString("\n")
}

func (w *Encoder) Heading2(m Text, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("== " + string(m) + " ==")
	w.main.WriteString("\n")
}

func (w *Encoder) Heading3(m Text, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("=== " + string(m) + " ===")
	w.main.WriteString("\n")
}

func (w *Encoder) Heading4(m Text, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("==== " + string(m) + " ====")
	w.main.WriteString("\n")
}

func (w *Encoder) UnorderedList(items []Text) {
	for _, item := range items {
		w.main.WriteString("*" + string(item) + "\n")
	}
}

func (w *Encoder) OrderedList(items []Text) {
	for _, item := range items {
		w.main.WriteString("#" + string(item) + "\n")
	}
}

func (w *Encoder) DefinitionList(items [][2]Text) {
	for _, item := range items {
		w.main.WriteString(fmt.Sprintf("%s\n", string(item[0])))
		if len(item[1]) > 0 {
			w.main.WriteString(text.PrefixLines(string(item[1]), ":"))
			w.main.WriteString("\n")
		}
		w.main.WriteString("\n")
	}
}

func (w *Encoder) BlockQuote(m Text) {
	w.main.WriteString("<blockquote>\n")
	w.main.WriteString(m.String())
	w.main.WriteString("</blockquote>\n")
}

func (w *Encoder) Pre(s string) {
	w.main.WriteString("<pre>\n")
	w.main.WriteString(s)
	w.main.WriteString("</pre>\n")
}

func (w *Encoder) Markdown(s string) {
	// m.ToHTML(&w.main)
}

func (w *Encoder) EncodeItalic(m Text) Text {
	return "''" + m + "''"
}

func (w *Encoder) EncodeBold(m Text) Text {
	return "'''" + m + "'''"
}

func (w *Encoder) EncodeLink(text Text, url string) Text {
	return w.EncodeText(fmt.Sprintf("[%s %s]", url, text))
}

func (w *Encoder) EncodeModelLink(text Text, m any) Text {
	buf := new(strings.Builder)
	w.writeModelLink(buf, "", text.String(), "", m)
	return w.EncodeText(buf.String())
}

func (w *Encoder) EncodeModelLinkDedupe(firstText Text, subsequentText Text, m any) Text {
	// Only encode the first mention of a link
	if w.seenLinks == nil {
		w.seenLinks = make(map[any]bool)
	}

	var name Text
	if !w.seenLinks[m] {
		name = firstText
	} else {
		name = subsequentText
	}

	buf := new(strings.Builder)
	w.writeModelLink(buf, "", name.String(), "", m)
	return w.EncodeText(buf.String())
}

func (w *Encoder) EncodeModelLinkNamed(m any, nc render.NameChooser, pov *model.POV) Text {
	// Only encode the first mention of a link
	if w.seenLinks == nil {
		w.seenLinks = make(map[any]bool)
	}

	buf := new(strings.Builder)
	var prefix, name, suffix string
	if !w.seenLinks[m] {
		prefix, name, suffix = nc.FirstUseSplit(m, pov)
	} else {
		prefix, name, suffix = nc.SubsequentSplit(m, pov)
		w.seenLinks[m] = true
	}
	w.writeModelLink(buf, prefix, name, suffix, m)

	return w.EncodeText(buf.String())
}

func (w *Encoder) writeModelLink(buf io.StringWriter, prefix string, text string, suffix string, v any) {
	buf.WriteString(prefix)
	if p, ok := v.(*model.Person); ok && p.WikiTreeID != "" {
		buf.WriteString(fmt.Sprintf("[[%s|%s]]", p.WikiTreeID, text))
	} else {
		buf.WriteString(text)
	}
	buf.WriteString(suffix)
}

func (w *Encoder) EncodeWithCitations(s Text, citations []*model.GeneralCitation) Text {
	sups := Text("")
	for i, cit := range citations {
		if i > 0 && sups != "" {
			sups += ","
		}
		sups += w.encodeCitationDetail(cit)
	}
	return s + sups
}

func (w *Encoder) encodeCitationDetail(c *model.GeneralCitation) Text {
	var detail string

	detail = text.FinishSentence(c.String())

	if w.citationMap == nil {
		w.citationMap = make(map[string]int)
	}
	if c.ID != "" {
		idx, exists := w.citationMap[c.ID]
		if exists {
			return w.EncodeText(fmt.Sprintf(`<ref name="cit_%d" />`, idx))
		}
	}

	w.citationidx++
	idx := w.citationidx
	if c.ID != "" {
		w.citationMap[c.ID] = idx
	}

	return w.EncodeText(fmt.Sprintf(`<ref name="cit_%d">%s</ref>`, idx, detail))
}

func (w *Encoder) ParaWithFigure(s Text, link string, alt string, caption Text) {
}

func (w *Encoder) Timeline(rows []render.TimelineRow[Text]) {
	yr := ""
	for _, row := range rows {
		if len(row.Details) == 0 {
			continue
		}
		if yr != row.Year {
			w.main.WriteString(fmt.Sprintf("'''%s'''\n", row.Year))
			yr = row.Year
		}
		for _, d := range row.Details {
			if row.Date != "" {
				w.main.WriteString(fmt.Sprintf(":'''%s''' %s\n", row.Date, d))
			} else {
				w.main.WriteString(fmt.Sprintf(": %s\n", d))
			}
		}

	}
}

func (w *Encoder) FactList(items []render.FactEntry[Text]) {
	for _, item := range items {
		w.main.WriteString(fmt.Sprintf("%s\n", string(item.Category)))
		w.main.WriteString("\n")
	}
}

func hasExcludedTranscriptionSource(c *model.GeneralCitation) bool {
	// avoid text that might have problematic copyright
	if c.Source == nil || c.Source.RepositoryName == "" {
		return false
	}

	if c.Source.RepositoryName == "The British Newspaper Archive" {
		return true
	}

	return false
}

func (w *Encoder) EncodeText(ss ...string) Text {
	if len(ss) == 0 {
		return ""
	} else if len(ss) == 1 {
		return Text(ss[0])
	}
	return Text(strings.Join(ss, ""))
}

func (w *Encoder) Figure(link string, alt string, caption Text, highlight *model.Region) {
}

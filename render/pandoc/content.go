package pandoc

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

type Content struct {
	main      strings.Builder
	seenLinks map[string]bool
}

func (p *Content) WriteTo(w io.Writer) (int64, error) {
	bb := new(bytes.Buffer)
	bb.WriteString(p.main.String())
	return bb.WriteTo(w)
}

func (p *Content) EncodeText(ss ...string) Text {
	if len(ss) == 0 {
		return ""
	} else if len(ss) == 1 {
		return Text(ss[0])
	}
	return Text(strings.Join(ss, ""))
}

func (e *Content) Text() Text {
	s := new(strings.Builder)
	e.WriteTo(s)
	return Text(s.String())
}

func (e *Content) String() string {
	s := new(strings.Builder)
	e.WriteTo(s)
	return s.String()
}

func (w *Content) Para(m Text) {
	w.main.WriteString("\n")
	w.main.WriteString(string(m))
	w.main.WriteString("\n")
}

func (w *Content) EmptyPara() {
	w.main.WriteString("\n")
}

func (w *Content) Heading1(m Text, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("# " + string(m))
	w.main.WriteString("\n")
}

func (w *Content) Heading2(m Text, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("## " + string(m))
	w.main.WriteString("\n")
}

func (w *Content) Heading3(m Text, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("### " + string(m))
	w.main.WriteString("\n")
}

func (w *Content) Heading4(m Text, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("#### " + string(m))
	w.main.WriteString("\n")
}

func (w *Content) UnorderedList(items []Text) {
	for _, item := range items {
		w.main.WriteString("\n* " + string(item) + "\n")
	}
}

func (w *Content) OrderedList(items []Text) {
	for _, item := range items {
		w.main.WriteString("\n1. " + string(item) + "\n")
	}
}

func (w *Content) DefinitionList(items [][2]Text) {
	for _, item := range items {
		w.main.WriteString(fmt.Sprintf("%s\n", string(item[0])))
		if len(item[1]) > 0 {
			w.main.WriteString(text.PrefixLines(string(item[1]), ":"))
			w.main.WriteString("\n")
		}
		w.main.WriteString("\n")
	}
}

func (w *Content) BlockQuote(m Text) {
	w.main.WriteString(text.PrefixLines(m.String(), "> "))
}

func (w *Content) Pre(s string) {
	w.main.WriteString("<pre>\n")
	w.main.WriteString(s)
	w.main.WriteString("</pre>\n")
}

func (w *Content) Markdown(s string) {
	// m.ToHTML(&w.main)
}

func (w *Content) EncodeItalic(m Text) Text {
	return "*" + m + "*"
}

func (w *Content) EncodeBold(m Text) Text {
	return "**" + m + "**"
}

func (w *Content) EncodeLink(text Text, url string) Text {
	return w.EncodeText(fmt.Sprintf("[%s %s]", url, text))
}

func (w *Content) EncodeModelLink(text Text, m any) Text {
	var pageref string
	switch mt := m.(type) {
	case *model.Family:
		pageref = mt.ID
	}

	buf := new(strings.Builder)
	w.writeModelLink(buf, pageref, "", text.String(), "")
	return w.EncodeText(buf.String())
}

func (w *Content) EncodeModelLinkDedupe(firstText Text, subsequentText Text, m any) Text {
	var pageref string
	switch mt := m.(type) {
	case *model.Family:
		pageref = mt.ID
	}

	var name Text
	if !w.seenLinks[pageref] {
		name = firstText
	} else {
		name = subsequentText
	}

	buf := new(strings.Builder)
	w.writeModelLink(buf, pageref, "", name.String(), "")
	return w.EncodeText(buf.String())
}

func (w *Content) EncodeModelLinkNamed(m any, nc render.NameChooser, pov *model.POV) Text {
	var pageref string
	switch mt := m.(type) {
	case *model.Family:
		pageref = mt.ID
	}

	var prefix, name, suffix string
	if !w.seenLinks[pageref] {
		prefix, name, suffix = nc.FirstUseSplit(m, pov)
	} else {
		prefix, name, suffix = nc.SubsequentSplit(m, pov)
	}

	buf := new(strings.Builder)
	w.writeModelLink(buf, pageref, prefix, name, suffix)
	return w.EncodeText(buf.String())
}

func (w *Content) writeModelLink(buf io.StringWriter, pageref string, prefix string, text string, suffix string) {
	if w.seenLinks == nil {
		w.seenLinks = make(map[string]bool)
	}

	buf.WriteString(prefix)
	if pageref == "" {
		buf.WriteString(text)
	} else {
		w.seenLinks[pageref] = true
		buf.WriteString(text + " (page \\pageref{" + pageref + "})")
	}
	buf.WriteString(suffix)
}

func (w *Content) EncodeWithCitations(s Text, citations []*model.GeneralCitation) Text {
	sups := Text("")
	for i, cit := range citations {
		if i > 0 && sups != "" {
			sups += ","
		}
		sups += w.encodeCitationDetail(cit)
	}
	return s + sups
}

func (w *Content) encodeCitationDetail(c *model.GeneralCitation) Text {
	// var detail string

	// detail = text.FinishSentence(c.String())

	// if w.citationMap == nil {
	// 	w.citationMap = make(map[string]int)
	// }
	// if c.ID != "" {
	// 	idx, exists := w.citationMap[c.ID]
	// 	if exists {
	// 		return w.EncodeText(fmt.Sprintf(`<ref name="cit_%d" />`, idx))
	// 	}
	// }

	// w.citationidx++
	// idx := w.citationidx
	// if c.ID != "" {
	// 	w.citationMap[c.ID] = idx
	// }

	// return w.EncodeText(fmt.Sprintf(`<ref name="cit_%d">%s</ref>`, idx, detail))
	return ""
}

func (w *Content) Timeline([]render.TimelineRow[Text]) {
}

func (w *Content) Image(link string, alt string) {
	w.main.WriteString("![" + alt + "](" + link + ")")
}

// Requires implicit_figures extension
func (w *Content) Figure(link string, alt string, caption Text, highlight *model.Region) {
	w.main.WriteString("\n")
	w.main.WriteString("![" + caption.String() + "](" + link + ")")
	w.main.WriteString("\n")
}

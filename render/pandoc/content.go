package pandoc

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

type Content struct {
	main strings.Builder
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
		w.main.WriteString("*" + string(item) + "\n")
	}
}

func (w *Content) OrderedList(items []Text) {
	for _, item := range items {
		w.main.WriteString("#" + string(item) + "\n")
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
	w.main.WriteString("<blockquote>\n")
	w.main.WriteString(m.String())
	w.main.WriteString("</blockquote>\n")
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
	return "''" + m + "''"
}

func (w *Content) EncodeBold(m Text) Text {
	return "'''" + m + "'''"
}

func (w *Content) EncodeLink(text Text, url string) Text {
	return w.EncodeText(fmt.Sprintf("[%s %s]", url, text))
}

func (w *Content) EncodeModelLink(text Text, m any) Text {
	buf := new(strings.Builder)
	// w.writeModelLink(buf, text, m)
	return w.EncodeText(buf.String())
}

func (w *Content) EncodeModelLinkDedupe(firstText Text, subsequentText Text, m any) Text {
	// // Only encode the first mention of a link
	// if w.seenLinks == nil {
	// 	w.seenLinks = make(map[any]bool)
	// }

	buf := new(strings.Builder)
	// if !w.seenLinks[m] {
	// 	w.writeModelLink(buf, firstText, m)
	// } else {
	// 	w.writeModelLink(buf, subsequentText, m)
	// 	w.seenLinks[m] = true
	// }

	return w.EncodeText(buf.String())
}

// func (w *Page) writeModelLink(buf io.StringWriter, text Text, v any) {
// 	if p, ok := v.(*model.Person); ok && p.WikiTreeID != "" {
// 		buf.WriteString(fmt.Sprintf("[[%s|%s]]", p.WikiTreeID, text))
// 		return
// 	}

// 	// TODO:review whether to use Render instead
// 	buf.WriteString(text.String())
// }

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

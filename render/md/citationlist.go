package md

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"sort"
)

// CitationList tracks citations referenced during encoding and renders them
// as a footnotes section.
type CitationList struct {
	Suppress        bool
	citationAnchors map[string]citationAnchor
	sourceMap       map[string]sourceEntry
}

func (cl *CitationList) encodeCitationLink(sourceName string, citationText Text, citationID string) string {
	if cl.Suppress {
		return ""
	}

	if cl.citationAnchors == nil {
		cl.citationAnchors = make(map[string]citationAnchor)
	}

	ret := ""

	var anchor string
	ca, ok := cl.citationAnchors[citationID]
	if ok {
		anchor = ca.anchor
	} else {
		if cl.sourceMap == nil {
			cl.sourceMap = make(map[string]sourceEntry)
		}
		if sourceName == "" {
			sourceName = GeneralNotesSourceName
		}

		se, ok := cl.sourceMap[sourceName]
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
					index: len(cl.sourceMap),
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

		cl.sourceMap[sourceName] = se

		ca := citationAnchor{
			anchor:  anchor,
			display: anchor,
		}

		cl.citationAnchors[citationID] = ca
	}

	ret = fmt.Sprintf("<a href=\"#%s\">%s</a>", anchor, anchor)

	return ret
}

// WriteTo writes the accumulated citations and notes section. It returns the
// number of bytes written and any error encountered.
func (cl *CitationList) WriteTo(w io.Writer) (int64, error) {
	if len(cl.sourceMap) == 0 {
		return 0, nil
	}

	bb := new(bytes.Buffer)

	sources := make([]sourceEntry, 0, len(cl.sourceMap))
	for id := range cl.sourceMap {
		sources = append(sources, cl.sourceMap[id])
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

	return bb.WriteTo(w)
}

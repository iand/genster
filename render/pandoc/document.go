// Package wt provides types and functions for encoding pandoc flavoured markdown
package pandoc

import (
	"bytes"
	"fmt"
	"io"
)

type Document struct {
	Content
	frontMatter map[string]any
}

func (d *Document) WriteTo(w io.Writer) (int64, error) {
	bb := new(bytes.Buffer)
	// tagRanks := map[string]byte{
	// 	MarkdownTagID:      4,
	// 	MarkdownTagTitle:   3,
	// 	MarkdownTagLayout:  2,
	// 	MarkdownTagSummary: 1,
	// }

	// if len(d.frontMatter) > 0 {
	// 	bb.WriteString("---\n")

	// 	keys := make([]string, 0, len(d.frontMatter))
	// 	for k := range d.frontMatter {
	// 		keys = append(keys, k)
	// 	}
	// 	sort.Slice(keys, func(i, j int) bool {
	// 		ri := tagRanks[keys[i]]
	// 		rj := tagRanks[keys[j]]
	// 		if ri != rj {
	// 			return ri > rj
	// 		}
	// 		return keys[i] < keys[j]
	// 	})

	// 	for _, k := range keys {
	// 		bb.WriteString(k)
	// 		bb.WriteString(": ")

	// 		switch tv := d.frontMatter[k].(type) {
	// 		case string:
	// 			if safeString.MatchString(tv) && !numericString.MatchString(tv) {
	// 				bb.WriteString(tv)
	// 			} else {
	// 				bb.WriteString(fmt.Sprintf("%q", tv))
	// 			}
	// 			bb.WriteString("\n")
	// 		case []string:
	// 			bb.WriteString("\n")
	// 			for _, v := range tv {
	// 				bb.WriteString("- ")
	// 				if safeString.MatchString(v) && !numericString.MatchString(v) {
	// 					bb.WriteString(v)
	// 				} else {
	// 					bb.WriteString(fmt.Sprintf("%q", v))
	// 				}
	// 				bb.WriteString("\n")
	// 			}
	// 		case []map[string]string:
	// 			bb.WriteString("\n")
	// 			for _, v := range tv {
	// 				bb.WriteString("- ")
	// 				indent := false

	// 				for subkey, subval := range v {
	// 					if indent {
	// 						bb.WriteString("  ")
	// 					}
	// 					indent = true

	// 					if safeString.MatchString(subkey) && !numericString.MatchString(subkey) {
	// 						bb.WriteString(subkey)
	// 					} else {
	// 						bb.WriteString(fmt.Sprintf("%q", subkey))
	// 					}
	// 					bb.WriteString(": ")
	// 					if safeString.MatchString(subval) && !numericString.MatchString(subval) {
	// 						bb.WriteString(subval)
	// 					} else {
	// 						bb.WriteString(fmt.Sprintf("%q", subval))
	// 					}
	// 					bb.WriteString("\n")
	// 				}
	// 			}
	// 		default:
	// 			panic(fmt.Sprintf("unknown front matter type for key %s: %T", k, tv))
	// 		}

	// 	}
	// 	bb.WriteString("---\n")
	// }
	// bb.WriteString("\n")

	n, err := bb.WriteTo(w)
	if err != nil {
		return n, fmt.Errorf("write front matter: %w", err)
	}

	n1, err := d.Content.WriteTo(w)
	n += n1
	if err != nil {
		return n, fmt.Errorf("write body: %w", err)
	}
	return n, nil
}

func (d *Document) AppendText(t Text) {
	d.Content.main.WriteString(t.String())
}

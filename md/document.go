package md

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

var (
	safeString    = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericString = regexp.MustCompile(`^[0-9]+$`)
)

const (
	MarkdownTagTitle       = "title"
	MarkdownTagSummary     = "summary"
	MarkdownTagLayout      = "layout"
	MarkdownTagTags        = "tags"
	MarkdownTagCategory    = "category"
	MarkdownTagID          = "id"
	MarkdownTagImage       = "image"
	MarkdownTagBasePath    = "basepath"
	MarkdownTagNextPage    = "next"
	MarkdownTagPrevPage    = "prev"
	MarkdownTagFirstPage   = "first"
	MarkdownTagLastPage    = "last"
	MarkdownTagIndexPage   = "index"
	MarkdownTagAliases     = "aliases"
	MarkdownTagLinks       = "links"
	MarkdownTagDescendants = "descendants"
)

type LinkBuilder interface {
	LinkFor(v any) string
}

type Document struct {
	Encoder
	frontMatter map[string]any
}

func (b *Document) String() string {
	s := new(strings.Builder)
	b.WriteTo(s)
	return s.String()
}

func (b *Document) WriteTo(w io.Writer) (int64, error) {
	bb := new(bytes.Buffer)
	tagRanks := map[string]byte{
		MarkdownTagID:      4,
		MarkdownTagTitle:   3,
		MarkdownTagLayout:  2,
		MarkdownTagSummary: 1,
	}

	if len(b.frontMatter) > 0 {
		bb.WriteString("---\n")

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
			bb.WriteString(k)
			bb.WriteString(": ")

			switch tv := b.frontMatter[k].(type) {
			case string:
				if safeString.MatchString(tv) && !numericString.MatchString(tv) {
					bb.WriteString(tv)
				} else {
					bb.WriteString(fmt.Sprintf("%q", tv))
				}
				bb.WriteString("\n")
			case []string:
				bb.WriteString("\n")
				for _, v := range tv {
					bb.WriteString("- ")
					if safeString.MatchString(v) && !numericString.MatchString(v) {
						bb.WriteString(v)
					} else {
						bb.WriteString(fmt.Sprintf("%q", v))
					}
					bb.WriteString("\n")
				}
			case []map[string]string:
				bb.WriteString("\n")
				for _, v := range tv {
					bb.WriteString("- ")
					indent := false

					for subkey, subval := range v {
						if indent {
							bb.WriteString("  ")
						}
						indent = true

						if safeString.MatchString(subkey) && !numericString.MatchString(subkey) {
							bb.WriteString(subkey)
						} else {
							bb.WriteString(fmt.Sprintf("%q", subkey))
						}
						bb.WriteString(": ")
						if safeString.MatchString(subval) && !numericString.MatchString(subval) {
							bb.WriteString(subval)
						} else {
							bb.WriteString(fmt.Sprintf("%q", subval))
						}
						bb.WriteString("\n")
					}
				}
			default:
				panic(fmt.Sprintf("unknown front matter type for key %s: %T", k, tv))
			}

		}
		bb.WriteString("---\n")
	}
	bb.WriteString("\n")

	n, err := bb.WriteTo(w)
	if err != nil {
		return n, fmt.Errorf("write front matter: %w", err)
	}

	n1, err := b.Encoder.WriteTo(w)
	n += n1
	if err != nil {
		return n, fmt.Errorf("write body: %w", err)
	}

	return n, nil
}

func (b *Document) SetFrontMatterField(k, v string) {
	if b.frontMatter == nil {
		b.frontMatter = make(map[string]any)
	}
	b.frontMatter[k] = v
}

func (b *Document) appendFrontMatterField(k, v string) {
	if b.frontMatter == nil {
		b.frontMatter = make(map[string]any)
	}

	val, ok := b.frontMatter[k]
	if !ok {
		b.frontMatter[k] = []string{k}
		return
	}

	ss := val.([]string)
	ss = append(ss, v)
	b.frontMatter[k] = ss
}

func (b *Document) appendFrontMatterFieldDict(k string, v map[string]string) {
	if b.frontMatter == nil {
		b.frontMatter = make(map[string]any)
	}

	val, ok := b.frontMatter[k]
	if !ok {
		b.frontMatter[k] = []map[string]string{v}
		return
	}

	ms := val.([]map[string]string)
	ms = append(ms, v)
	b.frontMatter[k] = ms
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

func (b *Document) Category(s string) {
	b.SetFrontMatterField(MarkdownTagCategory, s)
}

func (b *Document) ID(s string) {
	b.SetFrontMatterField(MarkdownTagID, s)
}

func (b *Document) Image(s string) {
	b.SetFrontMatterField(MarkdownTagImage, s)
}

func (b *Document) BasePath(s string) {
	b.SetFrontMatterField(MarkdownTagBasePath, s)
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

func (b *Document) AddAlias(s string) {
	if s == "" {
		return
	}
	b.appendFrontMatterField(MarkdownTagAliases, s)
}

func (b *Document) AddLink(title string, link string) {
	if title == "" {
		return
	}
	b.appendFrontMatterFieldDict(MarkdownTagLinks, map[string]string{
		"title": title,
		"link":  link,
	})
}

func (b *Document) AddDescendant(name string, link string, detail string) {
	if name == "" {
		return
	}
	b.appendFrontMatterFieldDict(MarkdownTagDescendants, map[string]string{
		"name":   name,
		"link":   link,
		"detail": detail,
	})
}

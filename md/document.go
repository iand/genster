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
	MarkdownTagTitle     = "title"
	MarkdownTagSummary   = "summary"
	MarkdownTagLayout    = "layout"
	MarkdownTagTags      = "tags"
	MarkdownTagCategory  = "category"
	MarkdownTagID        = "id"
	MarkdownTagBasePath  = "basepath"
	MarkdownTagNextPage  = "next"
	MarkdownTagPrevPage  = "prev"
	MarkdownTagFirstPage = "first"
	MarkdownTagLastPage  = "last"
	MarkdownTagIndexPage = "index"
	MarkdownTagAliases   = "aliases"
)

type LinkBuilder interface {
	LinkFor(v any) string
}

type Document struct {
	Encoder
	frontMatter map[string][]string
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
			vs := b.frontMatter[k]
			bb.WriteString(k)
			bb.WriteString(": ")
			if len(vs) == 1 {
				if safeString.MatchString(vs[0]) && !numericString.MatchString(vs[0]) {
					bb.WriteString(vs[0])
				} else {
					bb.WriteString(fmt.Sprintf("%q", vs[0]))
				}
				bb.WriteString("\n")
			} else {
				bb.WriteString("\n")
				for _, v := range vs {
					bb.WriteString("- ")
					if safeString.MatchString(v) && !numericString.MatchString(v) {
						bb.WriteString(v)
					} else {
						bb.WriteString(fmt.Sprintf("%q", v))
					}
					bb.WriteString("\n")
				}
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

func (b *Document) Category(s string) {
	b.SetFrontMatterField(MarkdownTagCategory, s)
}

func (b *Document) ID(s string) {
	b.SetFrontMatterField(MarkdownTagID, s)
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

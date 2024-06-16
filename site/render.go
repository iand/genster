package site

import (
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

func RenderText(t model.Text, enc render.Page) error {
	if t.Formatted {
		enc.Pre(t.Text)
		enc.Pre("")
	} else if t.Markdown {
		txt := EncodeText(t, enc)
		enc.RawMarkdown(txt)
		enc.EmptyPara()
	} else {
		enc.Para(text.FormatSentence(t.Text))
		enc.EmptyPara()
	}

	return nil
}

func EncodeText(t model.Text, enc render.MarkupBuilder) string {
	if len(t.Links) == 0 {
		return t.Text
	}

	text := []rune(t.Text)

	// Ensure links are ordered by start position
	// Overlapping links are not supported
	sort.Slice(t.Links, func(i, j int) bool {
		return t.Links[i].Start < t.Links[j].Start
	})
	formatted := ""
	cursor := 0
	for _, l := range t.Links {
		formatted += string(text[cursor:l.Start])
		linktext := string(text[l.Start:l.End])
		formatted += enc.EncodeModelLink(linktext, l.Object)
		cursor = l.End
	}
	formatted += string(text[cursor:])
	return formatted
}

func RenderFacts(facts []model.Fact, pov *model.POV, enc render.MarkupBuilder) error {
	enc.EmptyPara()

	categories := make([]string, 0)
	factsByCategory := make(map[string][]*model.Fact)

	for _, f := range facts {
		f := f // avoid shadowing
		fl, ok := factsByCategory[f.Category]
		if ok {
			fl = append(fl, &f)
			factsByCategory[f.Category] = fl
			continue
		}

		categories = append(categories, f.Category)
		factsByCategory[f.Category] = []*model.Fact{&f}
	}

	sort.Strings(categories)

	factlist := make([][2]string, 0, len(categories))
	for _, cat := range categories {
		fl, ok := factsByCategory[cat]
		if !ok {
			continue
		}
		if len(fl) == 0 {
			factlist = append(factlist, [2]string{cat, fl[0].Detail})
			continue
		}
		buf := new(strings.Builder)
		for i, f := range fl {
			if i > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(enc.EncodeWithCitations(f.Detail, f.Citations))
		}
		factlist = append(factlist, [2]string{cat, buf.String()})
	}

	enc.DefinitionList(factlist)
	return nil
}

func CleanTags(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	tags := make([]string, 0, len(ss))
	for _, s := range ss {
		tag := Tagify(s)
		if seen[tag] {
			continue
		}
		tags = append(tags, tag)
		seen[tag] = true
	}
	sort.Strings(tags)
	return tags
}

func Tagify(s string) string {
	s = strings.ToLower(s)
	parts := strings.Fields(s)
	s = strings.Join(parts, "-")
	return s
}

func RenderNames(names []*model.Name, enc render.MarkupBuilder) error {
	enc.EmptyPara()

	namelist := make([]string, 0, len(names))
	for _, n := range names {
		namelist = append(namelist, enc.EncodeWithCitations(n.Name, n.Citations))
	}

	enc.UnorderedList(namelist)
	return nil
}

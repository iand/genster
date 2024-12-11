package site

import (
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
)

func RenderFacts[T render.EncodedText](facts []model.Fact, pov *model.POV, enc render.ContentBuilder[T]) error {
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

	factlist := make([][2]T, 0, len(categories))
	for _, cat := range categories {
		fl, ok := factsByCategory[cat]
		if !ok {
			continue
		}
		if len(fl) == 0 {
			factlist = append(factlist, [2]T{
				enc.EncodeText(cat),
				enc.EncodeText(fl[0].Detail),
			})
			continue
		}
		buf := new(strings.Builder)
		for i, f := range fl {
			if i > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(enc.EncodeWithCitations(enc.EncodeText(f.Detail), f.Citations).String())
		}
		factlist = append(factlist, [2]T{
			enc.EncodeText(cat),
			enc.EncodeText(buf.String()),
		})
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

func RenderNames[T render.EncodedText](names []*model.Name, enc render.ContentBuilder[T]) error {
	enc.EmptyPara()

	namelist := make([]T, 0, len(names))
	for _, n := range names {
		namelist = append(namelist, enc.EncodeWithCitations(enc.EncodeText(n.Name), n.Citations))
	}

	enc.UnorderedList(namelist)
	return nil
}

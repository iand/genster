package site

import (
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
)

func RenderFacts[T render.EncodedText](p *model.Person, pov *model.POV, enc render.ContentBuilder[T]) error {
	var items []render.FactEntry[T]

	notknown := enc.EncodeText("(not known)")

	eventDetails := func(ev model.TimelineEvent) []T {
		dets := make([]T, 0, 2)
		dets = append(dets, enc.EncodeText(ev.GetDate().String()))
		if !ev.GetPlace().IsUnknown() {
			dets = append(dets, enc.EncodeText(ev.GetPlace().FullName))
		}
		return dets
	}

	// Birth
	fe := render.FactEntry[T]{
		Category: "Birth",
	}
	if birth, ok := model.FindFirstEvent(p.Timeline, model.IsOwnBirthEvent(p)); ok {
		fe.Details = eventDetails(birth)
	} else {
		fe.Details = append(fe.Details, notknown)
	}
	items = append(items, fe)

	// Baptism
	fe = render.FactEntry[T]{
		Category: "Baptism",
	}
	if bap, ok := model.FindFirstEvent(p.Timeline, model.IsOwnBaptismEvent(p)); ok {
		fe.Details = eventDetails(bap)
	} else {
		fe.Details = append(fe.Details, notknown)
	}
	items = append(items, fe)

	// Death
	fe = render.FactEntry[T]{
		Category: "Death",
	}
	if death, ok := model.FindFirstEvent(p.Timeline, model.IsOwnDeathEvent(p)); ok {
		fe.Details = eventDetails(death)
	} else {
		fe.Details = append(fe.Details, notknown)
	}
	items = append(items, fe)

	// Burial
	fe = render.FactEntry[T]{
		Category: "Burial",
	}
	if bur, ok := model.FindFirstEvent(p.Timeline, model.IsOwnBurialEvent(p)); ok {
		fe.Details = eventDetails(bur)
	} else {
		fe.Details = append(fe.Details, notknown)
	}
	items = append(items, fe)

	// Other names
	if len(p.KnownNames) > 0 {
		fe := render.FactEntry[T]{
			Category: "Names and variations",
		}
		for _, n := range p.KnownNames {
			fe.Details = append(fe.Details, enc.EncodeWithCitations(enc.EncodeText(n.Name), n.Citations))
		}
		items = append(items, fe)
	}

	// Add miscellaneous facts
	categories := make([]string, 0)
	factsByCategory := make(map[string][]*model.Fact)

	for _, f := range p.MiscFacts {
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

	for _, cat := range categories {
		fl, ok := factsByCategory[cat]
		if !ok {
			continue
		}

		fe := render.FactEntry[T]{
			Category: cat,
		}

		for _, f := range fl {
			fe.Details = append(fe.Details, enc.EncodeWithCitations(enc.EncodeText(f.Detail), f.Citations))
		}
		items = append(items, fe)
	}

	enc.FactList(items)

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

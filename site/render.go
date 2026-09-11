package site

import (
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
)

func RenderVitalFacts[T render.EncodedText](p *model.Person, pov *model.POV, enc render.ContentBuilder[T]) error {
	var items []render.FactEntry[T]

	notknown := enc.EncodeText("(not known)")

	eventDetails := func(ev model.TimelineEvent) []T {
		dets := make([]T, 0, 2)
		dets = append(dets, enc.EncodeWithCitations(enc.EncodeText(ev.GetDate().String()), ev.GetCitations()))
		if !ev.GetPlace().IsUnknown() {
			dets = append(dets, enc.EncodeModelLink(enc.EncodeText(ev.GetPlace().FullName), ev.GetPlace()))
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
	if p.CauseOfDeath != nil {
		fe.Details = append(fe.Details, enc.EncodeWithCitations(enc.EncodeText("Death attributed to "+p.CauseOfDeath.Detail), p.CauseOfDeath.Citations))
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

	enc.FactList(items)

	return nil

}

func RenderMiscFacts[T render.EncodedText](p *model.Person, pov *model.POV, enc render.ContentBuilder[T]) error {
	var items []render.FactEntry[T]

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

func RenderPersonParents[T render.EncodedText](p *model.Person, pov *model.POV, enc render.ContentBuilder[T]) error {
	var items []render.FactEntry[T]

	notknown := enc.EncodeText("(not known)")

	// Father
	fe := render.FactEntry[T]{
		Category: "Father",
	}
	if !p.Father.IsUnknown() {
		fe.Details = append(fe.Details, enc.EncodeModelLink(enc.EncodeText(p.Father.PreferredUniqueName), p.Father))
	} else {
		fe.Details = append(fe.Details, notknown)
	}
	items = append(items, fe)

	// Mother
	fe = render.FactEntry[T]{
		Category: "Mother",
	}
	if !p.Mother.IsUnknown() {
		fe.Details = append(fe.Details, enc.EncodeModelLink(enc.EncodeText(p.Mother.PreferredUniqueName), p.Mother))
	} else {
		fe.Details = append(fe.Details, notknown)
	}
	items = append(items, fe)

	// TODO: possible add adoptive parents

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

func RenderPersonFamily[T render.EncodedText](p *model.Person, f *model.Family, pov *model.POV, enc render.ContentBuilder[T]) error {
	// nc := narrative.DefaultNameChooser{}
	other := f.OtherParent(p)
	otherName := enc.EncodeModelLink(enc.EncodeText(other.PreferredUniqueName), p)

	var what string
	switch f.Bond {
	case model.FamilyBondMarried:
		what = "Spouse"
	default:
		what = "Partner"
	}
	enc.Heading3(enc.EncodeText(what+" "+otherName.String()), "")

	// var detail text.Para
	// startDate := f.BestStartDate
	// var event string
	// if !startDate.IsUnknown() {
	// 	// detail.ReplaceSentence(intro.Pronoun(seq, s.Start(), s.Principal))
	// 	event = what + " " + otherName.String() + " " + narrative.NamedWhen(startDate)
	// 	if age, ok := p.AgeInYearsAt(startDate); ok && (age < 18 || age > 45) {
	// 		event += " " + narrative.AgeQualifier(age)
	// 	}
	// } else {
	// 	event = what + " " + otherName.String()
	// }
	// if f.BestStartEvent != nil && !f.BestStartEvent.GetPlace().IsUnknown() {
	// 	event = narrative.WhatWherePov(event, f.BestStartEvent.GetPlace(), enc, nc, pov)
	// }
	// if f.BestStartEvent != nil {
	// 	detail.Continue(enc.EncodeWithCitations(enc.EncodeText(event), f.BestStartEvent.GetCitations()).String())
	// } else {
	// 	detail.Continue(event)
	// }
	// enc.Heading4(enc.EncodeText(detail.Text()), "")

	return nil

}

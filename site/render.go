package site

import (
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
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

// RenderPersonFamilies renders the person's families, in chronological order, as
// a grouped list with each union's children listed beneath it.
func RenderPersonFamilies[T render.EncodedText](p *model.Person, pov *model.POV, enc render.ContentBuilder[T]) error {
	groups := make([]render.Grouping[T], 0, len(p.Families))
	for _, f := range p.Families {
		grp := render.Grouping[T]{
			Heading: familyHeading(f, p, enc),
			Notes:   familyNotes(f, pov, enc),
			Ordered: len(f.Children) > 0,
		}
		for _, c := range f.Children {
			line := enc.EncodeModelLink(enc.EncodeText(c.PreferredUniqueName), c)
			// A redacted child's details are withheld.
			if !c.Redacted {
				if note := childNote(c); note != "" {
					line = enc.EncodeText(text.AppendRelated(line.String(), note))
				}
			}
			grp.Items = append(grp.Items, line)
		}
		if len(f.Children) == 0 && f.AllChildrenKnown {
			grp.Items = append(grp.Items, enc.EncodeItalic(enc.EncodeText("No known children")))
		}
		groups = append(groups, grp)
	}

	enc.GroupedList(groups)

	return nil
}

// familyHeading composes the heading line for one of p's families, naming the
// other parent and indicating whether they were a spouse or a partner.
func familyHeading[T render.EncodedText](f *model.Family, p *model.Person, enc render.ContentBuilder[T]) T {
	other := f.OtherParent(p)
	married := f.Bond == model.FamilyBondMarried || f.Bond == model.FamilyBondLikelyMarried

	switch {
	case other.IsUnknown() && married:
		return enc.EncodeText("Unknown spouse")
	case other.IsUnknown():
		return enc.EncodeText("Unknown partner")
	default:
		role := "Partner"
		if married {
			role = "Spouse"
		}
		name := enc.EncodeBold(enc.EncodeModelLink(enc.EncodeText(other.PreferredUniqueName), other))
		return enc.EncodeText(role, " ", name.String())
	}
}

// familyNotes returns the marriage and divorce detail lines for a family, each
// with its date, place and citations. It returns nothing when neither is a
// recorded event, so a date derived from another fact, such as the birth of a
// child, is not presented as a union date.
func familyNotes[T render.EncodedText](f *model.Family, pov *model.POV, enc render.ContentBuilder[T]) []T {
	var notes []T

	if f.BestStartEvent != nil {
		note := "Married:"
		if f.BestStartDate != nil && !f.BestStartDate.IsUnknown() {
			note += " " + f.BestStartDate.String()
		}
		note += placeSuffix(f.BestStartEvent.GetPlace(), pov, enc)
		notes = append(notes, enc.EncodeWithCitations(enc.EncodeText(note), f.BestStartEvent.GetCitations()))
	}

	if f.EndReason == model.FamilyEndReasonDivorce && f.BestEndDate != nil && !f.BestEndDate.IsUnknown() {
		note := "Divorced: " + f.BestEndDate.String()
		if f.BestEndEvent != nil {
			note += placeSuffix(f.BestEndEvent.GetPlace(), pov, enc)
			notes = append(notes, enc.EncodeWithCitations(enc.EncodeText(note), f.BestEndEvent.GetCitations()))
		} else {
			notes = append(notes, enc.EncodeText(note))
		}
	}

	return notes
}

// childNote returns a short descriptive clause for a child, preferring the
// notable fact over the epithet, or an empty string when neither is set.
func childNote(c *model.Person) string {
	if c.Notable != "" {
		return c.Notable
	}
	return c.Epithet
}

// placeSuffix returns a leading-space suffix that links the place relative to the
// point of view, or an empty string when the place is unknown.
func placeSuffix[T render.EncodedText](pl *model.Place, pov *model.POV, enc render.ContentBuilder[T]) string {
	if pl.IsUnknown() {
		return ""
	}
	return " " + pl.InAt() + " " + enc.EncodeModelLinkNamed(pl, narrative.TimelineNameChooser{}, pov).String()
}

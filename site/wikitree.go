package site

import (
	"fmt"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/render/wt"
	"github.com/iand/genster/text"
)

func RenderWikiTreePage(s *Site, p *model.Person) (render.Page[md.Text], error) {
	pov := &model.POV{Person: p}
	_ = pov

	doc := s.NewDocument()
	doc.Layout(PageLayoutPerson.String())
	doc.ID(p.ID)
	doc.Title(p.PreferredUniqueName)
	doc.SetFrontMatterField("gender", p.Gender.Noun())
	if l := s.LinkForFormat(p, "markdown"); l != "" {
		doc.SetFrontMatterField("markdownformat", l)
	}

	if p.WikiTreeID != "" {
		doc.SetFrontMatterField("wikitreeid", p.WikiTreeID)
	}

	if p.Redacted {
		doc.Summary("information withheld to preserve privacy")
		return doc, nil
	}

	doc.Para(doc.EncodeModelLink(doc.EncodeText("Main page for "+p.PreferredFamiliarName), p))
	if p.WikiTreeID != "" {
		doc.Para(doc.EncodeLink(doc.EncodeText("WikiTree page for "+p.WikiTreeID), "https://www.wikitree.com/wiki/"+p.WikiTreeID))
	}

	if p.BestBirthlikeEvent != nil {
		doc.EmptyPara()

		birth := p.BestBirthlikeEvent.What()
		date := p.BestBirthlikeEvent.GetDate()
		if !date.IsUnknown() {
			birth = text.JoinSentenceParts(birth, date.String())
		}

		pl := p.BestBirthlikeEvent.GetPlace()
		if !pl.IsUnknown() {
			birth = text.JoinSentenceParts(birth, pl.InAt(), pl.PreferredFullName)
		}

		doc.Para(md.Text(text.UpperFirst(birth)))
	}

	if p.BestDeathlikeEvent != nil {
		doc.EmptyPara()

		death := p.BestDeathlikeEvent.What()
		date := p.BestDeathlikeEvent.GetDate()
		if !date.IsUnknown() {
			death = text.JoinSentenceParts(death, date.String())
		}

		pl := p.BestDeathlikeEvent.GetPlace()
		if !pl.IsUnknown() {
			death = text.JoinSentenceParts(death, pl.InAt(), pl.PreferredFullName)
		}

		doc.Para(md.Text(text.UpperFirst(death)))
	}

	encodeWikiTreeLink := func(p *model.Person) string {
		return doc.EncodeLink(doc.EncodeText(p.PreferredUniqueName), fmt.Sprintf(s.WikiTreeLinkPattern, p.ID)).String()
	}

	doc.EmptyPara()
	if p.Father.IsUnknown() {
		doc.Para("Father: unknown")
	} else {
		doc.Para(md.Text("Father: " + encodeWikiTreeLink(p.Father)))
	}

	doc.EmptyPara()
	if p.Mother.IsUnknown() {
		doc.Para("Mother: unknown")
	} else {
		doc.Para(md.Text("Mother: " + encodeWikiTreeLink(p.Mother)))
	}

	for seq, f := range p.Families {
		otherName := ""
		other := f.OtherParent(p)
		if other.IsUnknown() {
			continue
		} else {
			otherName = encodeWikiTreeLink(other)
		}

		action := ""
		switch f.Bond {
		case model.FamilyBondMarried:
			action += "married"
		case model.FamilyBondLikelyMarried:
			action += ChooseFrom(seq, "likely married", "probably married")
		default:
			action += "met"
		}

		rel := action
		rel += " " + otherName

		startDate := f.BestStartDate
		if !startDate.IsUnknown() {
			rel += ", " + startDate.String()
		}
		if f.BestStartEvent != nil {
			pl := f.BestStartEvent.GetPlace()
			if !pl.IsUnknown() {
				rel = text.JoinSentenceParts(rel, pl.InAt(), pl.PreferredFullName)
			}
		}

		doc.EmptyPara()
		doc.Para(md.Text(text.UpperFirst(rel)))

	}
	var children string
	var ancestorChild string
	var ancestorChildCitationPrefix string
	var ancestorChildCitations []*model.GeneralCitation

	wtenc := &wt.Encoder{}

	for seq, ch := range p.Children {
		childName := ""
		if ch.IsUnknown() {
			continue
		} else {
			childName = encodeWikiTreeLink(ch)
		}

		if ch.IsDirectAncestor() {
			ancestorChild = childName
			if ch.BestBirthlikeEvent != nil {
				ancestorChildCitations = ch.BestBirthlikeEvent.GetCitations()
				ancestorChildCitationPrefix = text.UpperFirst(ch.BestBirthlikeEvent.Type()) + " of " + ch.Gender.RelationToParentNoun()
			} else {
				ancestorChildCitations = nil
			}
		}

		if seq > 0 {
			children += ", "
		}
		children += childName
	}
	if children != "" {
		doc.EmptyPara()
		doc.Para(md.Text("Children: " + children))
	}

	if ancestorChild != "" {
		doc.EmptyPara()
		doc.Para(md.Text("Related via: " + ancestorChild))
		cits := make([]md.Text, 0, len(ancestorChildCitations))
		for _, cit := range ancestorChildCitations {
			cits = append(cits, doc.EncodeText(fmt.Sprintf("'''%s''': %s", ancestorChildCitationPrefix, cit.String())))
		}
		doc.UnorderedList(cits)
	}

	if p.Olb != "" {
		wtenc.Para(wtenc.EncodeItalic(wt.Text(text.FormatSentence(p.Olb))))
	}

	wtenc.Heading2("Biography", "")
	summary := PersonSummary(p, wtenc, FullNameChooser{}, wt.Text(p.PreferredFullName), true, true, false, false)
	wtenc.Para(summary)

	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(p.Timeline)),
	}
	for _, ev := range p.Timeline {
		if ev.GetDate().IsUnknown() && ev.GetPlace().IsUnknown() {
			continue
		}
		if !ev.DirectlyInvolves(p) {
			continue
		}
		switch ev.(type) {
		case *model.BirthEvent:
			t.Events = append(t.Events, ev)
		case *model.BaptismEvent:
			t.Events = append(t.Events, ev)
		case *model.CensusEvent:
			t.Events = append(t.Events, ev)
		case *model.MarriageEvent:
			t.Events = append(t.Events, ev)
		case *model.DeathEvent:
			t.Events = append(t.Events, ev)
		case *model.BurialEvent:
			t.Events = append(t.Events, ev)
		case *model.CremationEvent:
			t.Events = append(t.Events, ev)
		case *model.ProbateEvent:
			t.Events = append(t.Events, ev)
		}
	}

	if len(t.Events) > 0 {
		wtenc.Heading2("Timeline", "")
		fmtr := &WikiTreeTimelineEntryFormatter[wt.Text]{
			pov:      pov,
			nc:       FullNameChooser{},
			enc:      wtenc,
			omitDate: true,
			logger:   logging.Default(),
		}
		if err := RenderTimeline(t, pov, wtenc, fmtr); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	doc.Pre(wtenc.String())

	return doc, nil
}

type WikiTreeTimelineEntryFormatter[T render.EncodedText] struct {
	pov      *model.POV
	nc       NameChooser
	enc      render.TextEncoder[T]
	omitDate bool
	logger   *logging.Logger
}

func (t *WikiTreeTimelineEntryFormatter[T]) Title(seq int, ev model.TimelineEvent) string {
	var title string

	var what string

	switch tev := ev.(type) {
	case *model.MarriageEvent:
		spouse := tev.GetOther(t.pov.Person)
		what = "Marriage to " + t.nc.FirstUse(spouse)
	case *model.CensusEvent:
		what = "Appeared in census"
	default:
		what = ev.Type()
	}

	if t.omitDate {
		title = WhatWhere(what, ev.GetPlace(), t.enc, t.nc)
	} else {
		title = WhatWhenWhere(what, ev.GetDate(), ev.GetPlace(), t.enc, t.nc)
	}

	if title == "" {
		t.logger.Debug("timeline: ignored event type", "type", fmt.Sprintf("%T", ev))
		return ""
	}
	title = t.enc.EncodeWithCitations(t.enc.EncodeText(title), ev.GetCitations()).String()
	return text.FormatSentence(title)
}

func (t *WikiTreeTimelineEntryFormatter[T]) Detail(seq int, ev model.TimelineEvent) string {
	switch tev := ev.(type) {
	default:
		return tev.GetDetail()
	}
}

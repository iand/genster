package site

import (
	"fmt"

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
	for seq, ch := range p.Children {
		childName := ""
		if ch.IsUnknown() {
			continue
		} else {
			childName = encodeWikiTreeLink(ch)
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

	wtenc := &wt.Encoder{}

	if p.Olb != "" {
		wtenc.Para(wtenc.EncodeItalic(wt.Text(text.FormatSentence(p.Olb))))
	}

	wtenc.Heading2("Biography", "")

	summary := PersonSummary(p, wtenc, FullNameChooser{}, wt.Text(p.PreferredFullName), true, true, false, false)
	wtenc.Para(summary)

	doc.Pre(wtenc.String())

	return doc, nil
}

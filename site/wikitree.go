package site

import (
	"fmt"
	"io"
	"strings"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/text"
)

func RenderWikiTreePage(s *Site, p *model.Person) (render.Page[render.Markdown], error) {
	pov := &model.POV{Person: p}

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

	doc.Para(render.Markdown(doc.EncodeModelLink(doc.EncodeText("Main page for "+p.PreferredFamiliarName), p)))

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

		doc.Para(render.Markdown(text.UpperFirst(birth)))
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

		doc.Para(render.Markdown(text.UpperFirst(death)))
	}

	encodeWikiTreeLink := func(p *model.Person) string {
		return doc.EncodeLink(doc.EncodeText(p.PreferredUniqueName), fmt.Sprintf(s.WikiTreeLinkPattern, p.ID)).String()
	}

	doc.EmptyPara()
	if p.Father.IsUnknown() {
		doc.Para("Father: unknown")
	} else {
		doc.Para(render.Markdown("Father: " + encodeWikiTreeLink(p.Father)))
	}

	doc.EmptyPara()
	if p.Mother.IsUnknown() {
		doc.Para("Mother: unknown")
	} else {
		doc.Para(render.Markdown("Mother: " + encodeWikiTreeLink(p.Mother)))
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
		doc.Para(render.Markdown(text.UpperFirst(rel)))

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
		doc.Para(render.Markdown("Children: " + children))
	}

	tldoc := &WikiTreeEncoder{}

	if p.Olb != "" {
		tldoc.EmptyPara()
		tldoc.Para(render.Markdown(tldoc.EncodeBold("One line bio:").String() + " " + p.Olb))
	}

	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(p.Timeline)),
	}
	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) {
			continue
		}
		if ev.GetDate().IsUnknown() && ev.GetPlace().IsUnknown() {
			continue
		}
		t.Events = append(t.Events, ev)
	}

	if len(t.Events) > 0 {
		tldoc.EmptyPara()
		tldoc.Heading2("Timeline", "")
		if err := RenderTimeline(t, pov, tldoc); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	doc.Pre(tldoc.String())

	doc.EmptyPara()
	doc.Para("Annotation stub:")

	wikiTreeID := p.WikiTreeID
	if wikiTreeID == "" {
		wikiTreeID = "TBD"
	}

	ann := fmt.Sprintf(`    {
      "id": "%s",
      "comment": "%s",
      "replace": {
        "wikitreeid": "%s"
      }
    },`, p.ID, p.PreferredUniqueName, wikiTreeID)
	doc.Pre(ann)

	return doc, nil
}

var _ render.PageBuilder[render.Markdown] = (*WikiTreeEncoder)(nil)

type WikiTreeEncoder struct {
	main strings.Builder

	citationidx int
	citationMap map[string]int
}

func (w *WikiTreeEncoder) String() string {
	s := new(strings.Builder)
	s.WriteString(w.main.String())
	s.WriteString("\n")

	if w.citationidx > 0 {
		s.WriteString("== Sources ==\n")
		s.WriteString("<references />")
	}
	return s.String()
}

func (w *WikiTreeEncoder) Para(m render.Markdown) {
	w.main.WriteString("\n")
	w.main.WriteString(string(m))
	w.main.WriteString("\n")
}

func (w *WikiTreeEncoder) EmptyPara() {
	w.main.WriteString("\n")
}

func (w *WikiTreeEncoder) Heading2(m render.Markdown, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("== " + string(m) + " ==")
	w.main.WriteString("\n")
}

func (w *WikiTreeEncoder) Heading3(m render.Markdown, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("=== " + string(m) + " ===")
	w.main.WriteString("\n")
}

func (w *WikiTreeEncoder) Heading4(m render.Markdown, id string) {
	w.main.WriteString("\n")
	w.main.WriteString("==== " + string(m) + " ====")
	w.main.WriteString("\n")
}

func (w *WikiTreeEncoder) UnorderedList(items []render.Markdown) {
	for _, item := range items {
		w.main.WriteString("*" + string(item) + "\n")
	}
}

func (w *WikiTreeEncoder) OrderedList(items []render.Markdown) {
	for _, item := range items {
		w.main.WriteString("#" + string(item) + "\n")
	}
}

func (w *WikiTreeEncoder) DefinitionList(items [][2]render.Markdown) {
	for _, item := range items {
		w.main.WriteString(fmt.Sprintf("%s\n", string(item[0])))
		if len(item[1]) > 0 {
			w.main.WriteString(text.PrefixLines(string(item[1]), ":"))
			w.main.WriteString("\n")
		}
		w.main.WriteString("\n")
	}
}

func (w *WikiTreeEncoder) BlockQuote(m render.Markdown) {
	w.main.WriteString("<blockquote>\n")
	m.ToHTML(&w.main)
	w.main.WriteString("</blockquote>\n")
}

func (w *WikiTreeEncoder) Pre(s string) {
	w.main.WriteString("<pre>\n")
	w.main.WriteString(s)
	w.main.WriteString("</pre>\n")
}

func (w *WikiTreeEncoder) Markdown(s string) {
	// m.ToHTML(&w.main)
}

func (w *WikiTreeEncoder) EncodeItalic(m render.Markdown) render.Markdown {
	return render.Markdown("''" + m + "''")
}

func (w *WikiTreeEncoder) EncodeBold(m render.Markdown) render.Markdown {
	return render.Markdown("'''" + m + "'''")
}

func (w *WikiTreeEncoder) EncodeLink(text render.Markdown, url string) render.Markdown {
	return w.EncodeText(fmt.Sprintf("[%s %s]", url, text))
}

func (w *WikiTreeEncoder) EncodeModelLink(text render.Markdown, m any) render.Markdown {
	buf := new(strings.Builder)
	w.writeModelLink(buf, text, m)
	return w.EncodeText(buf.String())
}

func (w *WikiTreeEncoder) EncodeModelLinkDedupe(firstText render.Markdown, subsequentText render.Markdown, m any) render.Markdown {
	buf := new(strings.Builder)
	w.writeModelLink(buf, firstText, m)
	return w.EncodeText(buf.String())
}

func (w *WikiTreeEncoder) writeModelLink(buf io.StringWriter, text render.Markdown, v any) {
	if p, ok := v.(*model.Person); ok && p.WikiTreeID != "" {
		buf.WriteString(fmt.Sprintf("[[%s|%s]]", p.WikiTreeID, text))
		return
	}

	// TODO:review whether to use Render instead
	buf.WriteString(text.String())
}

func (w *WikiTreeEncoder) EncodeWithCitations(s render.Markdown, citations []*model.GeneralCitation) render.Markdown {
	sups := render.Markdown("")
	for i, cit := range citations {
		if i > 0 && sups != "" {
			sups += ","
		}
		sups += w.encodeCitationDetail(cit)
	}
	return s + sups
}

func (w *WikiTreeEncoder) encodeCitationDetail(c *model.GeneralCitation) render.Markdown {
	var detail string

	detail = text.AppendIndependentClause(detail, text.StripNewlines(c.Detail))

	if !hasExcludedTranscriptionSource(c) {
		if len(c.TranscriptionText) > 0 {
			for _, t := range c.TranscriptionText {
				detail = text.AppendIndependentClause(detail, `"`+w.EncodeItalic(w.EncodeText(text.StripNewlines(t.Text))).String()+`"`)
			}
		}
	}

	if c.Source != nil && c.Source.Title != "" {
		detail = text.AppendIndependentClause(detail, text.StripNewlines(c.Source.Title))
	}

	var repo string
	if c.Source != nil {
		if c.Source.RepositoryName != "" {
			if c.Source.RepositoryLink != "" {
				repo = w.EncodeLink(w.EncodeText(text.StripNewlines(c.Source.RepositoryName)), c.Source.RepositoryLink).String()
			} else {
				repo = text.StripNewlines(c.Source.RepositoryName)
			}
		} else {
			if c.Source.RepositoryLink != "" {
				repo = text.StripNewlines(c.Source.RepositoryLink)
			}
		}
	}

	if repo != "" {
		detail = text.AppendIndependentClause(detail, w.EncodeItalic(w.EncodeText(repo)).String())
	}
	detail = text.FinishSentence(detail)

	if c.URL != nil {
		detail = text.AppendIndependentClause(detail, w.EncodeLink(w.EncodeText(c.URL.Title), c.URL.URL).String())
		detail = text.FinishSentence(detail)
	}

	detail = text.FinishSentence(detail)

	if w.citationMap == nil {
		w.citationMap = make(map[string]int)
	}
	if c.ID != "" {
		idx, exists := w.citationMap[c.ID]
		if exists {
			return w.EncodeText(fmt.Sprintf(`<ref name="cit_%d" />`, idx))
		}
	}

	w.citationidx++
	idx := w.citationidx
	if c.ID != "" {
		w.citationMap[c.ID] = idx
	}

	return w.EncodeText(fmt.Sprintf(`<ref name="cit_%d">%s</ref>`, idx, detail))
}

func (w *WikiTreeEncoder) ParaWithFigure(s render.Markdown, link string, alt string, caption render.Markdown) {
}

func (w *WikiTreeEncoder) Timeline(rows []render.TimelineRow[render.Markdown]) {
}

func hasExcludedTranscriptionSource(c *model.GeneralCitation) bool {
	// avoid text that might have problematic copyright
	if c.Source == nil || c.Source.RepositoryName == "" {
		return false
	}

	if c.Source.RepositoryName == "The British Newspaper Archive" {
		return true
	}

	return false
}

func (w *WikiTreeEncoder) EncodeText(ss ...string) render.Markdown {
	if len(ss) == 0 {
		return ""
	} else if len(ss) == 1 {
		return render.Markdown(ss[0])
	}
	return render.Markdown(strings.Join(ss, ""))
}

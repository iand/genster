package site

import (
	"fmt"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
)

func RenderFamilyLinePage(s *Site, fl *model.FamilyLine) (render.Document[md.Text], error) {
	doc := s.NewDocument()
	doc.Layout(PageLayoutFamily.String())
	doc.Category(PageCategoryFamily)
	doc.SetSitemapDisable()
	doc.ID(fl.ID)
	doc.Title(fl.Name)

	eventsLink := fmt.Sprintf(s.FamilyLineEventsLinkPattern, fl.ID)
	doc.Para(doc.EncodeLink("View timeline", eventsLink))

	nc := &narrative.DefaultNameChooser{}
	for _, f := range fl.Families {
		doc.Heading2(doc.EncodeText(f.PreferredUniqueName), f.ID)
		n := BuildFamilyNarrative(f, false)
		n.Render(doc, nc)
	}
	return doc, nil
}

func RenderFamilyLineEventsPage(s *Site, fl *model.FamilyLine) (render.Document[md.Text], error) {
	doc := s.NewDocument()
	doc.Layout(PageLayoutFamily.String())
	doc.Category(PageCategoryFamily)
	doc.SetSitemapDisable()
	doc.ID(fl.ID + "-events")
	doc.Title(fl.Name + " Events")

	narrativeLink := fmt.Sprintf(s.FamilyLineLinkPattern, fl.ID)
	doc.Para(doc.EncodeLink("View narrative", narrativeLink))

	// Collect all unique people from the family line
	seen := make(map[*model.Person]bool)
	var people []*model.Person
	for _, f := range fl.Families {
		for _, p := range []*model.Person{f.Father, f.Mother} {
			if p == nil || p.IsUnknown() || p.Redacted || seen[p] {
				continue
			}
			seen[p] = true
			people = append(people, p)
		}
		for _, p := range f.Children {
			if p == nil || p.IsUnknown() || p.Redacted || seen[p] {
				continue
			}
			seen[p] = true
			people = append(people, p)
		}
	}

	// Collect all timeline events, deduplicating by pointer identity
	var allEvents []model.TimelineEvent
	for _, p := range people {
		allEvents = append(allEvents, p.Timeline...)
	}
	allEvents = model.CollapseEventList(allEvents)

	// Filter to events with known dates or places
	t := &model.Timeline{
		Events: make([]model.TimelineEvent, 0, len(allEvents)),
	}
	for _, ev := range allEvents {
		if !IncludeInTimeline(ev) {
			continue
		}
		if !ev.GetDate().IsUnknown() || !ev.GetPlace().IsUnknown() {
			t.Events = append(t.Events, ev)
		}
	}

	if len(t.Events) > 0 {
		doc.Heading2("Timeline", "")
		pov := &model.POV{
			Person: model.UnknownPerson(),
			Place:  model.UnknownPlace(),
		}

		fmtr := &citationIDFormatter[md.Text]{
			inner: NewNarrativeTimelineEntryFormatter(pov, doc, logging.Default(), &narrative.TimelineNameChooser{}, false),
		}

		if err := RenderTimeline(t, pov, doc, fmtr); err != nil {
			return doc, fmt.Errorf("render timeline: %w", err)
		}
	}

	return doc, nil
}

// citationIDFormatter wraps a TimelineEntryFormatter and appends the first
// citation's GrampsID in parentheses after the event text.
type citationIDFormatter[T render.EncodedText] struct {
	inner TimelineEntryFormatter[T]
}

func (f *citationIDFormatter[T]) Title(seq int, ev model.TimelineEvent) string {
	title := f.inner.Title(seq, ev)
	if title == "" {
		return ""
	}
	cits := ev.GetCitations()
	model.SortCitationsBySourceQuality(cits)
	for _, cit := range cits {
		if cit.GrampsID != "" {
			title += " (" + cit.GrampsID + ")"
			break
		}
	}
	return title
}

func (f *citationIDFormatter[T]) Detail(seq int, ev model.TimelineEvent) string {
	return f.inner.Detail(seq, ev)
}

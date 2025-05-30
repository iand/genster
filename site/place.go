package site

import (
	"fmt"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/text"
)

func RenderPlacePage(s *Site, p *model.Place) (render.Document[md.Text], error) {
	pov := &model.POV{Place: p}

	doc := s.NewDocument()

	doc.Title(p.Name)
	doc.Layout(PageLayoutPlace.String())
	doc.Category(PageCategoryPlace)
	doc.SetSitemapDisable()

	if p.UpdateTime != nil {
		doc.LastUpdated(*p.UpdateTime)
	}
	doc.SetFrontMatterField("placetype", p.PlaceType.String())
	doc.SetFrontMatterField("buildingkind", p.BuildingKind.String())
	doc.ID(p.ID)
	doc.AddTags(CleanTags(p.Tags))

	name := p.Name + " is a" + text.MaybeAn(p.PlaceType.String())

	if !p.Parent.IsUnknown() {
		name += " " + p.Parent.InAt() + " " + doc.EncodeModelLink(doc.EncodeText(p.Parent.FullName), p.Parent).String()
	}

	doc.Para(doc.EncodeItalic(doc.EncodeText(text.FinishSentence(name))))

	for _, t := range p.Comments {
		narrative.RenderText(t, doc)
	}

	t := &model.Timeline{
		Events: p.Timeline,
	}

	if len(p.Timeline) > 0 {
		doc.EmptyPara()
		doc.Heading2("Timeline", "")

		fmtr := &NarrativeTimelineEntryFormatter[md.Text]{
			pov:    pov,
			enc:    doc,
			logger: logging.Default(),
			nc:     &narrative.TimelineNameChooser{},
		}

		if err := RenderTimeline(t, pov, doc, fmtr); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.Links) > 0 {
		doc.Heading2("Links", "")
		for _, l := range p.Links {
			doc.Para(doc.EncodeLink(doc.EncodeText(l.Title), l.URL))
		}
	}

	if len(p.ResearchNotes) > 0 {
		doc.Heading2("Research Notes", "")
		for _, t := range p.ResearchNotes {
			narrative.RenderText(t, doc)
		}
	}

	// TODO: link to https://maps.nls.uk/geo/explore/#zoom=14&lat=52.32243&lon=1.26273&layers=161&b=1

	return doc, nil
}

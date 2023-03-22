package site

import (
	"fmt"

	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

func RenderPlacePage(s *Site, p *model.Place) (*md.Document, error) {
	pov := &model.POV{Place: p}

	d := s.NewDocument()
	b := d.Body()

	d.Title(p.PreferredName)
	d.Section(md.PageLayoutPlace)
	d.ID(p.ID)
	d.AddTags(CleanTags(p.Tags))

	desc := p.PreferredName + " is a" + text.MaybeAn(p.PlaceType.String())

	if !p.Parent.IsUnknown() {
		desc += " in " + b.EncodeModelLinkDedupe(p.Parent.PreferredUniqueName, p.Parent.PreferredName, p.Parent)
	}

	b.Para(text.FinishSentence(desc))

	t := &model.Timeline{
		Events: p.Timeline,
	}

	if len(p.Timeline) > 0 {
		b.EmptyPara()
		b.Heading2("Timeline")

		if err := RenderTimeline(t, pov, b); err != nil {
			return nil, fmt.Errorf("render timeline narrative: %w", err)
		}
	}

	if len(p.Links) > 0 {
		b.Heading2("Links")
		for _, l := range p.Links {
			b.Para(b.EncodeLink(l.Title, l.URL))
		}
	}

	return d, nil
}

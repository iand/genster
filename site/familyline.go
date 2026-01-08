package site

import (
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

	nc := &narrative.DefaultNameChooser{}
	for _, f := range fl.Families {
		doc.Heading2(doc.EncodeText(f.PreferredUniqueName), f.ID)
		n := BuildFamilyNarrative(f, false)
		n.Render(doc, nc)
	}
	return doc, nil
}

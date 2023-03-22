package site

import (
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
)

func RenderSourcePage(s *Site, sr *model.Source) (*md.Document, error) {
	d := s.NewDocument()
	b := d.Body()

	d.Title(sr.Title)
	d.Section(md.PageLayoutSource)
	d.ID(sr.ID)
	d.AddTags(CleanTags(sr.Tags))

	b.Para("Citations from this source")

	return d, nil
}

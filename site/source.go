package site

import (
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
)

func RenderSourcePage(s *Site, so *model.Source) (render.Page[render.Markdown], error) {
	doc := s.NewDocument()
	doc.SuppressCitations = true

	// doc.Title(so.Title)
	doc.Layout(PageLayoutSource.String())
	doc.Category(PageCategorySource)
	doc.ID(so.ID)

	doc.Title(so.Title)

	if len(so.RepositoryRefs) > 0 {
		repos := make([]render.Markdown, 0, len(so.RepositoryRefs))
		for _, rr := range so.RepositoryRefs {
			// 	rr := c.Source.RepositoryRefs[0]

			s := ""
			if rr.Repository.ShortName != "" {
				s += rr.Repository.ShortName
			} else {
				s += rr.Repository.Name
			}
			if rr.CallNo != "" {
				s += ". " + rr.CallNo
			}

			if s != "" {
				repos = append(repos, render.Markdown(s))
			}
		}

		if len(repos) > 0 {
			doc.EmptyPara()
			doc.Para("Obtainable from:")
			doc.UnorderedList(repos)
		}
	}

	return doc, nil
}

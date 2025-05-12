package gramps

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/iand/gdate"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

func (l *Loader) populateRepositoryFacts(m ModelFinder, gr *grampsxml.Repository) error {
	id := pval(gr.ID, gr.Handle)
	r := m.FindRepository(l.ScopeName, id)
	r.Name = gr.Rname

	return nil
}

func (l *Loader) populateSourceFacts(m ModelFinder, gs *grampsxml.Source) error {
	id := pval(gs.ID, gs.Handle)
	s := m.FindSource(l.ScopeName, id)

	s.Title = pval(gs.Stitle, "unknown")
	s.Author = pval(gs.Sauthor, "")

	if len(gs.Reporef) > 0 {
		// TODO: handle multiple repos
		repo, ok := l.RepositoriesByHandle[gs.Reporef[0].Hlink]
		if ok {
			s.RepositoryName = repo.Rname
		}

		for _, grr := range gs.Reporef {
			repo, ok := l.RepositoriesByHandle[grr.Hlink]
			if !ok {
				logging.Warn("could not find repository", "hlink", grr.Hlink)
			}
			r := m.FindRepository(l.ScopeName, pval(repo.ID, repo.Handle))

			rr := model.RepositoryRef{
				Repository: r,
				CallNo:     pval(grr.Callno, ""),
			}

			s.RepositoryRefs = append(s.RepositoryRefs, rr)
		}

	}

	return nil
}

func (l *Loader) parseCitationRecords(m ModelFinder, gcrs []grampsxml.Citationref, logger *slog.Logger) []*model.GeneralCitation {
	cits := make([]*model.GeneralCitation, 0)
	for _, gcr := range gcrs {
		gc, ok := l.CitationsByHandle[gcr.Hlink]
		if !ok {
			logger.Warn("could not find citation", "hlink", gcr.Hlink)
		}
		pc, err := l.parseCitation(m, gc, logger)
		if err != nil {
			logger.Error("dropping citation due to parse error", "error", err, "citation_handle", gcr.Hlink)
			continue
		}
		changeTime, err := changeToTime(gc.Change)
		if err == nil {
			pc.UpdateTime = &changeTime
		}
		createdTime, err := createdTimeFromHandle(gc.Handle)
		if err == nil {
			pc.CreateTime = &createdTime
		}

		cits = append(cits, pc)
	}
	return cits
}

func (l *Loader) parseCitation(m ModelFinder, gc *grampsxml.Citation, logger *slog.Logger) (*model.GeneralCitation, error) {
	id := pval(gc.ID, gc.Handle)
	cit, done := m.FindCitation(l.ScopeName, id)
	if done {
		return cit, nil
	}
	cit.Detail = pval(gc.Page, "")
	cit.GrampsID = pval(gc.ID, "")

	dt, err := CitationDate(gc, gdate.Parser{})
	if err != nil {
		return nil, fmt.Errorf("citation date: %w", err)
	}
	cit.Date = dt

	if gc.Sourceref != nil {
		gs, ok := l.SourcesByHandle[gc.Sourceref.Hlink]
		if ok {
			cit.Source = m.FindSource(l.ScopeName, pval(gs.ID, gs.Handle))
		}
	}

	for _, gnr := range gc.Noteref {
		gn, ok := l.NotesByHandle[gnr.Hlink]
		if !ok {
			continue
		}
		if pval(gn.Priv, false) {
			logger.Debug("skipping citation note marked as private", "handle", gn.Handle)
			continue
		}
		switch strings.ToLower(gn.Type) {
		case "transcript":
			cit.TranscriptionText = append(cit.TranscriptionText, l.parseNote(gn, m))
		case "general":
			cit.Comments = append(cit.Comments, l.parseNote(gn, m))
		case "research":
			// research notes are always assumed to be markdown
			t := l.parseNote(gn, m)
			t.Markdown = true
			cit.ResearchNotes = append(cit.ResearchNotes, t)
		}
	}

	for _, gor := range gc.Objref {
		if pval(gor.Priv, false) {
			logger.Debug("skipping citation object marked as private", "handle", gor.Hlink)
			continue
		}
		gob, ok := l.ObjectsByHandle[gor.Hlink]
		if ok {
			mo := m.FindMediaObject(gob.File.Src)
			mo.Citations = append(mo.Citations, cit)

			cmo := &model.CitedMediaObject{
				Object: mo,
			}
			if gor.Region != nil && gor.Region.Corner1x != nil && gor.Region.Corner1y != nil && gor.Region.Corner2x != nil && gor.Region.Corner2y != nil {
				cmo.Highlight = &model.Region{
					Left:   *gor.Region.Corner1x,
					Bottom: 100 - *gor.Region.Corner2y,
					Width:  *gor.Region.Corner2x - *gor.Region.Corner1x,
					Height: *gor.Region.Corner2y - *gor.Region.Corner1y,
				}
			}

			cit.MediaObjects = append(cit.MediaObjects, cmo)
		}
	}

	for _, att := range gc.Srcattribute {
		if pval(att.Priv, false) {
			logger.Debug("skipping citation attribute marked as private", "type", att.Type)
			continue
		}
		switch strings.ToLower(att.Type) {
		case "url":
			cit.URL = model.LinkFromURL(att.Value)
		}
	}

	return cit, nil
}

func CitationDate(gc *grampsxml.Citation, dp gdate.Parser) (*model.Date, error) {
	if gc.Dateval != nil {
		dt, err := ParseDateval(*gc.Dateval, dp)
		if err != nil {
			return nil, fmt.Errorf("parse date value: %w", err)
		}
		return dt, nil
	} else if gc.Daterange != nil {
		dt, err := ParseDaterange(*gc.Daterange, dp)
		if err != nil {
			return nil, fmt.Errorf("parse date range: %w", err)
		}
		return dt, nil
	} else if gc.Datespan != nil {
		dt, err := ParseDatespan(*gc.Datespan, dp)
		if err != nil {
			return nil, fmt.Errorf("parse date span: %w", err)
		}
		return dt, nil
	} else if gc.Datestr != nil {
		dt, err := dp.Parse(gc.Datestr.Val)
		if err != nil {
			return nil, fmt.Errorf("parse date value: %w", err)
		}
		return &model.Date{Date: dt}, nil
	}
	return model.UnknownDate(), nil
}

func (l *Loader) parseNote(gn *grampsxml.Note, m ModelFinder) model.Text {
	txt := model.Text{
		ID:   pval(gn.ID, ""),
		Text: gn.Text,
	}
	if pval(gn.Format, false) {
		txt.Formatted = true
	}

	styleOffset := 0
	if strings.HasPrefix(txt.Text, "### ") {
		bef, aft, ok := strings.Cut(txt.Text, "\n")
		if ok {
			styleOffset = -len(bef) - 1 // -1 for the newline
			txt.Title = bef[4:]
			txt.Text = aft
		}
	}

	for _, st := range gn.Style {
		switch st.Name {
		case "link":
			if st.Value != nil {
				obj, ok := l.resolveGrampsLink(*st.Value, m)
				if ok {
					for _, r := range st.Range {
						// TODO: adjust start/end to account for any additional formatting
						txt.Links = append(txt.Links, model.ObjectLink{
							Object: obj,
							Start:  r.Start + styleOffset,
							End:    r.End + styleOffset,
						})
					}
				}
			}
		}
	}

	if len(txt.Links) > 0 {
		txt.Markdown = true
	}

	// <note handle="_f8e8f631a9b39636b2dd988731b" change="1717158291" id="N0182" type="Person Note">
	//   <text>On her marriage certificate she stated her father to be George Palmer. However her father Daniel died while she was very young and she never knew him. Having met her descendants and seen her family bible I'm confident of the relationship and believe she was simply unaware of her father's name.</text>
	//   <style name="link" value="gramps://Person/handle/f8b62146b55125f15e326031dd7">
	//     <range start="90" end="96"/>
	//   </style>
	// </note>

	return txt
}

func (l *Loader) resolveGrampsLink(link string, m ModelFinder) (any, bool) {
	if !strings.HasPrefix(link, "gramps://") {
		return nil, false
	}
	link = link[9:]

	kind, link, found := strings.Cut(link, "/")
	if !found {
		return nil, false
	}

	reftype, ref, found := strings.Cut(link, "/")
	if !found {
		return nil, false
	}

	if reftype != "handle" {
		return nil, false
	}

	// handles begin with an underscore
	ref = "_" + ref

	switch kind {
	case "Person":
		gp, ok := l.PeopleByHandle[ref]
		if !ok {
			return nil, false
		}
		id := pval(gp.ID, gp.Handle)
		p := m.FindPerson(l.ScopeName, id)
		return p, true
	case "Place":
		gp, ok := l.PlacesByHandle[ref]
		if !ok {
			return nil, false
		}
		id := pval(gp.ID, gp.Handle)
		p := m.FindPlace(l.ScopeName, id)
		return p, true
	}

	return nil, false
}

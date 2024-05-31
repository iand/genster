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

func (l *Loader) parseCitationRecords(m ModelFinder, gcrs []grampsxml.Citationref, logger *slog.Logger) ([]*model.GeneralCitation, []*model.Anomaly) {
	cits := make([]*model.GeneralCitation, 0)
	anomalies := make([]*model.Anomaly, 0)
	for _, gcr := range gcrs {
		gc, ok := l.CitationsByHandle[gcr.Hlink]
		if !ok {
			logger.Warn("could not find citation", "hlink", gcr.Hlink)
		}
		pc, err := l.parseCitation(m, gc, logger)
		if err != nil {
			anomalies = append(anomalies, &model.Anomaly{
				Category: "Gramps",
				Text:     err.Error(),
				Context:  "Citation",
			})
			continue
		}
		cits = append(cits, pc)
	}
	return cits, anomalies
}

func (l *Loader) parseCitation(m ModelFinder, gc *grampsxml.Citation, logger *slog.Logger) (*model.GeneralCitation, error) {
	id := pval(gc.ID, gc.Handle)
	cit, done := m.FindCitation(l.ScopeName, id)
	if done {
		return cit, nil
	}
	cit.Detail = pval(gc.Page, "")

	dt, err := CitationDate(gc)
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
		switch strings.ToLower(gn.Type) {
		case "transcript":
			cit.TranscriptionText = append(cit.TranscriptionText, noteToText(gn))
		case "citation":
			cit.TranscriptionText = append(cit.TranscriptionText, noteToText(gn))
		}
	}

	for _, gor := range gc.Objref {
		gob, ok := l.ObjectsByHandle[gor.Hlink]
		if ok {
			mo := m.FindMediaObject(gob.File.Src)
			mo.Citations = append(mo.Citations, cit)
			cit.MediaObjects = append(cit.MediaObjects, mo)
		}
	}

	return cit, nil
}

func CitationDate(gc *grampsxml.Citation) (*model.Date, error) {
	dp := &gdate.Parser{
		AssumeGROQuarter: false,
	}

	if gc.Dateval != nil {
		dt, err := ParseDateval(*gc.Dateval)
		if err != nil {
			return nil, fmt.Errorf("parse date value: %w", err)
		}
		return dt, nil
	} else if gc.Daterange != nil {
		dt, err := ParseDaterange(*gc.Daterange)
		if err != nil {
			return nil, fmt.Errorf("parse date range: %w", err)
		}
		return dt, nil
	} else if gc.Datespan != nil {
		dt, err := ParseDatespan(*gc.Datespan)
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

func noteToText(gn *grampsxml.Note) model.Text {
	txt := model.Text{
		Text: gn.Text,
	}
	if pval(gn.Format, false) {
		txt.Formatted = true
	}
	return txt
}

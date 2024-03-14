package gramps

import (
	"log/slog"

	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

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
	cit := &model.GeneralCitation{
		ID:     id,
		Detail: pval(gc.Page, ""),
	}

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
		if gn.Type != "Transcript" {
			continue
		}
		cit.TranscriptionText = append(cit.TranscriptionText, gn.Text)

	}

	return cit, nil
}

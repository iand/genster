package gedcom

import (
	"github.com/iand/gedcom"
	"golang.org/x/exp/slog"
)

func (l *Loader) populateSourceFacts(m ModelFinder, sr *gedcom.SourceRecord) error {
	slog.Debug("populating from source record", "xref", sr.Xref)
	so := m.FindSource(l.ScopeName, sr.Xref)
	so.Title = sr.Title

	return nil
}

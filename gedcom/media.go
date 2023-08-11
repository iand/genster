package gedcom

import "github.com/iand/gedcom"

func (l *Loader) populateMediaFacts(m ModelFinder, mr *gedcom.MediaRecord) error {
	l.MediaRecordsByXref[mr.Xref] = mr
	return nil
}

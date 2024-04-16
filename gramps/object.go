package gramps

import (
	"fmt"

	"github.com/iand/grampsxml"
)

func (l *Loader) populateObjectFacts(m ModelFinder, gob *grampsxml.Object) error {
	mo := m.FindMediaObject(gob.File.Src)
	mo.MediaType = gob.File.Mime

	var ext string
	switch mo.MediaType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	default:
		return fmt.Errorf("unsupported media type: %v", mo.MediaType)
	}
	mo.FileName = fmt.Sprintf("%s.%s", mo.ID, ext)
	return nil
}

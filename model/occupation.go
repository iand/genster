package model

import (
	"github.com/iand/gdate"
)

type Occupation struct {
	StartDate   gdate.Date
	EndDate     gdate.Date
	Place       *Place
	Title       string
	Detail      string
	Citations   []*GeneralCitation
	Occurrences int
}

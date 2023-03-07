package model

import (
	"github.com/iand/gdate"
)

type Source struct {
	ID string // canonical id
	// Page  string // path to page in site
	Title string
	Tags  []string
}

type GeneralCitation struct {
	Source            *Source
	Detail            string
	ID                string
	TranscriptionDate gdate.Date
	TranscriptionText []string
	URL               *Link
	MediaObjects      []MediaObject
}

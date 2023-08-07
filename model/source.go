package model

type Source struct {
	ID                  string // canonical id
	Title               string
	SearchLink          string // link to online search interface
	RepositoryName      string
	RepositoryLink      string
	EventsCiting        []TimelineEvent
	Tags                []string
	IsCivilRegistration bool // indicates whether this source holds civil registration records such as births marriages and deaths
	IsCensus            bool // indicates whether this source holds census records
}

type GeneralCitation struct {
	Source            *Source
	Detail            string
	ID                string
	TranscriptionDate *Date
	TranscriptionText []string
	URL               *Link
	MediaObjects      []MediaObject
}

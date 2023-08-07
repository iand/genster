package model

type Source struct {
	ID                  string // canonical id
	Unknown             bool   // true if this source is known to have existed but no other information is known
	Title               string
	SearchLink          string // link to online search interface
	RepositoryName      string
	RepositoryLink      string
	EventsCiting        []TimelineEvent
	Tags                []string
	IsCivilRegistration bool // indicates whether this source holds civil registration records such as births marriages and deaths
	IsCensus            bool // indicates whether this source holds census records
}

func (s *Source) IsUnknown() bool {
	if s == nil {
		return true
	}
	return s.Unknown
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

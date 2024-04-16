package model

import (
	"strings"
)

type Source struct {
	ID                  string // canonical id
	Unknown             bool   // true if this source is known to have existed but no other information is known
	Title               string
	SearchLink          string // link to online search interface
	RepositoryName      string
	RepositoryLink      string
	RepositoryRefs      []RepositoryRef
	EventsCiting        []TimelineEvent
	Tags                []string
	IsCivilRegistration bool // indicates whether this source holds civil registration records such as births marriages and deaths
	IsCensus            bool // indicates whether this source holds census records
	IsUnreliable        bool // indicates whether this source is of dubious reliability
}

func (s *Source) IsUnknown() bool {
	if s == nil {
		return true
	}
	return s.Unknown
}

type GeneralCitation struct {
	Date              *Date
	Source            *Source
	Detail            string
	ID                string
	TranscriptionDate *Date
	TranscriptionText []Text
	URL               *Link
	MediaObjects      []*MediaObject
	EventsCited       []TimelineEvent
	PeopleCited       []*Person
}

type Text struct {
	Text      string
	Formatted bool
}

func (c *GeneralCitation) String() string {
	s := ""
	if c.Source != nil && c.Source.Title != "" {
		s = c.Source.Title

		if c.Detail != "" {
			if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "!") && !strings.HasSuffix(s, "?") {
				s += "; "
			}
			s += c.Detail
		}
	} else {
		s = c.Detail
	}

	return s
}

type Repository struct {
	ID        string // canonical id
	Name      string
	ShortName string
}

type RepositoryRef struct {
	Repository *Repository
	CallNo     string
}

type CitationMatcher func(*GeneralCitation) bool

// FilterCitationList returns a new slice that includes only the ciitations that match the
// supplied CitationMatcher
func FilterCitationList(cits []*GeneralCitation, include CitationMatcher) []*GeneralCitation {
	switch len(cits) {
	case 0:
		return []*GeneralCitation{}
	case 1:
		if include(cits[0]) {
			return []*GeneralCitation{cits[0]}
		}
		return []*GeneralCitation{}
	default:
		l := make([]*GeneralCitation, 0, len(cits))
		for _, c := range cits {
			if include(c) {
				l = append(l, c)
			}
		}
		return l
	}
}

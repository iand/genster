package model

import (
	"strings"
	"time"
)

type Source struct {
	ID                  string // canonical id
	Unknown             bool   // true if this source is known to have existed but no other information is known
	Title               string
	Author              string
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
	GrampsID          string // the original gramps id, if any
	TranscriptionDate *Date
	TranscriptionText []Text
	Comments          []Text
	ResearchNotes     []Text
	URL               *Link
	MediaObjects      []*CitedMediaObject
	EventsCited       []TimelineEvent
	PeopleCited       []*Person
	Redacted          bool       // true if the citation's details should be redacted
	LastUpdated       *time.Time // time of last update, if known
}

type CitedMediaObject struct {
	Object    *MediaObject
	Highlight *Region
}

// An ObjectLink is a specification of where to insert a link to an object in some text
type ObjectLink struct {
	Object any
	Start  int // rune index
	End    int // rune index
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

func (c *GeneralCitation) SourceTitle() string {
	if c.Source == nil {
		return ""
	}
	return c.Source.Title
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

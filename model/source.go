package model

import (
	"sort"
	"strings"
	"time"
)

type SourceQuality string

const (
	SourceQualityUnknown   SourceQuality = ""
	SourceQualityPrimary   SourceQuality = "primary"
	SourceQualitySecondary SourceQuality = "secondary"
	SourceQualityTertiary  SourceQuality = "tertiary"
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
	Quality             SourceQuality
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
	UpdateTime        *time.Time // time of last update, if known
	CreateTime        *time.Time // time of creation
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

func (c *GeneralCitation) Created() (time.Time, bool) {
	if c.CreateTime == nil {
		return time.Time{}, false
	}
	return *c.CreateTime, true
}

func (c *GeneralCitation) Updated() (time.Time, bool) {
	if c.UpdateTime == nil {
		return time.Time{}, false
	}
	return *c.UpdateTime, true
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

// sourceQualityOrder maps source quality to a sort rank, lower is better.
var sourceQualityOrder = map[SourceQuality]int{
	SourceQualityPrimary:   0,
	SourceQualitySecondary: 1,
	SourceQualityTertiary:  2,
	SourceQualityUnknown:   3,
}

// SortCitationsBySourceQuality sorts citations in place by the quality of their
// source in order: primary, secondary, tertiary, unknown. Sources marked as
// unreliable are sorted last.
func SortCitationsBySourceQuality(cits []*GeneralCitation) {
	sort.SliceStable(cits, func(i, j int) bool {
		si := cits[i].Source
		sj := cits[j].Source

		ui := si != nil && si.IsUnreliable
		uj := sj != nil && sj.IsUnreliable
		if ui != uj {
			return uj
		}

		qi := SourceQualityUnknown
		if si != nil {
			qi = si.Quality
		}
		qj := SourceQualityUnknown
		if sj != nil {
			qj = sj.Quality
		}
		return sourceQualityOrder[qi] < sourceQualityOrder[qj]
	})
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

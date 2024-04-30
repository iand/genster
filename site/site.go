package site

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
)

type PageLayout string

const (
	PageLayoutPerson         PageLayout = "person"
	PageLayoutPlace          PageLayout = "place"
	PageLayoutSource         PageLayout = "source"
	PageLayoutCitation       PageLayout = "citation"
	PageLayoutListInferences PageLayout = "listinferences"
	PageLayoutListAnomalies  PageLayout = "listanomalies"
	PageLayoutListTodo       PageLayout = "listtodo"
	PageLayoutListPeople     PageLayout = "listpeople"
	PageLayoutListPlaces     PageLayout = "listplaces"
	PageLayoutListSources    PageLayout = "listsources"
	PageLayoutListSurnames   PageLayout = "listsurnames"
	PageLayoutCalendar       PageLayout = "calendar"
	PageLayoutTreeOverview   PageLayout = "treeoverview"
	PageLayoutChartAncestors PageLayout = "chartancestors"
	PageLayoutChartTrees     PageLayout = "charttrees"
)

func (p PageLayout) String() string { return string(p) }

const (
	PageSectionPerson   = "person"
	PageSectionPlace    = "place"
	PageSectionCitation = "citation"
	PageSectionSource   = "source"
	PageSectionList     = "list"
	PageSectionChart    = "chart"
	PageSectionMedia    = "media"
)

const (
	PageCategoryPerson   = "person"
	PageCategoryCitation = "citation"
	PageCategorySource   = "source"
	PageCategoryPlace    = "place"
)

type Site struct {
	BaseURL   string
	Tree      *tree.Tree
	Calendars map[int]*Calendar

	PersonDir           string
	PersonLinkPattern   string
	PersonFilePattern   string
	SourceDir           string
	SourceLinkPattern   string
	SourceFilePattern   string
	CitationDir         string
	CitationLinkPattern string
	CitationFilePattern string
	FamilyLinkPattern   string
	CalendarLinkPattern string
	FamilyFilePattern   string
	PlaceDir            string
	PlaceLinkPattern    string
	PlaceFilePattern    string
	CalendarFilePattern string
	MediaDir            string
	MediaLinkPattern    string
	MediaFilePattern    string

	ListInferencesDir string
	ListAnomaliesDir  string
	ListTodoDir       string
	ListPeopleDir     string
	ListPlacesDir     string
	ListSourcesDir    string
	ListSurnamesDir   string

	ChartAncestorsDir string
	ChartTreesDir     string

	IncludePrivate     bool
	TimelineExperiment bool

	GenerateHugo bool

	GenerateWikiTree    bool
	WikiTreeDir         string
	WikiTreeLinkPattern string
	WikiTreeFilePattern string

	// PublishSet is the set of objects that will have pages written
	PublishSet *PublishSet
}

func NewSite(baseURL string, hugoIndexNaming bool, t *tree.Tree) *Site {
	indexPage := "index.md"
	if hugoIndexNaming {
		indexPage = "_index.md"
	}
	s := &Site{
		BaseURL:   baseURL,
		Tree:      t,
		Calendars: make(map[int]*Calendar),

		GenerateHugo: hugoIndexNaming,

		PersonDir:         PageSectionPerson,
		PersonLinkPattern: path.Join(baseURL, PageSectionPerson, "/%s/"),
		PersonFilePattern: path.Join(PageSectionPerson, "/%s/", indexPage),

		SourceDir:         PageSectionSource,
		SourceLinkPattern: path.Join(baseURL, PageSectionSource, "/%s/"),
		SourceFilePattern: path.Join(PageSectionSource, "/%s/", indexPage),

		CitationDir:         PageSectionCitation,
		CitationLinkPattern: path.Join(baseURL, PageSectionCitation, "/%s/"),
		CitationFilePattern: path.Join(PageSectionCitation, "/%s/", indexPage),

		FamilyLinkPattern: path.Join(baseURL, "family/%s/"),
		FamilyFilePattern: path.Join("/family/%s/", indexPage),

		PlaceDir:         PageSectionPlace,
		PlaceLinkPattern: path.Join(baseURL, PageSectionPlace, "/%s/"),
		PlaceFilePattern: path.Join(PageSectionPlace, "/%s/", indexPage),

		// ListDir:         PageSectionList,
		// ListPagePattern: path.Join(basePath, PageSectionList, "/%s/"),
		// ListFilePattern: "/place/%s/index.md",

		WikiTreeDir:         PageSectionPerson,
		WikiTreeLinkPattern: path.Join(baseURL, PageSectionPerson, "/%s/wikitree"),
		WikiTreeFilePattern: path.Join(PageSectionPerson, "/%s/wikitree.md"),

		CalendarLinkPattern: path.Join(baseURL, "calendar/%02d/"),
		CalendarFilePattern: "/calendar/%02d.md",

		MediaDir:         PageSectionMedia,
		MediaLinkPattern: path.Join(baseURL, PageSectionMedia, "/%s"),
		MediaFilePattern: path.Join(PageSectionMedia, "/%s"),

		ListInferencesDir: path.Join(PageSectionList, "inferences"),
		ListAnomaliesDir:  path.Join(PageSectionList, "anomalies"),
		ListTodoDir:       path.Join(PageSectionList, "todo"),
		ListPeopleDir:     path.Join(PageSectionList, "people"),
		ListPlacesDir:     path.Join(PageSectionList, "places"),
		ListSourcesDir:    path.Join(PageSectionList, "sources"),
		ListSurnamesDir:   path.Join(PageSectionList, "surnames"),

		ChartAncestorsDir: path.Join(PageSectionChart, "ancestors"),
		ChartTreesDir:     path.Join(PageSectionChart, "trees"),

		PublishSet: nil,
	}

	return s
}

func (s *Site) Generate() error {
	if err := s.Tree.Generate(!s.IncludePrivate); err != nil {
		return err
	}
	for _, p := range s.Tree.People {
		GenerateOlb(p)
		s.ScanPersonTodos(p)
		s.ScanPersonForAnomalies(p)
		s.AssignTags(p)
	}

	return nil
}

func (s *Site) AssignTags(p *model.Person) error {
	if p.RelationToKeyPerson != nil && p.RelationToKeyPerson.IsDirectAncestor() {
		p.Tags = append(p.Tags, "Direct Ancestor")
	}

	if p.PreferredFamilyName != model.UnknownNamePlaceholder && p.PreferredFamilyName != "" {
		p.Tags = append(p.Tags, p.PreferredFamilyName)
	}

	if p.Pauper {
		p.Tags = append(p.Tags, "Pauper")
	}

	if p.BornInWorkhouse {
		p.Tags = append(p.Tags, "Born in workhouse")
	}

	if p.DiedInWorkhouse {
		p.Tags = append(p.Tags, "Died in workhouse")
	}

	if p.DiedInChildbirth {
		p.Tags = append(p.Tags, "Died in childbirth")
	}

	if p.Twin {
		p.Tags = append(p.Tags, "Twin")
	}

	if p.Blind {
		p.Tags = append(p.Tags, "Blind")
	}

	if p.Deaf {
		p.Tags = append(p.Tags, "Deaf")
	}

	if p.Illegitimate {
		p.Tags = append(p.Tags, "Illegitimate")
	}

	switch p.CauseOfDeath {
	case model.CauseOfDeathSuicide:
		p.Tags = append(p.Tags, "Died by suicide")
	case model.CauseOfDeathLostAtSea:
		p.Tags = append(p.Tags, "Lost at sea")
	case model.CauseOfDeathKilledInAction:
		p.Tags = append(p.Tags, "Killed in action")
	case model.CauseOfDeathDrowned:
		p.Tags = append(p.Tags, "Drowned")
	}

	// if y, ok := gdate.AsYear(ev.Date); ok {
	// 	decade := (y.Year() / 10) * 10
	// 	p.Tags = append(p.Tags, fmt.Sprintf("born in %ds", decade))
	// }

	// if y, ok := gdate.AsYear(ev.Date); ok {
	// 	decade := (y.Year() / 10) * 10
	// 	p.Tags = append(p.Tags, fmt.Sprintf("died in %ds", decade))
	// }

	if p.WikiTreeID != "" {
		p.Tags = append(p.Tags, "WikiTree")
	}

	return nil
}

func (s *Site) BuildCalendar() error {
	monthEvents := make(map[int]map[model.TimelineEvent]struct{})

	for _, p := range s.Tree.People {
		for _, ev := range p.Timeline {
			_, indiv := ev.(model.IndividualTimelineEvent)
			_, party := ev.(model.PartyTimelineEvent)

			if !indiv && !party {
				continue
			}

			_, m, _, ok := ev.GetDate().YMD()
			if !ok {
				continue
			}

			if _, ok := ev.(*model.BirthEvent); ok {
			} else if _, ok := ev.(*model.BaptismEvent); ok {
			} else if _, ok := ev.(*model.DeathEvent); ok {
			} else if _, ok := ev.(*model.MarriageEvent); ok {
			} else if _, ok := ev.(*model.BurialEvent); ok {
			} else {
				continue
			}

			// Ensure unique events only
			evs, ok := monthEvents[m]
			if !ok {
				evs = make(map[model.TimelineEvent]struct{})
			}
			evs[ev] = struct{}{}
			monthEvents[m] = evs
		}
	}

	for m, evset := range monthEvents {
		c := new(Calendar)
		for ev := range evset {
			c.Events = append(c.Events, ev)
		}
		s.Calendars[m] = c
	}

	return nil
}

func normalizePlaceName(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	var seenChar bool
	var prevWasSpace bool
	var prevWasComma bool
	for _, c := range s {
		if !unicode.IsGraphic(c) {
			continue
		}
		if unicode.IsSpace(c) {
			// collapse whitespace
			if prevWasSpace || !seenChar {
				continue
			}
			prevWasSpace = true
			continue
		}

		if c == ',' {
			if prevWasComma || !seenChar {
				continue
			}
			prevWasComma = true
			prevWasSpace = true
			continue
		}

		if unicode.IsPunct(c) || unicode.IsSymbol(c) {
			continue
		}

		if prevWasComma {
			b.WriteRune(',')
			prevWasComma = false
		}
		if prevWasSpace {
			b.WriteRune(' ')
			prevWasSpace = false
		}
		b.WriteRune(unicode.ToLower(c))
		seenChar = true
	}
	return b.String()
}

func (s *Site) LinkFor(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		if vt.Redacted {
			return ""
		}
		if _, ok := s.PublishSet.People[vt.ID]; !ok {
			return ""
		}
		return fmt.Sprintf(s.PersonLinkPattern, vt.ID)
	case *model.GeneralCitation:
		if _, ok := s.PublishSet.Citations[vt.ID]; !ok {
			return ""
		}
		return fmt.Sprintf(s.CitationLinkPattern, vt.ID)
	case *model.Source:
		if _, ok := s.PublishSet.Sources[vt.ID]; !ok {
			return ""
		}
		return fmt.Sprintf(s.SourceLinkPattern, vt.ID)
	case *model.Family:
		if _, ok := s.PublishSet.Families[vt.ID]; !ok {
			return ""
		}
		return fmt.Sprintf(s.FamilyLinkPattern, vt.ID)
	case *model.Place:
		if _, ok := s.PublishSet.Places[vt.ID]; !ok {
			return ""
		}
		return fmt.Sprintf(s.PlaceLinkPattern, vt.ID)
	case *model.MediaObject:
		if _, ok := s.PublishSet.MediaObjects[vt.ID]; !ok {
			return ""
		}
		return fmt.Sprintf(s.MediaLinkPattern, vt.FileName)
	default:
		return ""
	}
}

func (s *Site) LinkForFormat(v any, format string) string {
	switch format {
	case "markdown":
		return s.LinkFor(v)
	case "wikitree":
		switch vt := v.(type) {
		case *model.Person:
			if vt.Redacted {
				return ""
			}
			if _, ok := s.PublishSet.People[vt.ID]; !ok {
				return ""
			}
			return fmt.Sprintf(s.WikiTreeLinkPattern, vt.ID)
		}
	}

	return ""
}

func (s *Site) ScanPersonForAnomalies(p *model.Person) {
	var birthEvents []model.TimelineEvent
	var baptismEvents []model.TimelineEvent
	var deathEvents []model.TimelineEvent
	var burialEvents []model.TimelineEvent

	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) {
			continue
		}
		anoms := ScanTimelineEventForAnomalies(ev)
		if len(anoms) > 0 {
			p.Anomalies = append(p.Anomalies, anoms...)
		}

		switch ev.(type) {
		case *model.BirthEvent:
			birthEvents = append(birthEvents, ev)
		case *model.BaptismEvent:
			baptismEvents = append(baptismEvents, ev)
		case *model.DeathEvent:
			deathEvents = append(deathEvents, ev)
		case *model.BurialEvent:
			burialEvents = append(burialEvents, ev)
		}
	}

	describeEvents := func(typ string, evs []model.TimelineEvent) string {
		dates := make(map[string]int)
		for _, ev := range evs {
			if ev.GetDate().IsUnknown() {
				dates["with an unknown date"]++
			} else {
				dates["dated "+ev.GetDate().String()]++
			}
		}
		var details []string
		for dt, count := range dates {
			if len(details) == 0 {
				details = append(details, text.CardinalNoun(count)+" "+typ+" "+text.MaybePluralise("event", count)+" "+dt)
			} else {
				details = append(details, text.CardinalNoun(count)+" "+dt)
			}
		}

		return "This person has " + text.JoinList(details)
	}

	if len(birthEvents) > 1 {
		txt := describeEvents("birth", birthEvents)
		p.Anomalies = append(p.Anomalies, &model.Anomaly{
			Category: model.AnomalyCategoryEvent,
			Text:     txt,
			Context:  "Multiple birth events",
		})
	}
	if len(deathEvents) > 1 {
		txt := describeEvents("death", deathEvents)

		p.Anomalies = append(p.Anomalies, &model.Anomaly{
			Category: model.AnomalyCategoryEvent,
			Text:     txt,
			Context:  "Multiple death events",
		})
	}
	if len(burialEvents) > 1 {
		txt := describeEvents("burial", burialEvents)

		p.Anomalies = append(p.Anomalies, &model.Anomaly{
			Category: model.AnomalyCategoryEvent,
			Text:     txt,
			Context:  "Multiple burial events",
		})
	}
	// TODO: mayeb remove this, people can have multiple baptisms
	if len(baptismEvents) > 1 {
		txt := describeEvents("baptism", baptismEvents)

		p.Anomalies = append(p.Anomalies, &model.Anomaly{
			Category: model.AnomalyCategoryEvent,
			Text:     txt,
			Context:  "Multiple baptism events",
		})
	}
}

type simplevalue interface {
	~int64 | ~float64 | ~string | ~int
}

type Tuple[K simplevalue, V simplevalue] struct {
	K K
	V V
}

func (t *Tuple[K, V]) String() string {
	return fmt.Sprintf("%v (%v)", t.K, t.V)
}

// FlattenMapByKeyAsc flattens a map into a slice of Tuples
func FlattenMap[M ~map[K]V, K simplevalue, V simplevalue](m M) []Tuple[K, V] {
	list := make([]Tuple[K, V], 0, len(m))
	for k, v := range m {
		list = append(list, Tuple[K, V]{K: k, V: v})
	}

	return list
}

// FlattenMapByKeyAsc flattens a map and sorts by the key ascending
func FlattenMapByKeyAsc[M ~map[K]V, K simplevalue, V simplevalue](m M) []Tuple[K, V] {
	list := FlattenMap(m)
	sort.Slice(list, func(i, j int) bool { return list[i].K < list[j].K })
	return list
}

// FlattenMapByValueDesc flattens a map and sorts by the value descending
func FlattenMapByValueDesc[M ~map[K]V, K simplevalue, V simplevalue](m M) []Tuple[K, V] {
	list := FlattenMap(m)
	sort.Slice(list, func(i, j int) bool { return list[i].V > list[j].V })
	return list
}

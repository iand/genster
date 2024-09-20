package site

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gosimple/slug"
	"github.com/iand/gedcom"
	"github.com/iand/genster/chart"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
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
	GedcomDir         string

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
		GedcomDir:         path.Join(PageSectionChart, "gedcom"),

		PublishSet: nil,
	}

	return s
}

func (s *Site) Generate() error {
	if err := s.Tree.Generate(!s.IncludePrivate); err != nil {
		return err
	}
	for _, p := range s.Tree.People {
		// GenerateOlb(p)
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

	switch p.ModeOfDeath {
	case model.ModeOfDeathSuicide:
		p.Tags = append(p.Tags, "Died by suicide")
	case model.ModeOfDeathLostAtSea:
		p.Tags = append(p.Tags, "Lost at sea")
	case model.ModeOfDeathKilledInAction:
		p.Tags = append(p.Tags, "Killed in action")
	case model.ModeOfDeathDrowned:
		p.Tags = append(p.Tags, "Drowned")
	case model.ModeOfDeathExecuted:
		p.Tags = append(p.Tags, "Executed")
	}
	// if y, ok := gdate.AsYear(ev.Date); ok {
	// 	decade := (y.Year() / 10) * 10
	// 	p.Tags = append(p.Tags, fmt.Sprintf("born in %ds", decade))
	// }

	// if y, ok := gdate.AsYear(ev.Date); ok {
	// 	decade := (y.Year() / 10) * 10
	// 	p.Tags = append(p.Tags, fmt.Sprintf("died in %ds", decade))
	// }
	return nil
}

func (s *Site) BuildCalendar() error {
	monthEvents := make(map[int]map[model.TimelineEvent]struct{})

	for _, p := range s.Tree.People {
		for _, ev := range p.Timeline {
			_, indiv := ev.(model.IndividualTimelineEvent)
			_, party := ev.(model.UnionTimelineEvent)

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
		if vt.PlaceType == model.PlaceTypeCategory {
			return ""
		}
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

func (s *Site) LinkForSurnameListPage(surname string) string {
	return filepath.Join(s.BaseURL, s.ListSurnamesDir, slug.Make(surname))
}

func (s *Site) RedirectPath(id string) string {
	return "/r/" + id
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

func (s *Site) WritePages(contentDir string, mediaDir string) error {
	for _, p := range s.PublishSet.People {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderPersonPage(s, p)
		if err != nil {
			return fmt.Errorf("render person page: %w", err)
		}

		if err := writePage(d, contentDir, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
			return fmt.Errorf("write person page: %w", err)
		}

		if s.GenerateWikiTree {
			d, err := RenderWikiTreePage(s, p)
			if err != nil {
				return fmt.Errorf("render wikitree page: %w", err)
			}

			if err := writePage(d, contentDir, fmt.Sprintf(s.WikiTreeFilePattern, p.ID)); err != nil {
				return fmt.Errorf("write wikitree page: %w", err)
			}
		}
	}

	for _, p := range s.PublishSet.Places {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderPlacePage(s, p)
		if err != nil {
			return fmt.Errorf("render place page: %w", err)
		}

		if err := writePage(d, contentDir, fmt.Sprintf(s.PlaceFilePattern, p.ID)); err != nil {
			return fmt.Errorf("write place page: %w", err)
		}
	}

	for _, c := range s.PublishSet.Citations {
		if s.LinkFor(c) == "" {
			continue
		}
		d, err := RenderCitationPage(s, c)
		if err != nil {
			return fmt.Errorf("render citation page: %w", err)
		}
		if err := writePage(d, contentDir, fmt.Sprintf(s.CitationFilePattern, c.ID)); err != nil {
			return fmt.Errorf("write citation page: %w", err)
		}
	}

	// Not publishing sources at this time
	// for _, so := range s.PublishSet.Sources {
	// 	if s.LinkFor(so) == "" {
	// 		continue
	// 	}
	// 	d, err := RenderSourcePage(s, so)
	// 	if err != nil {
	// 		return fmt.Errorf("render source page: %w", err)
	// 	}
	// 	if err := writePage(d, contentDir, fmt.Sprintf(s.SourceFilePattern, so.ID)); err != nil {
	// 		return fmt.Errorf("write source page: %w", err)
	// 	}
	// }

	for _, mo := range s.PublishSet.MediaObjects {
		// TODO: redaction

		// var ext string
		// switch mo.MediaType {
		// case "image/jpeg":
		// 	ext = "jpg"
		// case "image/png":
		// 	ext = "png"
		// case "image/gif":
		// 	ext = "gif"
		// default:
		// 	return fmt.Errorf("unsupported media type: %v", mo.MediaType)
		// }

		fname := filepath.Join(mediaDir, fmt.Sprintf("%s/%s", s.MediaDir, mo.FileName))

		if err := CopyFile(fname, mo.SrcFilePath); err != nil {
			return fmt.Errorf("copy media object: %w", err)
		}
	}

	s.BuildCalendar()

	for month, c := range s.Calendars {
		d, err := c.RenderPage(s)
		if err != nil {
			return fmt.Errorf("generate markdown: %w", err)
		}

		fname := fmt.Sprintf(s.CalendarFilePattern, month)

		f, err := CreateFile(filepath.Join(contentDir, fname))
		if err != nil {
			return fmt.Errorf("create calendar file: %w", err)
		}
		if _, err := d.WriteTo(f); err != nil {
			return fmt.Errorf("write calendar markdown: %w", err)
		}
		f.Close()
	}

	if err := s.WritePersonListPages(contentDir); err != nil {
		return fmt.Errorf("write people list pages: %w", err)
	}

	if err := s.WritePlaceListPages(contentDir); err != nil {
		return fmt.Errorf("write place list pages: %w", err)
	}

	// Not publishing sources at this time
	// if err := s.WriteSourceListPages(contentDir); err != nil {
	// 	return fmt.Errorf("write source list pages: %w", err)
	// }

	if err := s.WriteSurnameListPages(contentDir); err != nil {
		return fmt.Errorf("write surname list pages: %w", err)
	}

	if err := s.WriteInferenceListPages(contentDir); err != nil {
		return fmt.Errorf("write inferences pages: %w", err)
	}

	if err := s.WriteAnomalyListPages(contentDir); err != nil {
		return fmt.Errorf("write anomalies pages: %w", err)
	}

	if err := s.WriteTodoListPages(contentDir); err != nil {
		return fmt.Errorf("write todo pages: %w", err)
	}

	if err := s.WriteTreeOverview(contentDir); err != nil {
		return fmt.Errorf("write tree overview: %w", err)
	}

	if err := s.WriteChartAncestors(contentDir); err != nil {
		return fmt.Errorf("write ancestor chart: %w", err)
	}

	if err := s.WriteGedcom(contentDir); err != nil {
		return fmt.Errorf("write ancestor chart: %w", err)
	}

	// if err := s.WriteChartTrees(root); err != nil {
	// 	return fmt.Errorf("write chart trees: %w", err)
	// }

	return nil
}

func (s *Site) NewDocument() *md.Document {
	doc := &md.Document{}
	doc.LastUpdated(s.PublishSet.LastUpdated)
	doc.BasePath(s.BaseURL)
	doc.SetLinkBuilder(s)
	return doc
}

func (s *Site) NewMarkdownBuilder() render.PageBuilder[md.Text] {
	enc := &md.Encoder{}
	enc.SetLinkBuilder(s)

	return enc
}

func (s *Site) WriteTreeOverview(root string) error {
	doc := s.NewDocument()
	if s.Tree.Name != "" {
		doc.Title(s.Tree.Name)
	} else {
		doc.Title("Tree Overview")
	}
	doc.Layout(PageLayoutTreeOverview.String())

	fname := "index.md"
	if s.GenerateHugo {
		fname = "_index.md"
	}

	desc := s.Tree.Description

	if desc != "" {
		doc.Para(md.Text(text.FormatSentence(desc)))
	}

	peopleDesc := ""

	numberOfPeople := s.PublishSet.NumberOfPeople()
	if numberOfPeople > 0 {
		doc.EmptyPara()
		peopleDesc = text.FormatSentence(fmt.Sprintf("There are %d people in this tree", numberOfPeople))
	}

	// ancestorSurnames := FlattenMapByValueDesc(s.PublishSet.AncestorSurnameDistribution())
	ancestorSurnames := s.PublishSet.AncestorSurnameGroupList()
	if len(ancestorSurnames) > 0 {
		list := make([]string, 16)
		for i := range ancestorSurnames {
			if i > len(list)-1 {
				break
			}
			list[i] = doc.EncodeLink(doc.EncodeText(ancestorSurnames[i]), s.LinkForSurnameListPage(ancestorSurnames[i])).String()
		}
		detail := text.JoinSentenceParts("The principle surnames are ", text.JoinList(list))
		peopleDesc = text.JoinSentences(peopleDesc, text.FormatSentence(detail))
		peopleDesc = text.JoinSentences(peopleDesc, text.FormatSentence(text.JoinSentenceParts("See", doc.EncodeLink("all surnames...", s.ListSurnamesDir).String())))
	}

	if peopleDesc != "" {
		doc.EmptyPara()
		doc.Para(md.Text(peopleDesc))
	}

	doc.EmptyPara()
	doc.Para(md.Text(text.JoinSentenceParts("See a", doc.EncodeLink("full list of ancestors", s.ChartAncestorsDir).String(), "for", doc.EncodeModelLink(doc.EncodeText(s.Tree.KeyPerson.PreferredFamiliarFullName), s.Tree.KeyPerson).String())))

	// Featured people
	featuredPeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		if s.LinkFor(p) == "" {
			return false
		}
		return p.Featured
	}, 8)
	if len(featuredPeople) > 0 {
		model.SortPeopleByName(featuredPeople)
		doc.EmptyPara()
		doc.Heading2("Featured", "")
		items := make([]md.Text, len(featuredPeople))
		for i, p := range featuredPeople {
			items[i] = md.Text(text.AppendRelated(doc.EncodeModelLink(doc.EncodeText(p.PreferredUniqueName), p).String(), p.Olb))
		}
		doc.UnorderedList(items)
	}

	// Currently puzzling over
	puzzlePeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		if s.LinkFor(p) == "" {
			return false
		}
		return p.Puzzle && !p.Featured
	}, 8)
	if len(puzzlePeople) > 0 {
		model.SortPeopleByName(puzzlePeople)
		doc.EmptyPara()
		doc.Heading2("Currently puzzling over", "")
		doc.Para("These people are the focus of current research or are brick walls that we can't currently move past.")
		items := make([]md.Text, len(puzzlePeople))
		for i, p := range puzzlePeople {
			desc := p.Olb
			for _, rn := range p.ResearchNotes {
				if rn.Title != "" {
					desc = rn.Title
					break
				}
			}
			items[i] = md.Text(text.AppendRelated(doc.EncodeModelLink(doc.EncodeText(p.PreferredUniqueName), p).String(), desc))
		}
		doc.UnorderedList(items)
	}

	// People with research notes
	rnPeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		if s.LinkFor(p) == "" {
			return false
		}
		if len(p.ResearchNotes) == 0 {
			return false
		}
		for _, pp := range puzzlePeople {
			if pp.SameAs(p) {
				return false
			}
		}
		for _, pp := range featuredPeople {
			if pp.SameAs(p) {
				return false
			}
		}
		return true
	}, 3)
	if len(rnPeople) > 0 {
		model.SortPeopleByName(rnPeople)
		doc.EmptyPara()
		detail := text.JoinSentenceParts("Other people with research notes:", EncodePeopleListInline(rnPeople, func(p *model.Person) string {
			return p.PreferredFamiliarFullName
		}, doc))
		doc.Para(md.Text(text.FormatSentence(detail)))
	}

	doc.Heading2("Statistics and Records", "")

	// Oldest people
	earliestPeople := s.PublishSet.EarliestBorn(3)
	if len(earliestPeople) > 0 {
		doc.EmptyPara()
		detail := text.JoinSentenceParts("The earliest known births are:", EncodePeopleListInline(earliestPeople, func(p *model.Person) string {
			dt := p.BestBirthDate()
			yr, ok := dt.Year()
			if !ok {
				return p.PreferredFamiliarFullName
			}
			return fmt.Sprintf("%s (b. %d)", p.PreferredFamiliarFullName, yr)
		}, doc))
		doc.Para(md.Text(text.FormatSentence(detail)))

	}

	// Oldest people
	oldestPeople := s.PublishSet.OldestPeople(3)
	if len(oldestPeople) > 0 {
		doc.EmptyPara()
		detail := text.JoinSentenceParts("The people who lived the longest:", EncodePeopleListInline(oldestPeople, func(p *model.Person) string {
			age, _ := p.AgeInYearsAtDeath()
			return fmt.Sprintf("%s (%d years)", p.PreferredFamiliarFullName, age)
		}, doc))
		doc.Para(md.Text(text.FormatSentence(detail)))

	}

	// Greatest number of children people
	greatestChildrenPeople := s.PublishSet.GreatestChildren(6)
	if len(greatestChildrenPeople) > 0 {
		greatestChildrenPeopleDedupe := make([]*model.Person, 0, len(greatestChildrenPeople))
		for _, p := range greatestChildrenPeople {
			skipAddPerson := false
			for k, other := range greatestChildrenPeopleDedupe {
				if p.Gender == other.Gender {
					continue
				}
				if len(p.Children) != len(other.Children) {
					continue
				}
				// check if they were married
				married := false
				for _, sp := range p.Spouses {
					if sp.SameAs(other) {
						married = true
						break
					}
				}

				if married {
					skipAddPerson = true
					// keep the female
					if p.Gender == model.GenderFemale {
						greatestChildrenPeopleDedupe[k] = p
					}
					break
				}
			}

			if !skipAddPerson {
				greatestChildrenPeopleDedupe = append(greatestChildrenPeopleDedupe, p)
			}
		}

		if len(greatestChildrenPeopleDedupe) > 3 {
			greatestChildrenPeopleDedupe = greatestChildrenPeopleDedupe[:3]
		}

		doc.EmptyPara()
		detail := text.JoinSentenceParts("The people with the largest number of children:", EncodePeopleListInline(greatestChildrenPeopleDedupe, func(p *model.Person) string {
			return fmt.Sprintf("%s (%d)", p.PreferredFamiliarFullName, len(p.Children))
		}, doc))
		doc.Para(md.Text(text.FormatSentence(detail)))

	}

	var notes string
	if !s.Tree.KeyPerson.IsUnknown() {
		doc.EmptyPara()

		detail := text.JoinSentenceParts("In this family tree,", doc.EncodeModelLink(doc.EncodeText(s.Tree.KeyPerson.PreferredFamiliarFullName), s.Tree.KeyPerson).String(), "acts as the primary reference point, with all relationships defined in relation to", s.Tree.KeyPerson.Gender.ObjectPronoun())
		notes = text.JoinSentences(notes, text.FormatSentence(detail))
		notes = text.JoinSentences(notes, text.FormatSentence(text.JoinSentenceParts("Names suffixed by the", md.DirectAncestorMarker, "symbol indicate direct ancestors")))
	}

	if !s.IncludePrivate {
		detail := text.JoinSentenceParts("The tree excludes information on people who are possibly alive or who have died within the past twenty years")
		notes = text.JoinSentences(notes, text.FormatSentence(detail))
	}

	if len(notes) > 0 {
		doc.EmptyPara()
		doc.Heading3("Notes", "")
		doc.Para(md.Text(text.FormatSentence(notes)))
	}

	if err := writePage(doc, root, fname); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	return nil
}

func (s *Site) WriteChartAncestors(root string) error {
	fname := "index.md"
	if s.GenerateHugo {
		fname = "_index.md"
	}

	generations := 8
	ancestors := s.PublishSet.Ancestors(s.Tree.KeyPerson, generations)

	doc := s.NewDocument()
	doc.Title("Ancestors of " + s.Tree.KeyPerson.PreferredFamiliarFullName)
	doc.Summary(text.JoinSentenceParts("This is a full list of ancestors of", doc.EncodeModelLink(doc.EncodeText(s.Tree.KeyPerson.PreferredFamiliarFullName), s.Tree.KeyPerson).String()))
	doc.Layout(PageLayoutChartAncestors.String())

	g := 0
	doc.Heading3("Generation 1", "")

	doc.Para(md.Text(text.JoinSentenceParts("1.", doc.EncodeLink(doc.EncodeText(s.Tree.KeyPerson.PreferredFamiliarFullName), doc.LinkBuilder.LinkFor(s.Tree.KeyPerson)).String())))
	for i := range ancestors {
		ig := -1
		idx := i + 2
		for idx > 0 {
			idx >>= 1
			ig++
		}
		if ig != g {
			g = ig
			if g == 1 {
				doc.Heading3("Generation 2: Parents", "p")
			} else if g == 2 {
				doc.Heading3("Generation 3: Grandparents", "gp")
			} else if g == 3 {
				doc.Heading3("Generation 4: Great-Grandparents", "ggp")
			} else if g == 4 {
				doc.Heading3("Generation 5: Great-Great-Grandparents", "gggp")
			} else {
				doc.Heading3(md.Text(fmt.Sprintf("Generation %d: %dx Great-Grandparents", g+1, g-2)), "")
			}
		}
		if ancestors[i] != nil {
			detail := text.JoinSentenceParts(fmt.Sprintf("%d.", i+2), doc.EncodeBold(doc.EncodeLink(doc.EncodeText(ancestors[i].PreferredFullName), doc.LinkBuilder.LinkFor(ancestors[i]))).String())

			var adds []string
			if ancestors[i].PrimaryOccupation != "" {
				adds = append(adds, ancestors[i].PrimaryOccupation)
			}
			if ancestors[i].BestBirthlikeEvent != nil && !ancestors[i].BestBirthlikeEvent.GetDate().IsUnknown() {
				adds = append(adds, EventWhatWhenWhere(ancestors[i].BestBirthlikeEvent, doc, DefaultNameChooser{}))
			}
			if ancestors[i].BestDeathlikeEvent != nil && !ancestors[i].BestDeathlikeEvent.GetDate().IsUnknown() {
				adds = append(adds, EventWhatWhenWhere(ancestors[i].BestDeathlikeEvent, doc, DefaultNameChooser{}))
			}

			detail = text.AppendClause(detail, text.JoinList(adds))
			doc.Para(md.Text(detail))
		} else {

			name := "Not known"
			// Odd numbers are female, even numbers are male.
			// The child of entry n is found at (n-2)/2 if n is even and (n-3)/2 if n is odd.

			if i%2 == 0 {
				// male
				lb := (i - 2) / 2
				if lb >= 0 && ancestors[lb] != nil {
					name += " (father of " + ancestors[lb].PreferredFullName + ")"
				} else {
					lb = (lb - 2) / 2
					if lb >= 0 && ancestors[lb] != nil {
						name += " (grandfather of " + ancestors[lb].PreferredFullName + ")"
					} else {
						name += " (male)"
					}
				}
			} else {
				// female
				lb := (i - 3) / 2
				if lb >= 0 && ancestors[lb] != nil {
					name += " (mother of " + ancestors[lb].PreferredFullName + ")"
				} else {
					lb = (lb - 2) / 2
					if lb >= 0 && ancestors[lb] != nil {
						name += " (grandmother of " + ancestors[lb].PreferredFullName + ")"
					} else {
						name += " (female)"
					}
				}
			}

			doc.Para(md.Text(text.JoinSentenceParts(fmt.Sprintf("%d.", i+2), name)))
		}
	}

	baseDir := filepath.Join(root, s.ChartAncestorsDir)
	if err := writePage(doc, baseDir, fname); err != nil {
		return fmt.Errorf("failed to write ancestor overview: %w", err)
	}

	return nil
}

func (s *Site) WriteChartTrees(root string) error {
	fname := "index.md"
	if s.GenerateHugo {
		fname = "_index.md"
	}

	generations := 8
	ancestors := s.PublishSet.Ancestors(s.Tree.KeyPerson, generations)

	doc := s.NewDocument()
	doc.Title("Family Trees")
	doc.Summary(text.JoinSentenceParts("This is a list of family trees generated for various people"))
	doc.Layout(PageLayoutChartTrees.String())

	// index 14-29 are great-great grandparents, only produce chart if they have no known parents
	for i := 14; i <= 29; i++ {
		if ancestors[i] == nil {
			continue
		}

		if ancestors[i].Father != nil || ancestors[i].Mother != nil {
			continue
		}

		fname := filepath.Join(s.ChartTreesDir, ancestors[i].ID+".svg")
		if err := s.WriteDescendantTree(filepath.Join(root, fname), ancestors[i], 2); err != nil {
			return fmt.Errorf("failed to write descendant tree: %w", err)
		}
		doc.Para(doc.EncodeBold(doc.EncodeLink(doc.EncodeText(ancestors[i].PreferredUniqueName), ancestors[i].ID+".svg")))
	}

	// index 30-61 are great-great-great grandparents, only produce chart if they have no known parents, but at a greater depth
	for i := 30; i <= 61; i++ {
		if ancestors[i] == nil {
			continue
		}

		if ancestors[i].Father != nil || ancestors[i].Mother != nil {
			continue
		}

		fname := filepath.Join(s.ChartTreesDir, ancestors[i].ID+".svg")
		if err := s.WriteDescendantTree(filepath.Join(root, fname), ancestors[i], 3); err != nil {
			return fmt.Errorf("failed to write descendant tree: %w", err)
		}
		doc.Para(doc.EncodeBold(doc.EncodeLink(doc.EncodeText(ancestors[i].PreferredUniqueName), ancestors[i].ID+".svg")))
	}

	// produce chart for each member of a later generation
	for i := 62; i < len(ancestors); i++ {
		if ancestors[i] == nil {
			continue
		}
		fname := filepath.Join(s.ChartTreesDir, ancestors[i].ID+".svg")
		if err := s.WriteDescendantTree(filepath.Join(root, fname), ancestors[i], 4); err != nil {
			return fmt.Errorf("failed to write descendant tree: %w", err)
		}
		doc.Para(doc.EncodeBold(doc.EncodeLink(doc.EncodeText(ancestors[i].PreferredUniqueName), ancestors[i].ID+".svg")))
	}

	baseDir := filepath.Join(root, s.ChartTreesDir)
	if err := writePage(doc, baseDir, fname); err != nil {
		return fmt.Errorf("failed to write chart trees index: %w", err)
	}

	return nil
}

func (s *Site) WriteDescendantTree(fname string, p *model.Person, depth int) error {
	ch, err := chart.BuildDescendantChart(s.Tree, p, 3, depth, true, true)
	if err != nil {
		return fmt.Errorf("build lineage: %w", err)
	}

	ch.Title = "Descendants of " + p.PreferredUniqueName
	ch.Notes = []string{}
	ch.Notes = append(ch.Notes, time.Now().Format("Generated _2 January 2006"))
	if !s.Tree.KeyPerson.IsUnknown() {
		ch.Notes = append(ch.Notes, "(★ denotes a direct ancestor of "+s.Tree.KeyPerson.PreferredFamiliarFullName+")")
	}

	opts := gtree.DefaultLayoutOptions()
	lay, err := ch.Layout(opts)
	if err != nil {
		return fmt.Errorf("layout chart: %w", err)
	}

	svg, err := gtree.SVG(lay)
	if err != nil {
		return fmt.Errorf("render SVG: %w", err)
	}
	f, err := CreateFile(fname)
	if err != nil {
		return fmt.Errorf("create SVG file: %w", err)
	}
	defer f.Close()
	if err = os.WriteFile(fname, []byte(svg), 0o666); err != nil {
		return fmt.Errorf("write svg: %w", err)
	}
	return nil
}

func (s *Site) WriteGedcom(root string) error {
	fname := "all.ged"

	g, err := s.BuildGedcom()
	if err != nil {
		return fmt.Errorf("generate gedcom: %w", err)
	}
	f, err := CreateFile(filepath.Join(root, fname))
	if err != nil {
		return fmt.Errorf("create gedcom file: %w", err)
	}
	defer f.Close()

	enc := gedcom.NewEncoder(f)
	if err := enc.Encode(g); err != nil {
		return fmt.Errorf("encode gedcom: %w", err)
	}

	return nil
}

func groupRelation(rel *model.Relation) (string, int) {
	var group string
	var groupPriority int
	distance := rel.Distance()
	if distance < 5 {
		group = "Close relations"
		groupPriority = 1
	} else if rel.IsDirectAncestor() {
		group = "Direct ancestors"
		groupPriority = 2
	} else if rel.IsCloseToDirectAncestor() {
		group = "Family of ancestors"
		groupPriority = 3
	} else if distance < 12 {
		group = "Distant relations"
		groupPriority = 4
	} else {
		group = "Others"
		groupPriority = 5
	}

	return group, groupPriority
}

func writePage(p io.WriterTo, root string, fname string) error {
	f, err := CreateFile(filepath.Join(root, fname))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if _, err := p.WriteTo(f); err != nil {
		return fmt.Errorf("write file content: %w", err)
	}
	return f.Close()
}

func (s *Site) BuildPublishSet(m model.PersonMatcher) error {
	subset, err := NewPublishSet(s.Tree, m)
	if err != nil {
		return fmt.Errorf("build publish set: %w", err)
	}

	s.PublishSet = subset
	return nil
}

func (s *Site) BuildGedcom() (*gedcom.Gedcom, error) {
	includedPeople := make(map[string]*model.Person)
	for _, p := range s.PublishSet.People {
		includedPeople[p.ID] = p
	}

	if s.Tree.KeyPerson != nil {
		// Walk ancestors from key person and add to included list
		// until we hit people from the publish set
		queue := []*model.Person{
			s.Tree.KeyPerson,
		}
		for len(queue) > 0 {
			p := queue[0]
			queue = queue[1:]

			if _, ok := includedPeople[p.ID]; ok {
				// already included
				continue
			}
			includedPeople[p.ID] = p

			if !p.Father.IsUnknown() {
				queue = append(queue, p.Father)
			}
			if !p.Mother.IsUnknown() {
				queue = append(queue, p.Mother)
			}
		}
	}

	// Records to include in the gedcom
	irs := make(map[string]*gedcom.IndividualRecord)
	frs := make(map[string]*gedcom.FamilyRecord)

	includeIndividual := func(p *model.Person) *gedcom.IndividualRecord {
		ir, ok := irs[p.ID]
		if ok {
			return ir
		}
		ir = new(gedcom.IndividualRecord)
		ir.Xref = p.ID
		irs[p.ID] = ir

		n := new(gedcom.NameRecord)
		if p.Redacted && !p.RedactionKeepsName {
			n.Name = "private"
		} else {
			n.Name = strings.Replace(p.PreferredFullName, p.PreferredFamilyName, "/"+p.PreferredFamilyName+"/", 1)
		}
		ir.Name = append(ir.Name, n)

		if p.Gender.IsMale() {
			ir.Sex = "M"
		} else if p.Gender.IsFemale() {
			ir.Sex = "F"
		}

		if !p.Redacted {
			if p.BestBirthlikeEvent != nil {
				er := &gedcom.EventRecord{}

				ev := p.BestBirthlikeEvent

				switch ev.(type) {
				case *model.BirthEvent:
					er.Tag = "BIRT"
				case *model.BaptismEvent:
					er.Tag = "BAPM"
				default:
					panic(fmt.Sprintf("unhandled birthlike event type: %T", ev))
				}
				if !ev.GetDate().IsUnknown() {
					er.Date = ev.GetDate().Gedcom()
				}
				if !ev.GetPlace().IsUnknown() {
					er.Place = gedcom.PlaceRecord{
						Name: ev.GetPlace().PreferredFullName,
					}
				}
				ir.Event = append(ir.Event, er)
			}

			if p.BestDeathlikeEvent != nil {
				er := &gedcom.EventRecord{}

				ev := p.BestDeathlikeEvent

				switch ev.(type) {
				case *model.DeathEvent:
					er.Tag = "DEAT"
				case *model.BurialEvent:
					er.Tag = "BURI"
				case *model.CremationEvent:
					er.Tag = "CREM"
				default:
					panic(fmt.Sprintf("unhandled deathlike event type: %T", ev))
				}
				if !ev.GetDate().IsUnknown() {
					er.Date = ev.GetDate().Gedcom()
				}
				if !ev.GetPlace().IsUnknown() {
					er.Place = gedcom.PlaceRecord{
						Name: ev.GetPlace().PreferredFullName,
					}
				}
				ir.Event = append(ir.Event, er)
			}
		}
		return ir
	}

	includeFamily := func(f *model.Family) *gedcom.FamilyRecord {
		fr, ok := frs[f.ID]
		if ok {
			return fr
		}

		fr = new(gedcom.FamilyRecord)
		fr.Xref = f.ID
		frs[f.ID] = fr

		if !f.Father.IsUnknown() {
			fr.Husband = includeIndividual(f.Father)
			flr := new(gedcom.FamilyLinkRecord)
			flr.Family = fr
			fr.Husband.Family = append(fr.Husband.Family, flr)
		}
		if !f.Mother.IsUnknown() {
			fr.Wife = includeIndividual(f.Mother)
			flr := new(gedcom.FamilyLinkRecord)
			flr.Family = fr
			fr.Wife.Family = append(fr.Wife.Family, flr)
		}
		for _, c := range f.Children {
			if c.IsUnknown() {
				continue
			}
			cr := includeIndividual(c)
			fr.Child = append(fr.Child, cr)
			plr := new(gedcom.FamilyLinkRecord)
			plr.Family = fr
			cr.Parents = append(cr.Parents, plr)
		}

		return fr
	}

	// Seed the gedcom lists from the publish set
	for _, p := range includedPeople {
		includeIndividual(p)
		if p.ParentFamily != nil {
			includeFamily(p.ParentFamily)
		}
		// Only include families of people in publish set
		if !s.PublishSet.Includes(p) {
			continue
		}
		for _, f := range p.Families {
			includeFamily(f)
		}
	}

	sub := &gedcom.SubmitterRecord{
		Xref: "SUBM",
		Name: "Not known",
	}

	g := new(gedcom.Gedcom)

	g.Submitter = append(g.Submitter, sub)

	g.Header = &gedcom.Header{
		SourceSystem: gedcom.SystemRecord{
			Xref:       "genster",
			SourceName: "genster",
		},
		Submitter:    sub,
		Filename:     "all.ged",
		CharacterSet: "UTF-8",
		Language:     "English",
		Version:      "5.5.1",
		Form:         "LINEAGE-LINKED",
	}

	g.Trailer = &gedcom.Trailer{}

	kpid := ""
	if s.Tree.KeyPerson != nil {
		// put the key person first in the file for some systems
		// that use the first individual as the root
		if ir, ok := irs[s.Tree.KeyPerson.ID]; ok {
			g.Individual = append(g.Individual, ir)
			kpid = s.Tree.KeyPerson.ID
			g.Header.UserDefined = append(g.Header.UserDefined, gedcom.UserDefinedTag{
				Tag:  "_ROOT",
				Xref: kpid,
			})
		}
	}

	for id, ir := range irs {
		if id == kpid {
			continue
		}
		g.Individual = append(g.Individual, ir)
	}

	for _, fr := range frs {
		g.Family = append(g.Family, fr)
	}
	return g, nil
}

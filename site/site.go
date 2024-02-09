package site

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gosimple/slug"
	"github.com/iand/genster/chart"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
)

type PageLayout string

const (
	PageLayoutPerson         PageLayout = "person"
	PageLayoutPlace          PageLayout = "place"
	PageLayoutSource         PageLayout = "source"
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
	PageSectionPerson = "person"
	PageSectionPlace  = "place"
	PageSectionSource = "source"
	PageSectionList   = "list"
	PageSectionChart  = "chart"
)

const (
	PageCategoryPerson = "person"
	PageCategorySource = "source"
	PageCategoryPlace  = "place"
)

type Site struct {
	BaseURL   string
	Tree      *tree.Tree
	Calendars map[int]*Calendar

	PersonDir           string
	PersonPagePattern   string
	PersonFilePattern   string
	SourceDir           string
	SourcePagePattern   string
	SourceFilePattern   string
	FamilyPagePattern   string
	CalendarPagePattern string
	FamilyFilePattern   string
	PlaceDir            string
	PlacePagePattern    string
	PlaceFilePattern    string
	CalendarFilePattern string

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
	WikiTreePagePattern string
	WikiTreeFilePattern string

	SkippedPersonPages map[string]bool // map of person pages that should not be generated
}

func NewSite(baseURL string, t *tree.Tree) *Site {
	s := &Site{
		BaseURL:   baseURL,
		Tree:      t,
		Calendars: make(map[int]*Calendar),

		PersonDir:         PageSectionPerson,
		PersonPagePattern: path.Join(baseURL, PageSectionPerson, "/%s/"),
		PersonFilePattern: path.Join(PageSectionPerson, "/%s/index.md"),

		SourceDir:         PageSectionSource,
		SourcePagePattern: path.Join(baseURL, PageSectionSource, "/%s/"),
		SourceFilePattern: path.Join(PageSectionSource, "/%s/index.md"),

		FamilyPagePattern: path.Join(baseURL, "family/%s/"),
		FamilyFilePattern: path.Join("/family/%s/index.md"),

		PlaceDir:         PageSectionPlace,
		PlacePagePattern: path.Join(baseURL, PageSectionPlace, "/%s/"),
		PlaceFilePattern: path.Join(PageSectionPlace, "/%s/index.md"),

		// ListDir:         PageSectionList,
		// ListPagePattern: path.Join(basePath, PageSectionList, "/%s/"),
		// ListFilePattern: "/place/%s/index.md",

		WikiTreeDir:         "person",
		WikiTreePagePattern: path.Join(baseURL, "person/%s/wikitree"),
		WikiTreeFilePattern: path.Join(PageSectionPerson, "/%s/wikitree.md"),

		CalendarPagePattern: path.Join(baseURL, "calendar/%02d/"),
		CalendarFilePattern: "/calendar/%02d.md",

		ListInferencesDir: path.Join(PageSectionList, "inferences"),
		ListAnomaliesDir:  path.Join(PageSectionList, "anomalies"),
		ListTodoDir:       path.Join(PageSectionList, "todo"),
		ListPeopleDir:     path.Join(PageSectionList, "people"),
		ListPlacesDir:     path.Join(PageSectionList, "places"),
		ListSourcesDir:    path.Join(PageSectionList, "sources"),
		ListSurnamesDir:   path.Join(PageSectionList, "surnames"),

		ChartAncestorsDir: path.Join(PageSectionChart, "ancestors"),
		ChartTreesDir:     path.Join(PageSectionChart, "trees"),

		SkippedPersonPages: make(map[string]bool),
	}

	return s
}

func (s *Site) WritePages(root string) error {
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderPersonPage(s, p)
		if err != nil {
			return fmt.Errorf("render person page: %w", err)
		}

		if err := writePage(d, root, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
			return fmt.Errorf("write person page: %w", err)
		}

		if s.GenerateWikiTree {
			d, err = RenderWikiTreePage(s, p)
			if err != nil {
				return fmt.Errorf("render wikitree page: %w", err)
			}

			if err := writePage(d, root, fmt.Sprintf(s.WikiTreeFilePattern, p.ID)); err != nil {
				return fmt.Errorf("write wikitree page: %w", err)
			}
		}
	}

	for _, p := range s.Tree.Places {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderPlacePage(s, p)
		if err != nil {
			return fmt.Errorf("render place page: %w", err)
		}

		if err := writePage(d, root, fmt.Sprintf(s.PlaceFilePattern, p.ID)); err != nil {
			return fmt.Errorf("write place page: %w", err)
		}
	}

	for _, p := range s.Tree.Sources {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderSourcePage(s, p)
		if err != nil {
			return fmt.Errorf("render source page: %w", err)
		}
		if err := writePage(d, root, fmt.Sprintf(s.SourceFilePattern, p.ID)); err != nil {
			return fmt.Errorf("write source page: %w", err)
		}
	}
	s.BuildCalendar()

	for month, c := range s.Calendars {
		d, err := c.RenderPage(s)
		if err != nil {
			return fmt.Errorf("generate markdown: %w", err)
		}

		fname := fmt.Sprintf(s.CalendarFilePattern, month)

		f, err := CreateFile(filepath.Join(root, fname))
		if err != nil {
			return fmt.Errorf("create calendar file: %w", err)
		}
		if err := d.WriteMarkdown(f); err != nil {
			return fmt.Errorf("write calendar markdown: %w", err)
		}
		f.Close()
	}

	if err := s.WritePersonListPages(root); err != nil {
		return fmt.Errorf("write people list pages: %w", err)
	}

	if err := s.WritePlaceListPages(root); err != nil {
		return fmt.Errorf("write place list pages: %w", err)
	}

	if err := s.WriteSourceListPages(root); err != nil {
		return fmt.Errorf("write source list pages: %w", err)
	}

	if err := s.WriteSurnameListPages(root); err != nil {
		return fmt.Errorf("write surname list pages: %w", err)
	}

	if err := s.WriteInferenceListPages(root); err != nil {
		return fmt.Errorf("write inferences pages: %w", err)
	}

	if err := s.WriteAnomalyListPages(root); err != nil {
		return fmt.Errorf("write anomalies pages: %w", err)
	}

	if err := s.WriteTodoListPages(root); err != nil {
		return fmt.Errorf("write todo pages: %w", err)
	}

	if err := s.WriteTreeOverview(root); err != nil {
		return fmt.Errorf("write tree overview: %w", err)
	}

	if err := s.WriteChartAncestors(root); err != nil {
		return fmt.Errorf("write ancestor chart: %w", err)
	}

	// if err := s.WriteChartTrees(root); err != nil {
	// 	return fmt.Errorf("write chart trees: %w", err)
	// }

	return nil
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

func writePage(doc *md.Document, root string, fname string) error {
	f, err := CreateFile(filepath.Join(root, fname))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if err := doc.WriteMarkdown(f); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	return f.Close()
}

func (s *Site) LinkFor(v any) string {
	switch vt := v.(type) {
	case *model.Person:
		if s.SkippedPersonPages[vt.ID] {
			return ""
		}
		if vt.Redacted {
			return ""
		}
		return fmt.Sprintf(s.PersonPagePattern, vt.ID)
	case *model.Source:
		return fmt.Sprintf(s.SourcePagePattern, vt.ID)
	case *model.Family:
		return fmt.Sprintf(s.FamilyPagePattern, vt.ID)
	case *model.Place:
		return fmt.Sprintf(s.PlacePagePattern, vt.ID)
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
			return fmt.Sprintf(s.WikiTreePagePattern, vt.ID)
		}
	}

	return ""
}

func (s *Site) NewDocument() *md.Document {
	doc := &md.Document{}
	doc.BasePath(s.BaseURL)
	doc.SetLinkBuilder(s)

	return doc
}

func (s *Site) NewMarkdownBuilder() MarkdownBuilder {
	enc := &md.Encoder{}
	enc.SetLinkBuilder(s)

	return enc
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

func (s *Site) WriteAnomalyListPages(root string) error {
	baseDir := filepath.Join(root, s.ListAnomaliesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to anomalies index", "id", p.ID)
			continue
		}
		categories := make([]model.AnomalyCategory, 0)
		anomaliesByCategory := make(map[model.AnomalyCategory][]*model.Anomaly)

		for _, a := range p.Anomalies {
			a := a // avoid shadowing
			al, ok := anomaliesByCategory[a.Category]
			if ok {
				al = append(al, a)
				anomaliesByCategory[a.Category] = al
				continue
			}

			categories = append(categories, a.Category)
			anomaliesByCategory[a.Category] = []*model.Anomaly{a}
		}
		sort.Slice(categories, func(i, j int) bool { return categories[i] < categories[j] })

		if len(anomaliesByCategory) > 0 {
			b := s.NewMarkdownBuilder()
			b.Heading2(p.PreferredUniqueName)
			rel := "unknown relation"
			if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
				rel = p.RelationToKeyPerson.Name()
			}

			links := b.EncodeModelLink("View page", p)

			if p.EditLink != nil {
				links += " or " + b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL)
			}
			b.Para(text.FormatSentence(rel) + " " + links)
			for _, cat := range categories {
				al := anomaliesByCategory[cat]
				items := make([][2]string, 0, len(al))

				for _, a := range al {
					items = append(items, [2]string{
						a.Context,
						a.Text,
					})
				}
				b.DefinitionList(items)
			}

			group, groupPriority := groupRelation(p.RelationToKeyPerson)
			pn.AddEntryWithGroup(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.Markdown(), group, groupPriority)
		}

	}

	if err := pn.WritePages(s, baseDir, PageLayoutListAnomalies, "Anomalies", "Anomalies are errors or inconsistencies that have been detected in the underlying data used to generate this site."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteInferenceListPages(root string) error {
	baseDir := filepath.Join(root, s.ListInferencesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to inference index", "id", p.ID)
			continue
		}
		items := make([][2]string, 0)
		for _, inf := range p.Inferences {
			items = append(items, [2]string{
				inf.Type + " " + inf.Value,
				"because " + inf.Reason,
			})
		}

		if len(items) > 0 {
			b := s.NewMarkdownBuilder()
			b.Heading2(p.PreferredUniqueName)
			if p.EditLink != nil {
				b.Para(b.EncodeModelLink("View page", p) + " or " + b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL))
			} else {
				b.Para(b.EncodeModelLink("View page", p))
			}
			b.DefinitionList(items)
			pn.AddEntry(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.Markdown())
		}

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListInferences, "Inferences Made", "Inferences refer to hints and suggestions that help fill in missing information in the family tree."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteTodoListPages(root string) error {
	baseDir := filepath.Join(root, s.ListTodoDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to todo index", "id", p.ID)
			continue
		}
		categories := make([]model.ToDoCategory, 0)
		todosByCategory := make(map[model.ToDoCategory][]*model.ToDo)

		for _, a := range p.ToDos {
			a := a // avoid shadowing
			al, ok := todosByCategory[a.Category]
			if ok {
				al = append(al, a)
				todosByCategory[a.Category] = al
				continue
			}

			categories = append(categories, a.Category)
			todosByCategory[a.Category] = []*model.ToDo{a}
		}
		sort.Slice(categories, func(i, j int) bool {
			return categories[i] < categories[j]
		})

		if len(todosByCategory) > 0 {
			b := s.NewMarkdownBuilder()
			b.Heading2(p.PreferredUniqueName)
			rel := "unknown relation"
			if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
				rel = p.RelationToKeyPerson.Name()
			}

			links := b.EncodeModelLink("View page", p)

			if p.EditLink != nil {
				links += " or " + b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL)
			}
			b.Para(text.FormatSentence(rel) + " " + links)

			for _, cat := range categories {
				al := todosByCategory[cat]
				items := make([][2]string, 0, len(al))

				for _, a := range al {
					line := text.StripTerminator(text.UpperFirst(a.Goal))
					if a.Reason != "" {
						line += " (" + text.LowerFirst(a.Reason) + ")"
					} else {
						line = text.FinishSentence(line)
					}
					items = append(items, [2]string{
						a.Context,
						line,
					})
				}
				b.DefinitionList(items)

				// for _, a := range al {
				// 	line := b.EncodeItalic(a.Context) + ": " + text.StripTerminator(text.LowerFirst(a.Goal))
				// 	if a.Reason != "" {
				// 		line += " (" + text.LowerFirst(a.Reason) + ")"
				// 	} else {
				// 		line = text.FinishSentence(line)
				// 	}
				// 	items = append(items, line)
				// }
				// b.Heading4(cat.String())
				// b.UnorderedList(items)
			}

			group, groupPriority := groupRelation(p.RelationToKeyPerson)
			pn.AddEntryWithGroup(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.Markdown(), group, groupPriority)
		}

	}

	if err := pn.WritePages(s, baseDir, PageLayoutListTodo, "To Do", "These suggested tasks and projects are loose ends or incomplete areas in the tree that need further research."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WritePersonListPages(root string) error {
	baseDir := filepath.Join(root, s.ListPeopleDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to person index", "id", p.ID)
			continue
		}
		items := make([][2]string, 0)
		b := s.NewMarkdownBuilder()

		var rel string
		if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
			rel = b.EncodeBold(text.FormatSentence(p.RelationToKeyPerson.Name()))
		}

		items = append(items, [2]string{
			b.EncodeModelLink(p.PreferredUniqueName, p),
			text.JoinSentences(p.Olb, rel),
		})
		b.DefinitionList(items)
		pn.AddEntry(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListPeople, "People", "This is a full, alphabetical list of people in the tree."); err != nil {
		return err
	}
	return nil
}

func (s *Site) WritePlaceListPages(root string) error {
	baseDir := filepath.Join(root, s.ListPlacesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.Places {
		items := make([][2]string, 0)
		b := s.NewMarkdownBuilder()
		items = append(items, [2]string{
			b.EncodeModelLink(p.PreferredUniqueName, p),
		})
		b.DefinitionList(items)
		pn.AddEntry(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListPlaces, "Places", "This is a full, alphabetical list of places in the tree."); err != nil {
		return err
	}
	return nil
}

func (s *Site) WriteSourceListPages(root string) error {
	baseDir := filepath.Join(root, s.ListSourcesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, so := range s.Tree.Sources {
		items := make([][2]string, 0)
		b := s.NewMarkdownBuilder()
		items = append(items, [2]string{
			b.EncodeModelLink(so.Title, so),
		})
		b.DefinitionList(items)
		pn.AddEntry(so.Title+"~"+so.ID, so.Title, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListSources, "Sources", "This is a full, alphabetical list of sources cited in the tree."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteSurnameListPages(root string) error {
	peopleBySurname := make(map[string][]*model.Person)
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to surname index", "id", p.ID)
			continue
		}
		if p.PreferredFamilyName == model.UnknownNamePlaceholder {
			continue
		}
		peopleBySurname[p.PreferredFamilyName] = append(peopleBySurname[p.PreferredFamilyName], p)
	}

	surnames := make([]string, 0, len(peopleBySurname))

	for surname, people := range peopleBySurname {
		surnames = append(surnames, surname)
		model.SortPeople(people)

		pn := NewPaginator()
		pn.HugoStyle = s.GenerateHugo

		for _, p := range people {
			items := make([][2]string, 0)
			b := s.NewMarkdownBuilder()

			var rel string
			if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
				rel = b.EncodeBold(text.FormatSentence(p.RelationToKeyPerson.Name()))
			}

			items = append(items, [2]string{
				b.EncodeModelLink(p.PreferredUniqueName, p),
				text.JoinSentences(p.Olb, rel),
			})
			b.DefinitionList(items)
			pn.AddEntry(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.Markdown())
		}

		baseDir := filepath.Join(root, s.ListSurnamesDir, slug.Make(surname))
		if err := pn.WritePages(s, baseDir, PageLayoutListSurnames, surname, "This is a full, alphabetical list of people in the tree with the surname '"+surname+"'."); err != nil {
			return err
		}

	}

	sort.Slice(surnames, func(i, j int) bool { return surnames[i] < surnames[j] })
	indexPage := "index.md"
	if s.GenerateHugo {
		indexPage = "_index.md"
	}

	ancestorSurnames := s.Tree.AncestorSurnameDistribution()

	doc := s.NewDocument()
	doc.Title("Surnames")
	doc.Summary("This is a full, alphabetical list of the surnames of people in the tree.")
	doc.Layout(PageLayoutListSurnames.String())

	alist := make([]string, 0, len(ancestorSurnames))
	olist := make([]string, 0, len(surnames))
	for _, surname := range surnames {
		if _, ok := ancestorSurnames[surname]; ok {
			alist = append(alist, doc.EncodeLink(surname, "./"+slug.Make(surname)))
		} else {
			olist = append(olist, doc.EncodeLink(surname, "./"+slug.Make(surname)))
		}
	}
	doc.Heading3("Direct ancestors")
	doc.Para(strings.Join(alist, ", "))
	doc.Heading3("Other surnames")
	doc.Para(strings.Join(olist, ", "))

	baseDir := filepath.Join(root, s.ListSurnamesDir)
	if err := writePage(doc, baseDir, indexPage); err != nil {
		return fmt.Errorf("failed to write surname index: %w", err)
	}

	return nil
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
		doc.Para(text.FormatSentence(desc))
	}

	peopleDesc := ""

	numberOfPeople := s.Tree.NumberOfPeople()
	if numberOfPeople > 0 {
		doc.EmptyPara()
		peopleDesc = text.FormatSentence(fmt.Sprintf("There are %d people in this tree", numberOfPeople))
	}

	ancestorSurnames := FlattenMapByValueDesc(s.Tree.AncestorSurnameDistribution())
	if len(ancestorSurnames) > 0 {
		list := make([]string, 12)
		for i := range ancestorSurnames {
			if i > 11 {
				break
			}
			list[i] = doc.EncodeLink(ancestorSurnames[i].K, path.Join(s.BaseURL, s.ListSurnamesDir, slug.Make(ancestorSurnames[i].K)))
		}
		detail := text.JoinSentenceParts("The principle surnames are ", text.JoinList(list))
		peopleDesc = text.JoinSentences(peopleDesc, text.FormatSentence(detail))
		peopleDesc = text.JoinSentences(peopleDesc, text.FormatSentence(text.JoinSentenceParts("See", doc.EncodeLink("all surnames...", s.ListSurnamesDir))))
	}

	if peopleDesc != "" {
		doc.EmptyPara()
		doc.Para(peopleDesc)
	}

	doc.EmptyPara()
	doc.Para(text.JoinSentenceParts("See a", doc.EncodeLink("full list of ancestors", s.ChartAncestorsDir), "for", doc.EncodeModelLink(s.Tree.KeyPerson.PreferredFamiliarFullName, s.Tree.KeyPerson)))

	// Featured people
	featuredPeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		return p.Featured
	}, 8)
	if len(featuredPeople) > 0 {
		model.SortPeople(featuredPeople)
		doc.EmptyPara()
		doc.Heading2("Featured")
		items := make([]string, len(featuredPeople))
		for i, p := range featuredPeople {
			items[i] = text.AppendRelated(doc.EncodeModelLink(p.PreferredUniqueName, p), p.Olb)
		}
		doc.UnorderedList(items)
	}

	// Currently puzzling over
	puzzlePeople := s.Tree.ListPeopleMatching(func(p *model.Person) bool {
		return p.Puzzle && !p.Featured
	}, 8)
	if len(puzzlePeople) > 0 {
		model.SortPeople(puzzlePeople)
		doc.EmptyPara()
		doc.Heading2("Currently puzzling over")
		items := make([]string, len(puzzlePeople))
		for i, p := range puzzlePeople {
			items[i] = text.AppendRelated(doc.EncodeModelLink(p.PreferredUniqueName, p), p.Olb)
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
		model.SortPeople(rnPeople)
		doc.EmptyPara()
		detail := text.JoinSentenceParts("Other people with research notes:", EncodePeopleListInline(rnPeople, func(p *model.Person) string {
			return p.PreferredFamiliarFullName
		}, doc))
		doc.Para(text.FormatSentence(detail))
	}

	// Oldest people
	oldestPeople := s.Tree.OldestPeople(3)
	if len(oldestPeople) > 0 {
		doc.EmptyPara()
		detail := text.JoinSentenceParts("Oldest people:", EncodePeopleListInline(oldestPeople, func(p *model.Person) string {
			age, _ := p.AgeInYearsAtDeath()
			return fmt.Sprintf("%s (%d years)", p.PreferredFamiliarFullName, age)
		}, doc))
		doc.Para(text.FormatSentence(detail))

	}

	var notes string
	if !s.Tree.KeyPerson.IsUnknown() {
		doc.EmptyPara()

		detail := text.JoinSentenceParts("In this family tree,", doc.EncodeModelLink(s.Tree.KeyPerson.PreferredFamiliarFullName, s.Tree.KeyPerson), "acts as the primary reference point, with all relationships defined in relation to", s.Tree.KeyPerson.Gender.ObjectPronoun())
		notes = text.JoinSentences(notes, text.FormatSentence(detail))
		notes = text.JoinSentences(notes, text.FormatSentence(text.JoinSentenceParts("Names suffixed by the", md.DirectAncestorMarker, "symbol indicate direct ancestors")))
	}

	if !s.IncludePrivate {
		detail := text.JoinSentenceParts("The tree excludes information on people who are possibly alive or who have died within the past twenty years")
		notes = text.JoinSentences(notes, text.FormatSentence(detail))
	}

	if len(notes) > 0 {
		doc.EmptyPara()
		doc.Heading3("Notes")
		doc.Para(text.FormatSentence(notes))
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
	ancestors := s.Tree.Ancestors(s.Tree.KeyPerson, generations)

	doc := s.NewDocument()
	doc.Title("Ancestors of " + s.Tree.KeyPerson.PreferredFamiliarFullName)
	doc.Summary(text.JoinSentenceParts("This is a full list of ancestors of", doc.EncodeModelLink(s.Tree.KeyPerson.PreferredFamiliarFullName, s.Tree.KeyPerson)))
	doc.Layout(PageLayoutChartAncestors.String())

	g := 0
	doc.Heading3("Generation 1")

	doc.Para(text.JoinSentenceParts("1.", doc.EncodeLink(s.Tree.KeyPerson.PreferredFamiliarFullName, doc.LinkBuilder.LinkFor(s.Tree.KeyPerson))))
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
				doc.Heading3("Generation 2: Parents")
			} else if g == 2 {
				doc.Heading3("Generation 3: Grandparents")
			} else if g == 3 {
				doc.Heading3("Generation 4: Great-Grandparents")
			} else if g == 4 {
				doc.Heading3("Generation 5: Great-Great-Grandparents")
			} else {
				doc.Heading3(fmt.Sprintf("Generation %d: %dx Great-Grandparents", g+1, g-2))
			}
		}
		if ancestors[i] != nil {
			detail := text.JoinSentenceParts(fmt.Sprintf("%d.", i+2), doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredFullName, doc.LinkBuilder.LinkFor(ancestors[i]))))

			var adds []string
			if ancestors[i].PrimaryOccupation != "" {
				adds = append(adds, ancestors[i].PrimaryOccupation)
			}
			if ancestors[i].BestBirthlikeEvent != nil && !ancestors[i].BestBirthlikeEvent.GetDate().IsUnknown() {
				adds = append(adds, WhatWhenWhere(ancestors[i].BestBirthlikeEvent, doc))
			}
			if ancestors[i].BestDeathlikeEvent != nil && !ancestors[i].BestDeathlikeEvent.GetDate().IsUnknown() {
				adds = append(adds, WhatWhenWhere(ancestors[i].BestDeathlikeEvent, doc))
			}

			detail = text.AppendClause(detail, text.JoinList(adds))
			doc.Para(detail)
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

			doc.Para(text.JoinSentenceParts(fmt.Sprintf("%d.", i+2), name))
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
	ancestors := s.Tree.Ancestors(s.Tree.KeyPerson, generations)

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
		doc.Para(doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredUniqueName, ancestors[i].ID+".svg")))
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
		doc.Para(doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredUniqueName, ancestors[i].ID+".svg")))
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
		doc.Para(doc.EncodeBold(doc.EncodeLink(ancestors[i].PreferredUniqueName, ancestors[i].ID+".svg")))
	}

	baseDir := filepath.Join(root, s.ChartTreesDir)
	if err := writePage(doc, baseDir, fname); err != nil {
		return fmt.Errorf("failed to write chart trees index: %w", err)
	}

	return nil
}

func (s *Site) WriteDescendantTree(fname string, p *model.Person, depth int) error {
	ch, err := chart.BuildDescendantChart(s.Tree, p, 3, depth, true)
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
	lay := ch.Layout(opts)

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

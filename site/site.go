package site

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/gosimple/slug"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
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
)

func (p PageLayout) String() string { return string(p) }

const (
	PageSectionPerson = "person"
	PageSectionPlace  = "place"
	PageSectionSource = "source"
	PageSectionList   = "list"
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

	IncludePrivate     bool
	TimelineExperiment bool

	GenerateHugo bool

	GenerateWikiTree    bool
	WikiTreeDir         string
	WikiTreePagePattern string
	WikiTreeFilePattern string
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
	} else {
		p.Tags = append(p.Tags, "Unknown Surname")
	}

	if p.PreferredGivenName == model.UnknownNamePlaceholder || p.PreferredGivenName == "" {
		p.Tags = append(p.Tags, "Unknown Forename")
	}

	if len(p.Inferences) > 0 {
		p.Tags = append(p.Tags, "Has Inferences")
	}

	if len(p.Anomalies) > 0 {
		p.Tags = append(p.Tags, "Has Anomalies")
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

	if p.BestBirthlikeEvent == nil || p.BestBirthlikeEvent.GetDate().IsUnknown() {
		p.Tags = append(p.Tags, "Unknown Birthdate")
	}
	if p.BestDeathlikeEvent == nil || p.BestDeathlikeEvent.GetDate().IsUnknown() {
		p.Tags = append(p.Tags, "Unknown Deathdate")
	}

	if p.WikiTreeID != "" {
		p.Tags = append(p.Tags, "WikiTree")
	}

	if len(p.ResearchNotes) > 0 {
		p.Tags = append(p.Tags, "Has Research Notes")
	}

	return nil
}

func personid(p *model.Person) string {
	if p == nil {
		return "nil"
	}
	return p.ID
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
	for _, ev := range p.Timeline {
		if !ev.DirectlyInvolves(p) {
			continue
		}
		anoms := ScanTimelineEventForAnomalies(ev)
		if len(anoms) > 0 {
			for _, anom := range anoms {
				p.Anomalies = append(p.Anomalies, anom)
			}
		}
	}
}

func (s *Site) WriteAnomalyListPages(root string) error {
	baseDir := filepath.Join(root, s.ListAnomaliesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if p.Redacted {
			logging.Debug("not writing redacted person to anomalies index", "id", p.ID)
			continue
		}
		categories := make([]string, 0)
		anomaliesByCategory := make(map[string][]*model.Anomaly)

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
		sort.Strings(categories)

		if len(anomaliesByCategory) > 0 {
			b := s.NewMarkdownBuilder()
			b.Heading2(p.PreferredUniqueName)
			if p.EditLink != nil {
				b.Para(b.EncodeModelLink("View page", p) + " or " + b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL))
			} else {
				b.Para(b.EncodeModelLink("View page", p))
			}
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
			pn.AddEntryWithGroup(p.PreferredSortName, b.Markdown(), group, groupPriority)
		}

	}

	if err := pn.WritePages(s, baseDir, PageLayoutListAnomalies, "Anomalies"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteInferenceListPages(root string) error {
	baseDir := filepath.Join(root, s.ListInferencesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
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
			pn.AddEntry(p.PreferredSortName, b.Markdown())
		}

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListInferences, "Inferences Made"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteTodoListPages(root string) error {
	baseDir := filepath.Join(root, s.ListTodoDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
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
			if p.EditLink != nil {
				b.Para(b.EncodeModelLink("View page", p) + " or " + b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL))
			} else {
				b.Para(b.EncodeModelLink("View page", p))
			}
			for _, cat := range categories {
				al := todosByCategory[cat]
				items := make([]string, 0, len(al))

				for _, a := range al {
					line := b.EncodeItalic(a.Context) + ": " + text.StripTerminator(text.LowerFirst(a.Goal))
					if a.Reason != "" {
						line += " (" + text.LowerFirst(a.Reason) + ")"
					} else {
						line = text.FinishSentence(line)
					}
					items = append(items, line)
				}
				b.Heading4(cat.String())
				b.UnorderedList(items)
			}

			group, groupPriority := groupRelation(p.RelationToKeyPerson)
			pn.AddEntryWithGroup(p.PreferredSortName, b.Markdown(), group, groupPriority)
		}

	}

	if err := pn.WritePages(s, baseDir, PageLayoutListTodo, "To Do"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WritePersonListPages(root string) error {
	baseDir := filepath.Join(root, s.ListPeopleDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if p.Redacted {
			logging.Debug("not writing redacted person to person index", "id", p.ID)
			continue
		}
		items := make([][2]string, 0)
		b := s.NewMarkdownBuilder()
		items = append(items, [2]string{
			b.EncodeModelLink(p.PreferredUniqueName, p),
			p.Olb,
		})
		b.DefinitionList(items)
		pn.AddEntry(p.PreferredSortName, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListPeople, "People"); err != nil {
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
		pn.AddEntry(p.PreferredSortName, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListPlaces, "Places"); err != nil {
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
		pn.AddEntry(so.Title, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListSources, "Sources"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteSurnameListPages(root string) error {
	peopleBySurname := make(map[string][]*model.Person)
	for _, p := range s.Tree.People {
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
		sort.Slice(people, func(i, j int) bool {
			return people[i].PreferredSortName < people[j].PreferredSortName
		})

		pn := NewPaginator()
		pn.HugoStyle = s.GenerateHugo

		for _, p := range people {
			items := make([][2]string, 0)
			b := s.NewMarkdownBuilder()
			items = append(items, [2]string{
				b.EncodeModelLink(p.PreferredUniqueName, p),
				p.Olb,
			})
			b.DefinitionList(items)
			pn.AddEntry(p.PreferredSortName, b.Markdown())
		}

		baseDir := filepath.Join(root, s.ListSurnamesDir, slug.Make(surname))
		if err := pn.WritePages(s, baseDir, PageLayoutListSurnames, surname); err != nil {
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
	doc.Title("Tree Overview")

	fname := "index.md"
	if s.GenerateHugo {
		fname = "_index.md"
	}

	if !s.Tree.KeyPerson.IsUnknown() {
		doc.EmptyPara()

		detail := text.JoinSentence("In this family tree,", doc.EncodeModelLink(s.Tree.KeyPerson.PreferredFamiliarFullName, s.Tree.KeyPerson), "acts as the primary reference point, with all relationships defined in relation to", s.Tree.KeyPerson.Gender.ObjectPronoun())
		detail = text.FormatSentence(detail)
		doc.Para(text.FormatSentence(detail))
	}

	if !s.IncludePrivate {
		detail := text.JoinSentence("The tree excludes information on people who are possibly alive or who have died within the past twenty years")
		detail = text.FormatSentence(detail)
		doc.Para(text.FormatSentence(detail))
	}

	numberOfPeople := s.Tree.NumberOfPeople()
	if numberOfPeople > 0 {
		detail := fmt.Sprintf("There are %d people in this tree", numberOfPeople)
		doc.Para(text.FormatSentence(detail))
	}

	ancestorSurnames := flattenByKeyAsc(s.Tree.AncestorSurnameDistribution())
	if len(ancestorSurnames) > 0 {
		doc.EmptyPara()
		detail := text.JoinSentence("Ancestral surnames:", joinStringIntTuples(ancestorSurnames))
		doc.Para(text.FormatSentence(detail))
	}

	treeSurnames := flattenByKeyAsc(s.Tree.TreeSurnameDistribution())
	if len(treeSurnames) > 0 {
		doc.EmptyPara()
		detail := text.JoinSentence("Tree surnames:", joinStringIntTuples(treeSurnames))
		doc.Para(text.FormatSentence(detail))
	}

	oldestPeople := s.Tree.OldestPeople(3)
	if len(oldestPeople) > 0 {
		doc.EmptyPara()
		detail := text.JoinSentence("Oldest people:", EncodePeopleListInline(oldestPeople, func(p *model.Person) string {
			age, _ := p.AgeInYearsAtDeath()
			return fmt.Sprintf("%s (%d years)", p.PreferredUniqueName, age)
		}, doc))
		doc.Para(text.FormatSentence(detail))

	}

	rnPeople := s.Tree.ListPeopleMatching(model.PersonHasResearchNotes(), 3)
	if len(rnPeople) > 0 {
		doc.EmptyPara()
		doc.Heading2("People with research notes")
		items := make([]string, len(rnPeople))
		for i, p := range rnPeople {
			items[i] = doc.EncodeModelLink(p.PreferredUniqueName, p)
		}
		doc.UnorderedList(items)
	}

	if err := writePage(doc, root, fname); err != nil {
		return fmt.Errorf("write page: %w", err)
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
	} else if distance < 12 {
		group = "Distant relations"
		groupPriority = 3
	} else {
		group = "Others"
		groupPriority = 4
	}

	return group, groupPriority
}

type StringIntTuple struct {
	S string
	I int
}

func (t *StringIntTuple) String() string {
	return fmt.Sprintf("%s (%d)", t.S, t.I)
}

// flattenByKeyAsc flattens the map and sorts by the key ascending
func flattenByKeyAsc(m map[string]int) []StringIntTuple {
	list := make([]StringIntTuple, 0, len(m))
	for k, v := range m {
		list = append(list, StringIntTuple{S: k, I: v})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].S < list[j].S })

	return list
}

func joinStringIntTuples(l []StringIntTuple) string {
	ss := make([]string, len(l))
	for i := range l {
		ss[i] = l[i].String()
	}

	return text.JoinList(ss)
}

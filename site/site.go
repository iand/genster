package site

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
)

type Site struct {
	BasePath  string
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
	InferencesDir       string
	AnomaliesDir        string
	TodoDir             string
	IncludePrivate      bool
	TimelineExperiment  bool

	GenerateWikiTree    bool
	WikiTreeDir         string
	WikiTreePagePattern string
	WikiTreeFilePattern string
}

func NewSite(basePath string, t *tree.Tree) *Site {
	s := &Site{
		BasePath:  basePath,
		Tree:      t,
		Calendars: make(map[int]*Calendar),

		PersonDir:         "person",
		PersonPagePattern: path.Join(basePath, "person/%s/"),
		PersonFilePattern: "/person/%s/index.md",

		SourceDir:         "source",
		SourcePagePattern: path.Join(basePath, "source/%s/"),
		SourceFilePattern: "/source/%s/index.md",

		FamilyPagePattern: path.Join(basePath, "family/%s/"),
		FamilyFilePattern: "/family/%s/index.md",

		PlaceDir:         "place",
		PlacePagePattern: path.Join(basePath, "place/%s/"),
		PlaceFilePattern: "/place/%s/index.md",

		WikiTreeDir:         "person",
		WikiTreePagePattern: path.Join(basePath, "person/%s/wikitree"),
		WikiTreeFilePattern: "/person/%s/wikitree.md",

		CalendarPagePattern: path.Join(basePath, "calendar/%02d/"),
		CalendarFilePattern: "/calendar/%02d.md",
		InferencesDir:       "inferences",
		AnomaliesDir:        "anomalies",
		TodoDir:             "todo",
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

	if err := s.WritePersonIndexPages(root); err != nil {
		return fmt.Errorf("write people index pages: %w", err)
	}

	if err := s.WritePlaceIndexPages(root); err != nil {
		return fmt.Errorf("write place index pages: %w", err)
	}

	if err := s.WriteSourceIndexPages(root); err != nil {
		return fmt.Errorf("write source index pages: %w", err)
	}

	if err := s.WriteInferencesPages(root); err != nil {
		return fmt.Errorf("write inferences pages: %w", err)
	}

	if err := s.WriteAnomaliesPages(root); err != nil {
		return fmt.Errorf("write anomalies pages: %w", err)
	}

	if err := s.WriteTodoPages(root); err != nil {
		return fmt.Errorf("write todo pages: %w", err)
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

	if p.CauseOfDeath == model.CauseOfDeathSuicide {
		p.Tags = append(p.Tags, "Died by suicide")
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
	doc.BasePath(s.BasePath)
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

func (s *Site) WriteAnomaliesPages(root string) error {
	baseDir := filepath.Join(root, s.AnomaliesDir)
	pn := NewPaginator()
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
			pn.AddEntry(p.PreferredSortName, b.Markdown())
		}

	}

	if err := pn.WritePages(s, baseDir, "anomalies", "Anomalies"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteInferencesPages(root string) error {
	baseDir := filepath.Join(root, s.InferencesDir)
	pn := NewPaginator()
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
	if err := pn.WritePages(s, baseDir, "inferences", "Inferences Made"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteTodoPages(root string) error {
	baseDir := filepath.Join(root, s.TodoDir)
	pn := NewPaginator()
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
			pn.AddEntry(p.PreferredSortName, b.Markdown())
		}

	}

	if err := pn.WritePages(s, baseDir, "todo", "To Do"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WritePersonIndexPages(root string) error {
	baseDir := filepath.Join(root, s.PersonDir)
	pn := NewPaginator()
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
	if err := pn.WritePages(s, baseDir, s.PersonDir, "People"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WritePlaceIndexPages(root string) error {
	baseDir := filepath.Join(root, s.PlaceDir)
	pn := NewPaginator()
	for _, p := range s.Tree.Places {
		items := make([][2]string, 0)
		b := s.NewMarkdownBuilder()
		items = append(items, [2]string{
			b.EncodeModelLink(p.PreferredUniqueName, p),
		})
		b.DefinitionList(items)
		pn.AddEntry(p.PreferredSortName, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, s.PlaceDir, "Places"); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteSourceIndexPages(root string) error {
	baseDir := filepath.Join(root, s.SourceDir)
	pn := NewPaginator()
	for _, so := range s.Tree.Sources {
		items := make([][2]string, 0)
		b := s.NewMarkdownBuilder()
		items = append(items, [2]string{
			b.EncodeModelLink(so.Title, so),
		})
		b.DefinitionList(items)
		pn.AddEntry(so.Title, b.Markdown())

	}
	if err := pn.WritePages(s, baseDir, s.SourceDir, "Sources"); err != nil {
		return err
	}

	return nil
}

package site

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/iand/gdate"
	"github.com/iand/genster/md"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/iand/werr"
)

type Site struct {
	BasePath string
	Tree     *tree.Tree
	// People              map[string]*model.Person
	// Sources             map[string]*model.Source
	// Families            map[string]*model.Family
	// Places              map[string]*model.Place
	Calendars           map[int]*Calendar
	PersonPagePattern   string
	SourcePagePattern   string
	FamilyPagePattern   string
	PlacePagePattern    string
	CalendarPagePattern string
	PersonFilePattern   string
	SourceFilePattern   string
	FamilyFilePattern   string
	PlaceFilePattern    string
	CalendarFilePattern string
	InferencesFile      string
	AnomaliesFile       string
	IncludePrivate      bool
}

func NewSite(basePath string, t *tree.Tree) *Site {
	s := &Site{
		BasePath:            basePath,
		Tree:                t,
		Calendars:           make(map[int]*Calendar),
		PersonPagePattern:   path.Join(basePath, "person/%s/"),
		PersonFilePattern:   "/person/%s.md",
		SourcePagePattern:   path.Join(basePath, "source/%s/"),
		SourceFilePattern:   "/source/%s.md",
		FamilyPagePattern:   path.Join(basePath, "family/%s/"),
		FamilyFilePattern:   "/family/%s.md",
		PlacePagePattern:    path.Join(basePath, "place/%s/"),
		PlaceFilePattern:    "/place/%s.md",
		CalendarPagePattern: path.Join(basePath, "calendar/%02d/"),
		CalendarFilePattern: "/calendar/%02d.md",
		InferencesFile:      "/inferences.md",
		AnomaliesFile:       "/anomalies.md",
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
			return werr.Wrap(err) // fmt.Errorf("generate markdown: %w", err)
		}

		if err := writePage(d, root, fmt.Sprintf(s.PersonFilePattern, p.ID)); err != nil {
			return werr.Wrap(err)
		}
	}

	for _, p := range s.Tree.Places {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderPlacePage(s, p)
		if err != nil {
			return werr.Wrap(err) // fmt.Errorf("generate markdown: %w", err)
		}

		if err := writePage(d, root, fmt.Sprintf(s.PlaceFilePattern, p.ID)); err != nil {
			return werr.Wrap(err)
		}
	}

	for _, p := range s.Tree.Sources {
		if s.LinkFor(p) == "" {
			continue
		}
		d, err := RenderSourcePage(s, p)
		if err != nil {
			return werr.Wrap(err) // fmt.Errorf("generate markdown: %w", err)
		}
		if err := writePage(d, root, fmt.Sprintf(s.SourceFilePattern, p.ID)); err != nil {
			return werr.Wrap(err)
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
			return fmt.Errorf("create file: %w", err)
		}
		if err := d.WriteMarkdown(f); err != nil {
			return fmt.Errorf("write markdown: %w", err)
		}
		f.Close()
	}

	infd, err := RenderInferencesPage(s)
	if err != nil {
		return fmt.Errorf("render inferences page markdown: %w", err)
	}
	if err := writePage(infd, root, s.InferencesFile); err != nil {
		return fmt.Errorf("inferences: %w", err)
	}

	anomd, err := RenderAnomaliesPage(s)
	if err != nil {
		return fmt.Errorf("render inferences page markdown: %w", err)
	}
	if err := writePage(anomd, root, s.AnomaliesFile); err != nil {
		return fmt.Errorf("anomalies: %w", err)
	}

	return nil
}

func (s *Site) Generate() error {
	if err := s.Tree.Generate(!s.IncludePrivate); err != nil {
		return err
	}
	for _, p := range s.Tree.People {
		s.AssignTags(p)
		s.ScanPersonForAnomalies(p)
	}

	return nil
}

func (s *Site) AssignTags(p *model.Person) error {
	if p.RelationToKeyPerson != nil && p.RelationToKeyPerson.IsDirectAncestor() {
		p.Tags = append(p.Tags, "Direct Ancestor")
	}
	p.Tags = append(p.Tags, p.PreferredFamilyName)

	if len(p.Inferences) > 0 {
		p.Tags = append(p.Tags, "Has Inferences")
	}

	if len(p.Anomalies) > 0 {
		p.Tags = append(p.Tags, "Has Anomalies")
	}

	// if y, ok := gdate.AsYear(ev.Date); ok {
	// 	decade := (y.Year() / 10) * 10
	// 	p.Tags = append(p.Tags, fmt.Sprintf("born in %ds", decade))
	// }

	// if y, ok := gdate.AsYear(ev.Date); ok {
	// 	decade := (y.Year() / 10) * 10
	// 	p.Tags = append(p.Tags, fmt.Sprintf("died in %ds", decade))
	// }

	// if p.BestBirthlikeEvent == nil || gdate.IsUnknown(p.BestBirthlikeEvent.GetDate()) {
	// 	p.Tags = append(p.Tags, "Unknown Birthdate")
	// }
	// if p.BestDeathlikeEvent == nil || gdate.IsUnknown(p.BestDeathlikeEvent.GetDate()) {
	// 	p.Tags = append(p.Tags, "Unknown Deathdate")
	// }

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

			dt, ok := gdate.AsPrecise(ev.GetDate())
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
			evs, ok := monthEvents[dt.M]
			if !ok {
				evs = make(map[model.TimelineEvent]struct{})
			}
			evs[ev] = struct{}{}
			monthEvents[dt.M] = evs
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

func (s *Site) NewDocument() *md.Document {
	d := &md.Document{
		LinkBuilder: s,
	}

	return d
}

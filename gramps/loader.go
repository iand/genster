package gramps

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/iand/grampsxml"
)

var _ = logging.Debug

var startsWithNumber = regexp.MustCompile(`^[1-9]`)

type ModelFinder interface {
	FindPerson(scope string, id string) *model.Person
	FindSource(scope string, id string) *model.Source
	FindPlace(name string, id string) *model.Place
	FindFamilyByParents(father *model.Person, mother *model.Person) *model.Family
	AddAlias(alias string, canonical string)
}

type Loader struct {
	DB                   *grampsxml.Database
	ScopeName            string
	TagsByHandle         map[string]*grampsxml.Tag
	EventsByHandle       map[string]*grampsxml.Event
	PeopleByHandle       map[string]*grampsxml.Person
	FamiliesByHandle     map[string]*grampsxml.Family
	CitationsByHandle    map[string]*grampsxml.Citation
	SourcesByHandle      map[string]*grampsxml.Source
	PlacesByHandle       map[string]*grampsxml.Placeobj
	ObjectsByHandle      map[string]*grampsxml.Object
	RepositoriesByHandle map[string]*grampsxml.Repository
	NotesByHandle        map[string]*grampsxml.Note
	populatedPlaces      map[string]bool // which place handles have been populated to save repeated work when traversing the hierarchy
	censusEvents         map[string]*model.CensusEvent
}

func NewLoader(filename string) (*Loader, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("open gramps file: %w", err)
	}

	var db grampsxml.Database
	if err := xml.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("unmarshal gramps xml: %w", err)
	}

	l := &Loader{
		DB:                   &db,
		ScopeName:            filename,
		TagsByHandle:         make(map[string]*grampsxml.Tag),
		EventsByHandle:       make(map[string]*grampsxml.Event),
		PeopleByHandle:       make(map[string]*grampsxml.Person),
		FamiliesByHandle:     make(map[string]*grampsxml.Family),
		CitationsByHandle:    make(map[string]*grampsxml.Citation),
		SourcesByHandle:      make(map[string]*grampsxml.Source),
		PlacesByHandle:       make(map[string]*grampsxml.Placeobj),
		ObjectsByHandle:      make(map[string]*grampsxml.Object),
		RepositoriesByHandle: make(map[string]*grampsxml.Repository),
		NotesByHandle:        make(map[string]*grampsxml.Note),
		populatedPlaces:      make(map[string]bool),
		censusEvents:         make(map[string]*model.CensusEvent),
	}

	l.indexObjects()
	// l.readTags()

	// if id, ok := l.Attrs["ANCESTRY_TREE_ID"]; ok {
	// 	l.ScopeName = fmt.Sprintf("ANCESTRY_TREE_%s", id)
	// }

	return l, nil
}

func (l *Loader) Scope() string {
	return l.ScopeName
}

func (l *Loader) indexObjects() error {
	for _, v := range l.DB.Tags.Tag {
		l.TagsByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Events.Event {
		l.EventsByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.People.Person {
		l.PeopleByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Families.Family {
		l.FamiliesByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Citations.Citation {
		l.CitationsByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Sources.Source {
		l.SourcesByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Places.Place {
		l.PlacesByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Objects.Object {
		l.ObjectsByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Repositories.Repository {
		l.RepositoriesByHandle[v.Handle] = &v
	}
	for _, v := range l.DB.Notes.Note {
		l.NotesByHandle[v.Handle] = &v
	}
	return nil
}

// func (l *Loader) readAttrs() error {
// 	// Look for an ancestry tree identifier
// 	if l.Gedcom.Header.SourceSystem.BusinessName == "Ancestry.com" {
// 		for _, hud := range l.Gedcom.Header.SourceSystem.UserDefined {
// 			if hud.Tag == "_TREE" {
// 				if hud.Value != "" {
// 					l.Attrs["ANCESTRY_TREE_NAME"] = hud.Value
// 				}
// 				for _, tud := range hud.UserDefined {
// 					if tud.Value != "" {
// 						switch tud.Tag {
// 						case "RIN":
// 							l.Attrs["ANCESTRY_TREE_ID"] = tud.Value
// 						case "NOTE":
// 							l.Attrs["ANCESTRY_TREE_NOTE"] = tud.Value
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return nil
// }

// func (l *Loader) readTags() error {
// 	// Look for ancestry style tags using _MTTAG
// 	for _, ud := range l.Gedcom.UserDefined {
// 		if ud.Tag != "_MTTAG" {
// 			continue
// 		}
// 		for _, uds := range ud.UserDefined {
// 			if uds.Tag == "NAME" {
// 				l.Tags[ud.Xref] = uds.Value
// 				break
// 			}
// 		}
// 	}

// 	return nil
// }

func (l *Loader) Load(t *tree.Tree) error {
	// if name, ok := l.Attrs["ANCESTRY_TREE_NAME"]; ok {
	// 	t.Name = name
	// }
	// if desc, ok := l.Attrs["ANCESTRY_TREE_NOTE"]; ok {
	// 	t.Description = desc
	// }

	// for _, mr := range l.Gedcom.Media {
	// 	if err := l.populateMediaFacts(t, mr); err != nil {
	// 		return fmt.Errorf("media: %w", err)
	// 	}
	// }
	// logging.Info(fmt.Sprintf("loaded %d media records", len(l.Gedcom.Media)))

	// for _, sr := range l.Gedcom.Source {
	// 	if err := l.populateSourceFacts(t, sr); err != nil {
	// 		return fmt.Errorf("source: %w", err)
	// 	}
	// }
	// logging.Info(fmt.Sprintf("loaded %d source records", len(l.Gedcom.Source)))

	for _, p := range l.DB.Sources.Source {
		if err := l.populateSourceFacts(t, &p); err != nil {
			return fmt.Errorf("source: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d source records", len(l.DB.Sources.Source)))

	for _, p := range l.DB.Places.Place {
		if err := l.populatePlaceFacts(t, &p); err != nil {
			return fmt.Errorf("place: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d place records", len(l.DB.Places.Place)))

	for _, p := range l.DB.People.Person {
		if err := l.populatePersonFacts(t, &p); err != nil {
			return fmt.Errorf("person: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d person records", len(l.DB.People.Person)))

	for _, fr := range l.DB.Families.Family {
		if err := l.populateFamilyFacts(t, &fr); err != nil {
			return fmt.Errorf("family: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d family records", len(l.DB.Families.Family)))

	return nil
}

func pval[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}

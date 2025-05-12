package gramps

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/iand/grampsxml"
)

var _ = logging.Debug

type ModelFinder interface {
	FindPerson(scope string, id string) *model.Person
	FindCitation(scope string, id string) (*model.GeneralCitation, bool)
	FindSource(scope string, id string) *model.Source
	FindRepository(scope string, id string) *model.Repository
	FindPlace(name string, id string) *model.Place
	FindFamily(scope string, id string) *model.Family
	FindFamilyByParents(father *model.Person, mother *model.Person) *model.Family
	FindMediaObject(path string) *model.MediaObject
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
	multipartyEvents     map[string]model.MultipartyTimelineEvent
	unionEvents          map[string]model.UnionTimelineEvent
	timelineEvents       map[string]model.TimelineEvent
	familyNameGroups     map[string]string
}

func NewLoader(filename string, databaseName string) (*Loader, error) {
	db, err := openGrampsDB(filename)
	if err != nil {
		return nil, fmt.Errorf("open gramps file: %w", err)
	}

	scope := filename
	if databaseName != "" {
		scope = databaseName
	}

	l := &Loader{
		DB:                   db,
		ScopeName:            scope,
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
		multipartyEvents:     make(map[string]model.MultipartyTimelineEvent),
		unionEvents:          make(map[string]model.UnionTimelineEvent),
		timelineEvents:       make(map[string]model.TimelineEvent),

		familyNameGroups: make(map[string]string),
	}

	l.indexObjects()
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
	for _, v := range l.DB.Namemaps.Map {
		if v.Type != "group_as" {
			continue
		}
		l.familyNameGroups[v.Key] = v.Value
	}
	return nil
}

func (l *Loader) Load(t *tree.Tree) error {
	for _, o := range l.DB.Objects.Object {
		if err := l.populateObjectFacts(t, &o); err != nil {
			return fmt.Errorf("object: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d object records", len(l.DB.Objects.Object)))

	for _, p := range l.DB.Repositories.Repository {
		if err := l.populateRepositoryFacts(t, &p); err != nil {
			return fmt.Errorf("repository: %w", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d repository records", len(l.DB.Repositories.Repository)))

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

	for _, e := range l.DB.Events.Event {
		if pval(e.Priv, false) {
			logging.Debug("skipping event marked as private", "handle", e.Handle)
			continue
		}
		if err := l.populateEventFacts(t, &e); err != nil {
			logging.Error("failed to populate event facts", "error", err)
		}
	}
	logging.Info(fmt.Sprintf("loaded %d event records", len(l.DB.Events.Event)))

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

func openGrampsDB(fname string) (*grampsxml.Database, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	br := bufio.NewReader(f)
	b, err := br.Peek(2)
	if err != nil {
		return nil, fmt.Errorf("peeking leading bytes: %w", err)
	}

	var r io.Reader = br
	if b[0] == 31 && b[1] == 139 {
		r, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("reading gzip: %w", err)
		}
	}

	var db grampsxml.Database
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&db); err != nil {
		return nil, fmt.Errorf("unmarshal xml: %w", err)
	}

	return &db, nil
}

func pval[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}

func p[T any](v T) *T {
	return &v
}

func changeToTime(s string) (time.Time, error) {
	sec, err := strconv.Atoi(s)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(sec), 0), nil
}

func createdTimeFromHandle(h string) (time.Time, error) {
	if len(h) == 0 {
		return time.Time{}, errors.New("malformed handle")
	}
	if h[0] == '_' {
		h = h[1:]
	}
	if len(h) < 11 {
		return time.Time{}, errors.New("malformed handle")
	}
	n, err := strconv.ParseInt(h[:11], 16, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(n/10000, 0), nil
}

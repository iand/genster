package tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/flopp/go-coordsparser"
	"github.com/iand/genster/model"
	"golang.org/x/exp/slog"
)

type Annotations struct {
	person map[string]PersonMeta
	place  map[string][]PlaceAnnotation
	source map[string][]SourceAnnotation
	tree   map[string][]TreeAnnotation
}

type PersonMeta struct {
	Comment     string
	Annotations []PersonAnnotation
}

type PersonAnnotation struct {
	Field string
	Value any
	Kind  string
	Fn    personAnnotaterFunc
}

type PlaceAnnotation struct {
	Field string
	Value any
	Kind  string
	Fn    placeAnnotaterFunc
}

type SourceAnnotation struct {
	Field string
	Value any
	Kind  string
	Fn    sourceAnnotaterFunc
}

type TreeAnnotation struct {
	Field string
	Value any
	Kind  string
	Fn    treeAnnotaterFunc
}

func (o *Annotations) Replace(kind string, id string, field string, value any) error {
	switch kind {
	case "person":
		fn, ok := personReplacers[field]
		if !ok {
			return fmt.Errorf("unsupported annotation field for person: %s", field)
		}
		if o.person == nil {
			o.person = make(map[string]PersonMeta, 0)
		}
		pm := o.person[id]
		pm.Annotations = append(pm.Annotations, PersonAnnotation{
			Fn:    fn,
			Field: field,
			Kind:  "replace",
			Value: value,
		})
		o.person[id] = pm
	case "place":
		fn, ok := placeReplacers[field]
		if !ok {
			return fmt.Errorf("unsupported annotation field for place: %s", field)
		}
		if o.place == nil {
			o.place = make(map[string][]PlaceAnnotation, 0)
		}
		o.place[id] = append(o.place[id], PlaceAnnotation{
			Fn:    fn,
			Field: field,
			Kind:  "replace",
			Value: value,
		})
	case "source":
		fn, ok := sourceReplacers[field]
		if !ok {
			return fmt.Errorf("unsupported annotation field for source: %s", field)
		}
		if o.source == nil {
			o.source = make(map[string][]SourceAnnotation, 0)
		}
		o.source[id] = append(o.source[id], SourceAnnotation{
			Fn:    fn,
			Field: field,
			Kind:  "replace",
			Value: value,
		})
	case "tree":
		fn, ok := treeReplacers[field]
		if !ok {
			return fmt.Errorf("unsupported annotation field for tree: %s", field)
		}
		if o.tree == nil {
			o.tree = make(map[string][]TreeAnnotation, 0)
		}
		o.tree[id] = append(o.tree[id], TreeAnnotation{
			Fn:    fn,
			Field: field,
			Kind:  "replace",
			Value: value,
		})
	default:
		return fmt.Errorf("unsupported annotation kind: %s", kind)
	}

	return nil
}

func (o *Annotations) Add(kind string, id string, field string, value any) error {
	switch kind {
	case "person":
		fn, ok := personAdders[field]
		if !ok {
			return fmt.Errorf("unsupported annotation field for person: %s", field)
		}
		if o.person == nil {
			o.person = make(map[string]PersonMeta, 0)
		}
		pm := o.person[id]
		pm.Annotations = append(pm.Annotations, PersonAnnotation{
			Fn:    fn,
			Field: field,
			Kind:  "add",
			Value: value,
		})
		o.person[id] = pm

	case "place":
		fn, ok := placeAdders[field]
		if !ok {
			return fmt.Errorf("unsupported annotation field for place: %s", field)
		}
		if o.place == nil {
			o.place = make(map[string][]PlaceAnnotation, 0)
		}
		o.place[id] = append(o.place[id], PlaceAnnotation{
			Fn:    fn,
			Field: field,
			Kind:  "add",
			Value: value,
		})
	case "source":
		fn, ok := sourceAdders[field]
		if !ok {
			return fmt.Errorf("unsupported annotation field for source: %s", field)
		}
		if o.source == nil {
			o.source = make(map[string][]SourceAnnotation, 0)
		}
		o.source[id] = append(o.source[id], SourceAnnotation{
			Fn:    fn,
			Field: field,
			Kind:  "add",
			Value: value,
		})
	default:
		return fmt.Errorf("unsupported annotation kind: %s", kind)
	}

	return nil
}

func (o *Annotations) ApplyPerson(p *model.Person) error {
	pm, ok := o.person[p.ID]
	if !ok {
		return nil
	}

	for _, ann := range pm.Annotations {
		slog.Debug("annotation for person", "id", p.ID, "field", ann.Field, "value", ann.Value)
		if err := ann.Fn(p, ann.Value); err != nil {
			return fmt.Errorf("annotating value of person field %s: %w", ann.Field, err)
		}
	}

	return nil
}

func (o *Annotations) ApplyPlace(p *model.Place) error {
	anns, ok := o.place[p.ID]
	if !ok {
		return nil
	}

	for _, ann := range anns {
		slog.Debug("annotation for place", "id", p.ID, "field", ann.Field, "value", ann.Value)
		if err := ann.Fn(p, ann.Value); err != nil {
			return fmt.Errorf("annotating value of place field %s: %w", ann.Field, err)
		}
	}

	return nil
}

func (o *Annotations) ApplySource(p *model.Source) error {
	anns, ok := o.source[p.ID]
	if !ok {
		return nil
	}

	for _, ann := range anns {
		slog.Debug("annotation for source", "id", p.ID, "field", ann.Field, "value", ann.Value)
		if err := ann.Fn(p, ann.Value); err != nil {
			return fmt.Errorf("annotating value of source field %s: %w", ann.Field, err)
		}
	}

	return nil
}

func (o *Annotations) ApplyTree(t *Tree) error {
	anns, ok := o.tree[t.ID]
	if !ok {
		return nil
	}

	for _, ann := range anns {
		slog.Debug("annotation for tree", "id", t.ID, "field", ann.Field, "value", ann.Value)
		if err := ann.Fn(t, ann.Value); err != nil {
			return fmt.Errorf("annotating value of place field %s: %w", ann.Field, err)
		}
	}

	return nil
}

func (a *Annotations) UnmarshalJSON(data []byte) error {
	r := bytes.NewReader(data)
	d := json.NewDecoder(r)

	var aj AnnotationsJSON
	err := d.Decode(&aj)
	if err != nil {
		return err
	}

	for _, oa := range aj.People {
		if a.person == nil {
			a.person = make(map[string]PersonMeta, 0)
		}
		pm := a.person[oa.ID]
		pm.Comment = oa.Comment
		a.person[oa.ID] = pm

		for f, v := range oa.Replace {
			if err := a.Replace("person", oa.ID, f, v); err != nil {
				return err
			}
		}
		for f, v := range oa.Add {
			if err := a.Add("person", oa.ID, f, v); err != nil {
				return err
			}
		}
	}

	for _, oa := range aj.Places {
		for f, v := range oa.Replace {
			if err := a.Replace("place", oa.ID, f, v); err != nil {
				return err
			}
		}
		for f, v := range oa.Add {
			if err := a.Add("place", oa.ID, f, v); err != nil {
				return err
			}
		}
	}

	for _, oa := range aj.Sources {
		for f, v := range oa.Replace {
			if err := a.Replace("source", oa.ID, f, v); err != nil {
				return err
			}
		}
		for f, v := range oa.Add {
			if err := a.Add("source", oa.ID, f, v); err != nil {
				return err
			}
		}
	}

	for _, oa := range aj.Trees {
		for f, v := range oa.Replace {
			if err := a.Replace("tree", oa.ID, f, v); err != nil {
				return err
			}
		}
		for f, v := range oa.Add {
			if err := a.Add("tree", oa.ID, f, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Annotations) MarshalJSON() ([]byte, error) {
	aj := &AnnotationsJSON{
		People:  make([]ObjectAnnotationsJSON, 0, len(a.person)),
		Places:  make([]ObjectAnnotationsJSON, 0, len(a.place)),
		Sources: make([]ObjectAnnotationsJSON, 0, len(a.source)),
		Trees:   make([]ObjectAnnotationsJSON, 0, len(a.tree)),
	}

	for id, pm := range a.person {
		oa := ObjectAnnotationsJSON{
			ID:      id,
			Comment: pm.Comment,
			Replace: make(map[string]any),
			Add:     make(map[string]any),
		}

		for _, ann := range pm.Annotations {
			switch ann.Kind {
			case "replace":
				oa.Replace[ann.Field] = ann.Value
			case "add":
				oa.Add[ann.Field] = ann.Value
			default:
				slog.Warn("unsupported annotation kind, not writing to file: " + ann.Kind)
			}
		}

		aj.People = append(aj.People, oa)
	}
	sort.Slice(aj.People, func(i, j int) bool { return aj.People[i].ID < aj.People[j].ID })

	for id, anns := range a.place {
		oa := ObjectAnnotationsJSON{
			ID:      id,
			Replace: make(map[string]any),
			Add:     make(map[string]any),
		}

		for _, ann := range anns {
			switch ann.Kind {
			case "replace":
				oa.Replace[ann.Field] = ann.Value
			case "add":
				oa.Add[ann.Field] = ann.Value
			default:
				slog.Warn("unsupported annotation kind, not writing to file: " + ann.Kind)
			}
		}

		aj.Places = append(aj.Places, oa)
	}
	sort.Slice(aj.Places, func(i, j int) bool { return aj.Places[i].ID < aj.Places[j].ID })

	for id, anns := range a.source {
		oa := ObjectAnnotationsJSON{
			ID:      id,
			Replace: make(map[string]any),
			Add:     make(map[string]any),
		}

		for _, ann := range anns {
			switch ann.Kind {
			case "replace":
				oa.Replace[ann.Field] = ann.Value
			case "add":
				oa.Add[ann.Field] = ann.Value
			default:
				slog.Warn("unsupported annotation kind, not writing to file: " + ann.Kind)
			}
		}

		aj.Sources = append(aj.Sources, oa)
	}
	sort.Slice(aj.Sources, func(i, j int) bool { return aj.Sources[i].ID < aj.Sources[j].ID })

	for id, anns := range a.tree {
		oa := ObjectAnnotationsJSON{
			ID:      id,
			Replace: make(map[string]any),
			Add:     make(map[string]any),
		}

		for _, ann := range anns {
			switch ann.Kind {
			case "replace":
				oa.Replace[ann.Field] = ann.Value
			case "add":
				oa.Add[ann.Field] = ann.Value
			default:
				slog.Warn("unsupported annotation kind, not writing to file: " + ann.Kind)
			}
		}

		aj.Trees = append(aj.Trees, oa)
	}
	sort.Slice(aj.Trees, func(i, j int) bool { return aj.Trees[i].ID < aj.Trees[j].ID })

	return json.Marshal(aj)
}

func setString(f *string, v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("expected a string value")
	}

	*f = s
	return nil
}

func setBool(f *bool, v any) error {
	b, ok := v.(bool)
	if !ok {
		return fmt.Errorf("expected a boolean value")
	}

	*f = b
	return nil
}

func setCoordinates(latf *float64, longf *float64, v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("expected a string containing a pair of coordinates")
	}

	lat, long, err := coordsparser.Parse(s)
	if err != nil {
		return fmt.Errorf("failed to parse coordinates string: %s", s)
	}
	*latf = lat
	*longf = long
	return nil
}

func appendStringOrList(f *[]string, v any) error {
	switch tv := v.(type) {
	case string:
		*f = append(*f, tv)
	case []string:
		*f = append(*f, tv...)
	case []any:
		strs := make([]string, len(tv))
		for i := range tv {
			if s, ok := tv[i].(string); ok {
				strs[i] = s
			} else {
				return fmt.Errorf("expected a string value or a list of strings")
			}
		}
		*f = append(*f, strs...)
	default:
		return fmt.Errorf("expected a string value or a list of strings")
	}
	return nil
}

type (
	personAnnotaterFunc func(*model.Person, any) error
	placeAnnotaterFunc  func(*model.Place, any) error
	sourceAnnotaterFunc func(*model.Source, any) error
	treeAnnotaterFunc   func(*Tree, any) error
)

// all possible person replacers. use a map so the names of the overrides could be
// printed in a help command.
var personReplacers = map[string]personAnnotaterFunc{
	"nickname":                  func(p *model.Person, v any) error { return setString(&p.NickName, v) },
	"olb":                       func(p *model.Person, v any) error { return setString(&p.Olb, v) },
	"preferredfullname":         func(p *model.Person, v any) error { return setString(&p.PreferredFullName, v) },
	"preferredgivenname":        func(p *model.Person, v any) error { return setString(&p.PreferredGivenName, v) },
	"preferredfamiliarname":     func(p *model.Person, v any) error { return setString(&p.PreferredFamiliarName, v) },
	"preferredfamiliarfullname": func(p *model.Person, v any) error { return setString(&p.PreferredFamiliarFullName, v) },
	"preferredfamilyname":       func(p *model.Person, v any) error { return setString(&p.PreferredFamilyName, v) },
	"preferredsortname":         func(p *model.Person, v any) error { return setString(&p.PreferredSortName, v) },
	"preferreduniquename":       func(p *model.Person, v any) error { return setString(&p.PreferredUniqueName, v) },
	"wikitreeid":                func(p *model.Person, v any) error { return setString(&p.WikiTreeID, v) },
	// "causeofdeath":              func(p *model.Person, v any) error { return setString(&p.CauseOfDeath, v) },

	"possiblyalive": func(p *model.Person, v any) error { return setBool(&p.PossiblyAlive, v) },
	"unmarried":     func(p *model.Person, v any) error { return setBool(&p.Unmarried, v) },
	"childless":     func(p *model.Person, v any) error { return setBool(&p.Childless, v) },
	"illegitimate":  func(p *model.Person, v any) error { return setBool(&p.Illegitimate, v) },
	"redacted":      func(p *model.Person, v any) error { return setBool(&p.Redacted, v) },
	"featured":      func(p *model.Person, v any) error { return setBool(&p.Featured, v) },
}

// all possible place replacers
var placeReplacers = map[string]placeAnnotaterFunc{
	"preferredname": func(p *model.Place, v any) error { return setString(&p.PreferredName, v) },
	"latlong":       func(p *model.Place, v any) error { return setCoordinates(&p.Latitude, &p.Longitude, v) },
}

// all possible source replacers
var sourceReplacers = map[string]sourceAnnotaterFunc{
	"title":               func(s *model.Source, v any) error { return setString(&s.Title, v) },
	"searchlink":          func(s *model.Source, v any) error { return setString(&s.SearchLink, v) },
	"repositoryname":      func(s *model.Source, v any) error { return setString(&s.RepositoryName, v) },
	"repositorylink":      func(s *model.Source, v any) error { return setString(&s.RepositoryLink, v) },
	"iscivilregistration": func(s *model.Source, v any) error { return setBool(&s.IsCivilRegistration, v) },
	"iscensus":            func(s *model.Source, v any) error { return setBool(&s.IsCensus, v) },
}

// all possible tree replacers
var treeReplacers = map[string]treeAnnotaterFunc{
	"name":        func(t *Tree, v any) error { return setString(&t.Name, v) },
	"description": func(t *Tree, v any) error { return setString(&t.Description, v) },
}

// all possible person adders. use a map so the names of the overrides could be
// printed in a help command.
var personAdders = map[string]personAnnotaterFunc{
	"tags": func(p *model.Person, v any) error { return appendStringOrList(&p.Tags, v) },
}

// all possible place adders
var placeAdders = map[string]placeAnnotaterFunc{
	"tags": func(p *model.Place, v any) error { return appendStringOrList(&p.Tags, v) },
}

// all possible source adders
var sourceAdders = map[string]sourceAnnotaterFunc{
	"tags": func(s *model.Source, v any) error { return appendStringOrList(&s.Tags, v) },
}

type AnnotationsJSON struct {
	People  []ObjectAnnotationsJSON `json:"people,omitempty"`
	Places  []ObjectAnnotationsJSON `json:"places,omitempty"`
	Sources []ObjectAnnotationsJSON `json:"sources,omitempty"`
	Trees   []ObjectAnnotationsJSON `json:"trees,omitempty"`
}

type ObjectAnnotationsJSON struct {
	ID      string         `json:"id,omitempty"`
	Comment string         `json:"comment,omitempty"` // for use by user as a free form comment in the annotations file
	Replace map[string]any `json:"replace,omitempty"`
	Add     map[string]any `json:"add,omitempty"`
}

func LoadAnnotations(filename string) (*Annotations, error) {
	var a Annotations
	if filename == "" {
		return &a, nil
	}

	slog.Info("reading annotations", "filename", filename)
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &a, nil
		}
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	if err := d.Decode(&a); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	return &a, nil
}

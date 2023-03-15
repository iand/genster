package tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/iand/genster/model"
	"golang.org/x/exp/slog"
)

type Annotations struct {
	person map[string]map[string]string
	place  map[string]map[string]string
}

func (o *Annotations) Add(kind string, id string, field string, value string) error {
	switch kind {
	case "person":
		if _, ok := personOverrides[field]; !ok {
			return fmt.Errorf("unsupported override field for person: %s", field)
		}
		if o.person == nil {
			o.person = make(map[string]map[string]string, 0)
		}
		m, ok := o.person[id]
		if !ok {
			m = make(map[string]string)
		}
		m[strings.ToLower(field)] = value
		o.person[id] = m
	case "place":
		if _, ok := placeOverrides[field]; !ok {
			return fmt.Errorf("unsupported override field for place: %s", field)
		}
		if o.place == nil {
			o.place = make(map[string]map[string]string, 0)
		}
		m, ok := o.place[id]
		if !ok {
			m = make(map[string]string)
		}
		m[strings.ToLower(field)] = value
		o.place[id] = m
	default:
		return fmt.Errorf("unsupported override kind: %s", kind)
	}

	return nil
}

func (o *Annotations) ApplyPerson(p *model.Person) error {
	fvs, ok := o.person[p.ID]
	if !ok {
		return nil
	}

	for f, v := range fvs {
		if fn, ok := personOverrides[f]; ok {
			fn(p, v)
		}
	}

	return nil
}

func (o *Annotations) ApplyPlace(p *model.Place) error {
	fvs, ok := o.place[p.ID]
	if !ok {
		return nil
	}

	for f, v := range fvs {
		if fn, ok := placeOverrides[f]; ok {
			fn(p, v)
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
		for _, ann := range oa.Annotations {
			if err := a.Add("person", oa.ID, ann.Field, ann.Value); err != nil {
				return err
			}
		}
	}

	for _, oa := range aj.Places {
		for _, ann := range oa.Annotations {
			if err := a.Add("place", oa.ID, ann.Field, ann.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Annotations) MarshalJSON() ([]byte, error) {
	aj := &AnnotationsJSON{
		People: make([]ObjectAnnotationsJSON, 0, len(a.person)),
		Places: make([]ObjectAnnotationsJSON, 0, len(a.place)),
	}

	for id, fvs := range a.person {
		oa := ObjectAnnotationsJSON{
			ID: id,
		}

		for f, v := range fvs {
			oa.Annotations = append(oa.Annotations, AnnotationJSON{Field: f, Value: v})
		}
		sort.Slice(oa.Annotations, func(i, j int) bool { return oa.Annotations[i].Field < oa.Annotations[j].Field })

		aj.People = append(aj.People, oa)
	}
	sort.Slice(aj.People, func(i, j int) bool { return aj.People[i].ID < aj.People[j].ID })

	for id, fvs := range a.place {
		oa := ObjectAnnotationsJSON{
			ID: id,
		}

		for f, v := range fvs {
			oa.Annotations = append(oa.Annotations, AnnotationJSON{Field: f, Value: v})
		}
		sort.Slice(oa.Annotations, func(i, j int) bool { return oa.Annotations[i].Field < oa.Annotations[j].Field })

		aj.Places = append(aj.Places, oa)
	}
	sort.Slice(aj.Places, func(i, j int) bool { return aj.Places[i].ID < aj.Places[j].ID })

	return json.Marshal(aj)
}

type FV struct {
	Field string
	Value string
}

// all possible person overrides. use a map so the names of the overrides could be
// printed in a help command.
var personOverrides = map[string]func(p *model.Person, v string){
	"nickname": func(p *model.Person, v string) { p.NickName = v },
	"olb":      func(p *model.Person, v string) { p.Olb = v },
}

// all possible place overrides
var placeOverrides = map[string]func(p *model.Place, v string){
	"preferredname": func(p *model.Place, v string) { p.PreferredName = v },
}

type AnnotationsJSON struct {
	People []ObjectAnnotationsJSON `json:"people,omitempty"`
	Places []ObjectAnnotationsJSON `json:"places,omitempty"`
}

type ObjectAnnotationsJSON struct {
	ID          string           `json:"id,omitempty"`
	Annotations []AnnotationJSON `json:"annotations,omitempty"`
}

type AnnotationJSON struct {
	Field string `json:"field,omitempty"`
	Value string `json:"value,omitempty"`
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

func SaveAnnotations(filename string, a *Annotations) error {
	if filename == "" {
		return nil
	}

	slog.Info("writing annotations", "filename", filename)
	f, err := CreateFile(filename)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	d := json.NewEncoder(f)
	d.SetIndent("", "  ")
	if err := d.Encode(&a); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

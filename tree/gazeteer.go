package tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/iand/genster/identifier"
	"github.com/iand/genster/place"
	"golang.org/x/exp/slog"
)

var ErrInvalidPlaceName = errors.New("invalid place name")

type Gazeteer struct {
	places map[string]GazeteerPlace // map of canonical id to place information
	lookup map[string]string        // map of place names to canonical id
}

type GazeteerPlace struct {
	id       string
	idname   string // the name that was used to generate the id
	name     string // the administrative name of the place without hierarchy
	parentID string
}

// MatchPlace returns information about the named place, creating a new one if necessary
// The supplied name is assumed to be generally unstructured, but ordered hierarchically from most to least specific
// with each heierarchy level separated by a comma
func (g *Gazeteer) MatchPlace(original string, hints ...place.Hint) (GazeteerPlace, error) {
	if g.lookup == nil {
		g.lookup = map[string]string{}
	}
	if g.places == nil {
		g.places = map[string]GazeteerPlace{}
	}
	id, exists := g.lookup[original]
	if exists {
		gp, exists := g.places[id]
		if !exists {
			return GazeteerPlace{}, fmt.Errorf("could not find place with id %q in gazeteer", id)
		}

		return gp, nil
	}

	norm := place.Normalize(original)
	if norm == "" {
		return GazeteerPlace{}, ErrInvalidPlaceName
	}

	id, exists = g.lookup[norm]
	if exists {
		gp, exists := g.places[id]
		if !exists {
			return GazeteerPlace{}, fmt.Errorf("could not find place with id %q in gazeteer", id)
		}

		g.lookup[original] = id
		return gp, nil
	}

	id = g.makeID(norm)
	g.lookup[norm] = id
	g.lookup[original] = id
	gp := GazeteerPlace{
		id:     id,
		idname: norm,
	}
	g.places[id] = gp

	ph, ok := place.Parse(original)
	if !ok {
		return gp, nil
	}
	gp.name = ph.Name
	g.places[id] = gp

	parent, ok := ph.Parent()
	if !ok {
		return gp, nil
	}

	parentgp, err := g.MatchPlace(parent.String())
	if err != nil {
		return gp, nil
	}
	gp.parentID = parentgp.id
	g.places[id] = gp

	return gp, nil
}

func (g *Gazeteer) LookupPlace(id string) (GazeteerPlace, bool) {
	gp, ok := g.places[id]
	return gp, ok
}

func (g *Gazeteer) makeID(id string) string {
	return identifier.New("gazeteer", id)
}

func (g *Gazeteer) UnmarshalJSON(data []byte) error {
	g.places = map[string]GazeteerPlace{}
	g.lookup = map[string]string{}

	r := bytes.NewReader(data)
	d := json.NewDecoder(r)

	var jg GazeteerJSON
	err := d.Decode(&jg)
	if err != nil {
		return err
	}

	g.places = map[string]GazeteerPlace{}
	for id, info := range jg.Places {
		g.places[id] = GazeteerPlace{
			id:       id,
			name:     info.Name,
			idname:   info.IDName,
			parentID: info.ParentID,
		}
		g.lookup[info.IDName] = id
	}

	return nil
}

func (g *Gazeteer) MarshalJSON() ([]byte, error) {
	jg := &GazeteerJSON{
		Places: map[string]PlaceInfoJSON{},
	}

	for id, info := range g.places {
		jg.Places[id] = PlaceInfoJSON{
			Name:     info.name,
			IDName:   info.idname,
			ParentID: info.parentID,
		}
	}

	for name, id := range g.lookup {
		info, ok := jg.Places[id]
		if !ok {
			info = PlaceInfoJSON{}
		}
		if name != info.IDName {
			info.Matches = append(info.Matches, name)
		}
		jg.Places[id] = info
	}

	return json.Marshal(jg)
}

type GazeteerJSON struct {
	Places map[string]PlaceInfoJSON `json:"places"`
}

type PlaceInfoJSON struct {
	Name     string   `json:"name"`
	IDName   string   `json:"idname"`
	Matches  []string `json:"matches"`
	ParentID string   `json:"parentid"`
}

func LoadGazeteer(filename string) (*Gazeteer, error) {
	var g Gazeteer
	if filename == "" {
		return &g, nil
	}

	slog.Info("reading gazeteer", "filename", filename)
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &g, nil
		}
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	if err := d.Decode(&g); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	return &g, nil
}

func SaveGazeteer(filename string, g *Gazeteer) error {
	if filename == "" {
		return nil
	}

	slog.Info("writing gazeteer", "filename", filename)
	f, err := CreateFile(filename)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	d := json.NewEncoder(f)
	d.SetIndent("", "  ")
	if err := d.Encode(&g); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

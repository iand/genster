package tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

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
	idsrc    string          // the string that was used to generate the id
	name     string          // the administrative name of the place without hierarchy
	kind     place.PlaceKind // kind of place, empty if not known
	parentID string
}

// MatchPlace returns information about the named place, creating a new one if necessary
// The supplied name is assumed to be generally unstructured, but ordered hierarchically from most to least specific
// with each heierarchy level separated by a comma. Additional context may be supplied by the use of hints.
// For example the user could hint that the place could be a UK registration district
// based on the source being the general register office.
func (g *Gazeteer) MatchPlace(original string, hints ...place.Hint) (GazeteerPlace, error) {
	if g.lookup == nil {
		g.lookup = map[string]string{}
	}
	if g.places == nil {
		g.places = map[string]GazeteerPlace{}
	}

	norm := place.Normalize(original)
	if norm == "" {
		return GazeteerPlace{}, ErrInvalidPlaceName
	}

	hintIDs := make([]string, len(hints))
	for i := range hints {
		hintIDs[i] = hints[i].ID()
	}
	sort.Strings(hintIDs)
	hintstr := strings.Join(hintIDs, "|")

	lookupstr := norm
	if len(hints) > 0 {
		lookupstr += "|" + hintstr
	}

	// See if we have already processed this placename
	id, exists := g.lookup[lookupstr]
	if exists {
		gp, exists := g.places[id]
		if !exists {
			// this only happens if the gazeteer is corrupt or has been edited
			return GazeteerPlace{}, fmt.Errorf("could not find place with id %q in gazeteer", id)
		}

		return gp, nil
	}

	ph, ok := place.ParseHierarchy(original, hints...)
	if !ok {
		return GazeteerPlace{}, fmt.Errorf("could not parse hierarchy of %q", original)
	}

	gp, err := g.addPlaceHierarchy(ph)
	if err != nil {
		return GazeteerPlace{}, err
	}
	g.lookup[lookupstr] = gp.id
	return gp, nil
}

func (g *Gazeteer) addPlaceHierarchy(ph *place.PlaceHierarchy) (GazeteerPlace, error) {
	idsrc := ph.NormalizedWithHierarchy
	if len(ph.Kind) > 0 {
		idsrc += "|" + string(ph.Kind)
	}

	id := g.makeID(idsrc)

	gp, exists := g.places[id]
	if exists {
		return gp, nil
	}

	gp = GazeteerPlace{
		id:    id,
		idsrc: idsrc,
		name:  ph.Name.Name,
		kind:  ph.Kind,
	}
	g.places[id] = gp

	if ph.Parent != nil {
		parent, err := g.addPlaceHierarchy(ph.Parent)
		if err != nil {
			return GazeteerPlace{}, err
		}
		gp.parentID = parent.id
		g.places[id] = gp
	}

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
		gp := GazeteerPlace{
			id:       id,
			name:     info.Name,
			kind:     place.PlaceKind(info.Kind),
			idsrc:    info.IDSrc,
			parentID: info.ParentID,
		}
		g.places[id] = gp
		g.lookup[info.IDSrc] = id
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
			Kind:     string(info.kind),
			IDSrc:    info.idsrc,
			ParentID: info.parentID,
		}
	}

	for name, id := range g.lookup {
		info, ok := jg.Places[id]
		if !ok {
			info = PlaceInfoJSON{}
		}
		if name != info.IDSrc {
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
	Kind     string   `json:"kind,omitempty"`
	IDSrc    string   `json:"idsrc"`
	Matches  []string `json:"matches,omitempty"`
	ParentID string   `json:"parentid,omitempty"`
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

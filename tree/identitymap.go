package tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/iand/genster/identifier"
	"golang.org/x/exp/slog"
)

type IdentityMap struct {
	scopes       map[string]map[string]string // map of scopes each with scoped id mapped to canonical id
	replacements map[string]string            // map of ids to a canonical replacement, works for both scoped mappings and places
}

// ID returns the canonical identifier corresponding to the scoped identifier, creating a new one if necessary
// scope indicates the source of the supplied identifier, usually a gedcom file.
func (m *IdentityMap) ID(scope string, id string) string {
	if m.scopes == nil {
		m.scopes = map[string]map[string]string{}
	}
	sm, exists := m.scopes[scope]
	if !exists {
		generated := m.makeID(scope, id)
		m.scopes[scope] = map[string]string{
			id: generated,
		}
		return m.findCanonical(generated)
	}

	generated, exists := sm[id]
	if !exists {
		generated = m.makeID(scope, id)
		sm[id] = generated
		m.scopes[scope] = sm
	}

	return m.findCanonical(generated)
}

func (m *IdentityMap) findCanonical(canonical string) string {
	if m.replacements == nil {
		return canonical
	}
	if replacement, ok := m.replacements[canonical]; ok {
		return replacement
	}
	return canonical
}

// AddAlias adds an alias to a canonical identifier.
// This can be used to create more friendly identifiers.
func (m *IdentityMap) AddAlias(alias string, canonical string) {
	if alias == canonical {
		return
	}
	if m.replacements == nil {
		m.replacements = map[string]string{}
	}
	m.replacements[alias] = canonical
}

func (m *IdentityMap) makeID(scope string, id string) string {
	return identifier.New(scope, id)
}

func (m *IdentityMap) UnmarshalJSON(data []byte) error {
	r := bytes.NewReader(data)
	d := json.NewDecoder(r)

	var jm IdentityMapJSON
	err := d.Decode(&jm)
	if err != nil {
		return err
	}
	m.replacements = jm.Replace
	m.scopes = map[string]map[string]string{}

	for scope, mappings := range jm.Scopes {
		sm := map[string]string{}
		for _, mapping := range mappings {
			sm[mapping.ScopeID] = mapping.ID
		}
		m.scopes[scope] = sm
	}

	return nil
}

func (m *IdentityMap) MarshalJSON() ([]byte, error) {
	jm := &IdentityMapJSON{
		Scopes:  map[string][]IdentityMappingJSON{},
		Replace: m.replacements,
	}

	for scope, sm := range m.scopes {
		var mappings []IdentityMappingJSON
		for scopeid, id := range sm {
			mappings = append(mappings, IdentityMappingJSON{
				ScopeID: scopeid,
				ID:      id,
			})
		}
		jm.Scopes[scope] = mappings
	}

	return json.Marshal(jm)
}

type IdentityMapJSON struct {
	Scopes  map[string][]IdentityMappingJSON `json:"scopes"`
	Replace map[string]string                `json:"replace"`
	Places  map[string]string                `json:"places"`
}

type IdentityMappingJSON struct {
	ScopeID string `json:"scopeid"`
	ID      string `json:"id"`
}

func LoadIdentityMap(imFilename string) (*IdentityMap, error) {
	var im IdentityMap
	if imFilename == "" {
		return &im, nil
	}

	slog.Info("reading identity map", "filename", imFilename)

	imFile, err := os.Open(imFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &im, nil
		}
		return nil, fmt.Errorf("open identity map: %w", err)
	}
	defer imFile.Close()

	d := json.NewDecoder(imFile)
	if err := d.Decode(&im); err != nil {
		return nil, fmt.Errorf("read identity map: %w", err)
	}

	return &im, nil
}

func SaveIdentityMap(imFilename string, im *IdentityMap) error {
	if imFilename == "" {
		return nil
	}

	slog.Info("writing identity map", "filename", imFilename)
	imFile, err := CreateFile(imFilename)
	if err != nil {
		return fmt.Errorf("open identity map: %w", err)
	}
	defer imFile.Close()

	d := json.NewEncoder(imFile)
	d.SetIndent("", "  ")
	if err := d.Encode(&im); err != nil {
		return fmt.Errorf("write identity map: %w", err)
	}

	return nil
}

package tree

import (
	"fmt"
	"os"
	"path/filepath"
)

func DefaultConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "genster")
}

type Loader interface {
	Load(*Tree) error
	Scope() string
}

func LoadTree(id string, configDir string, loader Loader) (*Tree, error) {
	var identityMapFilename string
	var annotationsFilename string
	var surnamesFilename string
	if configDir != "" {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
		identityMapFilename = filepath.Join(configDir, "identitymap.json")
		annotationsFilename = filepath.Join(configDir, "annotations.json")
		surnamesFilename = filepath.Join(configDir, "surnames.json")
	}

	im, err := LoadIdentityMap(identityMapFilename)
	if err != nil {
		return nil, fmt.Errorf("load identity map: %w", err)
	}

	// Annotations are only read by genster, never written
	a, err := LoadAnnotations(annotationsFilename)
	if err != nil {
		return nil, fmt.Errorf("load annotations: %w", err)
	}

	// Surname groupings are only read by genster, never written
	sg, err := LoadSurnameGroups(surnamesFilename)
	if err != nil {
		return nil, fmt.Errorf("load surname groupings: %w", err)
	}

	t := NewTree(id, im, a, sg)

	if err := loader.Load(t); err != nil {
		return nil, fmt.Errorf("load data: %w", err)
	}

	if err := SaveIdentityMap(identityMapFilename, im); err != nil {
		return nil, fmt.Errorf("save identity map: %w", err)
	}

	return t, nil
}

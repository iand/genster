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
}

func LoadTree(configDir string, loader Loader) (*Tree, error) {
	var identityMapFilename string
	var gazeteerFilename string
	var annotationsFilename string
	if configDir != "" {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
		identityMapFilename = filepath.Join(configDir, "identitymap.json")
		gazeteerFilename = filepath.Join(configDir, "gazeteer.json")
		annotationsFilename = filepath.Join(configDir, "annotations.json")
	}

	im, err := LoadIdentityMap(identityMapFilename)
	if err != nil {
		return nil, fmt.Errorf("load identity map: %w", err)
	}

	g, err := LoadGazeteer(gazeteerFilename)
	if err != nil {
		return nil, fmt.Errorf("load gazeteer: %w", err)
	}

	a, err := LoadAnnotations(annotationsFilename)
	if err != nil {
		return nil, fmt.Errorf("load annotations: %w", err)
	}

	t := NewTree(im, g, a)

	if err := loader.Load(t); err != nil {
		return nil, fmt.Errorf("load gedcom: %w", err)
	}

	if err := SaveIdentityMap(identityMapFilename, im); err != nil {
		return nil, fmt.Errorf("save identity map: %w", err)
	}

	if err := SaveGazeteer(gazeteerFilename, g); err != nil {
		return nil, fmt.Errorf("save gazeteer: %w", err)
	}

	if err := SaveAnnotations(annotationsFilename, a); err != nil {
		return nil, fmt.Errorf("save annotations: %w", err)
	}

	return t, nil
}

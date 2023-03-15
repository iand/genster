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
	if configDir != "" {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
		identityMapFilename = filepath.Join(configDir, "identitymap.json")
		gazeteerFilename = filepath.Join(configDir, "gazeteer.json")
	}

	im, err := LoadIdentityMap(identityMapFilename)
	if err != nil {
		return nil, fmt.Errorf("load identity map: %w", err)
	}

	g, err := LoadGazeteer(gazeteerFilename)
	if err != nil {
		return nil, fmt.Errorf("load gazeteer: %w", err)
	}

	// im.ReplaceCanonical("ZUKZH4UEY66AK", "bertie-herbert-tew")

	overrides := new(Overrides)
	overrides.AddOverride("person", "TA5YTXSS52YDC", "nickname", "Andy")
	overrides.AddOverride("person", "S4GUXNLNCYIBY", "nickname", "Tizzie")
	overrides.AddOverride("person", "S4GULNLNCYC4U", "nickname", "Peggy")
	overrides.AddOverride("person", "ZU7RPKPLONKFC", "nickname", "Nel")
	overrides.AddOverride("person", "TC7G3QSRW4GAY", "nickname", "Bill")
	overrides.AddOverride("person", "TC7HJQSRW4MCU", "nickname", "Flo")

	t := NewTree(im, g, overrides)

	if err := loader.Load(t); err != nil {
		return nil, fmt.Errorf("load gedcom: %w", err)
	}

	if err := SaveIdentityMap(identityMapFilename, im); err != nil {
		return nil, fmt.Errorf("save identity map: %w", err)
	}

	if err := SaveGazeteer(gazeteerFilename, g); err != nil {
		return nil, fmt.Errorf("save gazeteer: %w", err)
	}

	return t, nil
}

package tree

import (
	"fmt"
)

type Loader interface {
	Load(*Tree) error
	Scope() string
}

func LoadTree(cfg *Config, loader Loader) (*Tree, error) {
	id := cfg.ID

	a := cfg.Annotations
	if a == nil {
		a = &Annotations{}
	}

	sg := cfg.SurnameGroups
	if sg == nil {
		sg = &SurnameGroups{}
	}

	t := NewTree(id, a, sg)

	if err := loader.Load(t); err != nil {
		return nil, fmt.Errorf("load data: %w", err)
	}

	if cfg.Name != "" {
		t.Name = cfg.Name
	}
	if cfg.Description != "" {
		t.Description = cfg.Description
	}

	return t, nil
}

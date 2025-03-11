package tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
)

type SurnameGroups struct {
	surnames map[string]*SurnameGroup
}

type SurnameGroup struct {
	Surname string   // canonical surname for this group
	Names   []string // sorted list of surnames in group other than the canonical one
}

func (g *SurnameGroup) String() string {
	return fmt.Sprintf("%s (%s)", g.Surname, strings.Join(g.Names, ", "))
}

func (sg *SurnameGroups) UnmarshalJSON(data []byte) error {
	r := bytes.NewReader(data)
	d := json.NewDecoder(r)

	var sj map[string]string
	err := d.Decode(&sj)
	if err != nil {
		return err
	}

	rev := make(map[string][]string)

	for name, canonical := range sj {
		rev[canonical] = append(rev[canonical], name)
	}

	if sg.surnames == nil {
		sg.surnames = make(map[string]*SurnameGroup, 0)
	}

	for canonical, names := range rev {
		g := &SurnameGroup{
			Surname: canonical,
		}
		sort.Strings(names)
		for _, n := range names {
			if n == canonical {
				continue
			}
			sg.surnames[n] = g
			g.Names = append(g.Names, n)
		}
		sg.surnames[canonical] = g

		slog.Warn("name group", "canonical", g.Surname, "names", strings.Join(g.Names, ", "))
	}

	return nil
}

func LoadSurnameGroups(filename string) (*SurnameGroups, error) {
	var sg SurnameGroups
	if filename == "" {
		return &sg, nil
	}

	slog.Info("reading surname groups", "filename", filename)
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &sg, nil
		}
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	if err := d.Decode(&sg); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	return &sg, nil
}

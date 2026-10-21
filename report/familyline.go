package report

import (
	"context"
	"fmt"
	"sort"

	"github.com/urfave/cli/v3"

	"github.com/iand/genster/gramps"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

var familylineCommand = &cli.Command{
	Name:   "familyline",
	Usage:  "Generate a report on a family line",
	Action: familyline,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gramps",
			Usage:       "Gramps xml file to read from",
			Required:    true,
			Destination: &familylineOpts.grampsFile,
		},
		&cli.StringFlag{
			Name:        "gramps-dbname",
			Usage:       "Name of the gramps database, used to keep IDs consistent between versions of the same database",
			Destination: &familylineOpts.grampsDatabaseName,
		},
		&cli.BoolFlag{
			Name:        "include-private",
			Usage:       "Include living people and people who died less than 20 years ago.",
			Value:       false,
			Destination: &familylineOpts.includePrivate,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Path to a KDL tree configuration file",
			Required:    true,
			Destination: &familylineOpts.treeConfig,
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "identifier of the key person, e.g. you",
			Required:    true,
			Destination: &familylineOpts.keyPersonID,
		},
	}, logging.Flags...),
}

var familylineOpts struct {
	grampsFile         string
	grampsDatabaseName string
	includePrivate     bool
	treeConfig         string
	keyPersonID        string
}

func familyline(ctx context.Context, cc *cli.Command) error {
	logging.Setup()

	l, err := gramps.NewLoader(familylineOpts.grampsFile, familylineOpts.grampsDatabaseName)
	if err != nil {
		return fmt.Errorf("load gramps: %w", err)
	}

	cfg, err := tree.ReadConfig(familylineOpts.treeConfig)
	if err != nil {
		return fmt.Errorf("read tree config: %w", err)
	}

	t, err := tree.LoadTree(cfg, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	keyPerson, ok := t.GetPerson(familylineOpts.keyPersonID)
	if !ok {
		keyPerson = t.FindPerson(l.Scope(), familylineOpts.keyPersonID)
	}
	t.SetKeyPerson(keyPerson)

	if err := t.Generate(!familylineOpts.includePrivate); err != nil {
		return fmt.Errorf("build tree: %w", err)
	}

	familyLines := tree.WalkFamilyLines(keyPerson)

	var flNames []string
	flMap := map[string]*model.FamilyLine{}
	for _, fl := range familyLines {
		baseName := fl.Lineage[0].PreferredFamilyName
		name := baseName
		idx := 2

		_, exists := flMap[name]
		for exists {
			name = fmt.Sprintf("%s %d", baseName, idx)
			idx++
			_, exists = flMap[name]
		}
		flNames = append(flNames, name)
		flMap[name] = fl
	}

	sort.Strings(flNames)

	for _, name := range flNames {
		fl := flMap[name]
		fmt.Printf("=== %s Family Line ===\n", name)
		for i, f := range fl.Families {
			if i == 0 {
				fmt.Printf("* %s (%s, %s)\n", f.PreferredFullName, fl.Lineage[i].PreferredUniqueName, fl.Lineage[i].GrampsID)
			}
			fmt.Printf("* %s\n", f.PreferredFullName)
		}
		for _, c := range fl.Families[len(fl.Families)-1].Children {
			if c.IsDirectAncestor() {
				fmt.Printf("-> %s (%s)\n", c.PreferredFullName, c.GrampsID)
				break
			}
		}
		fmt.Printf("\n")

	}

	return nil
}

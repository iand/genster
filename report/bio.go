package report

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/iand/genster/gramps"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/tree"
)

var bioOpts struct {
	grampsFile         string
	grampsDatabaseName string
}

var bioCommand = &cli.Command{
	Name:      "bio",
	Usage:     "Print generated biographies for the given people.",
	ArgsUsage: "[id...]",
	Action:    bioReport,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gramps",
			Usage:       "Gramps xml file to read from",
			Destination: &bioOpts.grampsFile,
		},
		&cli.StringFlag{
			Name:        "gramps-dbname",
			Usage:       "Name of the gramps database, used to keep IDs consistent between versions of the same database",
			Destination: &bioOpts.grampsDatabaseName,
		},
	}, logging.Flags...),
}

func bioReport(ctx context.Context, cc *cli.Command) error {
	logging.Setup()

	if bioOpts.grampsFile == "" {
		return fmt.Errorf("no gramps file specified")
	}

	ids := cc.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("no person identifiers specified")
	}

	l, err := gramps.NewLoader(bioOpts.grampsFile, bioOpts.grampsDatabaseName)
	if err != nil {
		return fmt.Errorf("load gramps: %w", err)
	}

	t, err := tree.LoadTree(&tree.Config{}, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	if err := t.Generate(false); err != nil {
		return fmt.Errorf("generate tree facts: %w", err)
	}

	for _, id := range ids {
		p, ok := t.GetPerson(id)
		if !ok {
			p = t.FindPerson(l.Scope(), model.NormalizeGrampsID(id))
		}
		if p.IsUnknown() {
			logging.Warn("person not found", "id", id)
			continue
		}

		fmt.Printf("# %s (%s)\n", p.PreferredFullName, id)
		fmt.Println(narrative.Bio(p))
		fmt.Println()
	}

	return nil
}

package report

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/iand/genster/debug"
	"github.com/iand/genster/gramps"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

var personOpts struct {
	grampsFile         string
	grampsDatabaseName string
}

var personCommand = &cli.Command{
	Name:      "person",
	Usage:     "Print a full report on the given people, including all families and children.",
	ArgsUsage: "[id...]",
	Action:    personReport,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gramps",
			Usage:       "Gramps xml file to read from",
			Destination: &personOpts.grampsFile,
		},
		&cli.StringFlag{
			Name:        "gramps-dbname",
			Usage:       "Name of the gramps database, used to keep IDs consistent between versions of the same database",
			Destination: &personOpts.grampsDatabaseName,
		},
	}, logging.Flags...),
}

func personReport(ctx context.Context, cc *cli.Command) error {
	logging.Setup()

	if personOpts.grampsFile == "" {
		return fmt.Errorf("no gramps file specified")
	}

	ids := cc.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("no person identifiers specified")
	}

	l, err := gramps.NewLoader(personOpts.grampsFile, personOpts.grampsDatabaseName)
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

		if err := debug.DumpPerson(p, os.Stdout); err != nil {
			return fmt.Errorf("dump person %s: %w", id, err)
		}
		fmt.Println()
	}

	return nil
}

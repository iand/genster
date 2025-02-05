package book

import (
	"fmt"
	"path/filepath"

	"github.com/iand/genster/gramps"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/urfave/cli/v2"
)

var bookopts struct {
	grampsFile         string
	grampsDatabaseName string
	treeID             string
	configDir          string
	outputDir          string
	outputFilename     string
	keyIndividual      string
	includePrivate     bool
	debug              bool
	relation           string
}

func checkFlags(cc *cli.Context) error {
	// switch bookopts.chartType {
	// case "descendant":
	// case "ancestor":
	// default:
	// 	return fmt.Errorf("unsupported chart type: %s", bookopts.chartType)
	// }
	return nil
}

var Command = &cli.Command{
	Name:   "book",
	Usage:  "Create a book from a gramps file",
	Action: bookCmd,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gramps",
			Usage:       "Gramps xml file to read from",
			Destination: &bookopts.grampsFile,
		},
		&cli.StringFlag{
			Name:        "gramps-dbname",
			Usage:       "Name of the gramps database, used to keep IDs consistent between versions of the same database",
			Destination: &bookopts.grampsDatabaseName,
		},
		&cli.StringFlag{
			Name:        "id",
			Usage:       "Identifier to give this tree (mainly to pick up configured annotations)",
			Destination: &bookopts.treeID,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where tree config can be found.",
			Destination: &bookopts.configDir,
		},
		&cli.StringFlag{
			Name:        "fname",
			Usage:       "output document filename",
			Destination: &bookopts.outputFilename,
		},
		&cli.StringFlag{
			Name:        "output",
			Usage:       "path to directory for temporary files",
			Destination: &bookopts.outputDir,
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "Identifier of the key individual",
			Destination: &bookopts.keyIndividual,
		},
		&cli.BoolFlag{
			Name:        "include-private",
			Usage:       "Include living people and people who died less than 20 years ago.",
			Value:       false,
			Destination: &bookopts.includePrivate,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "Include debug info as inline comments.",
			Value:       false,
			Destination: &bookopts.debug,
		},
		&cli.StringFlag{
			Name:        "relation",
			Usage:       "Only generate pages for people who are related to the key person. One of 'direct' (must be a direct ancestor), 'common' (must have a common ancestor) or 'any' (any relation). Ignored if no key person is specified.",
			Value:       "any",
			Destination: &bookopts.relation,
		},
	}, logging.Flags...),
}

func bookCmd(cc *cli.Context) error {
	if err := checkFlags(cc); err != nil {
		return err
	}

	logging.Setup()

	var l tree.Loader
	var err error

	if bookopts.grampsFile != "" {
		l, err = gramps.NewLoader(bookopts.grampsFile, bookopts.grampsDatabaseName)
		if err != nil {
			return fmt.Errorf("load gedcom: %w", err)
		}
	} else {
		return fmt.Errorf("no gramps file specified")
	}

	t, err := tree.LoadTree(bookopts.treeID, bookopts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	// Look for key individual, assume id is a genster id first
	keyIndividual, ok := t.GetPerson(bookopts.keyIndividual)
	if !ok {
		keyIndividual = t.FindPerson(l.Scope(), bookopts.keyIndividual)
	}

	t.SetKeyPerson(keyIndividual)

	redactLiving := !bookopts.includePrivate
	if err := t.Generate(redactLiving); err != nil {
		return err
	}

	b := NewBook(t)
	b.OutputDir = bookopts.outputDir
	b.IncludePrivate = bookopts.includePrivate
	b.IncludeDebugInfo = bookopts.debug

	inclusionFunc := func(p *model.Person) bool {
		if p.RelationToKeyPerson == nil {
			return false
		}

		// Always publish page if it's a direct ancestor
		if p.RelationToKeyPerson.IsDirectAncestor() {
			return true
		}

		// Don't ever publish page for someone who doesn't have a common ancestor
		// with the key person
		if !p.RelationToKeyPerson.HasCommonAncestor() {
			return false
		}

		return true
	}
	if err := b.BuildPublishSet(inclusionFunc); err != nil {
		return fmt.Errorf("build publish set: %w", err)
	}

	if err := b.BuildDocument(); err != nil {
		return fmt.Errorf("add pages: %w", err)
	}

	if err := b.WriteDocument(filepath.Join(bookopts.outputDir, bookopts.outputFilename)); err != nil {
		return fmt.Errorf("write document: %w", err)
	}

	// if genopts.rootDir != "" && genopts.mediaDir != "" {
	// 	if err := s.WriteDocument(genopts.rootDir, genopts.mediaDir); err != nil {
	// 		return fmt.Errorf("write pages: %w", err)
	// 	}
	// }

	return nil
}

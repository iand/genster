package site

import (
	"fmt"
	"strings"

	"github.com/iand/genster/debug"
	"github.com/iand/genster/gedcom"
	"github.com/iand/genster/gramps"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:   "gen",
	Usage:  "Generate a website from a gedcom file",
	Action: gen,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gedcom",
			Aliases:     []string{"g"},
			Usage:       "GEDCOM file to read from",
			Destination: &genopts.gedcomFile,
		},
		&cli.StringFlag{
			Name:        "gramps",
			Usage:       "Gramps xml file to read from",
			Destination: &genopts.grampsFile,
		},
		&cli.StringFlag{
			Name:        "id",
			Usage:       "Identifier to give this tree",
			Destination: &genopts.treeID,
		},
		&cli.StringFlag{
			Name:        "site",
			Aliases:     []string{"s"},
			Usage:       "Directory in which to write generated site",
			Destination: &genopts.rootDir,
		},
		&cli.StringFlag{
			Name:        "basepath",
			Aliases:     []string{"b"},
			Usage:       "Base URL path to use as a prefix to all links.",
			Value:       "/",
			Destination: &genopts.basePath,
		},
		&cli.StringFlag{
			Name:    "identity-map",
			Aliases: []string{"m"},
			Usage:   "Filename of identity mapping file",
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "Identifier of the key individual",
			Destination: &genopts.keyIndividual,
		},
		&cli.BoolFlag{
			Name:        "include-private",
			Usage:       "Include living people and people who died less than 20 years ago.",
			Value:       false,
			Destination: &genopts.includePrivate,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where config should be stored.",
			Destination: &genopts.configDir,
		},
		&cli.StringFlag{
			Name:        "inspect",
			Usage:       "Type and ID of an object to inspect. The internal data structure of the object will be printed to stdout. Use format '{object}/{id}' where object can be 'person', 'place' or 'source'.",
			Destination: &genopts.inspect,
		},
		&cli.BoolFlag{
			Name:        "wikitree",
			Usage:       "Generate pages that include wikitree markup for copy and paste.",
			Value:       false,
			Destination: &genopts.generateWikiTree,
		},
		&cli.BoolFlag{
			Name:        "hugo",
			Usage:       "Generate Hugo-specific markup and index pages.",
			Value:       true,
			Destination: &genopts.generateHugo,
		},
		&cli.StringFlag{
			Name:        "notes",
			Usage:       "Path to the folder where research notes are stored (in markdown format).",
			Destination: &genopts.notesDir,
		},
		&cli.StringFlag{
			Name:        "relation",
			Usage:       "Only include people who are related to the key person. One of 'direct' (must be a direct ancestor), 'common' (must have a common ancestor) or 'any' (any relation). Ignored if no key person is specified.",
			Value:       "any",
			Destination: &genopts.relation,
		},
	}, logging.Flags...),
}

var genopts struct {
	gedcomFile       string
	grampsFile       string
	treeID           string
	rootDir          string
	keyIndividual    string
	includePrivate   bool
	configDir        string
	notesDir         string
	basePath         string
	inspect          string
	generateWikiTree bool
	generateHugo     bool
	verbose          bool
	veryverbose      bool
	relation         string
}

func gen(cc *cli.Context) error {
	logging.Setup()

	var l tree.Loader
	var err error

	if genopts.gedcomFile != "" {
		l, err = gedcom.NewLoader(genopts.gedcomFile)
		if err != nil {
			return fmt.Errorf("load gedcom: %w", err)
		}
	} else if genopts.grampsFile != "" {
		l, err = gramps.NewLoader(genopts.grampsFile)
		if err != nil {
			return fmt.Errorf("load gedcom: %w", err)
		}
	} else {
		return fmt.Errorf("no gedcom or gramps file specified")
	}

	t, err := tree.LoadTree(genopts.treeID, genopts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	if genopts.notesDir != "" {
		nds, err := LoadNotes(genopts.notesDir)
		if err != nil {
			return fmt.Errorf("load notes: %w", err)
		}

		for _, nd := range nds {
			logging.Debug("found note", "filename", nd.Filename, "type", nd.Type)
			if nd.Type == "note" && nd.Person != "" {
				p, ok := t.GetPerson(nd.Person)
				if !ok {
					// logging.Warn("found research note for unknown person", "filename", nd.Filename, "person", nd.Person)
					continue
				}
				note := &model.Note{
					Title:         nd.Title,
					Author:        nd.Author,
					Date:          nd.Date,
					Markdown:      nd.Markdown,
					PrimaryPerson: p,
				}

				logging.Debug("found research note for person", "filename", nd.Filename, "id", nd.Person)
				p.ResearchNotes = append(p.ResearchNotes, note)

				for _, m := range nd.Mentions {
					mp, ok := t.GetPerson(m)
					if !ok {
						// logging.Warn("found research note for unknown person", "filename", nd.Filename, "person", nd.Person)
						continue
					}
					logging.Debug("found research note that mentions person", "filename", nd.Filename, "id", m)
					mp.ResearchNotes = append(mp.ResearchNotes, note)
				}
			}
		}

	}

	s := NewSite(genopts.basePath, t)
	s.IncludePrivate = genopts.includePrivate
	s.GenerateWikiTree = genopts.generateWikiTree
	s.GenerateHugo = genopts.generateHugo

	// Look for key individual, assume id is a genster id first
	keyIndividual, ok := t.GetPerson(genopts.keyIndividual)
	if !ok {
		keyIndividual = t.FindPerson(l.Scope(), genopts.keyIndividual)
	}

	t.SetKeyPerson(keyIndividual)

	if err := s.Generate(); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	skipPage := func(p *model.Person) (bool, error) {
		s.SkippedPersonPages[p.ID] = true
		return true, nil
	}

	switch genopts.relation {
	case "direct":
		logging.Info("only generating pages for direct ancestors")
		s.Tree.ApplyPeopleMatching(model.PersonIsNotDirectAncestor(), skipPage)
	case "common":
		logging.Info("only generating pages for people with common ancestors")
		s.Tree.ApplyPeopleMatching(model.PersonDoesNotHaveCommonAncestor(), skipPage)
	case "any":
		break
	default:
		return fmt.Errorf("unsupported relation option: %s", genopts.relation)
	}

	if genopts.inspect != "" {
		if strings.HasPrefix(genopts.inspect, "person/") {
			id := genopts.inspect[7:]
			p, ok := s.Tree.GetPerson(id)
			if !ok {
				return fmt.Errorf("no person found with id %s", id)
			}
			return debug.DumpPerson(p)
		} else {
			return fmt.Errorf("unrecognised object to inspect: %s", genopts.inspect)
		}
	}

	if genopts.rootDir != "" {
		if err := s.WritePages(genopts.rootDir); err != nil {
			return fmt.Errorf("write pages: %w", err)
		}
	}

	return nil
}

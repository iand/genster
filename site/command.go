package site

import (
	"fmt"
	"strings"

	"github.com/iand/genster/gedcom"
	"github.com/iand/genster/logging"
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
			Name:        "site",
			Aliases:     []string{"s"},
			Usage:       "Path to root of website",
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
	}, logging.Flags...),
}

var genopts struct {
	gedcomFile       string
	rootDir          string
	keyIndividual    string
	includePrivate   bool
	configDir        string
	basePath         string
	inspect          string
	generateWikiTree bool
	verbose          bool
	veryverbose      bool
}

func gen(cc *cli.Context) error {
	logging.Setup()

	if genopts.gedcomFile == "" {
		return fmt.Errorf("no gedcom file specified")
	}

	l, err := gedcom.NewLoader(genopts.gedcomFile)
	if err != nil {
		return fmt.Errorf("load gedcom: %w", err)
	}

	t, err := tree.LoadTree(genopts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	s := NewSite(genopts.basePath, t)
	s.IncludePrivate = genopts.includePrivate
	s.GenerateWikiTree = genopts.generateWikiTree

	// Look for key individual, assume id is a genster id first
	keyIndividual, ok := t.GetPerson(genopts.keyIndividual)
	if !ok {
		keyIndividual = t.FindPerson(l.ScopeName, genopts.keyIndividual)
	}

	t.SetKeyPerson(keyIndividual)

	if err := s.Generate(); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	if genopts.inspect != "" {
		if strings.HasPrefix(genopts.inspect, "person/") {
			id := genopts.inspect[7:]
			p, ok := s.Tree.GetPerson(id)
			if !ok {
				return fmt.Errorf("no person found with id %s", id)
			}
			return s.InspectPerson(p)
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

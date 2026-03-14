package site

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
			Name:        "gramps-dbname",
			Usage:       "Name of the gramps database, used to keep IDs consistent between versions of the same database",
			Destination: &genopts.grampsDatabaseName,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Path to a KDL tree configuration file",
			Required:    true,
			Destination: &genopts.treeConfig,
		},
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
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
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "Include debug info as inline comments.",
			Value:       false,
			Destination: &genopts.debug,
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
		&cli.StringFlag{
			Name:        "relation",
			Usage:       "Only generate pages for people who are related to the key person. One of 'direct' (must be a direct ancestor), 'common' (must have a common ancestor) or 'any' (any relation). Ignored if no key person is specified.",
			Value:       "any",
			Destination: &genopts.relation,
		},
		&cli.BoolFlag{
			Name:        "experiment-families",
			Usage:       "Enable experimental family pages.",
			Value:       true,
			Destination: &genopts.experimentFamilies,
		},
	}, logging.Flags...),
}

var genopts struct {
	gedcomFile         string
	grampsFile         string
	grampsDatabaseName string
	rootDir            string
	keyIndividual      string
	includePrivate     bool

	basePath           string
	inspect            string
	generateWikiTree   bool
	treeConfig         string
	verbose            bool
	veryverbose        bool
	relation           string
	debug              bool
	experimentFamilies bool
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
		l, err = gramps.NewLoader(genopts.grampsFile, genopts.grampsDatabaseName)
		if err != nil {
			return fmt.Errorf("load gedcom: %w", err)
		}
	} else {
		return fmt.Errorf("no gedcom or gramps file specified")
	}

	treeCfg, err := tree.ReadConfig(genopts.treeConfig)
	if err != nil {
		return fmt.Errorf("read tree config: %w", err)
	}

	t, err := tree.LoadTree(treeCfg, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	pageMap, err := walkDiaryPages("/home/iand/web/cozy.ac/history/content/diary", "/diary/")
	if err != nil {
		return fmt.Errorf("walk diary pages: %w", err)
	}

	for _, p := range t.People {
		if pages, ok := pageMap[p.GrampsID]; ok {
			p.DiaryLinks = append(p.DiaryLinks, pages...)
		}
		if pages, ok := pageMap[p.Slug]; ok {
			p.DiaryLinks = append(p.DiaryLinks, pages...)
		}
	}

	// for id, pages := range pageMap {
	// 	fmt.Printf("%s -->%q\n", id, pages)
	// }

	s := NewSite(genopts.basePath, t)
	s.IncludePrivate = genopts.includePrivate
	s.IncludeDebugInfo = genopts.debug
	s.ExperimentFamilies = genopts.experimentFamilies
	s.GenerateWikiTree = genopts.generateWikiTree

	// Look for key individual, assume id is a genster id first
	keyIndividual, ok := t.GetPerson(genopts.keyIndividual)
	if !ok {
		keyIndividual = t.FindPerson(l.Scope(), genopts.keyIndividual)
	}

	t.SetKeyPerson(keyIndividual)

	if err := s.Generate(); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	inclusionFunc := func(*model.Person) bool { return true }

	switch genopts.relation {
	case "direct":
		logging.Info("only generating pages for direct ancestors and people with common ancestors tagged as featured or publish")
		inclusionFunc = func(p *model.Person) bool {
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

			// If person tagged as publish or featured then include
			if p.Publish || p.Featured {
				return true
			}

			// If person is child in family tagged as publish then include
			if p.ParentFamily != nil && p.ParentFamily.PublishChildren {
				return true
			}

			return false
		}
	case "common":
		logging.Info("only generating pages for people with common ancestors those tagged as featured or publish")
		inclusionFunc = func(p *model.Person) bool {
			if p.RelationToKeyPerson == nil {
				return false
			}
			return p.RelationToKeyPerson.HasCommonAncestor()
		}
	case "any":
		break
	default:
		return fmt.Errorf("unsupported relation option: %s", genopts.relation)
	}

	s.BuildPublishSet(inclusionFunc)

	if genopts.inspect != "" {
		if strings.HasPrefix(genopts.inspect, "person/") {
			id := genopts.inspect[7:]
			p, ok := s.Tree.GetPerson(id)
			if !ok {
				return fmt.Errorf("no person found with id %s", id)
			}
			return debug.DumpPerson(p, os.Stdout)
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

var (
	reAliasLink = regexp.MustCompile(`\(/r/(.+?)\)`)
	reAliasTag  = regexp.MustCompile(` - /r/(.+?)\n`)
	reISODate   = regexp.MustCompile(`^(\d\d\d\d)-(\d\d)-(\d\d)$`)
)

// walkDiaryPages walks the directory tree for the research diary looking for
// links to person pages. It returns a map of person IDs to a slice of relative
// diary URLs
func walkDiaryPages(dir string, urlPrefix string) (map[string][]model.Link, error) {
	result := make(map[string][]model.Link)
	months := map[string]string{
		"01": "Jan",
		"02": "Feb",
		"03": "Mar",
		"04": "Apr",
		"05": "May",
		"06": "Jun",
		"07": "Jul",
		"08": "Aug",
		"09": "Sep",
		"10": "Oct",
		"11": "Nov",
		"12": "Dec",
	}

	isoToHuman := func(s string) (string, bool) {
		m := reISODate.FindStringSubmatch(s)
		if m == nil || len(m) != 4 {
			return "", false
		}

		return m[3] + " " + months[m[2]] + " " + m[1], true
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		// Determine relative URL
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) < 2 {
			return nil // unexpected structure
		}

		var link model.Link
		if d.Name() == "index.md" {
			var ok bool
			link.Title, ok = isoToHuman(parts[1])
			if !ok {
				return nil
			}
			link.URL = urlPrefix + filepath.ToSlash(filepath.Join(parts[0], parts[1])+"/")
		} else {
			var ok bool
			link.Title, ok = isoToHuman(strings.TrimSuffix(parts[1], ".md"))
			if !ok {
				return nil
			}
			link.URL = urlPrefix + filepath.ToSlash(strings.TrimSuffix(relPath, ".md")+"/")
		}

		// Read and process file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Find links like ./r/I0554 or /r/william-hinksman
		matches := reAliasLink.FindAllSubmatch(content, -1)
		for _, match := range matches {
			id := string(match[1])
			result[id] = append(result[id], link)
		}

		matches = reAliasTag.FindAllSubmatch(content, -1)
		for _, match := range matches {
			id := string(match[1])
			result[id] = append(result[id], link)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

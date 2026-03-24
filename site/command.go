package site

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
		&cli.StringFlag{
			Name:        "content",
			Usage:       "Path to the content directory whose diary, stories, and questions sub-folders are walked for person references.",
			Destination: &genopts.contentDir,
		},
		&cli.BoolFlag{
			Name:        "include-drafts",
			Usage:       "Include draft content pages when walking for person references.",
			Value:       false,
			Destination: &genopts.includeDrafts,
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
	contentDir         string
	includeDrafts      bool
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

	if genopts.contentDir != "" {
		pageMap, err := walkContentPages(genopts.contentDir, genopts.includeDrafts)
		if err != nil {
			return fmt.Errorf("walk content pages: %w", err)
		}
		for _, p := range t.People {
			if pages, ok := pageMap[p.GrampsID]; ok {
				p.Links = append(p.Links, pages...)
			}
			if pages, ok := pageMap[p.Slug]; ok {
				p.Links = append(p.Links, pages...)
			}
		}
	}

	s := NewSite(genopts.basePath, t)
	s.IncludePrivate = genopts.includePrivate
	s.IncludeDebugInfo = genopts.debug
	s.ExperimentFamilies = genopts.experimentFamilies
	s.GenerateWikiTree = genopts.generateWikiTree
	s.MapTilerAPIKey = os.Getenv("MAPTILER_API_KEY")

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
	reFMTitle   = regexp.MustCompile(`(?m)^title:\s*["']?(.+?)["']?\s*$`)
	reFMDraft   = regexp.MustCompile(`(?m)^draft:\s*["']?(\w+)["']?\s*$`)
)

var isoMonths = map[string]string{
	"01": "Jan", "02": "Feb", "03": "Mar", "04": "Apr",
	"05": "May", "06": "Jun", "07": "Jul", "08": "Aug",
	"09": "Sep", "10": "Oct", "11": "Nov", "12": "Dec",
}

func isoToHuman(s string) (string, bool) {
	m := reISODate.FindStringSubmatch(s)
	if m == nil {
		return "", false
	}
	return m[3] + " " + isoMonths[m[2]] + " " + m[1], true
}

// fmDraft reports whether the content's front-matter has draft set to a
// truthy value (true, "true", "yes", or "1").
func fmDraft(content []byte) bool {
	m := reFMDraft.FindSubmatch(content)
	if m == nil {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(string(m[1])))
	return v == "true" || v == "yes" || v == "1"
}

// fmTitle extracts the title value from a YAML front-matter block in content.
// Returns an empty string when no title can be found.
func fmTitle(content []byte) string {
	m := reFMTitle.FindSubmatch(content)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(string(m[1]))
}

// resolvePageTitle returns the display title for a content page and whether
// the page should be included in the results. stem is the filename stem
// (without .md extension) or directory name; content is the raw file bytes
// used for front-matter extraction when needed.
//
// When isoTitleOnly is true only ISO-date stems (YYYY-MM-DD) are accepted and
// the title is always formatted as "DD MMM YYYY". Non-ISO stems return
// ("", false), signalling that the page should be skipped.
//
// When isoTitleOnly is false an ISO-date stem is still formatted as
// "DD MMM YYYY" when present, otherwise the title falls back to the
// front-matter title field and then to the raw stem.
func resolvePageTitle(stem string, content []byte, isoTitleOnly bool) (string, bool) {
	if t, ok := isoToHuman(stem); ok {
		return t, true
	}
	if isoTitleOnly {
		return "", false
	}
	if t := fmTitle(content); t != "" {
		return t, true
	}
	return stem, true
}

type contentSection struct {
	subdir       string
	urlBase      string
	category     string
	isoTitleOnly bool // when true, title is always derived from ISO date (diary)
}

// walkContentPages walks the diary, stories, and questions subdirectories
// under contentDir looking for links to person pages. It returns a map of
// person IDs/slugs to a slice of links, each tagged with its section category.
// Diary links always use DD MMM YYYY titles and are sorted reverse-chronologically.
// Pages with draft:true in their front-matter are skipped unless includeDrafts is true.
func walkContentPages(contentDir string, includeDrafts bool) (map[string][]model.Link, error) {
	sections := []contentSection{
		{"diary", "/diary/", "diary", true},
		{"stories", "/stories/", "story", false},
		{"questions", "/questions/", "question", false},
	}

	result := make(map[string][]model.Link)
	for _, sec := range sections {
		dir := filepath.Join(contentDir, sec.subdir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		if err := walkSectionPages(dir, sec.urlBase, sec.category, sec.isoTitleOnly, includeDrafts, result); err != nil {
			return nil, fmt.Errorf("walk %s: %w", sec.subdir, err)
		}
		if sec.isoTitleOnly {
			// Sort diary links newest-first by URL (ISO dates sort lexicographically).
			for id, links := range result {
				slices.SortFunc(links, func(a, b model.Link) int {
					return strings.Compare(b.URL, a.URL)
				})
				result[id] = links
			}
		}
	}

	// Deduplicate: a person may be referenced by both GrampsID and slug, or
	// the same alias may appear more than once in a file. Keep the first
	// occurrence of each URL within each person's slice.
	for id, links := range result {
		seen := make(map[string]bool, len(links))
		deduped := links[:0]
		for _, l := range links {
			if !seen[l.URL] {
				seen[l.URL] = true
				deduped = append(deduped, l)
			}
		}
		result[id] = deduped
	}

	return result, nil
}

// walkSectionPages walks a single content section directory (e.g. diary/,
// stories/, questions/), finds .md files that reference person aliases, and
// records links for those people in result. When isoTitleOnly is true the
// title is always derived from the ISO date in the filename stem (diary
// behaviour); entries whose stem is not an ISO date are skipped. When false
// the title comes from front-matter, falling back to the stem.
func walkSectionPages(dir, urlBase, category string, isoTitleOnly, includeDrafts bool, result map[string][]model.Link) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if !includeDrafts && fmDraft(content) {
			return nil
		}

		// Skip section index files (_index.md / index.md at the top level of
		// this section dir — they are not content pages).
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(relPath)
		stem := strings.TrimSuffix(d.Name(), ".md")

		// Skip section-level index files (index.md / _index.md at the root of
		// this section dir). Subdirectory index files (e.g.
		// george-henry-chambers/index.md) are valid content pages.
		if (stem == "index" || stem == "_index") && filepath.Dir(relPath) == "." {
			return nil
		}

		// Derive URL and title.
		var link model.Link
		link.Category = category

		if stem == "index" || stem == "_index" {
			// Directory-style page: URL is the parent dir path.
			dirStem := filepath.Base(filepath.Dir(relPath))
			link.URL = urlBase + filepath.ToSlash(filepath.Dir(relPath)) + "/"
			var ok bool
			if link.Title, ok = resolvePageTitle(dirStem, content, isoTitleOnly); !ok {
				return nil
			}
		} else {
			// Leaf file: URL is derived from the stem.
			link.URL = urlBase + filepath.ToSlash(strings.TrimSuffix(relSlash, ".md")) + "/"
			var ok bool
			if link.Title, ok = resolvePageTitle(stem, content, isoTitleOnly); !ok {
				return nil
			}
		}

		// Find person references: markdown links like (/r/I0554) and
		// YAML front-matter list entries like "  - /r/william-hinksman".
		for _, match := range reAliasLink.FindAllSubmatch(content, -1) {
			id := string(match[1])
			result[id] = append(result[id], link)
		}
		for _, match := range reAliasTag.FindAllSubmatch(content, -1) {
			id := string(match[1])
			result[id] = append(result[id], link)
		}

		return nil
	})
}

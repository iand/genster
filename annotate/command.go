package annotate

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/iand/genster/gedcom"
	"github.com/iand/genster/gramps"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/tree"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:   "annotate",
	Usage:  "Replace citation references in markdown files with formatted citation links",
	Action: annotate,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gedcom",
			Aliases:     []string{"g"},
			Usage:       "GEDCOM file to read from",
			Destination: &opts.gedcomFile,
		},
		&cli.StringFlag{
			Name:        "gramps",
			Usage:       "Gramps xml file to read from",
			Destination: &opts.grampsFile,
		},
		&cli.StringFlag{
			Name:        "gramps-dbname",
			Usage:       "Name of the gramps database, used to keep IDs consistent between versions of the same database",
			Destination: &opts.grampsDatabaseName,
		},
		&cli.StringFlag{
			Name:        "id",
			Usage:       "Identifier to give this tree",
			Destination: &opts.treeID,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where config should be stored",
			Destination: &opts.configDir,
		},
		&cli.StringFlag{
			Name:        "dir",
			Aliases:     []string{"d"},
			Usage:       "Path to the folder of markdown files to annotate",
			Destination: &opts.markdownDir,
			Required:    true,
		},
		&cli.BoolFlag{
			Name:        "dry-run",
			Aliases:     []string{"n"},
			Usage:       "List changes without modifying files",
			Destination: &opts.dryRun,
		},
		&cli.BoolFlag{
			Name:        "undo",
			Usage:       "Restore original footnote references, removing generated citations",
			Destination: &opts.undo,
		},
		&cli.StringFlag{
			Name:        "basepath",
			Aliases:     []string{"b"},
			Usage:       "Base URL path for the tree site, used to generate citation links",
			Value:       "/",
			Destination: &opts.basePath,
		},
	}, logging.Flags...),
}

var opts struct {
	gedcomFile         string
	grampsFile         string
	grampsDatabaseName string
	treeID             string
	configDir          string
	markdownDir        string
	basePath           string
	dryRun             bool
	undo               bool
}

var (
	// footnoteDefRe matches a footnote definition line: [^label]: text
	footnoteDefRe = regexp.MustCompile(`(?m)^\[\^([^\]]+)\]:\s*(.+)\n?`)

	// footnoteRefRe matches any inline footnote reference: [^label]
	footnoteRefRe = regexp.MustCompile(`\[\^([^\]]+)\]`)

	// grampsIDRe identifies gramps citation IDs (C followed by 3+ digits).
	grampsIDRe = regexp.MustCompile(`^C\d{3,}$`)

	// prevCiteRe matches a previously generated inline citation with its
	// comment markers. Group 1 is the label, group 2 (optional) is the
	// footnote definition text for native footnotes.
	prevCiteRe = regexp.MustCompile(`<!-- cite \[\^([^\]]+)\](?:: (.+?))? -->.*?<!-- /cite -->`)

	// prevSectionRe matches a previously generated citations section.
	prevSectionRe = regexp.MustCompile(`(?s)\n?<!-- begin citations -->.*?<!-- end citations -->\n?`)
)

func annotate(cc *cli.Context) error {
	logging.Setup()

	if opts.undo {
		return undoAnnotations()
	}

	var l tree.Loader
	var err error

	if opts.gedcomFile != "" {
		l, err = gedcom.NewLoader(opts.gedcomFile)
		if err != nil {
			return fmt.Errorf("load gedcom: %w", err)
		}
	} else if opts.grampsFile != "" {
		l, err = gramps.NewLoader(opts.grampsFile, opts.grampsDatabaseName)
		if err != nil {
			return fmt.Errorf("load gramps: %w", err)
		}
	} else {
		return fmt.Errorf("no gedcom or gramps file specified")
	}

	t, err := tree.LoadTree(opts.treeID, opts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	// Build a lookup from GrampsID to citation.
	grampsIDToCitation := make(map[string]*model.GeneralCitation)
	for _, cit := range t.Citations {
		if cit.GrampsID != "" {
			grampsIDToCitation[cit.GrampsID] = cit
		}
	}

	lb := &citationLinkBuilder{
		citationLinkPattern: path.Join(opts.basePath, "trees", opts.treeID, "citation/%s/"),
	}

	// Find all markdown files in the directory tree.
	timestamp := time.Now().Unix()
	filesProcessed := 0
	err = filepath.WalkDir(opts.markdownDir, func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		changed, err := processFile(fpath, grampsIDToCitation, lb, opts.dryRun, timestamp)
		if err != nil {
			return fmt.Errorf("process %s: %w", fpath, err)
		}
		if changed {
			filesProcessed++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("%d file(s) processed\n", filesProcessed)
	return nil
}

func undoAnnotations() error {
	timestamp := time.Now().Unix()
	filesProcessed := 0
	err := filepath.WalkDir(opts.markdownDir, func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		changed, err := undoFile(fpath, opts.dryRun, timestamp)
		if err != nil {
			return fmt.Errorf("undo %s: %w", fpath, err)
		}
		if changed {
			filesProcessed++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("%d file(s) restored\n", filesProcessed)
	return nil
}

func undoFile(fpath string, dryRun bool, timestamp int64) (bool, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return false, fmt.Errorf("read file: %w", err)
	}

	content := string(data)
	changed := false

	// Strip the generated citation section.
	result := prevSectionRe.ReplaceAllString(content, "")
	if result != content {
		changed = true
	}

	// Restore inline citations to their original [^label] form and
	// collect native footnote definitions.
	var nativeDefs []string
	restored := prevCiteRe.ReplaceAllStringFunc(result, func(match string) string {
		changed = true
		submatch := prevCiteRe.FindStringSubmatch(match)
		if len(submatch) >= 2 {
			label := submatch[1]
			if len(submatch) >= 3 && submatch[2] != "" {
				nativeDefs = append(nativeDefs, fmt.Sprintf("[^%s]: %s", label, submatch[2]))
			}
			return "[^" + label + "]"
		}
		return match
	})

	if !changed {
		return false, nil
	}

	// Re-append native footnote definitions at the end.
	if len(nativeDefs) > 0 {
		restored = strings.TrimRight(restored, "\n") + "\n\n" + strings.Join(nativeDefs, "\n") + "\n"
	}

	fmt.Println(fpath)

	if dryRun {
		return true, nil
	}

	info, err := os.Stat(fpath)
	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}

	bakPath := fmt.Sprintf("%s.bak-%d", fpath, timestamp)
	if err := os.WriteFile(bakPath, data, info.Mode()); err != nil {
		return false, fmt.Errorf("write backup file: %w", err)
	}

	if err := os.WriteFile(fpath, []byte(restored), info.Mode()); err != nil {
		return false, fmt.Errorf("write file: %w", err)
	}

	return true, nil
}

func processFile(fpath string, grampsIDToCitation map[string]*model.GeneralCitation, lb md.LinkBuilder, dryRun bool, timestamp int64) (bool, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return false, fmt.Errorf("read file: %w", err)
	}

	content := string(data)

	// Strip any previously generated citation section.
	content = prevSectionRe.ReplaceAllString(content, "")

	// Restore previously processed inline citations to their original
	// [^label] form and recover native footnote definitions from the
	// embedded comments.
	footnoteDefs := make(map[string]string)
	content = prevCiteRe.ReplaceAllStringFunc(content, func(match string) string {
		submatch := prevCiteRe.FindStringSubmatch(match)
		if len(submatch) >= 2 {
			label := submatch[1]
			if len(submatch) >= 3 && submatch[2] != "" {
				footnoteDefs[label] = submatch[2]
			}
			return "[^" + label + "]"
		}
		return match
	})

	// Parse any remaining footnote definitions and remove them from content.
	cleaned := footnoteDefRe.ReplaceAllStringFunc(content, func(match string) string {
		submatch := footnoteDefRe.FindStringSubmatch(match)
		if len(submatch) >= 3 {
			footnoteDefs[submatch[1]] = strings.TrimSpace(submatch[2])
		}
		return ""
	})

	// Find all footnote references in the cleaned content.
	matches := footnoteRefRe.FindAllStringIndex(cleaned, -1)
	if len(matches) == 0 {
		return false, nil
	}

	fmt.Println(fpath)

	var enc md.Content
	enc.SetLinkBuilder(lb)

	// Process replacements from end to start so indices remain valid.
	result := cleaned
	changeCount := 0
	for i := len(matches) - 1; i >= 0; i-- {
		start := matches[i][0]
		end := matches[i][1]
		fullMatch := cleaned[start:end]

		submatch := footnoteRefRe.FindStringSubmatch(fullMatch)
		if len(submatch) < 2 {
			continue
		}
		label := submatch[1]

		var cit *model.GeneralCitation
		var comment string
		if grampsIDRe.MatchString(label) {
			c, ok := grampsIDToCitation[label]
			if !ok {
				return false, fmt.Errorf("citation %s not found", label)
			}
			cit = c
			comment = fmt.Sprintf("<!-- cite [^%s] -->", label)
		} else {
			defText, ok := footnoteDefs[label]
			if !ok {
				return false, fmt.Errorf("footnote definition [^%s] not found", label)
			}
			cit = &model.GeneralCitation{
				ID:     "footnote-" + label,
				Detail: defText,
			}
			comment = fmt.Sprintf("<!-- cite [^%s]: %s -->", label, defText)
		}

		title := cit.SourceTitle()
		if title == "" {
			title = cit.Detail
		}
		fmt.Printf("  %s: %s\n", label, title)

		citHTML := string(enc.EncodeWithCitations("", []*model.GeneralCitation{cit}))
		replacement := comment + citHTML + "<!-- /cite -->"
		result = result[:start] + replacement + result[end:]
		changeCount++
	}

	if changeCount == 0 {
		return false, nil
	}

	// Append the citations and notes section to the end of the document.
	var citationSection strings.Builder
	enc.Citations.WriteTo(&citationSection)
	if citationSection.Len() > 0 {
		result += "\n<!-- begin citations -->\n" + citationSection.String() + "<!-- end citations -->\n"
	}

	if dryRun {
		return true, nil
	}

	info, err := os.Stat(fpath)
	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}

	bakPath := fmt.Sprintf("%s.bak-%d", fpath, timestamp)
	if err := os.WriteFile(bakPath, data, info.Mode()); err != nil {
		return false, fmt.Errorf("write backup file: %w", err)
	}

	if err := os.WriteFile(fpath, []byte(result), info.Mode()); err != nil {
		return false, fmt.Errorf("write file: %w", err)
	}

	return true, nil
}

// citationLinkBuilder generates links for citations in the tree site.
type citationLinkBuilder struct {
	citationLinkPattern string
}

func (b *citationLinkBuilder) LinkFor(v any) string {
	switch vt := v.(type) {
	case *model.GeneralCitation:
		if vt.ID == "" {
			return ""
		}
		return fmt.Sprintf(b.citationLinkPattern, vt.ID)
	default:
		return ""
	}
}

/*
This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying UNLICENSE file.
*/

package chart

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	gegedcom "github.com/iand/genster/gedcom"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
)

func checkFlags(cc *cli.Context) error {
	switch chartopts.chartType {
	case "descendant":
	case "ancestor":
	default:
		return fmt.Errorf("unsupported chart type: %s", chartopts.chartType)
	}
	return nil
}

var chartopts struct {
	gedcomFile    string
	chartType     string
	treeID        string
	configDir     string
	keyPersonID   string
	startPersonID string
	title         string
	fontScale     float64

	outputFilename string
	outputFormat   string
	descendantId   string
	generations    int
	detail         int
	directOnly     bool
}

var Command = &cli.Command{
	Name:   "chart",
	Usage:  "Create a family tree chart.",
	Action: chartCmd,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gedcom",
			Aliases:     []string{"g", "input"},
			Usage:       "GEDCOM file to read from",
			Destination: &chartopts.gedcomFile,
		},
		&cli.StringFlag{
			Name:        "type",
			Aliases:     []string{"t"},
			Usage:       "Type of chart to produce, descendant or ancestor",
			Destination: &chartopts.chartType,
		},
		&cli.StringFlag{
			Name:        "id",
			Usage:       "Identifier to give this tree (mainly to pick up configured annotations)",
			Destination: &chartopts.treeID,
		},
		&cli.StringFlag{
			Name:        "output",
			Usage:       "output image filename",
			Destination: &chartopts.outputFilename,
		},
		&cli.StringFlag{
			Name:        "person",
			Usage:       "identifier of person to build tree from",
			Destination: &chartopts.startPersonID,
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "Identifier of the key individual",
			Destination: &chartopts.keyPersonID,
		},
		&cli.StringFlag{
			Name:        "title",
			Usage:       "Title to add to chart",
			Destination: &chartopts.title,
		},
		&cli.IntFlag{
			Name:        "gen",
			Usage:       "number of generations to draw",
			Value:       2,
			Destination: &chartopts.generations,
		},
		&cli.IntFlag{
			Name:        "detail",
			Usage:       "level of detail to include with each person (0:none,1:years,2:dates,3:full)",
			Value:       3,
			Destination: &chartopts.detail,
		},
		&cli.BoolFlag{
			Name:        "direct",
			Usage:       "only show children of direct ancestors (for descendant charts)",
			Value:       false,
			Destination: &chartopts.directOnly,
		},
		&cli.Float64Flag{
			Name:        "font-scale",
			Usage:       "scale all fonts by this factor",
			Value:       1.0,
			Destination: &chartopts.fontScale,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where tree config can be found.",
			Destination: &chartopts.configDir,
		},
	}, logging.Flags...),
}

func chartCmd(cc *cli.Context) error {
	if err := checkFlags(cc); err != nil {
		return err
	}

	logging.Setup()

	l, err := gegedcom.NewLoader(chartopts.gedcomFile)
	if err != nil {
		return fmt.Errorf("load gedcom: %w", err)
	}

	t, err := tree.LoadTree(chartopts.treeID, chartopts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	// Look for key person, if any. This is the person who is used to determine
	// whether a person in the tree is a direct ancestor
	// assume id is a genster id first
	var keyPerson *model.Person
	if chartopts.keyPersonID != "" {
		var ok bool
		keyPerson, ok = t.GetPerson(chartopts.keyPersonID)
		if !ok {
			keyPerson = t.FindPerson(l.ScopeName, chartopts.keyPersonID)
		}
		if keyPerson.IsUnknown() {
			return fmt.Errorf("key person not found")
		}
		t.SetKeyPerson(keyPerson)
	}

	if err := t.Generate(false); err != nil {
		return fmt.Errorf("generate tree facts: %w", err)
	}

	// Find the root of the tree, i.e. the earliest ancester we want to show on the tree
	// assume id is a genster id first
	startPerson, ok := t.GetPerson(chartopts.startPersonID)
	if !ok {
		// not a genster id, so look for a gedcom id
		startPerson = t.FindPerson(l.ScopeName, chartopts.startPersonID)
	}
	if startPerson.IsUnknown() {
		return fmt.Errorf("person with id %s not found", chartopts.startPersonID)
	}

	var lay gtree.Layout
	switch chartopts.chartType {
	case "descendant":
		ch, err := BuildDescendantChart(t, startPerson, chartopts.detail, chartopts.generations, chartopts.directOnly)
		if err != nil {
			return fmt.Errorf("build descendant chart: %w", err)
		}

		ch.Title = chartopts.title
		if ch.Title == "" {
			ch.Title = "Descendants of " + startPerson.PreferredUniqueName
		}
		ch.Notes = []string{}

		ch.Notes = append(ch.Notes, time.Now().Format("Generated _2 January 2006"))
		if !keyPerson.IsUnknown() {
			ch.Notes = append(ch.Notes, "(★ denotes a direct ancestor of "+keyPerson.PreferredFamiliarFullName+")")
		}

		opts := gtree.DefaultLayoutOptions()
		lay = ch.Layout(opts)
	case "ancestor":
		ch, err := BuildAncestorChart(t, startPerson, chartopts.detail, chartopts.generations)
		if err != nil {
			return fmt.Errorf("build ancestor chart: %w", err)
		}

		ch.Title = chartopts.title
		if ch.Title == "" {
			ch.Title = "Ancestors of " + startPerson.PreferredUniqueName
		}
		ch.Notes = []string{}

		ch.Notes = append(ch.Notes, time.Now().Format("Generated _2 January 2006"))

		opts := gtree.DefaultAncestorLayoutOptions()

		opts.TitleFontSize = scaleFont(opts.TitleFontSize, chartopts.fontScale)
		opts.TitleLineHeight = scaleFont(opts.TitleLineHeight, chartopts.fontScale)
		opts.NoteFontSize = scaleFont(opts.NoteFontSize, chartopts.fontScale)
		opts.NoteLineHeight = scaleFont(opts.NoteLineHeight, chartopts.fontScale)
		opts.HeadingFontSize = scaleFont(opts.HeadingFontSize, chartopts.fontScale)
		opts.HeadingLineHeight = scaleFont(opts.HeadingLineHeight, chartopts.fontScale)
		opts.DetailFontSize = scaleFont(opts.DetailFontSize, chartopts.fontScale)
		opts.DetailLineHeight = scaleFont(opts.DetailLineHeight, chartopts.fontScale)

		lay = ch.Layout(opts)

	default:
		return fmt.Errorf("unsupported chart type: %s", chartopts.chartType)

	}

	s, err := gtree.SVG(lay)
	if err != nil {
		return fmt.Errorf("render SVG: %w", err)
	}

	if chartopts.outputFilename != "" {
		err = os.WriteFile(chartopts.outputFilename, []byte(s), 0o666)
		if err != nil {
			return fmt.Errorf("failed writing output file: %w", err)
		}
	} else {
		fmt.Println(s)
	}

	return nil
}

type sequence struct {
	n int
}

func (s *sequence) next() int {
	n := s.n
	s.n++
	return n
}

func scaleFont(v gtree.Pixel, factor float64) gtree.Pixel {
	v = gtree.Pixel(float64(v) * factor)
	if v < 6 {
		v = 6
	}
	return v
}

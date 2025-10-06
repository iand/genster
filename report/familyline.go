package report

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	gegedcom "github.com/iand/genster/gedcom"
	"github.com/iand/genster/gramps"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/narrative"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/tree"
	"github.com/urfave/cli/v2"
)

var familylineCommand = &cli.Command{
	Name:   "familyline",
	Usage:  "Generate a report on a family line",
	Action: familyline,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gedcom",
			Aliases:     []string{"g", "input"},
			Usage:       "GEDCOM file to read from",
			Destination: &familylineOpts.gedcomFile,
		},
		&cli.StringFlag{
			Name:        "gramps",
			Usage:       "Gramps xml file to read from",
			Destination: &familylineOpts.grampsFile,
		},
		&cli.StringFlag{
			Name:        "gramps-dbname",
			Usage:       "Name of the gramps database, used to keep IDs consistent between versions of the same database",
			Destination: &familylineOpts.grampsDatabaseName,
		},
		&cli.StringFlag{
			Name:        "id",
			Usage:       "Identifier to give this tree (mainly for annotation support)",
			Destination: &familylineOpts.treeID,
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
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where config should be stored.",
			Destination: &familylineOpts.configDir,
		},
		&cli.StringFlag{
			Name:        "person",
			Aliases:     []string{"p"},
			Usage:       "identifier of person to produce report from",
			Destination: &familylineOpts.startPersonID,
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "identifier of the key person, e.g. you",
			Destination: &familylineOpts.keyPersonID,
		},
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
			Usage:       "name of file to output report to, default is stdout, use '-' for stdout",
			Destination: &familylineOpts.outputFile,
		},
		&cli.StringFlag{
			Name:        "surnames",
			Usage:       "comma separated list of surnames this report should include",
			Destination: &familylineOpts.surnames,
		},
	}, logging.Flags...),
}

var familylineOpts struct {
	gedcomFile         string
	grampsFile         string
	grampsDatabaseName string
	treeID             string
	includePrivate     bool
	configDir          string
	startPersonID      string
	keyPersonID        string
	outputFile         string
	surnames           string
}

func checkFlags(cc *cli.Context) error {
	if familylineOpts.outputFile == "" {
		familylineOpts.outputFile = "-"
	}
	return nil
}

func familyline(cc *cli.Context) error {
	if err := checkFlags(cc); err != nil {
		return err
	}

	logging.Setup()

	var l tree.Loader
	var err error

	if familylineOpts.gedcomFile != "" {
		l, err = gegedcom.NewLoader(familylineOpts.gedcomFile)
		if err != nil {
			return fmt.Errorf("load gedcom: %w", err)
		}
	} else if familylineOpts.grampsFile != "" {
		l, err = gramps.NewLoader(familylineOpts.grampsFile, familylineOpts.grampsDatabaseName)
		if err != nil {
			return fmt.Errorf("load gedcom: %w", err)
		}
	} else {
		return fmt.Errorf("no gedcom or gramps file specified")
	}

	t, err := tree.LoadTree(familylineOpts.treeID, familylineOpts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	if err := t.Generate(!familylineOpts.includePrivate); err != nil {
		return fmt.Errorf("build tree: %w", err)
	}

	// Look for key person, if any. This is the person who is used to determine
	// whether a person in the tree is a direct ancestor
	// assume id is a genster id first
	keyPerson, ok := t.GetPerson(familylineOpts.keyPersonID)
	if !ok {
		keyPerson = t.FindPerson(l.Scope(), familylineOpts.keyPersonID)
	}
	t.SetKeyPerson(keyPerson)

	if err := t.Generate(false); err != nil {
		return fmt.Errorf("generate tree facts: %w", err)
	}

	// Find the root of the tree, i.e. the earliest ancester we want to show on the tree
	// assume id is a genster id first
	startPerson, ok := t.GetPerson(familylineOpts.startPersonID)
	if !ok {
		// not a genster id, so look for a native id
		startPerson = t.FindPerson(l.Scope(), familylineOpts.startPersonID)
	}
	if startPerson.IsUnknown() {
		return fmt.Errorf("person with id %s not found", familylineOpts.startPersonID)
	}

	surnames := strings.Split(familylineOpts.surnames, ",")
	for i := range surnames {
		surnames[i] = strings.ToLower(strings.TrimSpace(surnames[i]))
	}

	outfile := os.Stdout
	if familylineOpts.outputFile != "-" {
		outfile, err = os.Create(familylineOpts.outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outfile.Close()
	}

	doc := &md.Document{}
	nc := &narrative.TimelineNameChooser{}

	type POVEvent struct {
		POV   *model.POV
		Event model.TimelineEvent
	}

	var timeline []*POVEvent

	fmt.Fprintf(outfile, "Surnames: %s\n\n", strings.Join(surnames, " "))

	fmt.Fprintf(outfile, "Lineage of Direct Ancestors: \n")
	people := startPerson.RelationToKeyPerson.Path()
	for i, p := range slices.Backward(people) {
		if slices.Index(surnames, strings.ToLower(p.PreferredFamilyName)) == -1 {
			break
		}
		if i < len(people)-1 {
			fmt.Fprintf(outfile, " -> ")
		}
		fmt.Fprintf(outfile, "%s", p.PreferredUniqueName)
		pov := &model.POV{Person: p}
		for _, ev := range p.Timeline {
			timeline = append(timeline, &POVEvent{POV: pov, Event: ev})
		}
	}
	fmt.Fprintf(outfile, "\n\n")

	sort.Slice(timeline, func(i, j int) bool {
		return model.EventSortsBefore(timeline[i].Event, timeline[j].Event)
	})

	fmt.Fprintf(outfile, "Timeline of Events\n")

	for i, pe := range timeline {
		dateText := ""
		dt := pe.Event.GetDate()
		if dt.IsUnknown() {
			dateText = "(unknown date)"
		} else {
			dateText = dt.When()
		}

		whoText := narrative.WhoPov(pe.Event, doc, nc, pe.POV)

		whatText := ""
		if dt.Derivation != model.DateDerivationStandard {
			whatText = model.ConditionalWhat(pe.Event, "probably")
		} else {
			whatText = model.What(pe.Event)
		}

		placeText := ""
		pl := pe.Event.GetPlace()
		if !pl.IsUnknown() {
			placeText = pl.Where()
		}

		fmt.Fprintf(outfile, "%d. %s - %s %s %s\n", i+1, dateText, whoText, whatText, placeText)
	}

	return nil
}

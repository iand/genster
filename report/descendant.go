package report

import (
	"fmt"

	"github.com/iand/genster/gedcom"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
	"github.com/iand/genster/tree"
	"github.com/urfave/cli/v2"
)

var descendantCommand = &cli.Command{
	Name:   "descendant",
	Usage:  "Generate a text report from a gedcom file",
	Action: descendant,
	Flags: append([]cli.Flag{
		&cli.StringFlag{
			Name:        "gedcom",
			Aliases:     []string{"g"},
			Usage:       "GEDCOM file to read from",
			Destination: &descendantOpts.gedcomFile,
		},
		&cli.BoolFlag{
			Name:        "include-private",
			Usage:       "Include living people and people who died less than 20 years ago.",
			Value:       false,
			Destination: &descendantOpts.includePrivate,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where config should be stored.",
			Destination: &descendantOpts.configDir,
		},
		&cli.StringFlag{
			Name:        "person",
			Aliases:     []string{"p"},
			Usage:       "identifier of person to build tree from",
			Destination: &descendantOpts.startPersonID,
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "identifier of the key person, e.g. you",
			Destination: &descendantOpts.keyPersonID,
		},
		&cli.IntFlag{
			Name:        "gen",
			Usage:       "number of descendant generations to draw",
			Value:       3,
			Destination: &descendantOpts.generations,
		},
		&cli.IntFlag{
			Name:        "detail",
			Usage:       "level of detail to include with each person (0:none,1:years,2:dates,3:full)",
			Value:       3,
			Destination: &descendantOpts.detail,
		},
		&cli.BoolFlag{
			Name:        "compact",
			Usage:       "Remove blank lines from output.",
			Value:       false,
			Destination: &descendantOpts.compact,
		},
	}, logging.Flags...),
}

var descendantOpts struct {
	gedcomFile     string
	includePrivate bool
	configDir      string
	startPersonID  string
	keyPersonID    string
	generations    int
	detail         int
	compact        bool
}

func descendant(cc *cli.Context) error {
	logging.Setup()

	if descendantOpts.gedcomFile == "" {
		return fmt.Errorf("no gedcom file specified")
	}

	var detailFn func(*model.Person) string
	switch descendantOpts.detail {
	case 0:
		detailFn = detailLevel0
	case 1:
		detailFn = detailLevel1
	case 2:
		detailFn = detailLevel2
	case 3:
		detailFn = detailLevel3
	default:
		return fmt.Errorf("Unsupported detail level: %d", descendantOpts.detail)
	}

	l, err := gedcom.NewLoader(descendantOpts.gedcomFile)
	if err != nil {
		return fmt.Errorf("load gedcom: %w", err)
	}

	t, err := tree.LoadTree(descendantOpts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	if err := t.Generate(!descendantOpts.includePrivate); err != nil {
		return fmt.Errorf("build tree: %w", err)
	}
	// Look for key individual, assume id is a genster id first
	keyIndividual, ok := t.GetPerson(descendantOpts.keyPersonID)
	if !ok {
		keyIndividual = t.FindPerson(l.ScopeName, descendantOpts.keyPersonID)
	}

	t.SetKeyPerson(keyIndividual)

	startPerson, ok := t.GetPerson(descendantOpts.startPersonID)
	if !ok {
		keyIndividual = t.FindPerson(l.ScopeName, descendantOpts.startPersonID)
	}

	printDescendants(startPerson, "", 1, detailFn, descendantOpts.compact, descendantOpts.generations)

	return nil
}

func printDescendants(p *model.Person, indent string, idx int, detailFn func(*model.Person) string, compact bool, generations int) {
	fmt.Printf("%s% 2d. %s\n", indent, idx, detailFn(p))
	if !compact {
		fmt.Println()
	}
	for _, f := range p.Families {
		o := f.OtherParent(p)
		fmt.Printf("%s    sp. %s\n", indent, detailFn(o))
		if !compact {
			fmt.Println()
		}
		if generations > 0 {
			for i, c := range f.Children {
				printDescendants(c, indent+"      ", i+1, detailFn, compact, generations-1)
			}
		}
	}
}

func detailLevel0(p *model.Person) string {
	return p.PreferredFullName
}

func detailLevel1(p *model.Person) string {
	// Just show birth and death years
	return p.PreferredUniqueName
}

func detailLevel2(p *model.Person) string {
	s := p.PreferredFullName
	bl := formatEventBrief(p.BestBirthlikeEvent)
	dl := formatEventBrief(p.BestDeathlikeEvent)

	if bl == "" && dl == "" {
		return s
	}

	s += " ("
	if bl != "" {
		s += bl
		if dl != "" {
			s += "; "
		}
	}
	if dl != "" {
		s += dl
	}
	s += ")"
	return s
}

func detailLevel3(p *model.Person) string {
	s := p.PreferredFullName
	bl := formatEventFull(p.BestBirthlikeEvent)
	dl := formatEventFull(p.BestDeathlikeEvent)

	if bl == "" && dl == "" {
		return s
	}

	s += " ("
	if bl != "" {
		s += bl
		if dl != "" {
			s += "; "
		}
	}
	if dl != "" {
		s += dl
	}
	s += ")"
	return s
}

func formatEventBrief(ev model.TimelineEvent) string {
	if ev == nil {
		return ""
	}
	details := ev.ShortDescription()

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		details += " " + pl.PreferredName
	}

	return details
}

func formatEventFull(ev model.TimelineEvent) string {
	if ev == nil {
		return ""
	}
	details := ev.What()
	date := ev.GetDate()
	if !date.IsUnknown() {
		details = text.JoinSentence(details, date.When())
	}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		details = text.JoinSentence(details, pl.PlaceType.InAt(), pl.PreferredName)
	}
	return details
}

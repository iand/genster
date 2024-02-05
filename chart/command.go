/*
This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying UNLICENSE file.
*/

package chart

import (
	"fmt"
	"image/color"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"

	gegedcom "github.com/iand/genster/gedcom"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/iand/gtree"
)

func checkFlags(cc *cli.Context) error {
	return nil
}

func parseColor(s string) (color.Color, error) {
	intColor, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		return color.RGBA{}, err
	}

	return color.RGBA{uint8((intColor & 0xFF0000) >> 16), uint8((intColor & 0x00FF00) >> 8), uint8((intColor & 0x0000FF)), 0xFF}, nil
}

var chartopts struct {
	gedcomFile    string
	treeID        string
	configDir     string
	keyPersonID   string
	startPersonID string

	outputFilename string
	outputFormat   string
	descendantId   string
	descendants    int
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
		&cli.IntFlag{
			Name:        "desc",
			Usage:       "number of descendant generations to draw",
			Value:       2,
			Destination: &chartopts.descendants,
		},
		&cli.IntFlag{
			Name:        "detail",
			Usage:       "level of detail to include with each person (0:none,1:years,2:dates,3:full)",
			Value:       3,
			Destination: &chartopts.detail,
		},
		&cli.BoolFlag{
			Name:        "direct",
			Usage:       "only show children of direct ancestors",
			Value:       false,
			Destination: &chartopts.directOnly,
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where config should be stored.",
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

	var personDetailFn func(*model.Person) []string
	var familyDetailFn func(*model.Family) []string
	switch chartopts.detail {
	case 0:
		personDetailFn = func(p *model.Person) []string {
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			return []string{name}
		}
		familyDetailFn = func(p *model.Family) []string {
			return []string{}
		}
	case 1:
		personDetailFn = func(p *model.Person) []string {
			var details []string
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			details = append(details, name)
			details = append(details, p.VitalYears)
			return details
		}
		familyDetailFn = func(f *model.Family) []string {
			var details []string
			startYear, ok := f.BestStartDate.Year()
			if ok {
				details = append(details, fmt.Sprintf("%d", startYear))
			}
			return details
		}
	case 2:
		personDetailFn = func(p *model.Person) []string {
			var details []string
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			details = append(details, name)
			if p.BestBirthlikeEvent != nil {
				details = append(details, p.BestBirthlikeEvent.ShortDescription())
			}
			if p.BestDeathlikeEvent != nil {
				details = append(details, p.BestDeathlikeEvent.ShortDescription())
			}

			return details
		}
		familyDetailFn = func(f *model.Family) []string {
			var details []string
			if f.BestStartEvent != nil {
				details = append(details, f.BestStartEvent.ShortDescription())
			}
			if f.BestEndEvent != nil {
				details = append(details, f.BestEndEvent.ShortDescription())
			}
			return details
		}
	case 3:
		personDetailFn = func(p *model.Person) []string {
			var details []string
			name := p.PreferredFullName
			if p.IsDirectAncestor() {
				name += "★"
			}
			details = append(details, name)
			if p.PrimaryOccupation != "" {
				details = append(details, p.PrimaryOccupation)
			}
			if p.BestBirthlikeEvent != nil {
				if p.BestBirthlikeEvent.GetPlace().IsUnknown() {
					details = append(details, p.BestBirthlikeEvent.ShortDescription())
				} else {
					details = append(details, p.BestBirthlikeEvent.ShortDescription()+", "+p.BestBirthlikeEvent.GetPlace().PreferredName)
				}
			}
			if p.BestDeathlikeEvent != nil {
				if p.BestDeathlikeEvent.GetPlace().IsUnknown() {
					details = append(details, p.BestDeathlikeEvent.ShortDescription())
				} else {
					details = append(details, p.BestDeathlikeEvent.ShortDescription()+", "+p.BestDeathlikeEvent.GetPlace().PreferredName)
				}
			}

			return details
		}
		familyDetailFn = func(f *model.Family) []string {
			var details []string
			if f.BestStartEvent != nil {
				if f.BestStartEvent.GetPlace().IsUnknown() {
					details = append(details, f.BestStartEvent.ShortDescription())
				} else {
					details = append(details, f.BestStartEvent.ShortDescription()+", "+f.BestStartEvent.GetPlace().PreferredName)
				}
			}
			if f.BestEndEvent != nil {
				details = append(details, f.BestEndEvent.ShortDescription())
				if f.BestEndEvent.GetPlace().IsUnknown() {
					details = append(details, f.BestEndEvent.ShortDescription())
				} else {
					details = append(details, f.BestEndEvent.ShortDescription()+", "+f.BestEndEvent.GetPlace().PreferredName)
				}
			}
			return details
		}
	default:
		return fmt.Errorf("unsupported detail level: %d", chartopts.detail)

	}

	lin := new(gtree.Lineage)
	lin.Root = descendants(startPerson, new(sequence), chartopts.descendants, chartopts.directOnly, personDetailFn, familyDetailFn)
	opts := gtree.DefaultLayoutOptions()
	lay := gtree.NewLayout(opts)
	lay.AddLineage(lin)
	lay.Reflow()

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

/*
This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying UNLICENSE file.
*/

package chart

import (
	"fmt"
	"image/color"
	"regexp"
	"sort"
	"strconv"

	"github.com/iand/gdate"
	gegedcom "github.com/iand/genster/gedcom"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"github.com/tdewolff/canvas/renderers/svg"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

var direct map[string]bool

var yearRe = regexp.MustCompile(`\b(\d\d\d\d)$`)

var (
	A5Paper = [2]int{210, 148}
	A4Paper = [2]int{297, 210}
	A3Paper = [2]int{420, 297}
	A2Paper = [2]int{594, 420}
	A1Paper = [2]int{841, 594}
)

func checkFlags(cc *cli.Context) error {
	var err error
	chartopts.lineColor, err = parseColor(chartopts.lineColorHex)
	if err != nil {
		return fmt.Errorf("%q is not a valid color", chartopts.lineColorHex)
	}

	chartopts.nameColor, err = parseColor(chartopts.nameColorHex)
	if err != nil {
		return fmt.Errorf("%q is not a valid color", chartopts.nameColorHex)
	}

	chartopts.detailsColor, err = parseColor(chartopts.detailsColorHex)
	if err != nil {
		return fmt.Errorf("%q is not a valid color", chartopts.detailsColorHex)
	}

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
	configDir     string
	keyPersonID   string
	startPersonID string

	outputFilename    string
	outputFormat      string
	descendantId      string
	generations       int
	detail            int
	horizontal        bool
	directOnly        bool
	lineWidth         int
	lineColor         color.Color
	lineColorHex      string
	nameSize          float64
	nameColor         color.Color
	nameColorHex      string
	detailsSize       float64
	detailsColor      color.Color
	detailsColorHex   string
	individualSpacing float64
	familySpacing     float64
	margin            float64
	dpi               float64
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
			Name:        "output",
			Usage:       "output image filename",
			Destination: &chartopts.outputFilename,
		},
		&cli.StringFlag{
			Name:        "format",
			Usage:       "output format (png or dot)",
			Value:       "png",
			Destination: &chartopts.outputFormat,
		},
		&cli.StringFlag{
			Name:        "person",
			Usage:       "identifier of person to build tree from",
			Destination: &chartopts.startPersonID,
		},
		&cli.StringFlag{
			Name:        "descendant",
			Usage:       "identifier of the main descendant, e.g. you",
			Destination: &chartopts.descendantId,
		},
		&cli.IntFlag{
			Name:        "gen",
			Usage:       "number of descendant generations to draw",
			Value:       3,
			Destination: &chartopts.generations,
		},
		&cli.IntFlag{
			Name:        "detail",
			Usage:       "level of detail to include with each person (0:none,1:years,2:dates,3:full)",
			Value:       3,
			Destination: &chartopts.detail,
		},
		&cli.BoolFlag{
			Name:        "hor",
			Usage:       "force all subtrees to be horizontal",
			Value:       false,
			Destination: &chartopts.horizontal,
		},
		&cli.BoolFlag{
			Name:        "direct",
			Usage:       "only show children of direct ancestors",
			Value:       false,
			Destination: &chartopts.directOnly,
		},
		&cli.IntFlag{
			Name:        "lwidth",
			Usage:       "width of lines (in mm)",
			Value:       1,
			Destination: &chartopts.lineWidth,
		},
		&cli.StringFlag{
			Name:        "lcolor",
			Usage:       "color of lines (e.g. FF0000)",
			Value:       "333333",
			Destination: &chartopts.lineColorHex,
		},
		&cli.StringFlag{
			Name:        "ncolor",
			Usage:       "color of names (e.g. FF0000)",
			Value:       "1D2951",
			Destination: &chartopts.nameColorHex,
		},
		&cli.StringFlag{
			Name:        "dcolor",
			Usage:       "color of details (e.g. FF0000)",
			Value:       "336633",
			Destination: &chartopts.detailsColorHex,
		},
		&cli.Float64Flag{
			Name:        "nsize",
			Usage:       "font size for names (in points)",
			Value:       16,
			Destination: &chartopts.nameSize,
		},
		&cli.Float64Flag{
			Name:        "dsize",
			Usage:       "font size for details (in points)",
			Value:       12,
			Destination: &chartopts.detailsSize,
		},
		&cli.Float64Flag{
			Name:        "ispace",
			Usage:       "spacing between individuals (in mm)",
			Value:       3,
			Destination: &chartopts.individualSpacing,
		},
		&cli.Float64Flag{
			Name:        "fspace",
			Usage:       "spacing between families (in mm)",
			Value:       5,
			Destination: &chartopts.familySpacing,
		},
		&cli.Float64Flag{
			Name:        "margin",
			Usage:       "margin to leave around tree (in mm)",
			Value:       15,
			Destination: &chartopts.margin,
		},
		&cli.Float64Flag{
			Name:        "dpi",
			Usage:       "dpi (dots per inch) of output",
			Value:       300,
			Destination: &chartopts.dpi,
		},
		&cli.StringFlag{
			Name:  "debug",
			Usage: "debug individual (i), family (f) and/or children (c), e.g. --debug=ic",
			Value: "",
		},
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       tree.DefaultConfigDir(),
			Usage:       "Path to the folder where config should be stored.",
			Destination: &chartopts.configDir,
		},
		&cli.StringFlag{
			Name:        "key",
			Aliases:     []string{"k"},
			Usage:       "Identifier of the key individual",
			Destination: &chartopts.keyPersonID,
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

	t, err := tree.LoadTree(chartopts.configDir, l)
	if err != nil {
		return fmt.Errorf("load tree: %w", err)
	}

	if err := t.Generate(false); err != nil {
		return fmt.Errorf("generate tree facts: %w", err)
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
		t.SetKeyPerson(keyPerson)
	}

	// Find the root of the tree, i.e. the earliest ancester we want to show of the tree
	startPerson, ok := t.GetPerson(chartopts.startPersonID)
	if !ok {
		startPerson = t.FindPerson(l.ScopeName, chartopts.startPersonID)
	}

	// familytree.RegisterFont(draw2d.FontData{"verdana", draw2d.FontFamilyMono, draw2d.FontStyleBold}, "/usr/share/fonts/truetype/msttcorefonts/verdanab.ttf")
	// familytree.RegisterFont(draw2d.FontData{"verdana", draw2d.FontFamilyMono, draw2d.FontStyleItalic}, "/usr/share/fonts/truetype/msttcorefonts/verdanai.ttf")
	// familytree.RegisterFont(draw2d.FontData{"georgia", draw2d.FontFamilyMono, draw2d.FontStyleBold}, "/usr/share/fonts/truetype/msttcorefonts/georgiab.ttf")
	// familytree.RegisterFont(draw2d.FontData{"impact", draw2d.FontFamilyMono, draw2d.FontStyleItalic}, "/usr/share/fonts/truetype/msttcorefonts/impact.ttf")
	// familytree.RegisterFont(draw2d.FontData{"arial", draw2d.FontFamilyMono, draw2d.FontStyleItalic}, "/usr/share/fonts/truetype/msttcorefonts/ariali.ttf")

	// data, err := ioutil.ReadFile(chartopts.gedcomFile)
	// if err != nil {
	// 	return fmt.Errorf("Failed opening input file: %w", err)
	// }

	// d := gedcom.NewDecoder(bytes.NewReader(data))

	// g, err := d.Decode()
	// if err != nil {
	// 	return fmt.Errorf("Failed decoding input file: %w", err)
	// }
	// direct = make(map[string]bool, 0)
	// if descendantId != "" {
	// 	findDirectAncestors(g, descendantId)
	// }

	// var rootRecord *gedcom.IndividualRecord
	// for _, i := range g.Individual {
	// 	if i.Xref == chartopts.startPersonID {
	// 		rootRecord = i
	// 		break
	// 	}
	// }

	// if rootRecord == nil {
	// 	return fmt.Errorf("Could not find a person with identifier '%s' in input file", chartopts.startPersonID)
	// }

	if chartopts.outputFilename == "" {
		chartopts.outputFilename = fmt.Sprintf("%s-%d.%s", chartopts.startPersonID, chartopts.generations, chartopts.outputFormat)
	}

	slog.Info("creating chart")

	switch chartopts.outputFormat {
	case "png", "svg", "pdf":
		root := descend(startPerson, chartopts.generations-1, chartopts.detail)
		c := NewChart()
		c.ForceHorizontal = chartopts.horizontal
		c.LineWidth = 1 // float64(lineWidth)
		c.LineColor = chartopts.lineColor
		c.IndividualNameFontSize = chartopts.nameSize
		c.IndividualNameColor = chartopts.nameColor
		c.IndividualDetailsFontSize = chartopts.detailsSize
		c.IndividualDetailsColor = chartopts.detailsColor
		c.IndividualSpacing = chartopts.individualSpacing
		c.FamilySpacing = chartopts.familySpacing

		// Temporary overrides
		c.IndividualSpacing = 15        // the horizontal distance between individuals in a family group (in mm)
		c.IndividualVerticalSpacing = 5 // the vertical distance between individuals in a family group (in mm)
		c.FamilySpacing = 10            // the horizontal distance between different family groups (in mm)
		c.IndividualPadding = 2         // the amount of whitespace padding between an individual and any line connecting to it (in mm)
		c.FamilyDetailsDrop = 10        // the vertical distance (in mm) between family details and the center point of the parents info (between the = and the date of marriage)
		c.ChildLineDropAbove = 10       // the vertical distance (in mm) between family details and the child line and between the child line and child details (between = and line that children hang from)
		c.ChildLineDropBelow = 10       // the vertical or horizontal distance (in mm) between the child line and the child details
		c.ChildLineOffset = 10          // the horizontal distance (in mm) between the start of an individual and the line descending to it
		c.TextVertSpacing = 10          // vertical spacing between lines of text   (in mm)
		c.SpouseSeparation = 1          // spacing between husband and wife  (in mm)
		c.LineWidth = 0.5               // Width of the line (in mm)
		c.IndividualDetailsWidth = 35

		debug := cc.String("debug")
		for _, r := range debug {
			switch r {
			case 'f':
				c.DebugFamilyOutline = true
			case 'i':
				c.DebugIndividualOutline = true
			case 'c':
				c.DebugChildrenOutline = true
			}
		}

		can := canvas.New(3*420-2*chartopts.margin, 2*297-2*chartopts.margin) // in millimetres

		if err := c.Draw(root, can); err != nil {
			return err
		}
		can.Fit(chartopts.margin)
		can.SetZIndex(-10)
		gc := canvas.NewContext(can)
		gc.SetFillColor(canvas.White)
		gc.DrawPath(0, 0, canvas.Rectangle(can.W, can.H))

		switch chartopts.outputFormat {
		case "png":
			return can.WriteFile(chartopts.outputFilename, renderers.PNG())
		case "svg":
			opts := svg.DefaultOptions
			opts.EmbedFonts = false
			return can.WriteFile(chartopts.outputFilename, renderers.SVG(&opts))
		case "pdf":
			return can.WriteFile(chartopts.outputFilename, renderers.PDF())
		}

		// fout, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY, 0o666)
		// if err != nil {
		// 	return fmt.Errorf("Failed writing output image: %w", err)
		// }
		// defer fout.Close()

		// if err = png.Encode(fout, imgOut); err != nil {
		// 	return fmt.Errorf("Failed encoding output image: %w", err)
		// }

	// case "dot":
	// 	root := descend(rootRecord, chartopts.generations-1, chartopts.detail)
	// 	c := NewDotChart()
	// 	dataOut, err := c.Draw(root)
	// 	if err != nil {
	// 		return fmt.Errorf("Failed drawing chart: %w", err)
	// 	}
	// 	err = os.WriteFile(chartopts.outputFilename, dataOut, 0o666)
	// 	if err != nil {
	// 		return fmt.Errorf("Failed writing output file: %w", err)
	// 	}
	default:

		return fmt.Errorf("Unsupported output format '%s'", chartopts.outputFormat)
	}

	return nil
}

func descend(p *model.Person, generations int, detail int) *Individual {
	slog.Debug("descending from person", "id", p.ID, "name", p.PreferredFullName)
	in := individual(p, 0, detail)

	numberOfSpouses := len(p.Families)
	for fidx, f := range p.Families {
		spouseIndex := 0
		if numberOfSpouses > 1 {
			spouseIndex = fidx + 1
		}

		fd := &Family{
			ID:     f.ID,
			Spouse: individual(f.OtherParent(p), spouseIndex, detail),
		}

		if f.Bond == model.FamilyBondMarried {
			if detail == 1 {
				start := f.BestStartEvent
				if start != nil {
					yr, ok := start.GetDate().Year()
					if ok {
						fd.Details = append(fd.Details, fmt.Sprintf("m. %d", yr))
					}
				}
			} else {
				var detailFunc func(model.TimelineEvent) []string
				if detail == 2 {
					detailFunc = formatEventBrief
				} else {
					detailFunc = formatEventFull
				}

				fd.Details = append(fd.Details, detailFunc(f.BestStartEvent)...)
				fd.Details = append(fd.Details, detailFunc(f.BestEndEvent)...)
			}
		}
		// _, isDirect := direct[p.Xref]
		// if p.RelationToKeyPerson.IsDirectAncestor() {
		for _, c := range f.Children {
			if generations >= 0 {
				fd.Children = append(fd.Children, descend(c, generations-1, detail))
			} else {
				// f.Children = append(f.Children, individual(crec, 0))
			}
		}
		// }

		in.Families = append(in.Families, fd)
	}

	sort.Slice(in.Families, func(i, j int) bool {
		return gdate.SortsBefore(in.Families[i].Date, in.Families[j].Date)
	})

	return in
}

func individual(p *model.Person, spouseIndex int, detail int) *Individual {
	name := p.PreferredFullName
	if spouseIndex > 0 {
		name = fmt.Sprintf("(%d) %s", spouseIndex, name)
	}

	slog.Debug("adding individual", "id", p.ID, "name", name)
	i := &Individual{
		ID:   p.ID,
		Name: name,
	}

	// if _, exists := direct[rec.Xref]; exists {
	// 	i.Direct = true
	// }

	if detail == 0 {
		// Only name is required
		return i
	}

	switch detail {
	case 1:
		// Just show birth and death years
		i.Details = append(i.Details, p.VitalYears)

	default:
		var detailFunc func(model.TimelineEvent) []string
		if detail == 2 {
			detailFunc = formatEventBrief
		} else {
			detailFunc = formatEventFull
		}
		i.Details = append(i.Details, detailFunc(p.BestBirthlikeEvent)...)
		i.Details = append(i.Details, detailFunc(p.BestDeathlikeEvent)...)
	}
	return i
}

// func formatName(names []*gedcom.NameRecord, spouseIndex int) string {
// 	name := "Unknown"
// 	if len(names) > 0 {
// 		pn := gedcom.SplitPersonalName(names[0].Name)
// 		name = pn.Full
// 	}

// 	if spouseIndex > 0 {
// 		return fmt.Sprintf("(%d) %s", spouseIndex, name)
// 	}
// 	return name
// }

// func formatPlace(e *gedcom.EventRecord) (string, bool) {
// 	// if e.Address.City != "" {
// 	// 	return fmt.Sprintf("%s", e.Address.City), true
// 	// } else {
// 	if e.Place.Name != "" {
// 		return fmt.Sprintf("%s", e.Place.Name), true
// 	} else {
// 		return "", false
// 	}
// 	// }
// }

func formatEventBrief(ev model.TimelineEvent) []string {
	if ev == nil {
		return []string{}
	}
	details := []string{ev.ShortDescription()}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		details = append(details, pl.PreferredName)
	}

	return details
}

func formatEventFull(ev model.TimelineEvent) []string {
	if ev == nil {
		return []string{}
	}
	details := []string{ev.ShortDescription()}

	pl := ev.GetPlace()
	if !pl.IsUnknown() {
		details = append(details, pl.PreferredFullName)
	}

	return details
}

// func findDirectAncestors(g *gedcom.Gedcom, id string) {
// 	for _, i := range g.Individual {
// 		if i.Xref == id {
// 			direct[id] = true
// 			markParentsDirect(i)
// 			return
// 		}
// 	}
// }

// func markParentsDirect(rec *gedcom.IndividualRecord) {
// 	for _, frec := range rec.Parents {
// 		if frec.Family.Husband != nil {
// 			direct[frec.Family.Husband.Xref] = true
// 			markParentsDirect(frec.Family.Husband)
// 		}

// 		if frec.Family.Wife != nil {
// 			direct[frec.Family.Wife.Xref] = true
// 			markParentsDirect(frec.Family.Wife)
// 		}
// 	}
// }

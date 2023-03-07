/*
This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying UNLICENSE file.
*/
package chart

import (
	"fmt"
	"image/color"

	"github.com/iand/gdate"
	// "github.com/iand/genster/model"
	"github.com/tdewolff/canvas"
)

type Chart struct {
	Dpi                       float64 // number of pixels per inch
	IndividualSpacing         float64 // the horizontal distance between individuals in a family group (in mm)
	IndividualVerticalSpacing float64 // the vertical distance between individuals in a family group when drawn vertically(in mm)
	FamilySpacing             float64 // the horizontal distance between different family groups (in mm)
	IndividualPadding         float64 // the amount of whitespace padding between an individual any line connecting to it (in mm)

	IndividualNameFont       *canvas.FontFace // Font to use for individual names
	IndividualNameFontSize   float64          // Font size to use for individual names (in points)
	IndividualNameColor      color.Color      // Color to use when drawing individual names
	IndividualNameTextHeight float64          // Calculated height (in mm) of a line of name text

	IndividualDetailsFont       *canvas.FontFace // Font to use for individual details
	IndividualDetailsFontSize   float64          // Font size to use for individual details (in points)
	IndividualDetailsColor      color.Color      // Color to use when drawing individual deails
	IndividualDetailsTextHeight float64          // Calculated height (in mm) of a line of detail text

	IndividualDetailsWidth float64 // width in mm

	FamilyDetailsDrop  float64 // the vertical distance (in mm) between family details and the center point of the parents info (between the = and the date of marriage)
	ChildLineDropAbove float64 // the vertical distance (in mm) between family details and the child line and between the child line and child details (between = and line that children hang from)
	ChildLineDropBelow float64 // the vertical or horizontal distance (in mm) between the child line and the child details
	ChildLineOffset    float64 // the horizontal distance (in mm) between the start of an individual and the line descending to it

	LineColor color.Color // Color to use when drawing lines
	LineWidth float64     // Width of the line (in mm)

	TextVertSpacing                float64 // vertical spacing between lines of text   (in mm)
	SpouseSeparation               float64 // spacing between husband and wife  (in mm)
	SpouseVerticalSeparation       float64 // vertical spacing between the main individual and the details of any marriage, and between a previous spouse and the next marriage details  (in mm)
	SpouseDetailVerticalSeparation float64 // vertical spacing between details of any marriage and the spouse  (in mm)

	ForceHorizontal bool // Whether to force all subtrees to be horizontally oriented

	PageHeight float64 // in mm
	PageWidth  float64 // in mm

	DebugFamilyOutline     bool
	DebugChildrenOutline   bool
	DebugIndividualOutline bool

	gc *canvas.Context
}

// NewChart creates a new chart with sensible defaults
func NewChart() *Chart {
	c := &Chart{
		Dpi:                       300,
		IndividualSpacing:         10,
		IndividualVerticalSpacing: 5,
		FamilySpacing:             10,
		IndividualPadding:         2,
		// IndividualNameFont:        draw2d.FontData{"verdana", draw2d.FontFamilyMono, draw2d.FontStyleBold},
		IndividualNameFontSize: 24,
		IndividualNameColor:    color.RGBA{0, 0, 0, 0xFF},
		// IndividualDetailsFont:     draw2d.FontData{"arial", draw2d.FontFamilyMono, draw2d.FontStyleItalic},
		IndividualDetailsFontSize: 18,
		IndividualDetailsColor:    color.RGBA{32, 32, 32, 0xFF},

		IndividualDetailsWidth: 30,

		LineColor: color.RGBA{32, 32, 32, 0xFF},
		LineWidth: 2.0,

		FamilyDetailsDrop:  10,
		ChildLineDropAbove: 10,
		ChildLineDropBelow: 10,
		ChildLineOffset:    10,

		SpouseSeparation:               24,
		SpouseVerticalSeparation:       5,
		SpouseDetailVerticalSeparation: 5,
		TextVertSpacing:                10,
	}

	return c
}

func (c *Chart) initContext(can *canvas.Canvas) error {
	// var err error

	fontGeorgia := canvas.NewFontFamily("georgia")
	// if err := fontGeorgia.LoadLocalFont("georgia", canvas.FontRegular); err != nil {
	// 	panic(err)
	// }
	// if err := fontGeorgia.LoadLocalFont("georgiab", canvas.FontBold); err != nil {
	// 	panic(err)
	// }

	if err := fontGeorgia.LoadFontFile("/usr/share/fonts/truetype/msttcorefonts/georgia.ttf", canvas.FontRegular); err != nil {
		panic(err)
	}
	if err := fontGeorgia.LoadFontFile("/usr/share/fonts/truetype/msttcorefonts/georgiab.ttf", canvas.FontBold); err != nil {
		panic(err)
	}

	fontArial := canvas.NewFontFamily("arial")
	if err := fontArial.LoadFontFile("/usr/share/fonts/truetype/msttcorefonts/arial.ttf", canvas.FontRegular); err != nil {
		panic(err)
	}
	if err := fontArial.LoadFontFile("/usr/share/fonts/truetype/msttcorefonts/ariali.ttf", canvas.FontItalic); err != nil {
		panic(err)
	}

	if c.IndividualNameFont == nil {
		c.IndividualNameFont = fontGeorgia.Face(c.IndividualNameFontSize, canvas.Black, canvas.FontBold, canvas.FontNormal)
		t := canvas.NewTextLine(c.IndividualNameFont, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", canvas.Left)
		bounds := t.Bounds()
		c.IndividualNameTextHeight = bounds.H
	}

	if c.IndividualDetailsFont == nil {
		c.IndividualDetailsFont = fontArial.Face(c.IndividualDetailsFontSize, canvas.Black, canvas.FontItalic, canvas.FontNormal)
		t := canvas.NewTextLine(c.IndividualDetailsFont, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", canvas.Left)
		bounds := t.Bounds()
		c.IndividualDetailsTextHeight = bounds.H
	}

	// if c.IndividualNameFont == nil {
	// 	c.IndividualNameFont, err = gg.LoadFontFace("/usr/share/fonts/truetype/msttcorefonts/georgiab.ttf", c.IndividualNameFontSize)
	// 	// c.IndividualNameFont, err = gg.LoadFontFace("/usr/share/fonts/truetype/msttcorefonts/verdanab.ttf", c.IndividualNameFontSize)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// }

	// if c.IndividualDetailsFont == nil {
	// 	c.IndividualDetailsFont, err = gg.LoadFontFace("/usr/share/fonts/truetype/msttcorefonts/ariali.ttf", c.IndividualDetailsFontSize)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// }

	return nil
}

func (c *Chart) Draw(in *Individual, can *canvas.Canvas) error {
	if err := c.initContext(can); err != nil {
		return err
	}

	c.PageHeight = can.H
	c.PageWidth = can.W

	c.gc = canvas.NewContext(can)

	root := NewIndividualContainer(in)
	root.Measure(c)

	// width, height := int(math.Ceil(root.MinWidth)), int(math.Ceil(root.MinHeight))

	// img := image.NewRGBA(image.Rect(0, 0, width, height))
	// draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)

	root.Draw(0, 0, c)

	return nil
}

func (c *Chart) MakeText(name string, details []string) *canvas.Text {
	if len(name) == 0 && len(details) == 0 {
		return canvas.NewTextLine(c.IndividualDetailsFont, "", canvas.Left)
	}

	rt := canvas.NewRichText(c.IndividualNameFont)
	rt.SetFace(c.IndividualNameFont)
	if len(name) > 0 {
		rt.WriteString(name)
		rt.WriteRune('\n')
	}
	rt.SetFace(c.IndividualDetailsFont)
	for _, s := range details {
		if len(s) > 0 {
			rt.WriteString(s)
			rt.WriteRune('\n')
		}
	}

	return rt.ToText(c.IndividualDetailsWidth, 0, canvas.Justify, canvas.Top, 0.0, 0.0)
}

func (c *Chart) MeasureTextArea(name string, details []string) (width float64, height float64) {
	t := c.MakeText(name, details)
	bounds := t.Bounds()
	return bounds.W, bounds.H
}

func (c *Chart) DrawLine(fromx, fromy, tox, toy float64) {
	var p canvas.Path
	p.MoveTo(fromx, c.PageHeight-fromy)
	p.LineTo(tox, c.PageHeight-toy)
	c.gc.DrawPath(0, 0, &p)
}

func (c *Chart) DrawIndividualText(x, y float64, in *Individual) {
	t := c.MakeText(in.Name, in.Details)
	c.gc.DrawText(x, c.PageHeight-y, t)
}

func (c *Chart) debugDrawLineHorz(x, y float64, col color.Color) {
	c.gc.Push()
	c.gc.SetStrokeColor(col)
	c.gc.SetStrokeWidth(1)

	c.DrawLine(x, y, x+20, y)
	c.gc.Pop()
}

func (c *Chart) debugDrawLineVert(x, y float64, col color.Color) {
	c.gc.Push()
	c.gc.SetStrokeColor(col)
	c.gc.SetStrokeWidth(1)

	c.DrawLine(x, y, x, y+20)
	c.gc.Pop()
}

func (c *Chart) debugDrawBox(x, y, w, h float64, col color.Color) {
	c.gc.Push()
	c.gc.SetFillColor(canvas.Transparent)
	c.gc.SetStrokeColor(col)
	c.gc.SetStrokeWidth(1)

	var p canvas.Path
	p.MoveTo(x, c.PageHeight-y)
	p.LineTo(x, c.PageHeight-(y+h))
	p.LineTo(x+w, c.PageHeight-(y+h))
	p.LineTo(x+w, c.PageHeight-y)
	p.Close()
	c.gc.DrawPath(0, 0, &p)
	c.gc.Pop()
}

type Individual struct {
	ID       string
	Name     string
	Details  []string
	Families []*Family
	Direct   bool
}

type Family struct {
	ID       string
	Date     gdate.Date
	Details  []string
	Spouse   *Individual
	Children []*Individual
}

type IndividualContainer struct {
	Individual       *Individual
	Families         []*FamilyContainer
	IndividualHeight float64 // height of just the individual name and details
	MinWidth         float64
	MinHeight        float64
	LeftPull         float64 // How much the family can be pulled over to the left to close up whitespace
	Text             *canvas.Text
}

func NewIndividualContainer(i *Individual) *IndividualContainer {
	ic := &IndividualContainer{
		Individual: i,
	}

	for index, f := range i.Families {
		fc := &FamilyContainer{
			Family:   f,
			Children: &ChildrenContainer{Individuals: make([]*IndividualContainer, 0)},
		}

		if f.Spouse != nil {
			fc.Spouse = &IndividualContainer{
				Individual: f.Spouse,
			}
		} else {
			fc.Spouse = UnknownIndividualContainer
		}

		if index == 0 {
			fc.Main = ic
		} else {
			fc.Main = EmptyIndividualContainer
		}

		if len(f.Children) > 0 {
			for _, child := range f.Children {
				fc.Children.Individuals = append(fc.Children.Individuals, NewIndividualContainer(child))
			}
		}
		ic.Families = append(ic.Families, fc)

	}

	return ic
}

var EmptyIndividualContainer = &IndividualContainer{
	Individual: &Individual{Name: ""},
}

var UnknownIndividualContainer = &IndividualContainer{
	Individual: &Individual{Name: "Unknown"},
}

func (ic *IndividualContainer) Measure(c *Chart) {
	ic.Text = c.MakeText(ic.Individual.Name, ic.Individual.Details)

	ic.MinWidth = c.IndividualDetailsWidth

	// Determine whether any family has children that must be drawn
	hasChildren := false
	for _, fc := range ic.Families {
		if fc.HasChildren() {
			hasChildren = true
			break
		}
	}

	if len(ic.Families) == 0 {
		// No spouses so height is simply the height of the main indidual's info
		bounds := ic.Text.Bounds()
		ic.MinHeight = bounds.H
	} else {
		familyMinWidth := 0.0
		for i, fc := range ic.Families {
			if !hasChildren && !c.ForceHorizontal {
				fc.DrawVertical = true
			}
			fc.Measure(c)

			if fc.DrawVertical {
				ic.MinHeight += fc.MinHeight
				familyMinWidth = max(familyMinWidth, fc.MinWidth)
			} else {
				// In some instances we can close up the gap between spouses
				if i > 0 {
					xAdjust := 0.0
					if len(fc.Children.Individuals) > 0 {
						for j := i - 1; j >= 0; j-- {
							if ic.Families[j].MinHeight >= fc.ParentHeight {
								break
							} else {
								xAdjust += ic.Families[j].MinWidth + c.IndividualSpacing
							}
						}

						if xAdjust > fc.MainOffset(c) {
							xAdjust = fc.MainOffset(c)
						}

					} else {
						if len(ic.Families[i-1].Children.Individuals) > 0 {
							if fc.MinHeight <= ic.Families[i-1].ParentHeight {
								xAdjust += ic.Families[i-1].MinWidth - ic.Families[i-1].MainOffset(c) - ic.Families[i-1].ParentWidth
							}
						}
					}

					fc.LeftPull = xAdjust

					familyMinWidth += (c.FamilySpacing)
					if fc.LeftPull < fc.MinWidth {
						familyMinWidth += fc.MinWidth - fc.LeftPull
					}

				} else {
					familyMinWidth += fc.MinWidth - fc.LeftPull
				}
				ic.MinHeight = max(ic.MinHeight, fc.MinHeight)
			}
		}

		ic.MinWidth = max(ic.MinWidth, familyMinWidth)
	}
}

// Returns the offset of the main individual from the left of the container
func (ic *IndividualContainer) MainOffset(c *Chart) float64 {
	if len(ic.Families) == 0 {
		return 0
	} else {
		return ic.Families[0].MainOffset(c)
	}
}

func (ic *IndividualContainer) Draw(x, y float64, c *Chart) {
	if c.DebugIndividualOutline {
		c.debugDrawBox(x, y, ic.MinWidth, ic.MinHeight, canvas.Yellow)
	}

	if len(ic.Families) == 0 {
		if ic.Individual != nil {
			c.DrawIndividualText(x, y, ic.Individual)
		}
	} else {
		for _, fc := range ic.Families {
			if fc.DrawVertical {
				fc.Draw(x, y, c)
				y += fc.MainHeight
				y += c.SpouseVerticalSeparation
				y += fc.DetailHeight
				y += c.SpouseDetailVerticalSeparation
				y += fc.SpouseHeight

				continue
			}

			x -= fc.LeftPull
			fc.Draw(x, y, c)
			x += fc.MinWidth
		}
	}
}

type FamilyContainer struct {
	Main     *IndividualContainer
	Spouse   *IndividualContainer
	Children *ChildrenContainer
	Family   *Family

	// dimensions of the entire family group including children
	MinWidth  float64
	MinHeight float64

	// dimensions of text information about the main individual in the family
	MainWidth  float64
	MainHeight float64

	// dimensions of marriage detail text
	DetailWidth  float64
	DetailHeight float64

	// dimensions of text information about the spouse in the family
	SpouseWidth  float64
	SpouseHeight float64

	// dimensions of text information about the main individual and the spouse
	ParentWidth  float64
	ParentHeight float64

	LeftPull float64 // How much the family can be pull over to the left to close up whitespace

	DrawVertical bool // when true this container will be rendered vertically
}

func (fc *FamilyContainer) HasChildren() bool {
	return fc != nil && fc.Children != nil && len(fc.Children.Individuals) > 0
}

func (fc *FamilyContainer) Measure(c *Chart) {
	if fc.Main == EmptyIndividualContainer {
		fc.MainWidth = (c.FamilySpacing) * 2
		fc.MainHeight = (c.FamilySpacing)
	} else {
		fc.MainWidth, fc.MainHeight = c.MeasureTextArea(fc.Main.Individual.Name, fc.Main.Individual.Details)
	}

	fc.SpouseWidth, fc.SpouseHeight = c.MeasureTextArea(fc.Spouse.Individual.Name, fc.Spouse.Individual.Details)

	fc.Spouse.MinWidth = fc.SpouseWidth
	fc.Spouse.MinHeight = fc.SpouseHeight
	fc.Spouse.IndividualHeight = fc.SpouseHeight

	fc.DetailWidth, fc.DetailHeight = c.MeasureTextArea("", fc.Family.Details)

	if fc.DrawVertical {
		fc.ParentWidth = max(max(fc.MainWidth, fc.SpouseWidth), fc.DetailWidth)
		fc.ParentHeight = fc.MainHeight + c.SpouseVerticalSeparation + fc.DetailHeight + c.SpouseDetailVerticalSeparation + fc.SpouseHeight
		fc.MinWidth = fc.ParentWidth
		fc.MinHeight = fc.ParentHeight
		return
	}

	fc.ParentWidth = fc.MainWidth + fc.SpouseWidth + max(c.SpouseSeparation, fc.DetailWidth)
	fc.ParentHeight = max(fc.MainHeight, fc.SpouseHeight)

	if fc.DetailWidth > 0 {
		if fc.MainHeight < (c.FamilyDetailsDrop)-2*(c.IndividualPadding) {
			// Details can start under main
			fc.ParentWidth -= (fc.DetailWidth/2 - c.SpouseSeparation/2)
		}

		if fc.SpouseHeight < (c.FamilyDetailsDrop)-2*(c.IndividualPadding) {
			// Details can extend under spouse
			fc.ParentWidth -= (fc.DetailWidth/2 - c.SpouseSeparation/2)
		}

		fc.ParentHeight = max(fc.ParentHeight, fc.DetailHeight+(c.FamilyDetailsDrop)+(c.IndividualPadding)+c.IndividualNameTextHeight/2+12+c.LineWidth+(c.TextVertSpacing))
	}

	fc.Children.Measure(c)
	fc.MinWidth = max(fc.ParentWidth, fc.Children.MinWidth)
	fc.MinHeight = fc.ParentHeight + fc.Children.MinHeight
}

// Returns the offset of the main individual from the left of the container
func (fc *FamilyContainer) MainOffset(c *Chart) float64 {
	if len(fc.Children.Individuals) == 0 {
		return 0
	}
	return fc.MinWidth/2 - fc.ParentWidth/2
}

func (fc *FamilyContainer) Draw(x, y float64, c *Chart) {
	if c.DebugFamilyOutline {
		c.debugDrawBox(x, y, fc.MinWidth, fc.MinHeight, canvas.Green)
	}

	if fc.DrawVertical {
		// Draw the main individuals in the family vertically
		xMain := x + fc.MainOffset(c)
		xSpouse := xMain
		xDetails := xMain

		// fc.Main is nil for a second or later marriage
		if fc.Main != nil {
			c.DrawIndividualText(xMain, y, fc.Main.Individual)
			y += fc.MainHeight
		}

		y += c.SpouseVerticalSeparation
		familyDetailsText := c.MakeText("", fc.Family.Details)
		c.gc.DrawText(xDetails, c.PageHeight-y, familyDetailsText)

		y += fc.DetailHeight
		y += c.SpouseDetailVerticalSeparation
		c.DrawIndividualText(xSpouse, y, fc.Spouse.Individual)
		return

	}

	xMain := x + fc.MainOffset(c)
	xSpouse := xMain + fc.ParentWidth - fc.SpouseWidth
	xLine := xSpouse - c.SpouseSeparation/2
	if fc.SpouseHeight >= (c.FamilyDetailsDrop)-2*(c.IndividualPadding) {
		xLine = xSpouse - fc.DetailWidth/2
	}

	xDetails := xLine - fc.DetailWidth/2

	if fc.Main != nil {
		c.DrawIndividualText(xMain, y, fc.Main.Individual)
	}
	c.DrawIndividualText(xSpouse, y, fc.Spouse.Individual)

	// Draw marriage symbol and descending line
	c.gc.SetStrokeColor(c.LineColor)
	c.gc.SetStrokeWidth(c.LineWidth)

	c.DrawLine(
		xLine-c.LineWidth*2, y+(c.IndividualPadding)+c.IndividualNameTextHeight*0.6-c.LineWidth,
		xLine+c.LineWidth*2, y+(c.IndividualPadding)+c.IndividualNameTextHeight*0.6-c.LineWidth,
	)

	c.DrawLine(
		xLine-c.LineWidth*2, y+(c.IndividualPadding)+c.IndividualNameTextHeight*0.6+c.LineWidth,
		xLine+c.LineWidth*2, y+(c.IndividualPadding)+c.IndividualNameTextHeight*0.6+c.LineWidth,
	)

	y += (c.IndividualPadding) + c.IndividualNameTextHeight/2 + 12

	if len(fc.Family.Details) > 0 || fc.HasChildren() {

		lineHeight := 0.0
		if len(fc.Family.Details) > 0 {
			c.DrawLine(
				xLine, y,
				xLine, y+(c.FamilyDetailsDrop)+c.LineWidth/2,
			)

			y += (c.FamilyDetailsDrop)

			// Draw marriage details
			c.gc.DrawText(xDetails, c.PageHeight-y, c.MakeText("", fc.Family.Details))

			y += fc.DetailHeight
			lineHeight = (c.ChildLineDropAbove)
		} else {
			// Drop 2 lines
			lineHeight = (c.FamilyDetailsDrop) + c.IndividualDetailsTextHeight*2.0 + (c.TextVertSpacing) + (c.ChildLineDropAbove)
		}

		if fc.Children != nil && len(fc.Children.Individuals) > 0 {
			// Draw desending line from marriage details
			c.DrawLine(
				xLine, y,
				xLine, y+lineHeight+c.LineWidth/2,
			)

			y += lineHeight

			xChildren := x + fc.MinWidth/2 - fc.Children.MinWidth/2

			fc.Children.Draw(xChildren, y, xLine, c)
		}
	}

	c.gc.Pop()
}

type ChildrenContainer struct {
	Individuals []*IndividualContainer
	MinWidth    float64
	MinHeight   float64
	Horizontal  bool
}

func (cc *ChildrenContainer) Measure(c *Chart) {
	if cc == nil || len(cc.Individuals) == 0 {
		return
	}

	// Decide if this container should be laid out with people in a horizontal row or a vertical column
	if c.ForceHorizontal || len(cc.Individuals) == 1 {
		cc.Horizontal = true
	} else {
		// Assume vertical
		cc.Horizontal = false

	indivloop:
		for _, ic := range cc.Individuals {
			// Family groups containing a direct ancestor should be horizontal
			if ic.Individual.Direct {
				cc.Horizontal = true
				break indivloop
			}
			// Family groups containing marriages with children should be horizontal
			for _, f := range ic.Families {
				if f.Spouse != nil && f.Spouse.Individual.Direct {
					cc.Horizontal = true
					break indivloop
				}
				if len(f.Children.Individuals) > 0 {
					cc.Horizontal = true
					break indivloop
				}

			}
		}
	}

	if cc.Horizontal {
		// pullAvailable := 0.0
		// pushRequired := 0.0
		// closeUpHeight := 0.0
		// leftUnderhang := 0.0

		for _, ic := range cc.Individuals {
			ic.Measure(c)

			// // In some instances we can close up the gap between siblings
			// if len(ic.Families) > 0 && len(ic.Families[0].Children.Individuals) > 0 {
			// 	// Families with children can slip under siblings on their left without children

			// 	if pushRequired > 0 && pullAvailable == 0 {
			// 		// Previous childless sibling has been slipped over top of an earlier sibling
			// 		// Push current person out to the right so they don't overlap
			// 		ic.LeftPull -= pushRequired
			// 	} else if i > 0 {

			// 		underAllSiblings := true
			// 		// Pull under all siblings to left that are not too tall
			// 		for j := i - 1; j >= 0; j-- {
			// 			if cc.Individuals[j].MinHeight > ic.Families[0].ParentHeight+(c.ChildLineDropAbove) {
			// 				underAllSiblings = false
			// 				break
			// 			} else {
			// 				if ic.LeftPull+cc.Individuals[j].MinWidth+(c.IndividualSpacing) > ic.MainOffset(c) {
			// 					underAllSiblings = false
			// 					break
			// 				} else {
			// 					ic.LeftPull += cc.Individuals[j].MinWidth + (c.IndividualSpacing)
			// 				}
			// 			}
			// 		}

			// 		if underAllSiblings && ic.LeftPull < ic.MainOffset(c) {
			// 			// This individual has slipped under all siblings to left
			// 			// Pull it as far left as possible
			// 			leftUnderhang = ic.MainOffset(c) - ic.LeftPull
			// 			ic.LeftPull = ic.MainOffset(c)
			// 		}

			// 	}

			// 	closeUpHeight = ic.Families[0].ParentHeight
			// 	allParentWidth := 0.0

			// 	for _, f := range ic.Families {
			// 		allParentWidth += f.ParentWidth + (c.FamilySpacing)
			// 		closeUpHeight = min(closeUpHeight, f.ParentHeight)
			// 	}

			// 	// Calculate how much space is left for siblings to the right to slide over this
			// 	// person's children
			// 	pullAvailable = ic.MinWidth - ic.MainOffset(c) - allParentWidth
			// 	pushRequired = pullAvailable

			// } else {
			// Current person does not have children

			// 		if ic.MinHeight <= closeUpHeight {
			// 			if pullAvailable > 0 {
			// 				ic.LeftPull = pullAvailable
			// 				pullAvailable = 0
			// 			}

			// 			if pushRequired > 0 {
			// 				pushRequired -= ic.MinWidth
			// 			}
			// 		} else {
			// 			ic.LeftPull -= pushRequired
			// 			pushRequired = 0
			// 			pullAvailable = 0
			// 		}
			// 	}
			// 	if pushRequired < 0.0 {
			// 		pushRequired = 0.0
			// 	}
		}

		// cc.Individuals[0].LeftPull = -leftUnderhang
		// cc.MinWidth = pushRequired - pullAvailable
		for _, ic := range cc.Individuals {
			cc.MinWidth += ic.MinWidth - ic.LeftPull
			cc.MinHeight = max(cc.MinHeight, ic.MinHeight)
		}
		// Allow some space between each individual
		if len(cc.Individuals) > 1 {
			cc.MinWidth += c.IndividualSpacing * float64(len(cc.Individuals)-1)
		}

		cc.MinHeight += (c.ChildLineDropBelow) + c.LineWidth
	} else {
		// Drawing children vertically
		for _, ic := range cc.Individuals {
			ic.Measure(c)
			cc.MinWidth = max(cc.MinWidth, ic.MinWidth)
			cc.MinHeight += ic.MinHeight
		}
		// Allow some space between each individual
		if len(cc.Individuals) > 1 {
			cc.MinHeight += c.IndividualVerticalSpacing * float64(len(cc.Individuals)-1)
		}

		cc.MinWidth += (c.ChildLineDropBelow) + c.LineWidth + (c.IndividualSpacing)
		cc.MinHeight += (c.ChildLineDropAbove) + (c.ChildLineDropBelow) + c.LineWidth
	}
}

func (cc *ChildrenContainer) Draw(x, y float64, xParentLine float64, c *Chart) {
	if c.DebugChildrenOutline {
		c.debugDrawBox(x, y, cc.MinWidth, cc.MinHeight, canvas.Blue)
	}

	// c.gc.Push()
	c.gc.SetStrokeColor(c.LineColor)
	c.gc.SetStrokeWidth(c.LineWidth)

	if cc.Horizontal {
		xLineLast := xParentLine
		for _, ic := range cc.Individuals {
			x -= ic.LeftPull

			xLine := x + ic.MainOffset(c) + (c.ChildLineOffset)

			// Draw descender line to child

			c.DrawLine(
				xLine, y-c.LineWidth/2,
				xLine, y+(c.ChildLineDropBelow)+c.LineWidth/2,
			)

			// Draw horizontal connnector to previous child
			c.DrawLine(
				xLine-c.LineWidth/2, y,
				xLineLast+c.LineWidth/2, y,
			)

			xLineLast = xLine

			ic.Draw(x, y+(c.ChildLineDropBelow), c)

			x += ic.MinWidth + (c.IndividualSpacing)

		}
	} else {
		// Draw in a vertical orientation

		yLineOffset := c.IndividualNameTextHeight * 0.66

		// Draw horizontal connnector to parent line
		c.DrawLine(
			x, y,
			xParentLine, y,
		)

		yLineLast := y
		y += (c.ChildLineDropBelow)

		for _, ic := range cc.Individuals {
			// Draw descender line
			c.DrawLine(
				x, yLineLast-c.LineWidth/2,
				x, y+yLineOffset+c.LineWidth/2,
			)

			// Draw line to child
			c.DrawLine(
				x-c.LineWidth/2, y+yLineOffset,
				x+(c.ChildLineDropBelow)+c.LineWidth/2-c.IndividualPadding, y+yLineOffset,
			)

			ic.Draw(x+(c.ChildLineDropBelow), y, c)

			yLineLast = y + yLineOffset
			y += ic.MinHeight + c.IndividualVerticalSpacing

		}
	}
	// c.gc.Pop()
}

type TextStyle struct {
	Face       *canvas.FontFace
	Style      canvas.FontStyle
	Size       float64 // in points
	Color      color.Color
	LineHeight float64 // in mm
}

func NewTextStyle(name string, size float64, color color.Color, style canvas.FontStyle) (*TextStyle, error) {
	ts := &TextStyle{
		Size:  size,
		Color: color,
		Style: style,
	}

	font := canvas.NewFontFamily("georgia")
	if err := font.LoadFontFile("/usr/share/fonts/truetype/msttcorefonts/georgia.ttf", canvas.FontRegular); err != nil {
		panic(err)
	}
	if err := font.LoadFontFile("/usr/share/fonts/truetype/msttcorefonts/georgiab.ttf", canvas.FontBold); err != nil {
		panic(err)
	}

	ts.Face = font.Face(size, color, canvas.FontBold, canvas.FontNormal)
	t := canvas.NewTextLine(ts.Face, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", canvas.Left)
	bounds := t.Bounds()
	ts.LineHeight = bounds.H

	return ts, nil
}

func debug(v ...interface{}) {
	fmt.Println(v...)
}

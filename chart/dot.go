package chart

import (
	"bytes"
	"fmt"
	"strings"
)

// Note: instead of dot, try https://pikchr.org
// Go port: https://github.com/gopikchr/gopikchr

type DotBuffer struct {
	buf    bytes.Buffer
	indent int
	err    error
}

func (d *DotBuffer) Println(s string) {
	if d.err != nil {
		return
	}
	if _, err := d.buf.WriteString(strings.Repeat("\t", d.indent)); err != nil {
		d.err = err
		return
	}
	if _, err := d.buf.WriteString(s); err != nil {
		d.err = err
		return
	}
	if _, err := d.buf.WriteString("\n"); err != nil {
		d.err = err
		return
	}
}

func (d *DotBuffer) Printf(f string, args ...interface{}) {
	if d.err != nil {
		return
	}
	d.Println(fmt.Sprintf(f, args...))
}

func (d *DotBuffer) Indent() {
	if d.err != nil {
		return
	}
	d.indent++
}

func (d *DotBuffer) Unindent() {
	if d.err != nil {
		return
	}
	d.indent--
}

type DotEdge struct {
	From    string
	To      string
	Comment string
	Attrs   []DotAttrFunc
}

type DotNode struct {
	ID      string
	Comment string
	Attrs   []DotAttrFunc
}

type DotAttrFunc func(b *DotBuffer)

func DotKV(k, v string) DotAttrFunc {
	return func(b *DotBuffer) {
		b.Printf(`%s=%s`, k, v)
	}
}

type DotChart struct {
	nodes       []*DotNode
	edges       []*DotEdge
	generations map[int][]string
}

func NewDotChart() *DotChart {
	return &DotChart{
		generations: make(map[int][]string),
	}
}

func (d *DotChart) Draw(root *Individual) ([]byte, error) {
	d.descend(root, 0, make(map[string]bool))

	b := &DotBuffer{}

	b.Printf(`digraph {`)
	b.Indent()
	b.Println(`// top to bottom layout`)
	b.Println(`rankdir="TB"`)
	b.Println(`// Use straight edges`)
	b.Println(`splines=ortho`)

	b.Println(``)
	b.Println(`graph [`)
	b.Indent()
	b.Println(`center=true`)
	b.Println(`margin=0.2`)
	b.Println(`nodesep=0.1`)
	b.Println(`ranksep=0.2`)
	b.Unindent()
	b.Println(`]`)

	b.Println(``)
	b.Println(`node [`)
	b.Indent()
	b.Println(`shape=none`)
	b.Println(`labelloc=t`)
	b.Unindent()
	b.Println(`]`)

	b.Println(``)
	b.Println(`edge [`)
	b.Indent()
	b.Println(`arrowhead=none`)
	b.Println(`penwidth=2`)
	b.Println(`color="#999999"`)
	b.Unindent()
	b.Println(`]`)

	for _, n := range d.nodes {
		b.Println(``)
		if n.Comment != "" {
			b.Printf(`// %s`, n.Comment)
		}
		b.Printf(`%s [`, n.ID)
		b.Indent()
		for _, afn := range n.Attrs {
			afn(b)
		}
		b.Unindent()
		b.Println(`]`)
	}

	for _, e := range d.edges {
		b.Println(``)
		if e.Comment != "" {
			b.Printf(`// %s`, e.Comment)
		}
		b.Printf(`%s -> %s [`, e.From, e.To)
		b.Indent()
		for _, afn := range e.Attrs {
			afn(b)
		}
		b.Unindent()
		b.Println(`]`)
	}

	for gen, ids := range d.generations {
		b.Println(``)
		b.Printf(`subgraph gen%d {`, gen)
		b.Indent()
		b.Println(`rank="same"`)
		for _, id := range ids {
			b.Println(id)
		}
		b.Unindent()
		b.Println(`}`)
	}

	b.Unindent()
	b.Println(`}`)

	if b.err != nil {
		return nil, b.err
	}
	return b.buf.Bytes(), nil
}

func (d *DotChart) descend(in *Individual, generation int, done map[string]bool) {
	if !done[in.ID] {
		d.addIndividualNode(in)
		d.generations[generation] = append(d.generations[generation], "in_"+in.ID)
		done[in.ID] = true
	}

	for index, fam := range in.Families {
		if index != 0 {
			continue
		}
		relNode, detailsNode := d.addFamilyNodes(in, fam, index+1)
		d.generations[generation] = append(d.generations[generation], relNode.ID)

		d.edges = append(d.edges, &DotEdge{
			From: "in_" + in.ID + ":e",
			To:   relNode.ID + ":w",
			Attrs: []DotAttrFunc{
				DotKV("color", `"black:invis:black"`),
			},
		})

		d.edges = append(d.edges, &DotEdge{
			From: relNode.ID + ":s",
			To:   detailsNode.ID + ":n",
		})

		if fam.Spouse != nil {

			d.edges = append(d.edges, &DotEdge{
				From: relNode.ID + ":e",
				To:   "in_" + fam.Spouse.ID + ":w",
				Attrs: []DotAttrFunc{
					DotKV("color", `"black:invis:black"`),
				},
			})

			d.descend(fam.Spouse, generation, done)
		}

		if len(fam.Children) == 0 {
			continue
		}
		famNode := &DotNode{
			ID:      fmt.Sprintf("fam_%s_%s_fam", in.ID, fam.ID),
			Comment: in.Name,
			Attrs: []DotAttrFunc{
				DotKV("shape", "point"),
				DotKV("style", "invis"),
			},
		}
		d.nodes = append(d.nodes, famNode)

		d.edges = append(d.edges, &DotEdge{
			From: detailsNode.ID + ":s",
			To:   famNode.ID + ":n",
			Attrs: []DotAttrFunc{
				DotKV("headclip", "false"),
			},
		})

		for _, ch := range fam.Children {
			d.descend(ch, generation+1, done)
			d.edges = append(d.edges, &DotEdge{
				From: famNode.ID + ":s",
				To:   "in_" + ch.ID + ":n",
				Attrs: []DotAttrFunc{
					DotKV("tailclip", "false"),
				},
			})

		}
	}
}

func (d *DotChart) addIndividualNode(in *Individual) *DotNode {
	n := &DotNode{
		ID:      "in_" + in.ID,
		Comment: in.Name,
		Attrs: []DotAttrFunc{
			DotKV("shape", "none"),
			func(b *DotBuffer) {
				b.Println(`label=<`)
				b.Indent()
				b.Println(`<TABLE CELLBORDER="0" BORDER="0">`)
				b.Indent()
				b.Printf(`<TR><TD ALIGN="LEFT" VALIGN="TOP" PORT="name"><FONT POINT-SIZE="14"><B>%s</B></FONT></TD></TR>`, in.Name)
				// for _, det := range in.Details {
				// 	b.Printf(`<TR><TD ALIGN="LEFT" VALIGN="TOP"><FONT POINT-SIZE="12">%s</FONT></TD></TR>`, det)
				// }
				b.Unindent()
				b.Println(`</TABLE>`)
				b.Unindent()
				b.Println(`>`)
			},
		},
	}

	d.nodes = append(d.nodes, n)
	return n
}

func (d *DotChart) addFamilyNodes(in *Individual, fam *Family, index int) (*DotNode, *DotNode) {
	relNode := &DotNode{
		ID:      fmt.Sprintf("fam_%s_%s_rel", in.ID, fam.ID),
		Comment: in.Name,
		Attrs: []DotAttrFunc{
			DotKV("shape", "none"),
			func(b *DotBuffer) {
				b.Println(`label=<`)
				b.Indent()
				b.Println(`<TABLE CELLBORDER="0" BORDER="0">`)
				b.Indent()
				b.Printf(`<TR><TD ALIGN="LEFT" VALIGN="TOP"><FONT POINT-SIZE="12">(%d)</FONT></TD></TR>`, index)
				b.Unindent()
				b.Println(`</TABLE>`)
				b.Unindent()
				b.Println(`>`)
			},
		},
	}
	d.nodes = append(d.nodes, relNode)

	detailsNode := &DotNode{
		ID:      fmt.Sprintf("fam_%s_%s_details", in.ID, fam.ID),
		Comment: in.Name,
		Attrs: []DotAttrFunc{
			DotKV("shape", "none"),
			func(b *DotBuffer) {
				b.Println(`label=<`)
				b.Indent()
				b.Println(`<TABLE CELLBORDER="0" BORDER="0">`)
				b.Indent()
				b.Printf(`<TR><TD ALIGN="LEFT" VALIGN="TOP"><FONT POINT-SIZE="12">m. 14 Dec 1999</FONT></TD></TR>`)
				b.Unindent()
				b.Println(`</TABLE>`)
				b.Unindent()
				b.Println(`>`)
			},
		},
	}
	d.nodes = append(d.nodes, detailsNode)

	return relNode, detailsNode
}

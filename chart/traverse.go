package chart

import (
	"github.com/iand/genster/model"
	"github.com/iand/gtree"
)

func descendants(p *model.Person, seq *sequence, generations int, directOnly bool, compact bool, personDetailFn func(*model.Person) ([]string, []string), familyDetailFn func(*model.Family) []string) *gtree.DescendantPerson {
	headings, details := personDetailFn(p)

	tp := &gtree.DescendantPerson{ID: seq.next(), Headings: headings, Details: details}
	if !directOnly || p.IsDirectAncestor() {
		if generations > 0 {
			for _, f := range p.Families {
				tf := new(gtree.DescendantFamily)
				tp.Families = append(tp.Families, tf)
				// Show spouses separately unless compact has been requested
				if !compact {
					tf.Details = familyDetailFn(f)
					o := f.OtherParent(p)
					if o != nil {
						oh, od := personDetailFn(o)
						tf.Other = &gtree.DescendantPerson{ID: seq.next(), Headings: oh, Details: od}
					}
				}
				// TODO: sort by date
				for _, c := range f.Children {
					tf.Children = append(tf.Children, descendants(c, seq, generations-1, directOnly, compact, personDetailFn, familyDetailFn))
				}
			}
		}
		// Show both parents of direct ancestor at bottom of tree
		if p.IsDirectAncestor() && generations == 0 {
			for _, f := range p.Families {
				if f.OtherParent(p).IsDirectAncestor() {
					tf := new(gtree.DescendantFamily)
					tf.Details = familyDetailFn(f)
					tp.Families = append(tp.Families, tf)
					oh, od := personDetailFn(f.OtherParent(p))
					tf.Other = &gtree.DescendantPerson{ID: seq.next(), Headings: oh, Details: od}
					break
				}
			}
		}
	}

	return tp
}

func ancestors(p *model.Person, seq *sequence, generation int, maxGeneration int, personDetailFn func(*model.Person, int) []string) *gtree.AncestorPerson {
	tp := &gtree.AncestorPerson{ID: seq.next(), Headings: []string{p.PreferredFullName}, Details: personDetailFn(p, generation)}
	if generation < maxGeneration {
		if p.Father != nil {
			tp.Father = ancestors(p.Father, seq, generation+1, maxGeneration, personDetailFn)
		}
		if p.Mother != nil {
			tp.Mother = ancestors(p.Mother, seq, generation+1, maxGeneration, personDetailFn)
		}
	}
	return tp
}

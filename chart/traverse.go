package chart

import (
	"github.com/iand/genster/model"
	"github.com/iand/gtree"
)

func descendants(p *model.Person, seq *sequence, generations int, directOnly bool, personDetailFn func(*model.Person) []string, familyDetailFn func(*model.Family) []string) *gtree.DescendantPerson {
	tp := &gtree.DescendantPerson{ID: seq.next(), Details: personDetailFn(p)}
	if !directOnly || p.IsDirectAncestor() {
		if generations > 0 {
			for _, f := range p.Families {
				tf := new(gtree.DescendantFamily)
				tf.Details = familyDetailFn(f)
				tp.Families = append(tp.Families, tf)
				o := f.OtherParent(p)
				if o != nil {
					tf.Other = &gtree.DescendantPerson{ID: seq.next(), Details: personDetailFn(o)}
				}
				// TODO: sort by date
				for _, c := range f.Children {
					tf.Children = append(tf.Children, descendants(c, seq, generations-1, directOnly, personDetailFn, familyDetailFn))
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
					tf.Other = &gtree.DescendantPerson{ID: seq.next(), Details: personDetailFn(f.OtherParent(p))}
					break
				}
			}
		}
	}

	return tp
}

func ancestors(p *model.Person, seq *sequence, generation int, maxGeneration int, personDetailFn func(*model.Person, int) []string) *gtree.AncestorPerson {
	tp := &gtree.AncestorPerson{ID: seq.next(), Details: personDetailFn(p, generation)}
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

package chart

import (
	"github.com/iand/genster/model"
	"github.com/iand/gtree"
)

func descendants(p *model.Person, seq *sequence, generations int, directOnly bool, personDetailFn func(*model.Person) []string, familyDetailFn func(*model.Family) []string) *gtree.Person {
	tp := &gtree.Person{ID: seq.next(), Details: personDetailFn(p)}
	if !directOnly || p.IsDirectAncestor() {
		if generations > 0 {
			for _, f := range p.Families {
				tf := new(gtree.Family)
				tf.Details = familyDetailFn(f)
				tp.Families = append(tp.Families, tf)
				o := f.OtherParent(p)
				if o != nil {
					tf.Other = &gtree.Person{ID: seq.next(), Details: personDetailFn(o)}
				}
				// TODO: sort by date
				for _, c := range f.Children {
					tf.Children = append(tf.Children, descendants(c, seq, generations-1, directOnly, personDetailFn, familyDetailFn))
				}
			}
		}
	}

	return tp
}

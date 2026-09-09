package genbio

import (
	"strconv"

	"github.com/iand/genster/model"
	"github.com/iand/genster/text"
)

// Bio returns a plain-text short biography of s.
func Bio(s Subject) string {
	return bio(s, NewVariant(s.seed()))
}

// BioFromPerson returns a plain-text short biography of p.
func BioFromPerson(p *model.Person) string {
	return Bio(FromPerson(p))
}

// phraser selects one phrasing from a set of alternatives. The production
// implementation is Variant; tests supply deterministic stand-ins.
type phraser interface {
	Pick(alternatives ...string) string
}

// bio assembles the biography paragraph using v to choose phrasing.
func bio(s Subject, v phraser) string {
	b := &builder{s: s, v: v, para: new(text.Para)}
	b.addBirth()
	b.addFamily()
	b.addDeath()
	return b.para.Text()
}

// builder accumulates the sentences of a single biography.
type builder struct {
	s     Subject
	v     phraser
	para  *text.Para
	named bool
}

// subjectVerb returns the subject of the next sentence joined to the supplied
// verb, using the full name on first mention and a pronoun thereafter. The verb
// is given in singular and plural forms to agree with a singular pronoun or the
// plural "they".
func (b *builder) subjectVerb(singular, plural string) string {
	if !b.named {
		b.named = true
		return text.JoinSentenceParts(b.s.PreferredFullName, singular)
	}
	verb := singular
	if b.s.Gender.IsUnknown() {
		verb = plural
	}
	return text.JoinSentenceParts(b.s.Gender.SubjectPronoun(), verb)
}

func (b *builder) addBirth() {
	ev := b.s.BestBirthlikeEvent
	if ev == nil {
		return
	}
	date := ev.GetDate()
	place := ev.GetPlace()
	hasDate := date != nil && !date.IsUnknown()
	hasPlace := place != nil && !place.IsUnknown()
	if !hasDate && !hasPlace {
		return
	}

	b.para.StartSentence(b.subjectVerb("was born", "were born"))
	if hasDate {
		b.para.Continue(date.When())
	}
	if hasPlace {
		b.para.Continue(place.Where())
	}
	b.para.FinishSentence()
}

func (b *builder) addFamily() {
	marriages := b.marriages()
	children := len(b.s.Children)
	if len(marriages) == 0 && children == 0 {
		return
	}

	if len(marriages) > 0 {
		b.para.StartSentence(b.subjectVerb("married", "married"))
		b.addSpouses(marriages)
		if children > 0 {
			b.para.Continue(b.v.Pick(
				"and had "+childCount(children),
				"and went on to have "+childCount(children),
			))
		}
		b.para.FinishSentence()
		return
	}

	b.para.StartSentence(b.subjectVerb("had", "had"), childCount(children))
	b.para.FinishSentence()
}

// addSpouses appends the spouse and marriage date of a first marriage, or a
// count when there was more than one marriage.
func (b *builder) addSpouses(marriages []marriage) {
	if len(marriages) == 1 {
		m := marriages[0]
		b.para.Continue(m.spouse)
		b.para.Continue(m.when)
		return
	}
	b.para.Continue(text.MultiplicativeAdverb(len(marriages)))
}

func (b *builder) addDeath() {
	ev := b.s.BestDeathlikeEvent
	if ev == nil {
		return
	}
	date := ev.GetDate()
	place := ev.GetPlace()
	hasDate := date != nil && !date.IsUnknown()
	hasPlace := place != nil && !place.IsUnknown()
	if !hasDate && !hasPlace {
		return
	}

	b.para.StartSentence(b.subjectVerb("died", "died"))
	if hasDate {
		b.para.Continue(date.When())
	}
	if hasPlace {
		b.para.Continue(place.Where())
	}
	if hasDate {
		if age, ok := b.s.AgeInYearsAt(date); ok && age >= 0 {
			b.para.AppendClause("aged " + strconv.Itoa(age))
		}
	}
	b.para.FinishSentence()
}

// marriage holds the phrasing components of a single marriage.
type marriage struct {
	spouse string
	when   string
}

// marriages returns the subject's marriages in family order.
func (b *builder) marriages() []marriage {
	var out []marriage
	for _, fam := range b.s.Families {
		if fam.Bond != model.FamilyBondMarried && fam.Bond != model.FamilyBondLikelyMarried {
			continue
		}
		var m marriage
		if other := fam.OtherParent(b.s.Person); other != nil && !other.IsUnknown() {
			m.spouse = other.PreferredFamiliarFullName
		}
		if fam.BestStartDate != nil && !fam.BestStartDate.IsUnknown() {
			m.when = fam.BestStartDate.When()
		}
		out = append(out, m)
	}
	return out
}

// childCount renders a number of children as a noun phrase.
func childCount(n int) string {
	if n == 1 {
		return "one child"
	}
	return text.CardinalNoun(n) + " children"
}

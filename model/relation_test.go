package model

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func makeSelfRelationFor(p *Person) *Relation {
	return &Relation{
		From:           p,
		To:             p,
		CommonAncestor: p,
		AncestorPath:   []*Person{p},
	}
}

func makeSelfRelation(id string) *Relation {
	return makeGenderedSelfRelation(id, GenderMale)
}

func makeGenderedSelfRelation(id string, gender Gender) *Relation {
	p := &Person{
		ID:     id,
		Gender: gender,
	}
	return makeSelfRelationFor(p)
}

func makeDirectAncestorRelation(path string) *Relation {
	return makeDirectAncestorRelationFrom(&Person{
		ID:     "self",
		Gender: GenderMale,
	}, path)
}

func makeDirectAncestorRelationFrom(from *Person, path string) *Relation {
	rel := &Relation{
		From: from,
	}

	rel.CommonAncestor = rel.From
	for _, r := range path {
		rel.FromGenerations++
		if r == 'm' {
			mother := &Person{
				ID:     rel.CommonAncestor.ID + "_m",
				Gender: GenderFemale,
			}
			rel.CommonAncestor.Mother = mother
			mother.Children = append(mother.Children, rel.CommonAncestor)
			rel.CommonAncestor = mother
		} else {
			father := &Person{
				ID:     rel.CommonAncestor.ID + "_f",
				Gender: GenderMale,
			}
			rel.CommonAncestor.Father = father
			father.Children = append(father.Children, rel.CommonAncestor)
			rel.CommonAncestor = father
		}
	}

	rel.To = rel.CommonAncestor

	return rel
}

func makeDirectDescendentRelation(path string, sex Gender) *Relation {
	rel := &Relation{
		To: &Person{
			Gender: sex,
		},
	}

	rel.CommonAncestor = rel.To
	for _, r := range path {
		rel.ToGenerations++
		if r == 'm' {
			mother := &Person{
				Gender: GenderFemale,
			}
			rel.CommonAncestor.Mother = mother
			mother.Children = append(mother.Children, rel.CommonAncestor)
			rel.CommonAncestor = mother
		} else {
			father := &Person{
				Gender: GenderMale,
			}
			rel.CommonAncestor.Father = father
			father.Children = append(father.Children, rel.CommonAncestor)
			rel.CommonAncestor = father
		}
	}

	rel.From = rel.CommonAncestor

	return rel
}

func makeCousinRelation(fromGenerations, toGenerations int) *Relation {
	return makeGenderedCousinRelation(fromGenerations, toGenerations, GenderMale)
}

func makeGenderedCousinRelation(fromGenerations, toGenerations int, sex Gender) *Relation {
	return &Relation{
		To: &Person{
			Gender: sex,
		},
		From: &Person{
			Gender: GenderMale,
		},
		CommonAncestor: &Person{
			Gender: GenderMale,
		},
		FromGenerations: fromGenerations,
		ToGenerations:   toGenerations,
	}
}

func makeSpouseRelation(closestDirectRelation *Relation, spouseRelation *Relation) *Relation {
	closestDirectRelation.To.Spouses = append(closestDirectRelation.To.Spouses, spouseRelation.From)
	spouseRelation.From.Spouses = append(spouseRelation.From.Spouses, closestDirectRelation.To)
	return &Relation{
		From:                  closestDirectRelation.From,
		To:                    spouseRelation.To,
		ClosestDirectRelation: closestDirectRelation,
		SpouseRelation:        spouseRelation,
	}
}

var testRelations = map[string]*Relation{
	"father":                           makeDirectAncestorRelation("f"),
	"mother":                           makeDirectAncestorRelation("m"),
	"grandfather":                      makeDirectAncestorRelation("ff"),
	"grandmother":                      makeDirectAncestorRelation("fm"),
	"great grandfather":                makeDirectAncestorRelation("mmf"),
	"great grandmother":                makeDirectAncestorRelation("mmm"),
	"son":                              makeDirectDescendentRelation("f", GenderMale),
	"daughter":                         makeDirectDescendentRelation("f", GenderFemale),
	"child":                            makeDirectDescendentRelation("f", GenderUnknown),
	"grandson":                         makeDirectDescendentRelation("mf", GenderMale),
	"granddaughter":                    makeDirectDescendentRelation("mf", GenderFemale),
	"grandchild":                       makeDirectDescendentRelation("mf", GenderUnknown),
	"great grandson":                   makeDirectDescendentRelation("mmf", GenderMale),
	"great granddaughter":              makeDirectDescendentRelation("mmf", GenderFemale),
	"great grandchild":                 makeDirectDescendentRelation("mmf", GenderUnknown),
	"first cousin":                     makeCousinRelation(2, 2),
	"second cousin":                    makeCousinRelation(3, 3),
	"third cousin":                     makeCousinRelation(4, 4),
	"fourth cousin":                    makeCousinRelation(5, 5),
	"fifth cousin":                     makeCousinRelation(6, 6),
	"sixth cousin":                     makeCousinRelation(7, 7),
	"seventh cousin":                   makeCousinRelation(8, 8),
	"eighth cousin":                    makeCousinRelation(9, 9),
	"ninth cousin":                     makeCousinRelation(10, 10),
	"uncle":                            makeGenderedCousinRelation(2, 1, GenderMale),
	"aunt":                             makeGenderedCousinRelation(2, 1, GenderFemale),
	"great uncle":                      makeGenderedCousinRelation(3, 1, GenderMale),
	"great aunt":                       makeGenderedCousinRelation(3, 1, GenderFemale),
	"great great uncle":                makeGenderedCousinRelation(4, 1, GenderMale),
	"great great aunt":                 makeGenderedCousinRelation(4, 1, GenderFemale),
	"first cousin twice removed":       makeCousinRelation(2, 4),
	"first cousin three times removed": makeCousinRelation(2, 5),
	"first cousin four times removed":  makeCousinRelation(2, 6),
	"second cousin once removed":       makeCousinRelation(3, 2),
	"husband":                          makeSpouseRelation(makeSelfRelation("self"), makeGenderedSelfRelation("husband", GenderMale)),
	"wife":                             makeSpouseRelation(makeSelfRelation("self"), makeGenderedSelfRelation("wife", GenderFemale)),
	"spouse":                           makeSpouseRelation(makeSelfRelation("self"), makeGenderedSelfRelation("spouse", GenderUnknown)),
	"father-in-law":                    makeSpouseRelation(makeSelfRelation("self"), makeDirectAncestorRelation("f")),
	"mother-in-law":                    makeSpouseRelation(makeSelfRelation("self"), makeDirectAncestorRelation("m")),
	"stepson":                          makeSpouseRelation(makeSelfRelation("self"), makeDirectDescendentRelation("m", GenderMale)),
	"stepdaughter":                     makeSpouseRelation(makeSelfRelation("self"), makeDirectDescendentRelation("m", GenderFemale)),
	"stepchild":                        makeSpouseRelation(makeSelfRelation("self"), makeDirectDescendentRelation("m", GenderUnknown)),

	"brother": {
		To: &Person{
			Gender: GenderMale,
		},
		From: &Person{
			Gender: GenderMale,
		},
		CommonAncestor: &Person{
			Gender: GenderMale,
		},
		FromGenerations: 1,
		ToGenerations:   1,
	},
	"sister": {
		To: &Person{
			Gender: GenderMale,
		},
		From: &Person{
			Gender: GenderFemale,
		},
		CommonAncestor: &Person{
			Gender: GenderMale,
		},
		FromGenerations: 1,
		ToGenerations:   1,
	},

	"sibling": {
		To: &Person{
			Gender: GenderMale,
		},
		From: &Person{},
		CommonAncestor: &Person{
			Gender: GenderMale,
		},
		FromGenerations: 1,
		ToGenerations:   1,
	},

	"step-mother": makeSpouseRelation(makeDirectAncestorRelation("f"), makeGenderedSelfRelation("wife", GenderFemale)),
	"step-father": makeSpouseRelation(makeDirectAncestorRelation("m"), makeGenderedSelfRelation("husband", GenderMale)),

	"wife of the grandfather":  makeSpouseRelation(makeDirectAncestorRelation("ff"), makeGenderedSelfRelation("wife", GenderFemale)),
	"wife of the first cousin": makeSpouseRelation(makeCousinRelation(2, 2), makeGenderedSelfRelation("wife", GenderFemale)),
}

func TestRelationName(t *testing.T) {
	for want, rel := range testRelations {
		t.Run(want, func(t *testing.T) {
			got := rel.Name()
			if got != want {
				t.Errorf("got %s, wanted %s", got, want)
			}
		})
	}
}

func TestRelationNameTree(t *testing.T) {
	id := &Person{ID: "id", Gender: GenderMale}
	id.RelationToKeyPerson = makeSelfRelationFor(id)

	md := &Person{ID: "md", Gender: GenderMale}
	md.RelationToKeyPerson = id.RelationToKeyPerson.ExtendToParent(md)

	sd := &Person{ID: "sd", Gender: GenderFemale}
	sd.RelationToKeyPerson = id.RelationToKeyPerson.ExtendToParent(sd)

	wc := &Person{ID: "wc", Gender: GenderMale}
	wc.RelationToKeyPerson = sd.RelationToKeyPerson.ExtendToParent(wc)

	fh := &Person{ID: "fh", Gender: GenderFemale}
	fh.RelationToKeyPerson = sd.RelationToKeyPerson.ExtendToParent(fh)

	sc := &Person{ID: "sc", Gender: GenderFemale}
	sc.RelationToKeyPerson = wc.RelationToKeyPerson.ExtendToChild(sc)

	jc := &Person{ID: "jc", Gender: GenderMale}
	jc.RelationToKeyPerson = wc.RelationToKeyPerson.ExtendToChild(jc)

	jwh := &Person{ID: "jwh", Gender: GenderMale}
	jwh.RelationToKeyPerson = fh.RelationToKeyPerson.ExtendToParent(jwh)

	jh := &Person{ID: "jh", Gender: GenderMale}
	jh.RelationToKeyPerson = jwh.RelationToKeyPerson.ExtendToChild(jh)

	testCases := []struct {
		p    *Person
		want string
	}{
		{id, "same person as"},
		{md, "father"},
		{sd, "mother"},
		{wc, "grandfather"},
		{fh, "grandmother"},
		{sc, "aunt"},
		{jc, "uncle"},
		{jwh, "great grandfather"},
		{jh, "great uncle"},
	}
	for _, tc := range testCases {
		t.Run(tc.p.ID, func(t *testing.T) {
			got := tc.p.RelationToKeyPerson.Name()
			if got != tc.want {
				t.Errorf("got %s, wanted %s", got, tc.want)
			}
		})
	}
}

func TestRelationsViaSpouse(t *testing.T) {
	testCases := []struct {
		rel  *Relation
		want string
	}{
		{
			rel:  makeSpouseRelation(makeGenderedSelfRelation("self", GenderFemale), makeGenderedSelfRelation("husband", GenderMale)),
			want: "husband",
		},
		{
			rel:  makeSpouseRelation(makeGenderedSelfRelation("self", GenderMale), makeGenderedSelfRelation("wife", GenderFemale)),
			want: "wife",
		},
		{
			rel:  makeSpouseRelation(makeSelfRelation("self"), makeGenderedSelfRelation("spouse", GenderUnknown)),
			want: "spouse",
		},
		{
			rel:  makeSpouseRelation(makeDirectAncestorRelation("ff"), makeGenderedSelfRelation("wife", GenderFemale)),
			want: "wife of the grandfather",
		},

		// TODO: Judith Margery King is first cousin once removed

	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			got := tc.rel.Name()
			if got != tc.want {
				t.Errorf("got %s, wanted %s", got, tc.want)
			}
		})
	}
}

func TestRelationDistanceByName(t *testing.T) {
	testCases := []struct {
		name string
		want int
	}{
		{
			name: "father",
			want: 1,
		},
		{
			name: "mother",
			want: 1,
		},
		{
			name: "grandfather",
			want: 2,
		},
		{
			name: "grandmother",
			want: 2,
		},
		{
			name: "great grandfather",
			want: 3,
		},
		{
			name: "great grandmother",
			want: 3,
		},
		{
			name: "son",
			want: 1,
		},
		{
			name: "daughter",
			want: 1,
		},
		{
			name: "child",
			want: 1,
		},
		{
			name: "grandson",
			want: 2,
		},
		{
			name: "granddaughter",
			want: 2,
		},
		{
			name: "grandchild",
			want: 2,
		},
		{
			name: "great grandson",
			want: 3,
		},
		{
			name: "great granddaughter",
			want: 3,
		},
		{
			name: "great grandchild",
			want: 3,
		},
		{
			name: "first cousin",
			want: 4,
		},
		{
			name: "second cousin",
			want: 7,
		},
		{
			name: "third cousin",
			want: 10,
		},
		{
			name: "fourth cousin",
			want: 13,
		},
		{
			name: "fifth cousin",
			want: 16,
		},
		{
			name: "sixth cousin",
			want: 19,
		},
		{
			name: "seventh cousin",
			want: 22,
		},
		{
			name: "eighth cousin",
			want: 25,
		},
		{
			name: "ninth cousin",
			want: 28,
		},
		{
			name: "uncle",
			want: 2,
		},
		{
			name: "aunt",
			want: 2,
		},
		{
			name: "great uncle",
			want: 3,
		},
		{
			name: "great aunt",
			want: 3,
		},
		{
			name: "great great uncle",
			want: 4,
		},
		{
			name: "great great aunt",
			want: 4,
		},
		{
			name: "first cousin twice removed",
			want: 8,
		},
		{
			name: "second cousin once removed",
			want: 5,
		},
		{
			name: "husband",
			want: 1,
		},
		{
			name: "wife",
			want: 1,
		},
		{
			name: "spouse",
			want: 1,
		},
		{
			name: "father-in-law",
			want: 2,
		},
		{
			name: "mother-in-law",
			want: 2,
		},
		{
			name: "stepson",
			want: 2,
		},
		{
			name: "stepdaughter",
			want: 2,
		},
		{
			name: "stepchild",
			want: 2,
		},

		{
			name: "first cousin three times removed",
			want: 10,
		},
		{
			name: "first cousin four times removed",
			want: 12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rel, ok := testRelations[tc.name]
			if !ok {
				t.Fatalf("relation %q not found", tc.name)
			}

			got := rel.Distance()
			if got != tc.want {
				t.Errorf("got %d, wanted %d", got, tc.want)
			}
		})
	}
}

func TestRelationExtendToSpouse(t *testing.T) {
	p := &Person{
		ID: "start",
	}

	self := &Relation{
		From:           p,
		To:             p,
		CommonAncestor: p,
	}

	sp := &Person{
		ID: "spouse",
	}

	spself := makeSelfRelationFor(sp)

	testCases := []struct {
		rel  *Relation
		want *Relation
	}{
		{
			rel: self,
		},
		{
			rel: makeDirectAncestorRelationFrom(p, "f"),
		},
		{
			rel: makeDirectAncestorRelationFrom(p, "ff"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.rel.Name(), func(t *testing.T) {
			got := tc.rel.ExtendToSpouse(sp)
			want := makeSpouseRelation(tc.rel, spself)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("ExtendToSpouse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRelationExtendToChild(t *testing.T) {
	p := &Person{
		ID: "start",
	}

	self := &Relation{
		From:           p,
		To:             p,
		CommonAncestor: p,
	}

	ch := &Person{
		ID: "child",
	}

	testCases := []struct {
		rel  *Relation
		want *Relation
	}{
		{
			rel: self,
			want: &Relation{
				From:            p,
				To:              ch,
				CommonAncestor:  p,
				FromGenerations: 0,
				ToGenerations:   1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.rel.Name(), func(t *testing.T) {
			got := tc.rel.ExtendToChild(ch)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ExtendToChild() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRelationExtendToParent(t *testing.T) {
	p := &Person{
		ID: "start",
	}

	self := &Relation{
		From:           p,
		To:             p,
		CommonAncestor: p,
	}

	pnew := &Person{
		ID: "new",
	}

	parent := &Person{
		ID: "parent",
	}

	testCases := []struct {
		rel  *Relation
		want *Relation
	}{
		{
			rel: self,
			want: &Relation{
				From:            p,
				To:              pnew,
				CommonAncestor:  pnew,
				FromGenerations: 1,
				ToGenerations:   0,
				AncestorPath:    []*Person{pnew},
			},
		},
		{
			rel: &Relation{
				From:            p,
				To:              parent,
				CommonAncestor:  parent,
				FromGenerations: 1,
				ToGenerations:   0,
				AncestorPath:    []*Person{parent},
			},
			want: &Relation{
				From:            p,
				To:              pnew,
				CommonAncestor:  pnew,
				FromGenerations: 2,
				ToGenerations:   0,
				AncestorPath:    []*Person{parent, pnew},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.rel.Name(), func(t *testing.T) {
			got := tc.rel.ExtendToParent(pnew)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ExtendToParent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsCloseToDirectAncestor(t *testing.T) {
	testCases := []struct {
		rel  *Relation
		want bool
	}{
		{
			rel:  testRelations["father"],
			want: true,
		},
		{
			rel:  testRelations["mother"],
			want: true,
		},
		{
			rel:  testRelations["grandfather"],
			want: true,
		},
		{
			rel:  testRelations["grandmother"],
			want: true,
		},
		{
			rel:  testRelations["great grandfather"],
			want: true,
		},
		{
			rel:  testRelations["great grandmother"],
			want: true,
		},
		{
			rel:  testRelations["child"],
			want: true,
		},
		{
			rel:  testRelations["spouse"],
			want: true,
		},
		{
			rel:  testRelations["wife of the grandfather"],
			want: true,
		},
		{
			rel:  testRelations["grandchild"],
			want: false,
		},
		{
			rel:  testRelations["great grandchild"],
			want: false,
		},
		{
			rel:  testRelations["first cousin"],
			want: false,
		},
		{
			rel:  testRelations["wife of the first cousin"],
			want: false,
		},
		{
			rel:  testRelations["second cousin"],
			want: false,
		},
		{
			rel:  testRelations["third cousin"],
			want: false,
		},
		{
			rel:  testRelations["uncle"],
			want: false,
		},
		{
			rel:  testRelations["aunt"],
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.rel.Name(), func(t *testing.T) {
			got := tc.rel.IsCloseToDirectAncestor()
			if got != tc.want {
				t.Errorf("got IsCloseToDirectAncestor()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	p := familyTree()

	testCases := []struct {
		rel  *Relation
		want []*Person
	}{
		{
			rel:  nil,
			want: nil,
		},
		{
			rel: p.RelationToKeyPerson,
			want: []*Person{
				p,
			},
		},
		{
			rel: p.Father.RelationToKeyPerson,
			want: []*Person{
				p,
				p.Father,
			},
		},
		{
			rel: p.Father.Father.RelationToKeyPerson,
			want: []*Person{
				p,
				p.Father,
				p.Father.Father,
			},
		},
		{
			rel: p.Father.Mother.Father.RelationToKeyPerson,
			want: []*Person{
				p,
				p.Father,
				p.Father.Mother,
				p.Father.Mother.Father,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.rel.Name(), func(t *testing.T) {
			got := tc.rel.Path()
			if len(got) != len(tc.want) {
				t.Errorf("got %d in path, want %d", len(got), len(tc.want))
				return
			}

			for i := range got {
				if !got[i].SameAs(tc.want[i]) {
					t.Errorf("got person %s at index %d, want %s", got[i].ID, i, tc.want[i].ID)
				}
			}
		})
	}
}

// familyTree returns a simple family tree with relations
func familyTree() *Person {
	addFather := func(ch *Person) *Person {
		p := &Person{ID: ch.ID + "f", Gender: GenderMale}
		ch.Father = p
		p.Children = append(p.Children, ch)
		p.RelationToKeyPerson = ch.RelationToKeyPerson.ExtendToParent(p)
		return p
	}
	addMother := func(ch *Person) *Person {
		p := &Person{ID: ch.ID + "m", Gender: GenderFemale}
		ch.Mother = p
		p.Children = append(p.Children, ch)
		p.RelationToKeyPerson = ch.RelationToKeyPerson.ExtendToParent(p)
		return p
	}

	p := &Person{ID: "p", Gender: GenderMale}
	p.RelationToKeyPerson = Self(p)

	pf := addFather(p) // father
	pm := addMother(p) // mother

	pff := addFather(pf) // paternal grandfather
	pfm := addMother(pf) // paternal grandmother
	pmf := addFather(pm) // maternal grandfather
	pmm := addMother(pm) // maternal grandmother

	pfff := addFather(pff) // paternal grandfather's father
	pffm := addMother(pff) // paternal grandfather's mother
	pfmf := addFather(pfm) // paternal grandmother's father
	pfmm := addMother(pfm) // paternal grandmother's mother
	pmff := addFather(pmf) // maternal grandfather's father
	pmfm := addMother(pmf) // maternal grandfather's mother
	pmmf := addFather(pmm) // maternal grandmother's father
	pmmm := addMother(pmm) // maternal grandmother's mother

	_ = pfff
	_ = pffm
	_ = pfmf
	_ = pfmm
	_ = pmff
	_ = pmfm
	_ = pmmf
	_ = pmmm

	return p
}

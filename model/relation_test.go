package model

import "testing"

func makeSelfRelation() *Relation {
	p := &Person{
		Gender: GenderMale,
	}

	return &Relation{
		From:           p,
		To:             p,
		CommonAncestor: p,
	}
}

func makeDirectAncestorRelation(path string) *Relation {
	rel := &Relation{
		From: &Person{
			Gender: GenderMale,
		},
	}

	rel.CommonAncestor = rel.From
	for _, r := range path {
		rel.FromGenerations++
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

func makeGenderedSpouseRelation(spouseRelation *Relation, toSpouseGenerations int, sex Gender) *Relation {
	return &Relation{
		From: spouseRelation.From,
		To: &Person{
			Gender: sex,
		},
		SpouseRelation:      spouseRelation,
		ToSpouseGenerations: toSpouseGenerations,
	}
}

func TestRelationsDirect(t *testing.T) {
	testCases := []struct {
		rel  *Relation
		want string
	}{
		{
			rel:  makeDirectAncestorRelation("f"),
			want: "father",
		},
		{
			rel:  makeDirectAncestorRelation("m"),
			want: "mother",
		},
		{
			rel:  makeDirectAncestorRelation("ff"),
			want: "grandfather",
		},
		{
			rel:  makeDirectAncestorRelation("mf"),
			want: "grandfather",
		},
		{
			rel:  makeDirectAncestorRelation("fm"),
			want: "grandmother",
		},
		{
			rel:  makeDirectAncestorRelation("mm"),
			want: "grandmother",
		},
		{
			rel:  makeDirectAncestorRelation("mmm"),
			want: "great grandmother",
		},
		{
			rel:  makeDirectAncestorRelation("fmm"),
			want: "great grandmother",
		},
		{
			rel:  makeDirectAncestorRelation("fmmf"),
			want: "great great grandfather",
		},
		{
			rel:  makeDirectDescendentRelation("f", GenderMale),
			want: "son",
		},
		{
			rel:  makeDirectDescendentRelation("m", GenderMale),
			want: "son",
		},
		{
			rel:  makeDirectDescendentRelation("f", GenderFemale),
			want: "daughter",
		},
		{
			rel:  makeDirectDescendentRelation("m", GenderFemale),
			want: "daughter",
		},
		{
			rel:  makeDirectDescendentRelation("mf", GenderMale),
			want: "grandson",
		},
		{
			rel:  makeDirectDescendentRelation("fm", GenderMale),
			want: "grandson",
		},
		{
			rel:  makeDirectDescendentRelation("mf", GenderFemale),
			want: "granddaughter",
		},
		{
			rel:  makeDirectDescendentRelation("fm", GenderFemale),
			want: "granddaughter",
		},
		{
			rel:  makeDirectDescendentRelation("mfm", GenderFemale),
			want: "great granddaughter",
		},
		{
			rel: &Relation{
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

			want: "brother",
		},
		{
			rel: &Relation{
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

			want: "sister",
		},
		{
			rel: &Relation{
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

			want: "sibling",
		},
		{
			rel:  makeCousinRelation(2, 2),
			want: "first cousin",
		},
		{
			rel:  makeCousinRelation(3, 3),
			want: "second cousin",
		},
		{
			rel:  makeCousinRelation(4, 4),
			want: "third cousin",
		},
		{
			rel:  makeCousinRelation(5, 5),
			want: "fourth cousin",
		},
		{
			rel:  makeCousinRelation(6, 6),
			want: "fifth cousin",
		},
		{
			rel:  makeCousinRelation(7, 7),
			want: "sixth cousin",
		},
		{
			rel:  makeCousinRelation(8, 8),
			want: "seventh cousin",
		},
		{
			rel:  makeCousinRelation(9, 9),
			want: "eighth cousin",
		},
		{
			rel:  makeCousinRelation(10, 10),
			want: "ninth cousin",
		},
		{
			rel:  makeGenderedCousinRelation(2, 1, GenderMale),
			want: "uncle",
		},
		{
			rel:  makeGenderedCousinRelation(2, 1, GenderFemale),
			want: "aunt",
		},
		{
			rel:  makeGenderedCousinRelation(3, 1, GenderMale),
			want: "great uncle",
		},
		{
			rel:  makeGenderedCousinRelation(3, 1, GenderFemale),
			want: "great aunt",
		},
		{
			rel:  makeGenderedCousinRelation(4, 1, GenderMale),
			want: "great great uncle",
		},
		{
			rel:  makeGenderedCousinRelation(4, 1, GenderFemale),
			want: "great great aunt",
		},
		{
			rel:  makeCousinRelation(2, 4),
			want: "first cousin twice removed",
		},
		{
			rel:  makeCousinRelation(3, 2),
			want: "second cousin once removed",
		},
		{
			rel:  makeCousinRelation(3, 2),
			want: "second cousin once removed",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), 0, GenderMale),
			want: "husband",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), 0, GenderFemale),
			want: "wife",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), 0, GenderUnknown),
			want: "spouse",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), 1, GenderMale),
			want: "father-in-law",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), 1, GenderFemale),
			want: "mother-in-law",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), 1, GenderUnknown),
			want: "parent-in-law",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), -1, GenderMale),
			want: "stepson",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), -1, GenderFemale),
			want: "stepdaughter",
		},
		{
			rel:  makeGenderedSpouseRelation(makeSelfRelation(), -1, GenderUnknown),
			want: "stepchild",
		},
		{
			rel:  makeGenderedSpouseRelation(makeDirectAncestorRelation("ff"), 0, GenderFemale),
			want: "wife of the grandfather",
		},
		{
			rel:  makeGenderedSpouseRelation(makeDirectAncestorRelation("fm"), 0, GenderMale),
			want: "husband of the grandmother",
		},
		// {
		// 	rel:  makeGenderedSpouseRelation(makeDirectAncestorRelation("m"), 0, SexMale),
		// 	want: "step-father",
		// },
		// {
		// 	rel:  makeGenderedSpouseRelation(makeDirectAncestorRelation("f"), 0, SexFemale),
		// 	want: "step-mother",
		// },
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

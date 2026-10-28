package narrative

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// A person with a known parent but no birth-like event gets a complete opening sentence.
func TestIntroStatementNoBirthEventHasVerb(t *testing.T) {
	father := &model.Person{ID: "f", PreferredUniqueName: "Charles Denney", PreferredFamiliarName: "Charles", Gender: model.GenderMale}
	p := &model.Person{ID: "p", PreferredGivenName: "Sarah", PreferredFamiliarName: "Sarah", Gender: model.GenderFemale, Father: father}

	s := &IntroStatement[md.Text]{Principal: p}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

	got := enc.String()
	want := "Sarah was the daughter of Charles Denney."
	if !strings.Contains(got, want) {
		t.Errorf("output does not contain %q\ngot:\n%s", want, got)
	}
}

func TestFamilyStatementSingleParentNoChildBirth(t *testing.T) {
	father := func() *model.Person {
		return &model.Person{ID: "f", PreferredGivenName: "Francisco", PreferredFamiliarName: "Francisco", Gender: model.GenderMale}
	}
	mother := func() *model.Person {
		return &model.Person{ID: "m", PreferredGivenName: "Jane", PreferredFamiliarName: "Jane", Gender: model.GenderFemale}
	}
	child := func(id, name string, g model.Gender) *model.Person {
		return &model.Person{ID: id, PreferredGivenName: name, PreferredFamiliarName: name, PreferredFullName: name + " Ximenes", Gender: g}
	}

	testCases := []struct {
		name       string
		principal  *model.Person
		family     func(p *model.Person) *model.Family
		nameMinSeq int
		want       string
		notWant    []string
	}{
		{
			name:      "father, unknown mother, unknown bond, one son",
			principal: father(),
			family: func(p *model.Person) *model.Family {
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnknown, Father: p, Children: []*model.Person{child("c1", "Joseph", model.GenderMale)}}
			},
			want:    "Francisco's only known child was a son:–",
			notWant: []string{"Had a", "son, Joseph"},
		},
		{
			name:      "father referred to by pronoun, one daughter",
			principal: father(),
			family: func(p *model.Person) *model.Family {
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnknown, Father: p, Children: []*model.Person{child("c1", "Catherine", model.GenderFemale)}}
			},
			nameMinSeq: 5,
			want:       "His only known child was a daughter:–",
			notWant:    []string{"Had a"},
		},
		{
			name:      "mother, unknown father, unmarried, one son",
			principal: mother(),
			family: func(p *model.Person) *model.Family {
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnmarried, Mother: p, Children: []*model.Person{child("c1", "Joseph", model.GenderMale)}}
			},
			want: "Jane's only known child was a son:–",
		},
		{
			name:      "father, known partner, unmarried, one son",
			principal: father(),
			family: func(p *model.Person) *model.Family {
				partner := &model.Person{ID: "p", PreferredFamiliarFullName: "Mary Smith", PreferredFamiliarName: "Mary", Gender: model.GenderFemale}
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnmarried, Father: p, Mother: partner, Children: []*model.Person{child("c1", "Joseph", model.GenderMale)}}
			},
			want:    "Francisco had a",
			notWant: []string{"Had a"},
		},
		{
			name:      "mother, unknown father, redacted daughter",
			principal: mother(),
			family: func(p *model.Person) *model.Family {
				c := &model.Person{ID: "c1", Redacted: true, Gender: model.GenderFemale, PreferredFullName: "(person born in the 1910s)", PreferredFamiliarName: "(person born in the 1910s)"}
				c.BestBirthlikeEvent = &model.BirthEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.UnknownDate()},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c},
				}
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnmarried, Mother: p, Children: []*model.Person{c}}
			},
			want:    "Jane had a daughter",
			notWant: []string{"who was born", "who."},
		},
		{
			name:      "mother, unknown father, son with undated birth and no place",
			principal: mother(),
			family: func(p *model.Person) *model.Family {
				c := child("c1", "Joseph", model.GenderMale)
				c.BestBirthlikeEvent = &model.BirthEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.UnknownDate()},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c},
				}
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnmarried, Mother: p, Children: []*model.Person{c}}
			},
			want:    "Jane had a son",
			notWant: []string{"who was born", "who."},
		},
		{
			name:      "mother, unknown father, son with dated birth keeps the clause",
			principal: mother(),
			family: func(p *model.Person) *model.Family {
				c := child("c1", "Joseph", model.GenderMale)
				c.BestBirthlikeEvent = &model.BaptismEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.Year(1840)},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c},
				}
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnmarried, Mother: p, Children: []*model.Person{c}}
			},
			want: "who was baptised in 1840",
		},
		{
			name:      "father, unknown mother, two sons",
			principal: father(),
			family: func(p *model.Person) *model.Family {
				return &model.Family{ID: "fam", Bond: model.FamilyBondUnknown, Father: p, Children: []*model.Person{
					child("c1", "Joseph", model.GenderMale),
					child("c2", "John", model.GenderMale),
				}}
			},
			want:    "Francisco had two sons:–",
			notWant: []string{"Had two"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.principal
			f := tc.family(p)
			p.Families = []*model.Family{f}

			s := &FamilyStatement[md.Text]{Principal: p, Family: f}
			intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}, NameMinSeq: tc.nameMinSeq}
			enc := &md.Content{}
			s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

			got := enc.String()
			if !strings.Contains(got, tc.want) {
				t.Errorf("output does not contain %q\ngot:\n%s", tc.want, got)
			}
			for _, nw := range tc.notWant {
				if strings.Contains(got, nw) {
					t.Errorf("output contains %q\ngot:\n%s", nw, got)
				}
			}
		})
	}
}

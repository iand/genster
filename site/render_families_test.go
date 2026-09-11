package site

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

func TestRenderPersonFamilies(t *testing.T) {
	p := &model.Person{PreferredUniqueName: "John Smith"}
	spouse := &model.Person{PreferredUniqueName: "Jane Doe"}
	child1 := &model.Person{PreferredUniqueName: "Mary Smith (1851–1920)", Notable: "The first to emigrate"}
	child2 := &model.Person{PreferredUniqueName: "William Smith (1853–1855)", Epithet: "a blacksmith"}

	married := &model.Family{
		Father:        p,
		Mother:        spouse,
		Bond:          model.FamilyBondMarried,
		BestStartDate: model.Year(1850),
		BestStartEvent: &model.MarriageEvent{
			GeneralEvent: model.GeneralEvent{Place: &model.Place{FullName: "Anytown"}},
		},
		BestEndDate:      model.Year(1861),
		EndReason:        model.FamilyEndReasonDivorce,
		Children:         []*model.Person{child1, child2},
		AllChildrenKnown: true,
	}

	// A union with an unknown other parent and a start date derived from a child's
	// birth rather than a recorded event.
	unknownPartner := &model.Family{
		Father:           &model.Person{Unknown: true},
		Mother:           p,
		Bond:             model.FamilyBondUnmarried,
		BestStartDate:    model.Year(1846),
		AllChildrenKnown: true,
	}

	p.Families = []*model.Family{married, unknownPartner}

	enc := &md.Content{}
	if err := RenderPersonFamilies(p, &model.POV{Person: p}, enc); err != nil {
		t.Fatalf("RenderPersonFamilies: %v", err)
	}

	got := enc.String()
	t.Logf("output:\n%s", got)

	wants := []string{
		`class="grouped-list"`,
		`<ol class="group-items">`,
		`<p class="group-note">`,
		"Spouse",
		"Jane Doe",
		"Married: 1850",
		"in Anytown",
		"Divorced: 1861",
		"Mary Smith (1851–1920)",
		"the first to emigrate",
		"William Smith (1853–1855)",
		"a blacksmith",
		"Unknown partner",
		"No known children",
	}
	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Errorf("output missing %q", w)
		}
	}

	// The unknown-partner union has no recorded start event, so its derived start
	// date is not shown as a union date.
	if strings.Contains(got, "1846") {
		t.Errorf("derived start date leaked into heading")
	}

	// The childless union's note is unnumbered.
	if strings.Contains(got, `<ol class="group-items">`+"\n<li><em>No known children") {
		t.Errorf("childless union rendered as an ordered list")
	}
}

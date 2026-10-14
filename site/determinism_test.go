package site

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

// TestGreatestChildrenTiesOrderedByID checks that people with the same number
// of children are chosen and listed in ID order.
func TestGreatestChildrenTiesOrderedByID(t *testing.T) {
	var first []string
	for run := range 10 {
		tr := tree.NewTree("test", nil, &tree.SurnameGroups{})
		for i := range 20 {
			p := tr.FindPerson("test", fmt.Sprintf("parent%02d", i))
			for j := range 3 {
				p.Children = append(p.Children, &model.Person{ID: fmt.Sprintf("c%02d%d", i, j)})
			}
		}
		ps, err := NewPublishSet(tr, func(*model.Person) bool { return true })
		if err != nil {
			t.Fatalf("NewPublishSet: %v", err)
		}
		var got []string
		for _, p := range ps.GreatestChildren(3) {
			got = append(got, p.ID)
		}
		if run == 0 {
			first = got
			var ids []string
			for id := range tr.People {
				ids = append(ids, id)
			}
			slices.Sort(ids)
			want := ids[:3]
			if !slices.Equal(got, want) {
				t.Fatalf("got %v, want the three lowest IDs %v", got, want)
			}
			continue
		}
		if !slices.Equal(got, first) {
			t.Fatalf("run %d: got %v, want %v", run, got, first)
		}
	}
}

// TestSortEventDaysTiesOrderedByText checks that calendar entries on the same
// day and year are ordered by their text.
func TestSortEventDaysTiesOrderedByText(t *testing.T) {
	days := []eventDay{
		{day: 2, year: 1900, text: "b"},
		{day: 1, year: 1900, text: "z"},
		{day: 2, year: 1900, text: "a"},
		{day: 2, year: 1850, text: "c"},
	}
	sortEventDays(days)
	var got []string
	for _, d := range days {
		got = append(got, d.text)
	}
	want := []string{"z", "c", "a", "b"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestSortChangesTiesOrderedByObject checks that changes at the same time are
// ordered by the object they concern.
func TestSortChangesTiesOrderedByObject(t *testing.T) {
	when := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	later := when.Add(time.Hour)
	changes := []*Change{
		{Time: when, Object: &model.Person{ID: "p2"}},
		{Time: when, Object: &model.Place{ID: "pl1"}},
		{Time: later, Object: &model.Person{ID: "p9"}},
		{Time: when, Object: &model.Person{ID: "p1"}},
		{Time: when, Object: &model.GeneralCitation{ID: "c1"}},
	}
	sortChanges(changes)
	var got []string
	for _, ch := range changes {
		got = append(got, changeObjectKey(ch.Object))
	}
	want := []string{"person/p9", "citation/c1", "person/p1", "person/p2", "place/pl1"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestBuildGedcomIndividualOrderIsDeterministic checks that the GEDCOM export
// lists individuals in the same order every time.
func TestBuildGedcomIndividualOrderIsDeterministic(t *testing.T) {
	var first []string
	for run := range 10 {
		tr := tree.NewTree("test", nil, &tree.SurnameGroups{})
		for i := range 30 {
			tr.FindPerson("test", fmt.Sprintf("p%02d", i))
		}
		s := NewSite("/trees/test/", tr)
		ps, err := NewPublishSet(tr, func(*model.Person) bool { return true })
		if err != nil {
			t.Fatalf("NewPublishSet: %v", err)
		}
		s.PublishSet = ps

		g, err := s.BuildGedcom()
		if err != nil {
			t.Fatalf("BuildGedcom: %v", err)
		}
		var got []string
		for _, ir := range g.Individual {
			got = append(got, ir.Xref)
		}
		if run == 0 {
			first = got
			continue
		}
		if !slices.Equal(got, first) {
			t.Fatalf("run %d: individual order differs from run 0", run)
		}
	}
}

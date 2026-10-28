package infer

import (
	"testing"

	"github.com/iand/genster/model"
)

// A person whose only dated event is qualified, such as "before 1805", is bounded by that
// date and cannot be alive now.
func TestInferPersonAliveOrDeadQualifiedEventDate(t *testing.T) {
	testCases := []struct {
		name      string
		date      *model.Date
		wantAlive bool
	}{
		{name: "before 1805", date: model.BeforeYear(1805), wantAlive: false},
		{name: "range 1800 to 1805", date: model.YearRange(1800, 1805), wantAlive: false},
		{name: "plain year 1805", date: model.Year(1805), wantAlive: false},
		{name: "after 1805 gives no upper bound", date: model.AfterYear(1805), wantAlive: true},
		{name: "before 1990", date: model.BeforeYear(1990), wantAlive: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "p"}
			p.Timeline = []model.TimelineEvent{
				&model.OccupationEvent{
					GeneralEvent:           model.GeneralEvent{Date: tc.date},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
				},
			}

			if err := InferPersonAliveOrDead(p, 2026); err != nil {
				t.Fatalf("InferPersonAliveOrDead: %v", err)
			}
			if p.PossiblyAlive != tc.wantAlive {
				t.Errorf("PossiblyAlive: got %v, want %v", p.PossiblyAlive, tc.wantAlive)
			}
		})
	}
}

// A person with no dated events of their own is bounded by the birth years of their spouse
// and parents.
func TestInferPersonAliveOrDeadRelativesBound(t *testing.T) {
	born := func(y int) *model.Person {
		r := &model.Person{ID: "r"}
		r.BestBirthlikeEvent = &model.BirthEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(y)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: r}}
		return r
	}

	testCases := []struct {
		name      string
		build     func(p *model.Person)
		wantAlive bool
	}{
		{
			name: "wife born 1790",
			build: func(p *model.Person) {
				p.Families = []*model.Family{{ID: "f", Father: p, Mother: born(1790)}}
			},
			wantAlive: false,
		},
		{
			name: "wife born 1990",
			build: func(p *model.Person) {
				p.Families = []*model.Family{{ID: "f", Father: p, Mother: born(1990)}}
			},
			wantAlive: true,
		},
		{
			name:      "mother born 1881",
			build:     func(p *model.Person) { p.Mother = born(1881) },
			wantAlive: true,
		},
		{
			name:      "father born 1700",
			build:     func(p *model.Person) { p.Father = born(1700) },
			wantAlive: false,
		},
		{
			name: "married 1903",
			build: func(p *model.Person) {
				f := &model.Family{ID: "f", Father: p}
				f.Timeline = []model.TimelineEvent{&model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1903)}, GeneralUnionEvent: model.GeneralUnionEvent{Husband: p}}}
				p.Families = []*model.Family{f}
			},
			wantAlive: false,
		},
		{
			name: "own marriage in 1908 implies birth by 1896",
			build: func(p *model.Person) {
				m := &model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1908)}, GeneralUnionEvent: model.GeneralUnionEvent{Husband: p}}
				f := &model.Family{ID: "f", Father: p, Timeline: []model.TimelineEvent{m}}
				p.Families = []*model.Family{f}
				p.Timeline = []model.TimelineEvent{m}
			},
			wantAlive: false,
		},
		{
			name: "married 1942",
			build: func(p *model.Person) {
				f := &model.Family{ID: "f", Father: p}
				f.Timeline = []model.TimelineEvent{&model.MarriageEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1942)}, GeneralUnionEvent: model.GeneralUnionEvent{Husband: p}}}
				p.Families = []*model.Family{f}
			},
			wantAlive: true,
		},
		{
			name: "child born 1890",
			build: func(p *model.Person) {
				p.Families = []*model.Family{{ID: "f", Father: p, Children: []*model.Person{born(1890)}}}
			},
			wantAlive: false,
		},
		{
			name: "own recent event is not overridden by an old spouse",
			build: func(p *model.Person) {
				p.Families = []*model.Family{{ID: "f", Father: p, Mother: born(1790)}}
				p.Timeline = []model.TimelineEvent{&model.OccupationEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1990)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}}
			},
			wantAlive: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "p"}
			tc.build(p)
			if err := InferPersonAliveOrDead(p, 2026); err != nil {
				t.Fatalf("InferPersonAliveOrDead: %v", err)
			}
			if p.PossiblyAlive != tc.wantAlive {
				t.Errorf("PossiblyAlive: got %v, want %v", p.PossiblyAlive, tc.wantAlive)
			}
		})
	}
}

// A dated burial or cremation shows that the person has died, as a death event does.
func TestInferPersonAliveOrDeadBurialProvesDeath(t *testing.T) {
	testCases := []struct {
		name string
		ev   func(p *model.Person) model.IndividualTimelineEvent
	}{
		{
			name: "burial",
			ev: func(p *model.Person) model.IndividualTimelineEvent {
				return &model.BurialEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1950)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
			},
		},
		{
			name: "cremation",
			ev: func(p *model.Person) model.IndividualTimelineEvent {
				return &model.CremationEvent{GeneralEvent: model.GeneralEvent{Date: model.Year(1950)}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "p"}
			ev := tc.ev(p)
			p.BestDeathlikeEvent = ev
			p.Timeline = []model.TimelineEvent{ev}

			if err := InferPersonAliveOrDead(p, 2026); err != nil {
				t.Fatalf("InferPersonAliveOrDead: %v", err)
			}
			if p.PossiblyAlive {
				t.Errorf("PossiblyAlive: got true, want false for a %s in 1950", tc.name)
			}
		})
	}
}

// A birth date qualified as "before" a year bounds the lifespan in the same way.
func TestInferPersonAliveOrDeadQualifiedBirthDate(t *testing.T) {
	p := &model.Person{ID: "p"}
	p.BestBirthlikeEvent = &model.BirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.BeforeYear(1790)},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}

	if err := InferPersonAliveOrDead(p, 2026); err != nil {
		t.Fatalf("InferPersonAliveOrDead: %v", err)
	}
	if p.PossiblyAlive {
		t.Errorf("PossiblyAlive: got true, want false for a birth before 1790")
	}
}

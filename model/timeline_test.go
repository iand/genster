package model

import "testing"

// TestCollapseEventListRemovesNonAdjacentDuplicates checks that the same event
// appearing twice is reduced to one, even when another event on the same date
// lies between the two copies.
func TestCollapseEventListRemovesNonAdjacentDuplicates(t *testing.T) {
	p := &Person{ID: "p"}
	q := &Person{ID: "q"}
	dup := &OccupationEvent{
		GeneralEvent:           GeneralEvent{Date: PreciseDate(1747, 3, 22)},
		GeneralIndividualEvent: GeneralIndividualEvent{Principal: p},
	}
	other := &OccupationEvent{
		GeneralEvent:           GeneralEvent{Date: PreciseDate(1747, 3, 22)},
		GeneralIndividualEvent: GeneralIndividualEvent{Principal: q},
	}

	got := CollapseEventList([]TimelineEvent{dup, other, dup})
	if len(got) != 2 {
		t.Fatalf("got %d events, want 2", len(got))
	}
}

// TestSortTimelineEventsKeepsInsertionOrderForTies checks that events on the
// same date keep the order they were added in.
func TestSortTimelineEventsKeepsInsertionOrderForTies(t *testing.T) {
	evs := make([]TimelineEvent, 0, 40)
	for i := range 40 {
		p := &Person{ID: string(rune('a' + i))}
		evs = append(evs, &BaptismEvent{
			GeneralEvent:           GeneralEvent{Date: PreciseDate(1800, 1, 1)},
			GeneralIndividualEvent: GeneralIndividualEvent{Principal: p},
		})
	}
	want := make([]TimelineEvent, len(evs))
	copy(want, evs)

	SortTimelineEvents(evs)
	for i := range evs {
		if evs[i] != want[i] {
			t.Fatalf("event %d moved: same-date events did not keep insertion order", i)
		}
	}
}

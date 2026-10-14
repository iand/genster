package model

import (
	"slices"
	"sort"
)

type Timeline struct {
	Events []TimelineEvent
}

type EventMatcher func(TimelineEvent) bool

var IsBirthEvent EventMatcher = func(ev TimelineEvent) bool {
	_, ok := ev.(*BirthEvent)
	return ok
}

var IsBaptismEvent EventMatcher = func(ev TimelineEvent) bool {
	_, ok := ev.(*BaptismEvent)
	return ok
}

var IsDeathEvent EventMatcher = func(ev TimelineEvent) bool {
	_, ok := ev.(*DeathEvent)
	return ok
}

var IsBurialEvent EventMatcher = func(ev TimelineEvent) bool {
	_, ok := ev.(*BurialEvent)
	return ok
}

func IsVitalEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		switch ev.(type) {
		case *BirthEvent, *BaptismEvent, *DeathEvent, *BurialEvent, *CremationEvent:
			return true
		default:
			return false
		}
	}
}

func IsOwnBirthEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*BirthEvent); ok {
			return ev.DirectlyInvolves(p)
		}
		return false
	}
}

func IsOwnPossibleBirthEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*PossibleBirthEvent); ok {
			return ev.DirectlyInvolves(p)
		}
		return false
	}
}

func IsOwnBaptismEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*BaptismEvent); ok {
			return ev.DirectlyInvolves(p)
		}
		return false
	}
}

func IsOwnDeathEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*DeathEvent); ok {
			return ev.DirectlyInvolves(p)
		}
		return false
	}
}

func IsOwnPossibleDeathEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*PossibleDeathEvent); ok {
			return ev.DirectlyInvolves(p)
		}
		return false
	}
}

func IsOwnBurialEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*BurialEvent); ok {
			return ev.DirectlyInvolves(p)
		}
		return false
	}
}

func IsOwnVitalEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		switch ev.(type) {
		case *BirthEvent, *BaptismEvent, *DeathEvent, *BurialEvent, *CremationEvent:
			return ev.DirectlyInvolves(p)
		default:
			return false
		}
	}
}

func AfterEvent(cutoff TimelineEvent) EventMatcher {
	if cutoff == nil || cutoff.GetDate().IsUnknown() {
		return func(ev TimelineEvent) bool {
			return true
		}
	}
	cdate := cutoff.GetDate()
	return func(ev TimelineEvent) bool {
		return cdate.SortsBefore(ev.GetDate())
	}
}

func FindFirstEvent(evs []TimelineEvent, include EventMatcher) (TimelineEvent, bool) {
	for _, ev := range evs {
		if include(ev) {
			return ev, true
		}
	}
	return nil, false
}

// FilterEventList returns a new slice that includes only the events that match the
// supplied EventMatcher
func FilterEventList(evs []TimelineEvent, include EventMatcher) []TimelineEvent {
	switch len(evs) {
	case 0:
		return []TimelineEvent{}
	case 1:
		if include(evs[0]) {
			return []TimelineEvent{evs[0]}
		}
		return []TimelineEvent{}
	default:
		l := make([]TimelineEvent, 0, len(evs))
		for _, ev := range evs {
			if include(ev) {
				l = append(l, ev)
			}
		}
		return l
	}
}

// SortTimelineEvents sorts events by date. Events that sort equally keep their
// existing order.
func SortTimelineEvents(evs []TimelineEvent) {
	sort.SliceStable(evs, func(i, j int) bool {
		return EventSortsBefore(evs[i], evs[j])
	})
}

// CollapseEventList returns a sorted slice that includes only unique events,
// keeping the first occurrence of each.
func CollapseEventList(evs []TimelineEvent) []TimelineEvent {
	seen := make(map[TimelineEvent]bool, len(evs))
	evs = slices.DeleteFunc(evs, func(ev TimelineEvent) bool {
		if seen[ev] {
			return true
		}
		seen[ev] = true
		return false
	})
	SortTimelineEvents(evs)
	return evs
	// switch len(evs) {
	// case 0:
	// 	return []TimelineEvent{}
	// case 1:
	// 	return []TimelineEvent{evs[0]}
	// case 2:
	// 	if evs[0] == evs[1] {
	// 		return []TimelineEvent{evs[0]}
	// 	}
	// 	return []TimelineEvent{evs[0], evs[1]}
	// default:
	// 	seen := make(map[TimelineEvent]bool)
	// 	l := make([]TimelineEvent, 0, len(evs))
	// 	for _, ev := range evs {
	// 		if seen[ev] {
	// 			continue
	// 		}
	// 		seen[ev] = true
	// 		l = append(l, ev)
	// 	}
	// 	return l
	// }
}

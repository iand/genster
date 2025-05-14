package model

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

func IsOwnBirthEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*BirthEvent); ok {
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

func IsOwnBurialEvent(p *Person) EventMatcher {
	return func(ev TimelineEvent) bool {
		if _, ok := ev.(*BurialEvent); ok {
			return ev.DirectlyInvolves(p)
		}
		return false
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

// CollapseEventList returns a new slice that includes only unique events
func CollapseEventList(evs []TimelineEvent) []TimelineEvent {
	switch len(evs) {
	case 0:
		return []TimelineEvent{}
	case 1:
		return []TimelineEvent{evs[0]}
	case 2:
		if evs[0] == evs[1] {
			return []TimelineEvent{evs[0]}
		}
		return []TimelineEvent{evs[0], evs[1]}
	default:
		seen := make(map[TimelineEvent]bool)
		l := make([]TimelineEvent, 0, len(evs))
		for _, ev := range evs {
			if seen[ev] {
				continue
			}
			seen[ev] = true
			l = append(l, ev)
		}
		return l
	}
}

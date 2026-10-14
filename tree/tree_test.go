package tree

import (
	"fmt"
	"slices"
	"testing"

	"github.com/iand/genster/model"
)

// TestGenerateAddsParentDeathToEveryChildTimeline checks that a parent's death
// reaches the timeline of every child, whatever order the people map is
// iterated in. With many children, some are visited before the mother, which
// exposes any dependence on map order. The scenario is repeated on fresh trees
// so the test cannot pass by a lucky iteration order.
func TestGenerateAddsParentDeathToEveryChildTimeline(t *testing.T) {
	const numChildren = 60
	for run := range 5 {
		tr := NewTree("test", nil, &SurnameGroups{})

		mother := tr.FindPerson("test", "mother")
		motherDeath := &model.DeathEvent{
			GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1920, 1, 1)},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: mother},
		}
		mother.Timeline = []model.TimelineEvent{
			&model.BirthEvent{
				GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1850, 1, 1)},
				GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: mother},
			},
			motherDeath,
		}

		children := make([]*model.Person, 0, numChildren)
		for i := range numChildren {
			c := tr.FindPerson("test", fmt.Sprintf("child%d", i))
			c.Mother = mother
			c.Timeline = []model.TimelineEvent{
				&model.BirthEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1880, 1, 1)},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c},
				},
				&model.DeathEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1960, 1, 1)},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c},
				},
			}
			children = append(children, c)
		}

		if err := tr.Generate(false); err != nil {
			t.Fatalf("run %d: Generate: %v", run, err)
		}

		missing := 0
		for _, c := range children {
			if !slices.Contains(c.Timeline, model.TimelineEvent(motherDeath)) {
				missing++
			}
		}
		if missing > 0 {
			t.Fatalf("run %d: mother's death missing from %d of %d child timelines", run, missing, numChildren)
		}
	}
}

// TestGenerateTimelineOrderIsDeterministic checks that events on the same date
// appear in the same order on a timeline every time a tree is generated.
func TestGenerateTimelineOrderIsDeterministic(t *testing.T) {
	const numChildren = 40
	var first []string
	for run := range 5 {
		tr := NewTree("test", nil, &SurnameGroups{})
		mother := tr.FindPerson("test", "mother")
		mother.Timeline = []model.TimelineEvent{
			&model.BirthEvent{
				GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1850, 1, 1)},
				GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: mother},
			},
		}
		for i := range numChildren {
			c := tr.FindPerson("test", fmt.Sprintf("child%02d", i))
			c.Mother = mother
			c.Timeline = []model.TimelineEvent{
				&model.BaptismEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1880, 6, 1)},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: c},
				},
			}
		}

		if err := tr.Generate(false); err != nil {
			t.Fatalf("run %d: Generate: %v", run, err)
		}

		var order []string
		for _, ev := range mother.Timeline {
			if bev, ok := ev.(*model.BaptismEvent); ok {
				order = append(order, bev.Principal.ID)
			}
		}
		if len(order) != numChildren {
			t.Fatalf("run %d: got %d child baptisms on mother's timeline, want %d", run, len(order), numChildren)
		}
		if run == 0 {
			first = order
			continue
		}
		if !slices.Equal(order, first) {
			t.Fatalf("run %d: child baptism order differs from run 0", run)
		}
	}
}

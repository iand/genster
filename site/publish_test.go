package site

import (
	"slices"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

// TestNewPublishSetFiltersEveryPlaceTimeline checks that every included place
// keeps only events of published people, including a place that is reached
// first as the parent of another place. Several child places and repeated
// runs expose any dependence on map iteration order.
func TestNewPublishSetFiltersEveryPlaceTimeline(t *testing.T) {
	for run := range 10 {
		tr := tree.NewTree("test", nil, &tree.SurnameGroups{})

		parish := &model.Place{ID: "parish"}
		tr.Places[parish.ID] = parish

		published := tr.FindPerson("test", "published")
		unpublished := tr.FindPerson("test", "unpublished")

		// an event of each person at the parish itself
		pubAtParish := &model.OccupationEvent{
			GeneralEvent:           model.GeneralEvent{Date: model.Year(1850), Place: parish},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: published},
		}
		unpubAtParish := &model.OccupationEvent{
			GeneralEvent:           model.GeneralEvent{Date: model.Year(1868), Place: parish},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: unpublished},
		}
		parish.Timeline = []model.TimelineEvent{pubAtParish, unpubAtParish}
		published.Timeline = []model.TimelineEvent{pubAtParish}
		unpublished.Timeline = []model.TimelineEvent{unpubAtParish}

		// events of the published person at several places within the parish
		for i := range 20 {
			street := &model.Place{ID: "street" + string(rune('a'+i)), Parent: parish}
			tr.Places[street.ID] = street
			ev := &model.OccupationEvent{
				GeneralEvent:           model.GeneralEvent{Date: model.Year(1851 + i), Place: street},
				GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: published},
			}
			street.Timeline = []model.TimelineEvent{ev}
			published.Timeline = append(published.Timeline, ev)
		}

		include := func(p *model.Person) bool { return p == published }
		ps, err := NewPublishSet(tr, include)
		if err != nil {
			t.Fatalf("run %d: NewPublishSet: %v", run, err)
		}
		if _, ok := ps.Places[parish.ID]; !ok {
			t.Fatalf("run %d: parish not included", run)
		}
		if slices.Contains(parish.Timeline, model.TimelineEvent(unpubAtParish)) {
			t.Fatalf("run %d: parish timeline includes an unpublished person's event", run)
		}
		if !slices.Contains(parish.Timeline, model.TimelineEvent(pubAtParish)) {
			t.Fatalf("run %d: parish timeline lost the published person's event", run)
		}
	}
}

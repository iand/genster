package gramps

import (
	"testing"

	"github.com/iand/genster/model"
)

func TestMostFrequentOccupationGroup(t *testing.T) {
	testCases := []struct {
		name   string
		counts map[model.OccupationGroup]int
		want   model.OccupationGroup
	}{
		{
			name:   "empty",
			counts: map[model.OccupationGroup]int{},
			want:   model.OccupationGroupUnknown,
		},
		{
			name: "most frequent wins",
			counts: map[model.OccupationGroup]int{
				model.OccupationGroupLabouring: 3,
				model.OccupationGroupCrafts:    1,
				model.OccupationGroupService:   2,
			},
			want: model.OccupationGroupLabouring,
		},
		{
			name: "tie broken by name",
			counts: map[model.OccupationGroup]int{
				model.OccupationGroupService:   2,
				model.OccupationGroupLabouring: 2,
				model.OccupationGroupMaritime:  2,
			},
			want: model.OccupationGroupLabouring,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// repeat to expose any dependence on map iteration order
			for range 20 {
				if got := mostFrequentOccupationGroup(tc.counts); got != tc.want {
					t.Fatalf("got %q, want %q", got, tc.want)
				}
			}
		})
	}
}

package tree

import (
	"testing"

	"github.com/iand/genster/model"
)

func calculatedYear(y int) *model.Date {
	d := model.Year(y)
	d.Derivation = model.DateDerivationCalculated
	return d
}

func estimatedYear(y int) *model.Date {
	d := model.Year(y)
	d.Derivation = model.DateDerivationEstimated
	return d
}

// A baptism replaces a less precise birth event as the best birth-like event, unless it is dated
// so long after the birth that it was a late baptism rather than one soon after birth. An
// estimated birth is not evidence of a late baptism, and a birth dated the same day as the
// baptism is taken to be copied from it.
func TestSelectPersonBestBirthEventLateBaptism(t *testing.T) {
	testCases := []struct {
		name      string
		birth     *model.Date
		baptism   *model.Date
		wantBirth bool
	}{
		{name: "calculated birth sixteen years before baptism", birth: calculatedYear(1791), baptism: model.PreciseDate(1807, 7, 13), wantBirth: true},
		{name: "calculated birth three years before baptism", birth: calculatedYear(1804), baptism: model.PreciseDate(1807, 7, 13), wantBirth: true},
		{name: "calculated birth a year before baptism", birth: calculatedYear(1806), baptism: model.PreciseDate(1807, 7, 13), wantBirth: false},
		{name: "calculated birth in the year of baptism", birth: calculatedYear(1807), baptism: model.PreciseDate(1807, 7, 13), wantBirth: false},
		{name: "birth year and baptism in the same year", birth: model.Year(1807), baptism: model.PreciseDate(1807, 7, 13), wantBirth: false},
		{name: "precise birth before baptism", birth: model.PreciseDate(1807, 7, 1), baptism: model.PreciseDate(1807, 7, 13), wantBirth: true},
		{name: "estimated birth five years before baptism", birth: estimatedYear(1754), baptism: model.PreciseDate(1759, 3, 4), wantBirth: false},
		{name: "birth on the day of baptism", birth: model.PreciseDate(1650, 3, 10), baptism: model.PreciseDate(1650, 3, 10), wantBirth: false},
	}

	for _, tc := range testCases {
		for _, order := range []string{"birth first", "baptism first"} {
			t.Run(tc.name+", "+order, func(t *testing.T) {
				tr := NewTree("test", nil, &SurnameGroups{})
				p := tr.FindPerson("test", "p")
				birth := &model.BirthEvent{GeneralEvent: model.GeneralEvent{Date: tc.birth}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
				baptism := &model.BaptismEvent{GeneralEvent: model.GeneralEvent{Date: tc.baptism}, GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p}}
				if order == "birth first" {
					p.Timeline = []model.TimelineEvent{birth, baptism}
				} else {
					p.Timeline = []model.TimelineEvent{baptism, birth}
				}

				if err := tr.SelectPersonBestBirthDeathEvents(p); err != nil {
					t.Fatalf("SelectPersonBestBirthDeathEvents: %v", err)
				}

				var want model.IndividualTimelineEvent = baptism
				if tc.wantBirth {
					want = birth
				}
				if p.BestBirthlikeEvent != want {
					t.Errorf("best birth-like event: got %s %s, want %s %s", p.BestBirthlikeEvent.Type(), p.BestBirthlikeEvent.GetDate(), want.Type(), want.GetDate())
				}
			})
		}
	}
}

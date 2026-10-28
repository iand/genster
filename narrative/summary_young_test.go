package narrative

import (
	"regexp"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// A child who died young and was born and died in one place is summarised without a
// sentence fragment starting "Was".
func TestPersonSummaryYoungDeathOnePlaceHasNoFragment(t *testing.T) {
	workhouse := &model.Place{ID: "wh", Name: "The Union Workhouse", FullContext: "Shipmeadow, Suffolk, England"}
	sandon := parish("Sandon", "Essex, England")

	testCases := []struct {
		name  string
		build func(p *model.Person)
		want  string
	}{
		{
			name: "died age thirteen, born and died in the workhouse",
			build: func(p *model.Person) {
				p.BestBirthlikeEvent = &model.BirthEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1835, 7, 12), Place: workhouse},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
				}
				p.BestDeathlikeEvent = &model.DeathEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1849, 6, 16), Place: workhouse},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
				}
			},
			want: "died age thirteen.",
		},
		{
			name: "died in infancy, baptised and buried in one parish",
			build: func(p *model.Person) {
				p.BestBirthlikeEvent = &model.BaptismEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1838, 6, 4), Place: sandon},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
				}
				p.BestDeathlikeEvent = &model.BurialEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1839, 3, 10), Place: sandon},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
				}
			},
			want: "died in infancy.",
		},
	}

	fragment := regexp.MustCompile(`\. Was\b`)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "m", PreferredGivenName: "Mary", PreferredFamiliarName: "Mary", Gender: model.GenderFemale}
			tc.build(p)
			p.Timeline = []model.TimelineEvent{p.BestBirthlikeEvent, p.BestDeathlikeEvent}

			for _, activeTense := range []bool{true, false} {
				enc := &md.Content{}
				got := PersonSummary(p, enc, &DefaultNameChooser{}, enc.EncodeText("Mary"), true, false, activeTense, true, true).String()
				checkContains(t, got, tc.want)
				if fragment.MatchString(got) {
					t.Errorf("activeTense=%v: summary has a fragment starting \"Was\": %s", activeTense, got)
				}
			}
		})
	}
}

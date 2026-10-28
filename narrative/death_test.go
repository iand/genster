package narrative

import (
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// A burial with an unknown date is described by its place alone.
func TestDeathStatementBurialUnknownDate(t *testing.T) {
	cemetery := &model.Place{ID: "bc", Name: "Bedford Cemetery", FullContext: "Foster Hill Road, Bedford, Bedfordshire, England"}

	p := &model.Person{ID: "a", PreferredGivenName: "Anastatia", PreferredFamiliarName: "Anastatia", Gender: model.GenderFemale}
	p.BestBirthlikeEvent = &model.BirthEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1868, 10, 30)},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
	p.BestDeathlikeEvent = &model.DeathEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1929, 10, 25)},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
	burial := &model.BurialEvent{
		GeneralEvent:           model.GeneralEvent{Date: model.UnknownDate(), Place: cemetery},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
	p.Timeline = []model.TimelineEvent{p.BestBirthlikeEvent, p.BestDeathlikeEvent, burial}

	s := &DeathStatement[md.Text]{Principal: p}
	intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
	enc := &md.Content{}
	s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

	checkContains(t, enc.String(), "buried in Bedford Cemetery in Foster Hill Road, Bedford, Bedfordshire, England.", "unknown date")
}

// An infant's death date follows "died" directly so it cannot be read as the birth date.
func TestDeathStatementInfantAgeClause(t *testing.T) {
	wallsend := &model.Place{ID: "hs", Name: "Hedley Street", FullContext: "Wallsend, Northumberland, England", PlaceType: model.PlaceTypeStreet}

	testCases := []struct {
		name    string
		birth   *model.Date
		death   *model.Date
		place   *model.Place
		want    string
		notWant []string
	}{
		{
			name:    "infant with death place",
			birth:   model.PreciseDate(1883, 6, 16),
			death:   model.PreciseDate(1883, 11, 13),
			place:   wallsend,
			want:    "Thomas died on 13 Nov 1883, less than a year after he was born, at Hedley Street in Wallsend, Northumberland, England.",
			notWant: []string{"born on 13 Nov"},
		},
		{
			name:  "infant with no death place",
			birth: model.PreciseDate(1883, 6, 16),
			death: model.PreciseDate(1883, 11, 13),
			want:  "Thomas died on 13 Nov 1883, less than a year after he was born.",
		},
		{
			name:  "adult keeps the age before the date",
			birth: model.PreciseDate(1823, 6, 16),
			death: model.PreciseDate(1883, 11, 13),
			place: wallsend,
			want:  "Thomas died at the age of sixty on 13 Nov 1883 at Hedley Street in Wallsend, Northumberland, England.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "t", PreferredGivenName: "Thomas", PreferredFamiliarName: "Thomas", Gender: model.GenderMale}
			p.BestBirthlikeEvent = &model.BirthEvent{
				GeneralEvent:           model.GeneralEvent{Date: tc.birth},
				GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
			}
			p.BestDeathlikeEvent = &model.DeathEvent{
				GeneralEvent:           model.GeneralEvent{Date: tc.death, Place: tc.place},
				GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
			}
			p.Timeline = []model.TimelineEvent{p.BestBirthlikeEvent, p.BestDeathlikeEvent}

			s := &DeathStatement[md.Text]{Principal: p}
			intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
			enc := &md.Content{}
			s.RenderDetail(0, intro, enc, &DefaultNameChooser{})

			checkContains(t, enc.String(), tc.want, tc.notWant...)
		})
	}
}

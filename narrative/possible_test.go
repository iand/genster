package narrative

import (
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// An unconfirmed birth candidate is introduced as unconfirmed, not after "Nothing is known".
func TestIntroStatementPossibleBirths(t *testing.T) {
	everton := parish("Everton", "Bedfordshire, England")
	cit := &model.GeneralCitation{ID: "c1", Detail: "Baptism entry"}

	testCases := []struct {
		name    string
		father  *model.Person
		cited   bool
		noCand  bool
		want    string
		notWant []string
	}{
		{
			name:    "no parents, uncited",
			want:    "No record of Henry's birth has been confirmed, but he may have been born on 24 Aug, 1727 in the parish of Everton in Bedfordshire, England.",
			notWant: []string{"Nothing is known", "It's possible"},
		},
		{
			name:    "no parents, cited",
			cited:   true,
			want:    "No record of Henry's birth has been confirmed, but there is some evidence that he was born on 24 Aug, 1727 in the parish of Everton in Bedfordshire, England",
			notWant: []string{"Nothing is known", "However"},
		},
		{
			name:    "parents known, cited",
			father:  &model.Person{ID: "f", PreferredUniqueName: "Charles Whitbread", PreferredFamiliarName: "Charles", Gender: model.GenderMale},
			cited:   true,
			want:    "Henry was the son of Charles Whitbread. His mother is not known. No record of his birth has been confirmed, but there is some evidence that he was born on 24 Aug, 1727",
			notWant: []string{"Nothing is known", "However"},
		},
		{
			name:   "no parents, no candidate",
			noCand: true,
			want:   "Nothing is known about the early life of Henry.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "h", PreferredGivenName: "Henry", PreferredFamiliarName: "Henry", Gender: model.GenderMale, Father: tc.father}
			s := &IntroStatement[md.Text]{Principal: p}
			if !tc.noCand {
				pb := &model.PossibleBirthEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1727, 8, 24), Place: everton},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
				}
				if tc.cited {
					pb.Citations = []*model.GeneralCitation{cit}
				}
				s.PossibleBirths = []*model.PossibleBirthEvent{pb}
			}

			intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
			enc := &md.Content{}
			s.RenderDetail(0, intro, enc, &DefaultNameChooser{})
			checkContains(t, enc.String(), tc.want, tc.notWant...)
		})
	}
}

// An unconfirmed death candidate is introduced as unconfirmed, not after "The time and place … is not known".
func TestDeathStatementPossibleDeaths(t *testing.T) {
	longstowe := parish("Longstowe", "Cambridgeshire, England")
	cit := &model.GeneralCitation{ID: "c1", Detail: "Burial entry"}

	testCases := []struct {
		name    string
		cited   bool
		noCand  bool
		want    string
		notWant []string
	}{
		{
			name:    "uncited",
			want:    "No record of Henry's death has been confirmed, but he may have died on 29 Dec, 1793 in the parish of Longstowe in Cambridgeshire, England.",
			notWant: []string{"is not known", "It's possible"},
		},
		{
			name:    "cited",
			cited:   true,
			want:    "No record of Henry's death has been confirmed, but there is some evidence that he died on 29 Dec, 1793 in the parish of Longstowe in Cambridgeshire, England",
			notWant: []string{"is not known", "However"},
		},
		{
			name:   "no candidate",
			noCand: true,
			want:   "The time and place of Henry's death is not known.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "h", PreferredGivenName: "Henry", PreferredFamiliarName: "Henry", Gender: model.GenderMale}
			s := &DeathStatement[md.Text]{Principal: p}
			if !tc.noCand {
				pd := &model.PossibleDeathEvent{
					GeneralEvent:           model.GeneralEvent{Date: model.PreciseDate(1793, 12, 29), Place: longstowe},
					GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
				}
				if tc.cited {
					pd.Citations = []*model.GeneralCitation{cit}
				}
				s.PossibleDeaths = []*model.PossibleDeathEvent{pd}
			}

			intro := &IntroGenerator[md.Text]{POV: &model.POV{Person: p}}
			enc := &md.Content{}
			s.RenderDetail(0, intro, enc, &DefaultNameChooser{})
			checkContains(t, enc.String(), tc.want, tc.notWant...)
		})
	}
}

package site

import (
	"testing"

	"github.com/iand/genster/model"
)

func TestScanPersonForAnomaliesBurialBeforeDeath(t *testing.T) {
	death := func(p *model.Person, d *model.Date) *model.DeathEvent {
		return &model.DeathEvent{
			GeneralEvent:           model.GeneralEvent{Date: d},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
		}
	}
	burial := func(p *model.Person, d *model.Date) *model.BurialEvent {
		return &model.BurialEvent{
			GeneralEvent:           model.GeneralEvent{Date: d},
			GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
		}
	}

	testCases := []struct {
		name     string
		death    *model.Date
		burials  []*model.Date
		wantText string // empty means no burial-before-death anomaly
	}{
		{
			name:     "burial before death",
			death:    model.PreciseDate(1900, 6, 17),
			burials:  []*model.Date{model.PreciseDate(1900, 1, 31)},
			wantText: "Burial dated 31 Jan 1900 is before death dated 17 Jun 1900",
		},
		{
			name:     "one burial before death and one after",
			death:    model.PreciseDate(1900, 6, 17),
			burials:  []*model.Date{model.PreciseDate(1900, 1, 31), model.PreciseDate(1900, 6, 20)},
			wantText: "Burial dated 31 Jan 1900 is before death dated 17 Jun 1900",
		},
		{
			name:    "burial after death",
			death:   model.PreciseDate(1900, 6, 17),
			burials: []*model.Date{model.PreciseDate(1900, 6, 20)},
		},
		{
			name:    "burial on day of death",
			death:   model.PreciseDate(1913, 1, 12),
			burials: []*model.Date{model.PreciseDate(1913, 1, 12)},
		},
		{
			name:    "year-only burial overlapping death",
			death:   model.PreciseDate(1900, 1, 30),
			burials: []*model.Date{model.Year(1900)},
		},
		{
			name:    "unknown burial date",
			death:   model.PreciseDate(1900, 6, 17),
			burials: []*model.Date{model.UnknownDate()},
		},
		{
			name:    "unknown death date",
			death:   model.UnknownDate(),
			burials: []*model.Date{model.PreciseDate(1900, 1, 31)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Person{ID: "p1"}
			p.Timeline = append(p.Timeline, death(p, tc.death))
			for _, d := range tc.burials {
				p.Timeline = append(p.Timeline, burial(p, d))
			}

			s := &Site{}
			s.ScanPersonForAnomalies(p)

			var found []*model.Anomaly
			for _, a := range p.Anomalies {
				if a.Context == "Burial before death" {
					found = append(found, a)
				}
			}

			if tc.wantText == "" {
				if len(found) != 0 {
					t.Errorf("got %d burial-before-death anomalies, want none: %q", len(found), found[0].Text)
				}
				return
			}
			if len(found) != 1 {
				t.Fatalf("got %d burial-before-death anomalies, want 1", len(found))
			}
			if found[0].Category != model.AnomalyCategoryEvent {
				t.Errorf("category: got %q, want %q", found[0].Category, model.AnomalyCategoryEvent)
			}
			if found[0].Text != tc.wantText {
				t.Errorf("text: got %q, want %q", found[0].Text, tc.wantText)
			}
		})
	}
}

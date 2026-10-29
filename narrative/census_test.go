package narrative

import (
	"strings"
	"testing"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
)

// Census entries in a timeline name the place without a verb, whatever their position.
func TestTimelineCensusEntryHasNoVerb(t *testing.T) {
	rushden := &model.Place{ID: "rh", Name: "Rushden Hill", FullName: "Rushden Hill, Rushden, Northamptonshire", NameWithRegion: "Rushden Hill, Northamptonshire", PlaceType: model.PlaceTypeStreet}
	p := &model.Person{ID: "s", PreferredUniqueName: "Sarah Ann Whitbread", Gender: model.GenderFemale}
	ev := &model.CensusEvent{
		GeneralEvent: model.GeneralEvent{Date: model.PreciseDate(1871, 4, 2), Place: rushden},
		Entries:      []*model.CensusEntry{{Principal: p}},
	}

	povs := map[string]*model.POV{
		"person": {Person: p},
		"place":  {Place: rushden},
	}
	wants := map[string]string{
		"person": "Recorded in the 1871 census at Rushden Hill",
		"place":  "recorded in the 1871 census here",
	}

	for name, pov := range povs {
		for seq := range 3 {
			enc := &md.Content{}
			fmtr := NewNarrativeTimelineEntryFormatter(pov, enc, logging.Default(), &TimelineNameChooser{}, false)
			got := fmtr.Title(seq, ev)
			if !strings.Contains(got, wants[name]) {
				t.Errorf("%s pov, seq %d: got %q, want it to contain %q", name, seq, got, wants[name])
			}
			if strings.Contains(got, "living") || strings.Contains(got, "residing") {
				t.Errorf("%s pov, seq %d: got %q, want no verb", name, seq, got)
			}
		}
	}
}

package infer

import (
	"testing"

	"github.com/iand/genster/model"
)

// TestRedactPersonalDetailsKeepsCitationsOfOthersEvents checks that redacting a
// person does not redact citations for events that belong to someone else,
// such as a parent's death copied onto the redacted person's timeline.
func TestRedactPersonalDetailsKeepsCitationsOfOthersEvents(t *testing.T) {
	parent := &model.Person{ID: "parent"}
	child := &model.Person{ID: "child", Mother: parent}

	parentDeathCit := &model.GeneralCitation{ID: "parent-death"}
	parentDeath := &model.DeathEvent{
		GeneralEvent: model.GeneralEvent{
			Date:      model.PreciseDate(1983, 10, 12),
			Citations: []*model.GeneralCitation{parentDeathCit},
		},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: parent},
	}
	parent.Timeline = []model.TimelineEvent{parentDeath}

	childBirthCit := &model.GeneralCitation{ID: "child-birth"}
	childBirth := &model.BirthEvent{
		GeneralEvent: model.GeneralEvent{
			Date:      model.PreciseDate(1960, 1, 1),
			Citations: []*model.GeneralCitation{childBirthCit},
		},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: child},
	}
	// The parent's death has been copied onto the child's timeline, as
	// tree.ExpandPersonTimeline does.
	child.Timeline = []model.TimelineEvent{childBirth, parentDeath}

	if _, err := RedactPersonalDetails(child); err != nil {
		t.Fatalf("RedactPersonalDetails: %v", err)
	}

	if !childBirthCit.Redacted {
		t.Errorf("citation for the redacted person's own birth was not redacted")
	}
	if parentDeathCit.Redacted {
		t.Errorf("citation for the parent's death was redacted")
	}
}

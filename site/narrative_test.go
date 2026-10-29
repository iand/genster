package site

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

// Hand-written narrative attached to a person's events appears in the Life Story: the text of
// a narrative event, and the narrative of any other event of their own.
func TestRenderPersonPageIncludesEventNarratives(t *testing.T) {
	tr := tree.NewTree("test", nil, &tree.SurnameGroups{})

	p := tr.FindPerson("test", "william")
	p.PreferredGivenName = "William"
	p.PreferredFamiliarName = "William"
	p.PreferredFullName = "William Hinksman"
	p.PreferredUniqueName = "William Hinksman"
	p.Gender = model.GenderMale

	narrativeEvent := &model.IndividualNarrativeEvent{
		GeneralEvent: model.GeneralEvent{
			Date:      model.Year(1790),
			Narrative: model.Text{Text: "William was apprenticed to a cordwainer in Whitchurch."},
		},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
	occupation := &model.OccupationEvent{
		GeneralEvent: model.GeneralEvent{
			Date:      model.Year(1801),
			Narrative: model.Text{Text: "The parish book names him as the overseer of the poor."},
		},
		GeneralIndividualEvent: model.GeneralIndividualEvent{Principal: p},
	}
	p.Timeline = []model.TimelineEvent{narrativeEvent, occupation}

	s := NewSite("/trees/test/", tr)
	ps, err := NewPublishSet(tr, func(*model.Person) bool { return true })
	if err != nil {
		t.Fatalf("NewPublishSet: %v", err)
	}
	s.PublishSet = ps

	doc, err := RenderPersonPage(s, p)
	if err != nil {
		t.Fatalf("RenderPersonPage: %v", err)
	}

	got := doc.String()
	lifeStory, _, _ := strings.Cut(got[strings.Index(got, "Life Story"):], "Timeline")
	for _, want := range []string{narrativeEvent.Narrative.Text, occupation.Narrative.Text} {
		if !strings.Contains(lifeStory, want) {
			t.Errorf("Life Story does not contain %q:\n%s", want, lifeStory)
		}
	}
}

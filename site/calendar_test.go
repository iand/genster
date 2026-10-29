package site

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

// A marriage to a person whose name is not known must not repeat the other
// partner's name through the description used as the unknown partner's title.
func TestCalendarMarriageToUnidentifiedPerson(t *testing.T) {
	tr := tree.NewTree("test", nil, &tree.SurnameGroups{})

	husband := tr.FindPerson("test", "husband")
	husband.PreferredFullName = "Stephen Learmouth"
	husband.PreferredUniqueName = "Stephen Learmouth (1675–1731)"
	husband.Gender = model.GenderMale

	wife := tr.FindPerson("test", "wife")
	wife.Unidentified = true
	wife.PreferredUniqueName = "Wife of Stephen Learmouth (1675–1731)"
	wife.Gender = model.GenderFemale

	marriage := &model.MarriageEvent{
		GeneralEvent:      model.GeneralEvent{Date: model.PreciseDate(1699, 5, 3)},
		GeneralUnionEvent: model.GeneralUnionEvent{Husband: husband, Wife: wife},
	}

	s := NewSite("/trees/test/", tr)
	ps, err := NewPublishSet(tr, func(*model.Person) bool { return true })
	if err != nil {
		t.Fatalf("NewPublishSet: %v", err)
	}
	s.PublishSet = ps

	c := &Calendar{Events: []model.TimelineEvent{marriage}}
	doc, err := c.RenderPage(s)
	if err != nil {
		t.Fatalf("RenderPage: %v", err)
	}

	got := doc.String()
	if strings.Contains(got, "Wife of") {
		t.Errorf("calendar uses the unknown partner's title:\n%s", got)
	}
	if !strings.Contains(got, "a woman whose name is not known") {
		t.Errorf("calendar does not describe the unknown partner:\n%s", got)
	}
}

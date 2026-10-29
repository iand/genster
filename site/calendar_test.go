package site

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

// A marriage where only one partner is a direct ancestor must say which one.
func TestCalendarMarriageWithOneDirectAncestor(t *testing.T) {
	testCases := []struct {
		name          string
		husbandDirect bool
		wifeDirect    bool
		want          string
	}{
		{name: "husband", husbandDirect: true, want: "He is a direct ancestor."},
		{name: "wife", wifeDirect: true, want: "She is a direct ancestor."},
		{name: "both", husbandDirect: true, wifeDirect: true, want: "Both are direct ancestors."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tr := tree.NewTree("test", nil, &tree.SurnameGroups{})

			husband := tr.FindPerson("test", "husband")
			husband.PreferredUniqueName = "John Smith (1700–1760)"
			husband.Gender = model.GenderMale
			if tc.husbandDirect {
				husband.RelationToKeyPerson = &model.Relation{To: husband, CommonAncestor: husband}
			}

			wife := tr.FindPerson("test", "wife")
			wife.PreferredUniqueName = "Mary Brown (1705–1770)"
			wife.Gender = model.GenderFemale
			if tc.wifeDirect {
				wife.RelationToKeyPerson = &model.Relation{To: wife, CommonAncestor: wife}
			}

			marriage := &model.MarriageEvent{
				GeneralEvent:      model.GeneralEvent{Date: model.PreciseDate(1725, 5, 3)},
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
			if !strings.Contains(got, tc.want) {
				t.Errorf("calendar does not contain %q:\n%s", tc.want, got)
			}
		})
	}
}

// A marriage to a redacted person names only the other partner.
func TestCalendarMarriageToRedactedPerson(t *testing.T) {
	tr := tree.NewTree("test", nil, &tree.SurnameGroups{})

	husband := tr.FindPerson("test", "husband")
	husband.PreferredUniqueName = "Norman Hall (1909–1985)"
	husband.Gender = model.GenderMale

	wife := tr.FindPerson("test", "wife")
	wife.Redacted = true
	wife.RedactNames("(person born in the 1910s)")
	wife.Gender = model.GenderFemale

	marriage := &model.MarriageEvent{
		GeneralEvent:      model.GeneralEvent{Date: model.PreciseDate(1942, 7, 24)},
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
	if strings.Contains(got, "(person born in the 1910s)") {
		t.Errorf("calendar names the redacted partner:\n%s", got)
	}
	if !strings.Contains(got, "Norman Hall (1909–1985)</a> married.") {
		t.Errorf("calendar does not describe the marriage:\n%s", got)
	}
}

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

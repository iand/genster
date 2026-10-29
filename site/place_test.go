package site

import (
	"strings"
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
)

// A ship or boat is described as a vessel.
func TestRenderPlacePageVessel(t *testing.T) {
	tr := tree.NewTree("test", nil, &tree.SurnameGroups{})
	pl := tr.FindPlace("test", "perseverance")
	pl.Name = "Perseverance"
	pl.FullName = "Perseverance"
	pl.PlaceType = model.PlaceTypeVessel

	s := NewSite("/trees/test/", tr)
	ps, err := NewPublishSet(tr, func(*model.Person) bool { return true })
	if err != nil {
		t.Fatalf("NewPublishSet: %v", err)
	}
	s.PublishSet = ps

	doc, err := RenderPlacePage(s, pl, nil)
	if err != nil {
		t.Fatalf("RenderPlacePage: %v", err)
	}
	if got := doc.String(); !strings.Contains(got, "Perseverance is a vessel.") {
		t.Errorf("page does not describe a vessel:\n%s", got)
	}
}

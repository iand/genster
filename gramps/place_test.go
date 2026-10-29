package gramps

import (
	"testing"

	"github.com/iand/genster/model"
	"github.com/iand/genster/tree"
	"github.com/iand/grampsxml"
)

func TestApplyGrampsPlaceType(t *testing.T) {
	testCases := []struct {
		grampsType string
		want       model.PlaceType
	}{
		{grampsType: "Ship", want: model.PlaceTypeVessel},
		{grampsType: "Canal Boat", want: model.PlaceTypeVessel},
		{grampsType: "Street", want: model.PlaceTypeStreet},
		{grampsType: "Locality", want: model.PlaceTypeUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.grampsType, func(t *testing.T) {
			pl := &model.Place{}
			applyGrampsPlaceType(pl, tc.grampsType)
			if pl.PlaceType != tc.want {
				t.Errorf("got %q, want %q", pl.PlaceType, tc.want)
			}
		})
	}
}

// loadPlaces populates the Gramps places, each given as handle, type, name and parent handle,
// and returns the tree's place for each handle.
func loadPlaces(t *testing.T, places ...[4]string) map[string]*model.Place {
	t.Helper()
	l := &Loader{
		ScopeName:       "test",
		PlacesByHandle:  map[string]*grampsxml.Placeobj{},
		populatedPlaces: map[string]bool{},
	}
	tr := tree.NewTree("test", nil, &tree.SurnameGroups{})
	for _, p := range places {
		id := "P" + p[0]
		gp := &grampsxml.Placeobj{ID: &id, Handle: p[0], Type: p[1], Pname: []grampsxml.Pname{{Value: p[2]}}}
		if p[3] != "" {
			gp.Placeref = []grampsxml.Placeref{{Hlink: p[3]}}
		}
		l.PlacesByHandle[gp.Handle] = gp
	}
	got := map[string]*model.Place{}
	for _, p := range places {
		gp := l.PlacesByHandle[p[0]]
		if !l.populatedPlaces[gp.Handle] {
			if err := l.populatePlaceFacts(tr, gp); err != nil {
				t.Fatalf("populatePlaceFacts(%s): %v", p[0], err)
			}
		}
		got[p[0]] = tr.FindPlace("test", persistentID(gp.ID, gp.Handle))
	}
	return got
}

// A numbered place is combined with the street, building or hamlet it is in.
func TestNumberedPlaceCombinedWithParent(t *testing.T) {
	testCases := []struct {
		name       string
		places     [][4]string
		wantName   string
		wantParent string
	}{
		{
			name: "numbered building in a street",
			places: [][4]string{
				{"t", "Town", "Rushden", ""},
				{"s", "Street", "High Street", "t"},
				{"n", "Building", "12", "s"},
			},
			wantName:   "12 High Street",
			wantParent: "Rushden",
		},
		{
			// Shaped like 164 North Street, Limehouse, where the number is typed as a street.
			name: "numbered street in a street",
			places: [][4]string{
				{"t", "Town", "Limehouse", ""},
				{"s", "Street", "North Street", "t"},
				{"n", "Street", "164", "s"},
			},
			wantName:   "164 North Street",
			wantParent: "Limehouse",
		},
		{
			// Shaped like 16 Rushden Hill, Rushden.
			name: "numbered building in a hamlet",
			places: [][4]string{
				{"p", "Parish", "Rushden", ""},
				{"h", "Hamlet", "Rushden Hill", "p"},
				{"n", "Building", "16", "h"},
			},
			wantName:   "16 Rushden Hill",
			wantParent: "Rushden",
		},
		{
			name: "named building in a hamlet",
			places: [][4]string{
				{"p", "Parish", "Marsworth", ""},
				{"h", "Hamlet", "Startops End", "p"},
				{"n", "Building", "Private House", "h"},
			},
			wantName:   "Private House",
			wantParent: "Startops End",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := loadPlaces(t, tc.places...)["n"]
			if got.Name != tc.wantName {
				t.Errorf("name: got %q, want %q", got.Name, tc.wantName)
			}
			if got.Parent == nil || got.Parent.Name != tc.wantParent {
				t.Errorf("parent: got %v, want %q", got.Parent, tc.wantParent)
			}
		})
	}
}

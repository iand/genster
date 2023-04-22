package place

import "testing"

func TestClassifyKind(t *testing.T) {
	testCases := []struct {
		in   string
		want PlaceKind
	}{
		{
			in:   "England",
			want: PlaceKindUKNation,
		},
		{
			in:   "sCOTLAnd",
			want: PlaceKindUKNation,
		},
		{
			in:   "Wales",
			want: PlaceKindUKNation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			pn := ClassifyName(tc.in)

			if pn.Kind != tc.want {
				t.Errorf("got %s, want %s", pn.Kind, tc.want)
			}
		})
	}
}

func TestClassifyPartOf(t *testing.T) {
	testCases := []struct {
		in   string
		want *PlaceName
	}{
		{
			in:   "England",
			want: countryUK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			pn := ClassifyName(tc.in)

			found := false
			for _, p := range pn.PartOf {
				if p.SameAs(tc.want) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("did not find %s", tc.want.Name)
			}
		})
	}
}

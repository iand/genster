package infer

import (
	"testing"

	"github.com/iand/genster/model"
)

// A couple with an unknown bond and more than one child were probably married; a recorded bond
// and a couple with fewer children are left as they are.
func TestInferFamilyBond(t *testing.T) {
	testCases := []struct {
		name     string
		bond     string
		children int
		want     string
	}{
		{name: "unknown with two children", bond: model.FamilyBondUnknown, children: 2, want: model.FamilyBondLikelyMarried},
		{name: "unknown with twelve children", bond: model.FamilyBondUnknown, children: 12, want: model.FamilyBondLikelyMarried},
		{name: "unknown with one child", bond: model.FamilyBondUnknown, children: 1, want: model.FamilyBondUnknown},
		{name: "unknown with no children", bond: model.FamilyBondUnknown, children: 0, want: model.FamilyBondUnknown},
		{name: "unmarried with two children", bond: model.FamilyBondUnmarried, children: 2, want: model.FamilyBondUnmarried},
		{name: "married with two children", bond: model.FamilyBondMarried, children: 2, want: model.FamilyBondMarried},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := &model.Family{ID: "f", Bond: tc.bond, Father: &model.Person{ID: "h"}, Mother: &model.Person{ID: "w"}}
			for i := 0; i < tc.children; i++ {
				f.Children = append(f.Children, &model.Person{ID: "c" + string(rune('a'+i))})
			}

			InferFamilyBond(f)

			if f.Bond != tc.want {
				t.Errorf("bond: got %q, want %q", f.Bond, tc.want)
			}
		})
	}
}

package infer

import "github.com/iand/genster/model"

// minChildrenForLikelyMarriage is the number of children that makes a couple whose bond is not
// recorded probably married.
const minChildrenForLikelyMarriage = 2

// InferFamilyBond infers that a couple whose bond is not recorded were probably married when
// they had more than one child, since most such families are marriages whose record has not
// been found. A recorded bond is left as it is.
func InferFamilyBond(f *model.Family) {
	if f.Bond != model.FamilyBondUnknown {
		return
	}

	seen := make(map[string]bool)
	for _, c := range f.Children {
		seen[c.ID] = true
	}
	if len(seen) >= minChildrenForLikelyMarriage {
		f.Bond = model.FamilyBondLikelyMarried
	}
}

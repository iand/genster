package model

import "fmt"

// Relation describes the relationship between the From person and the To
type Relation struct {
	From                  *Person
	To                    *Person
	CommonAncestor        *Person
	FromGenerations       int       // number of generations between the From person and CommonAncestor. 1 = common ancestor is parent of from
	ToGenerations         int       // number of generations between the To person and CommonAncestor. 1 = common ancestor is parent of to
	ClosestDirectRelation *Relation // if the To person does not share an ancestor with the From person, then this is the relation of From to a person who was the partner/spouse of a direct relation of To
	// ToSpouseGenerations   int       // number of generations between the To person and the spouse of SpouseRelation. 0 = spouse, 1 = parent of spouse
	SpouseRelation *Relation // the relationship of the To person to the spouse of the ClosestDirectRelation
}

func (r *Relation) IsSelf() bool {
	if r == nil {
		return false
	}
	return r.To.SameAs(r.From)
}

// IsDirectAncestor reports whether the To person is a direct ancestor of the From person
func (r *Relation) IsDirectAncestor() bool {
	if r == nil {
		return false
	}
	return r.To.SameAs(r.CommonAncestor)
}

// IsParent reports whether the To person is a parent of the From person
func (r *Relation) IsParent() bool {
	if r == nil {
		return false
	}
	return r.To.SameAs(r.CommonAncestor) && r.FromGenerations == 1
}

// IsDirectDescendant reports whether the To person is a direct descendant of the From person
func (r *Relation) IsDirectDescendant() bool {
	if r == nil {
		return false
	}
	return r.From.SameAs(r.CommonAncestor)
}

// IsChild reports whether the To person is a child of the From person
func (r *Relation) IsChild() bool {
	if r == nil {
		return false
	}
	return r.From.SameAs(r.CommonAncestor) && r.ToGenerations == 1
}

// IsCloseToDirectAncestor reports whether To person is a direct ancestor or a child or spouse of the From person.
func (r *Relation) IsCloseToDirectAncestor() bool {
	if r == nil || r.To.IsUnknown() {
		return false
	}

	if r.IsDirectAncestor() {
		return true
	}

	if !r.To.Father.IsUnknown() {
		prel := r.ExtendToParent(r.To.Father)
		if prel.IsDirectAncestor() {
			return true
		}
	}

	if !r.To.Mother.IsUnknown() {
		prel := r.ExtendToParent(r.To.Mother)
		if prel.IsDirectAncestor() {
			return true
		}
	}

	// if this person is not a direct relation but is the spouse of one
	if r.ClosestDirectRelation != nil && r.SpouseRelation.IsSelf() {
		if r.ClosestDirectRelation.IsDirectAncestor() {
			return true
		}
	}

	return false
}

// HasCommonAncestor reports whether the From and To person have a common ancestor, including
// if the To person is a direct ancestor of the From person
func (r *Relation) HasCommonAncestor() bool {
	if r == nil {
		return false
	}
	return !r.CommonAncestor.IsUnknown()
}

// Name returns the name of the relation between From and To, in the
// form that "To is the Name() of From"
func (r *Relation) Name() string {
	if r.CommonAncestor != nil {

		if r.To.SameAs(r.CommonAncestor) {
			// Direct ancestor

			if r.FromGenerations == 0 {
				return "same person as"
			}

			name := r.To.Gender.RelationToChildrenNoun()

			if r.FromGenerations == 1 {
				return name
			}

			name = "grand" + name

			if r.FromGenerations == 2 {
				return name
			}

			if r.FromGenerations < 5 {
				for i := 2; i < r.FromGenerations; i++ {
					name = "great " + name
				}
			} else {
				name = fmt.Sprintf("%dx great "+name, r.FromGenerations-2)
			}

			return name
		}

		if r.From.SameAs(r.CommonAncestor) {
			// Direct descendent

			if r.ToGenerations == 0 {
				return "self"
			}

			name := r.To.Gender.RelationToParentNoun()

			if r.ToGenerations == 1 {
				return name
			}

			name = "grand" + name

			if r.ToGenerations == 2 {
				return name
			}

			if r.FromGenerations < 5 {
				for i := 2; i < r.ToGenerations; i++ {
					name = "great " + name
				}
			} else {
				name = fmt.Sprintf("%dx great "+name, r.FromGenerations-2)
			}

			return name
		}

		if r.FromGenerations == 1 && r.ToGenerations == 1 {
			switch r.From.Gender {
			case GenderMale:
				return "brother"
			case GenderFemale:
				return "sister"
			default:
				return "sibling"
			}
		}

		if r.FromGenerations > 1 && r.ToGenerations == 1 {
			var name string

			switch r.To.Gender {
			case GenderMale:
				name = "uncle"
			case GenderFemale:
				name = "aunt"
			}

			if name != "" {
				for i := 2; i < r.FromGenerations; i++ {
					name = "great " + name
				}

				return name
			}
		}

		name := "cousin"

		degree := "first"
		removal := ""

		switch r.FromGenerations {
		case 3:
			degree = "second"
		case 4:
			degree = "third"
		case 5:
			degree = "fourth"
		case 6:
			degree = "fifth"
		case 7:
			degree = "sixth"
		case 8:
			degree = "seventh"
		case 9:
			degree = "eighth"
		case 10:
			degree = "ninth"
		case 11:
			degree = "tenth"
		case 12:
			degree = "eleventh"
		}
		if degree != "" {
			degree += " "
		}

		switch abs(r.FromGenerations - r.ToGenerations) {
		case 0:
		case 1:
			removal = "once"
		case 2:
			removal = "twice"
		case 3:
			removal = "three times"
		case 4:
			removal = "four times"
		case 5:
			removal = "five times"
		case 6:
			removal = "six times"
		case 7:
			removal = "seven times"
		case 8:
			removal = "eight times"
		}

		if removal != "" {
			removal = " " + removal + " removed"
		}

		return degree + name + removal

	}

	if r.ClosestDirectRelation != nil && r.SpouseRelation != nil {
		var rel string
		if r.SpouseRelation.IsSelf() {
			if r.ClosestDirectRelation.IsParent() {
				switch r.To.Gender {
				case GenderMale:
					return "step-father"
				case GenderFemale:
					return "step-mother"
				default:
					return "step-parent"
				}
			}
			switch r.To.Gender {
			case GenderMale:
				rel = "husband"
			case GenderFemale:
				rel = "wife"
			default:
				rel = "spouse"
			}
		} else if r.SpouseRelation.FromGenerations == 1 {
			switch r.To.Gender {
			case GenderMale:
				rel = "father-in-law"
			case GenderFemale:
				rel = "mother-in-law"
			default:
				rel = "parent-in-law"
			}
		} else if r.SpouseRelation.ToGenerations == 1 {
			switch r.To.Gender {
			case GenderMale:
				rel = "stepson"
			case GenderFemale:
				rel = "stepdaughter"
			default:
				rel = "stepchild"
			}
		}

		if rel != "" {
			if r.ClosestDirectRelation.From.SameAs(r.ClosestDirectRelation.To) {
				return rel
			}

			return rel + " of the " + r.ClosestDirectRelation.Name()
		}

	}

	return "unknown relation"
}

const MaxDistance = 1000

// Distance computes a score for how distant the relationship is.
// A self relationship has score 0, a parent/child/spouse has score 1
// A direct ancestor scores 1 per generation step
// A direct descendant scores 1 per generation step
// A descendant of a common ancestor scores 1 per generation step to the common ancestor and 2 per generation down beyond the first
// A non-direct relationship doubles the score of the closest direct relationship
func (r *Relation) Distance() int {
	if r == nil {
		return MaxDistance
	}
	if r.From.SameAs(r.To) {
		// self
		return 0
	}
	if r.CommonAncestor != nil {
		// Is the From person the common ancestor
		if r.From.SameAs(r.CommonAncestor) {
			// The To person is a descendant
			return r.ToGenerations
		}

		// Is the To person the common ancestor
		if r.To.SameAs(r.CommonAncestor) {
			// The To person is a direct ancestor
			return r.FromGenerations
		}

		// direct relationship
		return r.FromGenerations + 2*(r.ToGenerations-1)
	} else if r.ClosestDirectRelation != nil && r.SpouseRelation != nil {
		// indirect relationship
		if r.SpouseRelation.Distance() == 0 {
			return r.ClosestDirectRelation.Distance() + 1
		}
		return r.ClosestDirectRelation.Distance() + r.SpouseRelation.Distance()*2
	}

	return MaxDistance
}

// ExtendToSpouse produces a new relationship extended to the spouse of the To person
func (r *Relation) ExtendToSpouse(spouse *Person) *Relation {
	return &Relation{
		From:                  r.From,
		To:                    spouse,
		ClosestDirectRelation: r,
		SpouseRelation:        Self(spouse),
	}
}

// ExtendToChild produces a new relationship extended to the child of the To person
func (r *Relation) ExtendToChild(ch *Person) *Relation {
	if r.CommonAncestor != nil {
		return &Relation{
			From:            r.From,
			To:              ch,
			CommonAncestor:  r.CommonAncestor,
			FromGenerations: r.FromGenerations,
			ToGenerations:   r.ToGenerations + 1,
		}
	} else if r.ClosestDirectRelation != nil && r.SpouseRelation != nil {
		return &Relation{
			From:                  r.From,
			To:                    ch,
			ClosestDirectRelation: r.ClosestDirectRelation,
			SpouseRelation:        r.SpouseRelation.ExtendToChild(ch),
		}
	} else {
		return &Relation{
			From: r.From,
			To:   ch,
		}
	}
}

// ExtendToParent produces a new relationship extended to the parent of the To person
func (r *Relation) ExtendToParent(parent *Person) *Relation {
	if r.CommonAncestor != nil {
		if r.CommonAncestor.SameAs(r.To) {
			return &Relation{
				From:            r.From,
				To:              parent,
				CommonAncestor:  parent,
				FromGenerations: r.FromGenerations + 1,
				ToGenerations:   0,
			}
		} else {
			return &Relation{
				From:            r.From,
				To:              parent,
				CommonAncestor:  r.CommonAncestor,
				FromGenerations: r.FromGenerations,
				ToGenerations:   r.ToGenerations - 1,
			}
		}
		// TODO
		// } else if r.ClosestDirectRelation != nil && r.SpouseRelation != nil {
		// 	return &Relation{
		// 		From:                  r.From,
		// 		To:                    ch,
		// 		ClosestDirectRelation: r.ClosestDirectRelation,
		// 		SpouseRelation:        r.SpouseRelation.ExtendToChild(ch),
		// 	}
	} else {
		return &Relation{
			From: r.From,
			To:   parent,
		}
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func Parent(child *Person, parent *Person) *Relation {
	return &Relation{
		From:            child,
		To:              parent,
		CommonAncestor:  parent,
		FromGenerations: 1,
		ToGenerations:   0,
	}
}

func Self(p *Person) *Relation {
	return &Relation{
		From:            p,
		To:              p,
		CommonAncestor:  p,
		FromGenerations: 0,
		ToGenerations:   0,
	}
}

func Spouse(p *Person, spouse *Person) *Relation {
	return &Relation{
		From:                  p,
		To:                    spouse,
		ClosestDirectRelation: Self(p),
		SpouseRelation:        Self(spouse),
	}
}

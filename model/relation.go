package model

type Relation struct {
	From                *Person
	To                  *Person
	CommonAncestor      *Person
	FromGenerations     int       // number of generations between the From person and CommonAncestor. 1 = parent
	ToGenerations       int       // number of generations between the To person and CommonAncestor. 1 = parent
	SpouseRelation      *Relation // if the To person does not share an ancestor with the From person, then this is the relation of the spouse to From
	ToSpouseGenerations int       // number of generations between the To person and CommonSpouse. 0 = spouse, 1 = parent of spouse
}

func (r *Relation) IsDirectAncestor() bool {
	if r == nil {
		return false
	}
	return r.To.SameAs(r.CommonAncestor)
}

func (r *Relation) IsSelf() bool {
	if r == nil {
		return false
	}
	return r.To.SameAs(r.From)
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

			for i := 2; i < r.FromGenerations; i++ {
				name = "great " + name
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

			for i := 2; i < r.ToGenerations; i++ {
				name = "great " + name
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

	if r.SpouseRelation != nil {
		var rel string
		if r.ToSpouseGenerations == 0 {
			switch r.To.Gender {
			case GenderMale:
				rel = "husband"
			case GenderFemale:
				rel = "wife"
			default:
				rel = "spouse"
			}
		} else if r.ToSpouseGenerations == 1 {
			switch r.To.Gender {
			case GenderMale:
				rel = "father-in-law"
			case GenderFemale:
				rel = "mother-in-law"
			default:
				rel = "parent-in-law"
			}
		} else if r.ToSpouseGenerations == -1 {
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
			if r.SpouseRelation.From.SameAs(r.SpouseRelation.To) {
				return rel
			}

			return rel + " of the " + r.SpouseRelation.Name()
		}

	}

	return "unknown relation"
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

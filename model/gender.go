package model

const (
	GenderMale    = "man"
	GenderFemale  = "woman"
	GenderUnknown = "unknown"
)

type Gender string

func (g Gender) IsMale() bool {
	return g == GenderMale
}

func (g Gender) IsFemale() bool {
	return g == GenderFemale
}

func (g Gender) IsUnknown() bool {
	return g == GenderUnknown
}

func (g Gender) Opposite() Gender {
	switch g {
	case GenderMale:
		return GenderFemale
	case GenderFemale:
		return GenderMale
	default:
		return GenderUnknown
	}
}

func (g Gender) Noun() string {
	switch g {
	case GenderMale:
		return "man"
	case GenderFemale:
		return "woman"
	default:
		return "person"
	}
}

func (g Gender) ChildNoun() string {
	switch g {
	case GenderMale:
		return "boy"
	case GenderFemale:
		return "girl"
	default:
		return "child"
	}
}

// RelationToParentNoun returns the general noun to use for parent relation: son/daughter or child if sex is unknown  (this person is the ___ of their parent)
func (g Gender) RelationToParentNoun() string {
	switch g {
	case GenderMale:
		return "son"
	case GenderFemale:
		return "daughter"
	default:
		return "child"
	}
}

// RelationToChildrenNoun returns the general noun to use for child relation: father/mother or parent if sex is unknown  (this person is the ___ of their child)
func (g Gender) RelationToChildrenNoun() string {
	switch g {
	case GenderMale:
		return "father"
	case GenderFemale:
		return "mother"
	default:
		return "parent"
	}
}

// RelationToSpouseNoun returns the general noun to use for spouse relation: husband/wife or spouse if sex is unknown  (this person is the ___ of their spouse)
func (g Gender) RelationToSpouseNoun() string {
	switch g {
	case GenderMale:
		return "husband"
	case GenderFemale:
		return "wife"
	default:
		return "spouse"
	}
}

func (g Gender) RelationToSpouseNounPlural() string {
	switch g {
	case GenderMale:
		return "husbands"
	case GenderFemale:
		return "wives"
	default:
		return "spouses"
	}
}

func (g Gender) SubjectPronoun() string {
	switch g {
	case GenderMale:
		return "he"
	case GenderFemale:
		return "she"
	default:
		return "they"
	}
}

func (g Gender) SubjectPronounWithLink() string {
	switch g {
	case GenderMale:
		return "he was"
	case GenderFemale:
		return "she was"
	default:
		return "they were"
	}
}

func (g Gender) PossessivePronounSingular() string {
	switch g {
	case GenderMale:
		return "his"
	case GenderFemale:
		return "her"
	default:
		return "their"
	}
}

func (g Gender) PossessivePronounPlural() string {
	switch g {
	case GenderMale:
		return "his"
	case GenderFemale:
		return "hers"
	default:
		return "theirs"
	}
}

func (g Gender) ObjectPronoun() string {
	switch g {
	case GenderMale:
		return "him"
	case GenderFemale:
		return "her"
	default:
		return "them"
	}
}

func (g Gender) ReflexivePronoun() string {
	switch g {
	case GenderMale:
		return "himself"
	case GenderFemale:
		return "herself"
	default:
		return "themselves"
	}
}

// WidowWidower returns a word that fits the sentence:
// "when their spouse died, x was left a <widow/widower>"
func (g Gender) WidowWidower() string {
	switch g {
	case GenderMale:
		return "widower"
	case GenderFemale:
		return "widow"
	default:
		return "widow"
	}
}

type Link struct {
	Title string
	URL   string
}

package model

const (
	FactCategoryAKA                   = "Also known as"
	FactCategoryMilitaryServiceNumber = "Military service number"
)

type Fact struct {
	Category  string
	Detail    string
	Citations []*GeneralCitation
}

type AssociationKind string

const (
	AssociationKindTwin AssociationKind = "twin"
)

type Association struct {
	Kind      AssociationKind
	Other     *Person
	Citations []*GeneralCitation
}

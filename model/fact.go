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

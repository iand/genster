package model

const (
	FactCategoryAKA = "Also known as"
)

type Fact struct {
	Category  string
	Detail    string
	Citations []*GeneralCitation
}

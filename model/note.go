package model

type Note struct {
	Title         string
	Author        string
	Date          string
	Markdown      string
	PrimaryPerson *Person
}

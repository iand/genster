package model

type Text struct {
	ID        string
	Title     string
	Text      string
	Formatted bool
	Markdown  bool
	Links     []ObjectLink
}

package model

type ToDoCategory string

const (
	ToDoCategoryCitations ToDoCategory = "Citations"
	ToDoCategoryMissing   ToDoCategory = "Missing Information"
)

func (c ToDoCategory) String() string {
	return string(c)
}

// A ToDo is a task or loose end for an area of research
type ToDo struct {
	Category ToDoCategory
	Context  string
	Goal     string
	Reason   string
}

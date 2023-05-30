package model

type ToDoCategory string

const (
	ToDoCategoryCitations ToDoCategory = "Citation"
	ToDoCategoryMissing   ToDoCategory = "Missing Information"
	ToDoCategoryRecords   ToDoCategory = "Records"
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

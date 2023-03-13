package model

type Occupation struct {
	StartDate   *Date
	EndDate     *Date
	Place       *Place
	Title       string
	Detail      string
	Citations   []*GeneralCitation
	Occurrences int
}

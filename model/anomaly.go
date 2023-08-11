package model

type AnomalyCategory string

const (
	AnomalyCategoryCitation AnomalyCategory = "Citation"
	AnomalyCategoryEvent    AnomalyCategory = "Event"
	AnomalyCategoryName     AnomalyCategory = "Name"
)

func (c AnomalyCategory) String() string {
	return string(c)
}

// An Anomaly is something detected in existing data that can be corrected
// manually
type Anomaly struct {
	Category AnomalyCategory
	Text     string
	Context  string
}

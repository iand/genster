package model

// An Anomaly is something detected in existing data that can be corrected
// manually
type Anomaly struct {
	Category string
	Text     string
	Context  string
}

package chart

// New layout engine for charts

// An InfoBox represents the extent of the name and details of an individual on the chart.
type InfoBox struct{}

// A HRelBox represents the extent of the "=" symbol and details of a relationship between two individuals when
// represented horizontally. The details are printed below the "=" symbol.
// The distance between the top of the box and the start of the details must be greater than the height
// of both of the individuals in the relationship.
type HRelBox struct{}

// A VRelBox represents the extent of the "=" symbol and details of a relationship between two individuals when
// represented vertically. The details are printed to the right of the "=" symbol.
// The distance between the top of the box and the start of the details must be greater than the height
// of both of the individuals in the relationship.
type VRelBox struct{}

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


// An IndiBox represents the entire extent of an individual in its own generation row on the chart.
// This includes the RelBox and InfoBox of each relationship.
type IndiBox struct{}

// An IndiBoxNext represents the entire extent of an individual in the next generation row on the chart.
// This includes the IndiBoxSame of each child.
type IndiBoxNext struct{}

// A GenRow represents a row of individuals that belong to the same generation.
// It consists of all the IndiBox's for each individual.
type GenRow struct{}

// A ConnBox represents the extent of the connections between an individual and its children when 
// represented horizontally. It holds the lines that connect from the centre bottom of a HRelBox
// to the horizontal line that spans the width of the individual's IndiBoxNext, and the lines that
// drop down.
// It has a standard height.
type ConnBox struct{}

// A SideConnBox represents the extent of the connections between an individual and its children when 
// represented vertically.

// It has a standard width.
type SideConnBox struct{}

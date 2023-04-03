package model

import (
	"strconv"
	"time"

	"github.com/iand/gdate"
)

type Date struct {
	Date gdate.Date
}

func UnknownDate() *Date {
	return &Date{
		Date: &gdate.Unknown{},
	}
}

func PreciseDate(y, m, d int) *Date {
	if m < 1 || m > 12 {
		panic("month must be between 1 and 12")
	}
	return &Date{
		Date: &gdate.Precise{Y: y, M: m, D: d},
	}
}

func Year(y int) *Date {
	return &Date{
		Date: &gdate.Year{Y: y},
	}
}

func AboutYear(y int) *Date {
	return &Date{
		Date: &gdate.AboutYear{Y: y},
	}
}

func BeforeYear(y int) *Date {
	return &Date{
		Date: &gdate.BeforeYear{Y: y},
	}
}

func AfterYear(y int) *Date {
	return &Date{
		Date: &gdate.AfterYear{Y: y},
	}
}

// IsUnknown reports whether d is an Unknown date
func (d *Date) IsUnknown() bool {
	if d == nil {
		return true
	}

	_, ok := d.Date.(*gdate.Unknown)
	return ok
}

// IsEstimated reports whether d is a firm date or range of dates
func (d *Date) IsFirm() bool {
	if d == nil {
		return false
	}

	switch d.Date.(type) {
	case *gdate.Precise, *gdate.MonthYear, *gdate.Year, *gdate.YearQuarter:
		return true
	}

	return false
}

func (d *Date) String() string {
	if d == nil {
		return "unknown"
	}

	return d.Date.String()
}

func (d *Date) When() string {
	if d == nil {
		return "on an unknown date"
	}

	return d.Date.Occurrence()
}

func (d *Date) WhenYear() (string, bool) {
	yr, ok := d.Year()
	if !ok {
		return "", false
	}

	switch d.Date.(type) {
	case *gdate.BeforeYear:
		return "before " + strconv.Itoa(yr), true
	case *gdate.AfterYear:
		return "after " + strconv.Itoa(yr), true
	case *gdate.AboutYear:
		return "about " + strconv.Itoa(yr), true
	default:
		return "in " + strconv.Itoa(yr), true
	}
}

func (d *Date) Year() (int, bool) {
	if d == nil {
		return 0, false
	}

	yearer, ok := gdate.AsYear(d.Date)
	if !ok {
		return 0, false
	}

	return yearer.Year(), true
}

func (d *Date) YMD() (int, int, int, bool) {
	if d == nil {
		return 0, 0, 0, false
	}

	if p, ok := gdate.AsPrecise(d.Date); ok {
		return p.Y, p.M, p.D, true
	}
	return 0, 0, 0, false
}

func (d *Date) DateInYear(long bool) (string, bool) {
	if d == nil {
		return "", false
	}

	if a, ok := d.Date.(interface{ DateInYear(bool) string }); ok {
		return a.DateInYear(long), true
	}
	return "", false
}

func (d *Date) SameYear(other *Date) bool {
	dy, ok := d.Year()
	if !ok {
		return false
	}

	oy, ok := other.Year()
	if !ok {
		return false
	}

	return dy == oy
}

func (d *Date) SortsBefore(other *Date) bool {
	if d == nil {
		return false
	}
	if other == nil {
		return true
	}

	return gdate.SortsBefore(d.Date, other.Date)
}

func (d *Date) IntervalUntil(other *Date) *Interval {
	if d == nil || other == nil {
		return UnknownInterval()
	}

	return &Interval{
		Interval: gdate.IntervalBetween(d.Date, other.Date),
	}
}

func (d *Date) WholeYearsUntil(other *Date) (int, bool) {
	if d == nil || other == nil {
		return 0, false
	}

	in := d.IntervalUntil(other)
	if in.IsUnknown() {
		return 0, false
	}

	return in.WholeYears()
}

func IntervalSince(d *Date) *Interval {
	if d.IsUnknown() {
		return UnknownInterval()
	}
	now := time.Now()
	dt := &gdate.Precise{
		Y: now.Year(),
		M: int(now.Month()),
		D: now.Day(),
	}
	return &Interval{
		Interval: gdate.IntervalBetween(d.Date, dt),
	}
}

type Interval struct {
	Interval gdate.Interval
}

func UnknownInterval() *Interval {
	return &Interval{
		Interval: &gdate.UnknownInterval{},
	}
}

func (in *Interval) IsUnknown() bool {
	if in == nil {
		return true
	}

	return gdate.IsUnknownInterval(in.Interval)
}

func (in *Interval) WholeYears() (int, bool) {
	if in == nil {
		return 0, false
	}
	if yi, ok := gdate.AsYearsInterval(in.Interval); ok {
		return yi.Years(), true
	}
	return 0, false
}

func (in *Interval) ApproxDays() (int, bool) {
	if in == nil {
		return 0, false
	}

	if a, ok := in.Interval.(interface{ ApproxDays() int }); ok {
		return a.ApproxDays(), true
	}

	return 0, false
}

func (in *Interval) YMD() (int, int, int, bool) {
	if in == nil {
		return 0, 0, 0, false
	}

	if p, ok := gdate.AsPreciseInterval(in.Interval); ok {
		return p.Y, p.M, p.D, true
	}
	return 0, 0, 0, false
}

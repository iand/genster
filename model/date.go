package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iand/gdate"
)

type Date struct {
	Date       gdate.Date
	Derivation DateDerivation

	// Span indicates whether this date is a span of time
	// when true the date represents a span of time during which something continuously occurred (such as a residing at an address)
	// when false it represents a range of time during which a discrete event occurred
	Span bool
}

type DateDerivation int

const (
	DateDerivationStandard   DateDerivation = 0 // date is as given in source
	DateDerivationEstimated  DateDerivation = 1 // date was estimated from typical values (such as year of birth estimated from marriage)
	DateDerivationCalculated DateDerivation = 2 // date was calculated from another date (such as date of birth calculated using age at death)
)

func (d DateDerivation) Qualifier() string {
	switch d {
	case DateDerivationEstimated:
		return "estimated"
	case DateDerivationCalculated:
		return "calculated"
	default:
		return ""
	}
}

func (d DateDerivation) Abbrev() string {
	switch d {
	case DateDerivationEstimated:
		return "est."
	case DateDerivationCalculated:
		return "calc."
	default:
		return ""
	}
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

// YearRange returns a date that represents a range of time between two years, including the upper and lower year.
func YearRange(l, u int) *Date {
	return &Date{
		Date: &gdate.YearRange{Lower: l, Upper: u},
	}
}

// YearSpan returns a date that represents a span of time between two years, including the upper and lower year.
func YearSpan(l, u int) *Date {
	return &Date{
		Date: &gdate.YearRange{Lower: l, Upper: u},
		Span: true,
	}
}

func WithinDecade(y int) *Date {
	return &Date{
		Date: &gdate.YearRange{Lower: y, Upper: y + 9},
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

// IsEstimated reports whether d is a firm date or date range. Imprecise or estimated
// dates return false.
func (d *Date) IsFirm() bool {
	if d == nil {
		return false
	}

	if d.Derivation != DateDerivationStandard {
		return false
	}

	switch d.Date.(type) {
	case *gdate.Precise, *gdate.MonthYear, *gdate.Year, *gdate.YearQuarter:
		return true
	}

	return false
}

// IsMorePreciseThan reports whether d is more a more precise date than o.
func (d *Date) IsMorePreciseThan(o *Date) bool {
	if o.IsUnknown() {
		return true
	}
	if d.IsUnknown() {
		return false
	}

	// from here both dates are known

	if d.Derivation != DateDerivationStandard || o.Derivation != DateDerivationStandard {
		return d.Derivation == DateDerivationStandard
	}

	// from here both dates have standard derivation

	switch d.Date.(type) {
	case *gdate.Precise:
		switch o.Date.(type) {
		case *gdate.Precise:
			return false
		default:
			return true
		}
	case *gdate.MonthYear:
		switch o.Date.(type) {
		case *gdate.Precise:
			return false
		default:
			return true
		}
	case *gdate.YearQuarter:
		switch o.Date.(type) {
		case *gdate.Precise, *gdate.MonthYear:
			return false
		default:
			return true
		}
	case *gdate.Year:
		switch o.Date.(type) {
		case *gdate.Precise, *gdate.MonthYear, *gdate.YearQuarter:
			return false
		default:
			return true
		}
	default:
		switch o.Date.(type) {
		case *gdate.Precise, *gdate.MonthYear, *gdate.YearQuarter, *gdate.Year:
			return false
		default:
			return true
		}
	}
}

func (d *Date) String() string {
	if d == nil {
		return "unknown"
	}

	qual := d.Derivation.Qualifier()
	if qual != "" {
		qual += " "
	}

	return qual + d.Date.String()
}

func (d *Date) When() string {
	if d == nil {
		return "on an unknown date"
	}

	return d.Date.Occurrence()

	// if d.Date.Calendar() == gdate.Gregorian {
	// 	return d.Date.Occurrence()
	// }

	// cal := ""
	// switch d.Date.Calendar() {
	// case gdate.Julian, gdate.Julian25Mar:
	// 	cal = "julian"
	// default:
	// 	cal = "unknown calendar"
	// }

	// return fmt.Sprintf("%s (%s)", d.Date.Occurrence(), cal)
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

func (d *Date) AsYear() (*Date, bool) {
	if d == nil {
		return nil, false
	}

	y, ok := gdate.AsYear(d.Date)
	if !ok {
		return nil, false
	}
	return &Date{
		Date: y,
	}, true
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

func (d *Date) YM() (int, int, bool) {
	if d == nil {
		return 0, 0, false
	}

	if p, ok := gdate.AsPrecise(d.Date); ok {
		return p.Y, p.M, true
	}
	if p, ok := d.Date.(*gdate.MonthYear); ok {
		return p.Y, p.M, true
	}
	return 0, 0, false
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

func (d *Date) SameDate(other *Date) bool {
	if d.IsUnknown() || other.IsUnknown() {
		return false
	}
	p1, ok := gdate.AsPrecise(d.Date)
	if !ok {
		return false
	}

	p2, ok := gdate.AsPrecise(other.Date)
	if !ok {
		return false
	}

	return p1.Y == p2.Y && p1.M == p2.M && p1.D == p2.D
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

func (d *Date) DecadeStart() (int, bool) {
	if d == nil {
		return 0, false
	}

	yearer, ok := gdate.AsYear(d.Date)
	if !ok {
		return 0, false
	}

	return (yearer.Year() / 10) * 10, true
}

// Gedcom returns the date formatted for gedcom
func (d *Date) Gedcom() string {
	var prefix string
	switch d.Derivation {
	case DateDerivationEstimated:
		prefix = "EST "
	case DateDerivationCalculated:
		prefix = "CAL "
	}

	if d.Span {
		panic("unsupported span date type in Gedcom conversion")
	}

	switch dt := d.Date.(type) {
	case *gdate.Precise:
		return fmt.Sprintf("%s%d %s %d", prefix, dt.D, strings.ToUpper(shortMonthNames[dt.M]), dt.Y)
	case *gdate.MonthYear:
		return fmt.Sprintf("%s%s %d", prefix, strings.ToUpper(shortMonthNames[dt.M]), dt.Y)
	case *gdate.Year:
		return fmt.Sprintf("%s%d", prefix, dt.Y)
	case *gdate.YearQuarter:
		switch dt.Q {
		case 1:
			return fmt.Sprintf("%sBET 1 JAN %d AND 31 MAR %[2]d", prefix, dt.Y)
		case 2:
			return fmt.Sprintf("%sBET 1 APR %d AND 30 JUN %[2]d", prefix, dt.Y)
		case 3:
			return fmt.Sprintf("%sBET 1 JUL %d AND 30 SEP %[2]d", prefix, dt.Y)
		case 4:
			return fmt.Sprintf("%sBET 1 OCT %d AND 31 DEC %[2]d", prefix, dt.Y)
		default:
			panic(fmt.Sprintf("unsupported date quarter in Gedcom conversion: %v", dt.Q))
		}
	case *gdate.BeforeYear:
		return fmt.Sprintf("%sBEF %d", prefix, dt.Y)
	case *gdate.AfterYear:
		return fmt.Sprintf("%sAFT %d", prefix, dt.Y)
	case *gdate.AboutYear:
		return fmt.Sprintf("%sABT %d", prefix, dt.Y)
	default:
		panic(fmt.Sprintf("unsupported date type in Gedcom conversion: %T", dt))
	}
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

var shortMonthNames = []string{
	1:  "Jan",
	2:  "Feb",
	3:  "Mar",
	4:  "Apr",
	5:  "May",
	6:  "Jun",
	7:  "Jul",
	8:  "Aug",
	9:  "Sep",
	10: "Oct",
	11: "Nov",
	12: "Dec",
}

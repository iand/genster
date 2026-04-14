package gramps

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

func ptrStr(s string) *string { return &s }
func ptrBool(b bool) *bool    { return &b }

func TestEventDate(t *testing.T) {
	testCases := []struct {
		ev      *grampsxml.Event
		want    *model.Date
		wantErr bool
	}{
		{
			ev: &grampsxml.Event{},
			want: &model.Date{
				Date: &gdate.Unknown{},
			},
		},

		// Dateval
		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val: "1995-08-04",
				},
			},
			want: &model.Date{
				Date: &gdate.Precise{
					C: gdate.Gregorian,
					D: 4,
					M: 8,
					Y: 1995,
				},
			},
		},
		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val: "1985",
				},
			},
			want: &model.Date{
				Date: &gdate.Year{
					C: gdate.Gregorian,
					Y: 1985,
				},
			},
		},
		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val: "1921-03",
				},
			},
			want: &model.Date{
				Date: &gdate.MonthYear{
					C: gdate.Gregorian,
					M: 3,
					Y: 1921,
				},
			},
		},

		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:  "1985",
					Type: p("Before"),
				},
			},
			want: &model.Date{
				Date: &gdate.BeforeYear{
					C: gdate.Gregorian,
					Y: 1985,
				},
			},
		},

		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:  "1985",
					Type: p("After"),
				},
			},
			want: &model.Date{
				Date: &gdate.AfterYear{
					C: gdate.Gregorian,
					Y: 1985,
				},
			},
		},

		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:  "1985",
					Type: p("About"),
				},
			},
			want: &model.Date{
				Date: &gdate.AboutYear{
					C: gdate.Gregorian,
					Y: 1985,
				},
			},
		},

		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:  "1985-10-01",
					Type: p("Before"),
				},
			},
			want: &model.Date{
				Date: &gdate.BeforePrecise{
					C: gdate.Gregorian,
					Y: 1985,
					M: 10,
					D: 1,
				},
			},
		},
		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:  "1985-10-01",
					Type: p("After"),
				},
			},
			want: &model.Date{
				Date: &gdate.AfterPrecise{
					C: gdate.Gregorian,
					Y: 1985,
					M: 10,
					D: 1,
				},
			},
		},
		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:  "1985-10-01",
					Type: p("About"),
				},
			},
			wantErr: true,
		},

		// Datestr
		{
			ev: &grampsxml.Event{
				Datestr: &grampsxml.Datestr{
					Val: "1995-08-04",
				},
			},
			want: &model.Date{
				Date: &gdate.Precise{
					C: gdate.Gregorian,
					D: 4,
					M: 8,
					Y: 1995,
				},
			},
		},

		// Daterange
		{
			ev: &grampsxml.Event{
				Daterange: &grampsxml.Daterange{
					Start: "1936-10",
					Stop:  "1936-12",
				},
			},
			want: &model.Date{
				Date: &gdate.YearQuarter{
					C: gdate.Gregorian,
					Q: 4,
					Y: 1936,
				},
			},
		},

		// Dual-dated Julian: Gramps stores the New Style year in val.
		// val="1651-03-10", dualdated=1 → subtract 1 → OS year 1650.
		// Should display as "10 Mar 1650/51" and Year() should return 1651 (NS year).
		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:       "1651-03-10",
					Cformat:   ptrStr("Julian"),
					Dualdated: ptrBool(true),
				},
			},
			want: &model.Date{
				Date: &gdate.Precise{
					C: gdate.Julian25Mar,
					D: 10,
					M: 3,
					Y: 1650,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			dp := gdate.Parser{
				AssumeGROQuarter: false,
			}
			got, err := EventDate(tc.ev, dp)
			if tc.wantErr {
				if err != nil {
					return
				}
				t.Fatalf("missing expected error")
			}

			if err != nil {
				t.Fatalf("got unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want.Date, got.Date); diff != "" {
				t.Errorf("got date mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestParseDatevalSortOrder verifies that a Julian dual-dated date in early March
// sorts chronologically after a mid-year date in the same OS year.
//
// "10 Mar 1650/51 (Julian)" (Gramps: val="1651-03-10", dualdated=1) is OS March 10,
// 1650 = Julian March 10, 1651, which must sort AFTER "4 Jun 1650 (Julian)"
// (Gramps: val="1650-06-04").
func TestParseDatevalSortOrder(t *testing.T) {
	dp := gdate.Parser{AssumeGROQuarter: false}

	// 10 Mar 1650/51 OS — Gramps stores NS year 1651 in val.
	earlyMarch, err := ParseDateval(grampsxml.Dateval{
		Val:       "1651-03-10",
		Cformat:   ptrStr("Julian"),
		Dualdated: ptrBool(true),
	}, dp)
	if err != nil {
		t.Fatalf("ParseDateval(1651-03-10 Julian dualdated): %v", err)
	}

	// 4 Jun 1650 Julian — plain Julian date, no dual date.
	june, err := ParseDateval(grampsxml.Dateval{
		Val:     "1650-06-04",
		Cformat: ptrStr("Julian"),
	}, dp)
	if err != nil {
		t.Fatalf("ParseDateval(1650-06-04 Julian): %v", err)
	}

	marchJD, ok1 := earlyMarch.Date.(gdate.ComparableDate)
	juneJD, ok2 := june.Date.(gdate.ComparableDate)
	if !ok1 || !ok2 {
		t.Fatal("dates are not ComparableDate")
	}

	if marchJD.EarliestJulianDay() <= juneJD.EarliestJulianDay() {
		t.Errorf("10 Mar 1650/51 OS (JD=%d) should sort after 4 Jun 1650 (JD=%d)",
			marchJD.EarliestJulianDay(), juneJD.EarliestJulianDay())
	}

	// Display checks
	if got := earlyMarch.Date.String(); got != "10 Mar 1650/51" {
		t.Errorf("display: got %q, want %q", got, "10 Mar 1650/51")
	}
	if got := earlyMarch.Date.(interface{ Year() int }).Year(); got != 1651 {
		t.Errorf("Year(): got %d, want 1651", got)
	}
}

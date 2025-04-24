package gramps

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
	"github.com/iand/grampsxml"
)

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
			wantErr: true,
		},
		{
			ev: &grampsxml.Event{
				Dateval: &grampsxml.Dateval{
					Val:  "1985-10-01",
					Type: p("After"),
				},
			},
			wantErr: true,
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

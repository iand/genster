package model

import "testing"

func TestDateDefinitelyBeforeBy(t *testing.T) {
	testCases := []struct {
		name  string
		d     *Date
		other *Date
		days  int
		want  bool
	}{
		{name: "fifteen years apart", d: PreciseDate(1826, 12, 27), other: PreciseDate(1841, 6, 6), days: 365, want: true},
		{name: "ten months apart", d: PreciseDate(1860, 3, 1), other: PreciseDate(1861, 1, 10), days: 365, want: false},
		{name: "year-only dates two years apart", d: Year(1834), other: Year(1836), days: 365, want: true},
		{name: "year-only dates in consecutive years", d: Year(1834), other: Year(1835), days: 365, want: false},
		{name: "qualified date that may fall within the margin", d: PreciseDate(1860, 3, 1), other: BeforeYear(1870), days: 365, want: false},
		{name: "zero days is DefinitelyBefore", d: PreciseDate(1830, 5, 1), other: PreciseDate(1830, 5, 2), days: 0, want: true},
		{name: "unknown date", d: UnknownDate(), other: Year(1900), days: 365, want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.d.DefinitelyBeforeBy(tc.other, tc.days); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

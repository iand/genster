package site

import (
	"testing"
)

func TestNormalizePlaceName(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			input: "abc",
			want:  "abc",
		},
		{
			input: "ab c",
			want:  "ab c",
		},
		{
			input: "ab  c",
			want:  "ab c",
		},
		{
			name:  "leading space",
			input: " abc",
			want:  "abc",
		},
		{
			name:  "leading spaces",
			input: "    abc",
			want:  "abc",
		},
		{
			name:  "trailing space",
			input: "abc ",
			want:  "abc",
		},
		{
			name:  "trailing spaces",
			input: "abc    ",
			want:  "abc",
		},
		{
			input: "abc, def",
			want:  "abc, def",
		},
		{
			input: "abc,def",
			want:  "abc, def",
		},
		{
			input: "abc,  def",
			want:  "abc, def",
		},
		{
			input: "abc ,def",
			want:  "abc, def",
		},
		{
			input: "abc,,def",
			want:  "abc, def",
		},
		{
			input: "abc,",
			want:  "abc",
		},
		{
			input: "abc,,",
			want:  "abc",
		},
		{
			input: "abc, ,def",
			want:  "abc, def",
		},
		{
			input: ",,abc,def",
			want:  "abc, def",
		},
	}

	for _, tc := range testCases {
		name := tc.name
		if name == "" {
			name = tc.input
		}
		t.Run(name, func(t *testing.T) {
			got := normalizePlaceName(tc.input)
			if got != tc.want {
				t.Errorf("got %q, wanted %q", got, tc.want)
			}
		})
	}
}

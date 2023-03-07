package place

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		s    string
		alts []string
		err  bool
		want Place
	}{}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			pl, _ := Parse(tc.s)

			if diff := cmp.Diff(tc.want, pl); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tc.s, diff)
			}

			for _, alt := range tc.alts {
				if diff := cmp.Diff(tc.want, alt); diff != "" {
					t.Errorf("Parse(%q) mismatch (-want +got):\n%s", alt, diff)
				}
			}
		})
	}
}

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
		{
			input: "abc; def",
			want:  "abc, def",
		},
		{
			input: "",
			want:  "",
		},
		{
			input: ",",
			want:  "",
		},
		{
			input: "Abc,dEf",
			want:  "abc, def",
		},
	}

	for _, tc := range testCases {
		name := tc.name
		if name == "" {
			name = tc.input
		}
		t.Run(tc.name, func(t *testing.T) {
			got := Normalize(tc.input)
			if got != tc.want {
				t.Errorf("got %q, wanted %q", got, tc.want)
			}
		})
	}
}

func TestSplitPlaceName(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []string
	}{
		{
			input: "abc",
			want:  []string{"abc"},
		},
		{
			input: "ab c",
			want:  []string{"ab c"},
		},
		{
			input: "ab  c",
			want:  []string{"ab c"},
		},
		{
			name:  "leading space",
			input: " abc",
			want:  []string{"abc"},
		},
		{
			name:  "leading spaces",
			input: "    abc",
			want:  []string{"abc"},
		},
		{
			name:  "trailing space",
			input: "abc ",
			want:  []string{"abc"},
		},
		{
			name:  "trailing spaces",
			input: "abc    ",
			want:  []string{"abc"},
		},
		{
			input: "abc, def",
			want:  []string{"abc", "def"},
		},
		{
			input: "abc,def",
			want:  []string{"abc", "def"},
		},
		{
			input: "abc,  def",
			want:  []string{"abc", "def"},
		},
		{
			input: "abc ,def",
			want:  []string{"abc", "def"},
		},
		{
			input: "abc,,def",
			want:  []string{"abc", "def"},
		},
		{
			input: "abc,",
			want:  []string{"abc"},
		},
		{
			input: "abc,,",
			want:  []string{"abc"},
		},
		{
			input: "abc, ,def",
			want:  []string{"abc", "def"},
		},
		{
			input: ",,abc,def",
			want:  []string{"abc", "def"},
		},
		{
			input: "abc; def",
			want:  []string{"abc", "def"},
		},
		{
			input: "",
			want:  []string{},
		},
		{
			input: ",",
			want:  []string{},
		},
		{
			input: "Abc,dEf",
			want:  []string{"Abc", "dEf"},
		},
	}

	for _, tc := range testCases {
		name := tc.name
		if name == "" {
			name = tc.input
		}
		t.Run(tc.name, func(t *testing.T) {
			got := splitPlaceName(tc.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tc.input, diff)
			}
		})
	}
}

package text

import (
	"testing"
)

func TestJoinSentenceParts(t *testing.T) {
	testCases := []struct {
		name  string
		parts []string
		want  string
	}{
		// basic joining
		{
			name:  "no_parts",
			parts: nil,
			want:  "",
		},
		{
			name:  "single_part",
			parts: []string{"foo"},
			want:  "foo",
		},
		{
			name:  "two_parts",
			parts: []string{"foo", "bar"},
			want:  "foo bar",
		},
		{
			name:  "three_parts",
			parts: []string{"foo", "bar", "baz"},
			want:  "foo bar baz",
		},

		// empty and whitespace-only parts are skipped
		{
			name:  "single_empty",
			parts: []string{""},
			want:  "",
		},
		{
			name:  "all_empty",
			parts: []string{"", "", ""},
			want:  "",
		},
		{
			name:  "leading_empty",
			parts: []string{"", "foo"},
			want:  "foo",
		},
		{
			name:  "trailing_empty",
			parts: []string{"foo", ""},
			want:  "foo",
		},
		{
			name:  "middle_empty",
			parts: []string{"foo", "", "bar"},
			want:  "foo bar",
		},
		{
			name:  "whitespace_only_part",
			parts: []string{"  ", "foo"},
			want:  "foo",
		},
		{
			name:  "surrounding_whitespace_trimmed",
			parts: []string{"  foo  ", "  bar  "},
			want:  "foo bar",
		},

		// colon is appended without a preceding space
		{
			name:  "colon_as_last_part",
			parts: []string{"foo", ":"},
			want:  "foo:",
		},
		{
			name:  "colon_in_middle",
			parts: []string{"foo", ":", "bar"},
			want:  "foo: bar",
		},
		{
			name:  "colon_as_first_part",
			parts: []string{":", "foo"},
			want:  ": foo",
		},

		// case is never modified
		{
			name:  "capitalised_mid_part_unchanged",
			parts: []string{"was born", "He died"},
			want:  "was born He died",
		},
		{
			name:  "capitalised_first_part_unchanged",
			parts: []string{"He went", "to London"},
			want:  "He went to London",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := JoinSentenceParts(tc.parts...)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

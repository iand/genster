package text

import "testing"

func TestParaContinue(t *testing.T) {
	testCases := []struct {
		name  string
		build func(*Para)
		want  string
	}{
		// no-op calls leave the para empty
		{
			name:  "no_args",
			build: func(p *Para) { p.Continue() },
			want:  "",
		},
		{
			name:  "empty_string",
			build: func(p *Para) { p.Continue("") },
			want:  "",
		},
		{
			name:  "whitespace_only",
			build: func(p *Para) { p.Continue("  ") },
			want:  "",
		},
		{
			name:  "all_empty_parts",
			build: func(p *Para) { p.Continue("", "", "") },
			want:  "",
		},

		// single call on a fresh para: case is preserved, not modified
		{
			name:  "single_lowercase_part",
			build: func(p *Para) { p.Continue("foo") },
			want:  "foo",
		},
		{
			name:  "single_uppercase_part",
			build: func(p *Para) { p.Continue("Foo") },
			want:  "Foo",
		},
		{
			name:  "multiple_parts_in_one_call",
			build: func(p *Para) { p.Continue("foo", "bar", "baz") },
			want:  "foo bar baz",
		},

		// multiple calls accumulate into the same sentence
		{
			name: "two_calls",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("bar")
			},
			want: "foo bar",
		},
		{
			name: "three_calls",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("bar")
				p.Continue("baz")
			},
			want: "foo bar baz",
		},

		// empty and whitespace parts are skipped
		{
			name:  "empty_middle_part",
			build: func(p *Para) { p.Continue("foo", "", "bar") },
			want:  "foo bar",
		},
		{
			name:  "whitespace_middle_part",
			build: func(p *Para) { p.Continue("foo", "  ", "bar") },
			want:  "foo bar",
		},
		{
			name: "empty_call_between_content",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue()
				p.Continue("bar")
			},
			want: "foo bar",
		},
		{
			name: "empty_call_after_content",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("")
			},
			want: "foo",
		},

		// subsequent parts are appended as-is
		{
			name: "second_part_appended",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("bar baz")
			},
			want: "foo bar baz",
		},

		// FinishSentence uppercases and adds a period; the next open sentence
		// is raw again until it is finalized
		{
			name: "continue_after_finish",
			build: func(p *Para) {
				p.Continue("first")
				p.FinishSentence()
				p.Continue("second")
			},
			want: "First. second",
		},
		{
			name: "three_sentences",
			build: func(p *Para) {
				p.Continue("foo")
				p.FinishSentence()
				p.Continue("bar")
				p.FinishSentence()
				p.Continue("baz")
			},
			want: "Foo. Bar. baz",
		},

		// FinishSentence on empty para is a no-op
		{
			name: "finish_on_empty_then_continue",
			build: func(p *Para) {
				p.FinishSentence()
				p.Continue("foo")
			},
			want: "foo",
		},

		// empty leading parts are skipped; p.join emits a leading space for
		// index > 0 parts, which Current's TrimSpace removes
		{
			name:  "empty_first_part_leading_space_trimmed",
			build: func(p *Para) { p.Continue("", "he went") },
			want:  "he went",
		},
		{
			name:  "empty_first_part_capital_preserved",
			build: func(p *Para) { p.Continue("", "He went") },
			want:  "He went",
		},

		// when a call to Continue has an empty leading part and appends to existing
		// content, p.join's leading space combines with Continue's separator to
		// produce a double space
		{
			name: "empty_leading_part_mid_sentence_produces_double_space",
			build: func(p *Para) {
				p.Continue("was born")
				p.Continue("", "He died")
			},
			want: "was born  He died",
		},

		// case is never modified during accumulation
		{
			name: "capitalised_mid_part_preserved",
			build: func(p *Para) {
				p.Continue("was born")
				p.Continue("He died")
			},
			want: "was born He died",
		},
		{
			name:  "capitalised_mid_part_same_call_preserved",
			build: func(p *Para) { p.Continue("was born", "He died") },
			want:  "was born He died",
		},
		{
			name: "she_preserved",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("She went")
			},
			want: "foo She went",
		},
		{
			name: "they_preserved",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("They went")
			},
			want: "foo They went",
		},
		{
			name: "the_preserved",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("The bar")
			},
			want: "foo The bar",
		},

		// colon is appended without a preceding space, matching JoinSentenceParts
		{
			name: "colon_no_preceding_space",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue(":")
			},
			want: "foo:",
		},
		{
			name:  "colon_no_preceding_space_same_call",
			build: func(p *Para) { p.Continue("foo", ":") },
			want:  "foo:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var p Para
			tc.build(&p)
			got := p.Current()
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParaCurrent(t *testing.T) {
	testCases := []struct {
		name  string
		build func(*Para)
		want  string
	}{
		// empty para
		{
			name:  "empty",
			build: func(p *Para) {},
			want:  "",
		},

		// open sentence is returned as-is: no period, no uppercasing
		{
			name:  "open_sentence_no_period_no_uppercase",
			build: func(p *Para) { p.Continue("foo") },
			want:  "foo",
		},
		{
			name: "open_sentence_multiple_calls",
			build: func(p *Para) {
				p.Continue("foo")
				p.Continue("bar")
			},
			want: "foo bar",
		},

		// finished sentences carry their period and uppercase; the open trailing
		// sentence does not
		{
			name: "finished_plus_open",
			build: func(p *Para) {
				p.Continue("first")
				p.FinishSentence()
				p.Continue("second")
			},
			want: "First. second",
		},
		{
			name: "two_finished_sentences",
			build: func(p *Para) {
				p.Continue("first")
				p.FinishSentence()
				p.Continue("second")
				p.FinishSentence()
			},
			want: "First. Second.",
		},

		// Current does not mutate the para: Text() still finalizes correctly afterward
		{
			name: "current_does_not_prevent_text_from_adding_period",
			build: func(p *Para) {
				p.Continue("foo")
				_ = p.Current()
			},
			want: "foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var p Para
			tc.build(&p)
			got := p.Current()
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}

	// Verify non-mutation: calling Current() must not prevent Text() from
	// uppercasing and adding a period to the open sentence.
	t.Run("current_does_not_mutate", func(t *testing.T) {
		var p Para
		p.Continue("foo")
		if got := p.Current(); got != "foo" {
			t.Fatalf("Current() got %q, want %q", got, "foo")
		}
		if got := p.Text(); got != "Foo." {
			t.Errorf("Text() after Current() got %q, want %q", got, "Foo.")
		}
	})
}

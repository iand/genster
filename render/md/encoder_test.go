package md

import "testing"

// Paragraphs are HTML elements that separate themselves, so an empty paragraph adds no element,
// only a line break.
func TestContentEmptyParagraphs(t *testing.T) {
	c := &Content{}
	c.Para("First.")
	c.EmptyPara()
	c.Para("")
	c.Para("Second.")

	want := "<p>First.</p>\n\n<p>Second.</p>\n\n"
	if got := c.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

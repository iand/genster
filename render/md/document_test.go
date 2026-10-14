package md

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iand/genster/model"
)

// TestDocumentFrontMatterDictKeysSorted checks that dictionary entries in front
// matter are written with their keys in sorted order, so output is stable.
func TestDocumentFrontMatterDictKeysSorted(t *testing.T) {
	want := "links: \n- category: website\n  link: \"https://example.com/a\"\n  title: Example\n"

	for i := range 20 {
		d := &Document{}
		d.AddLink("Example", "https://example.com/a", model.LinkCategory("website"))

		var buf bytes.Buffer
		if _, err := d.WriteTo(&buf); err != nil {
			t.Fatalf("WriteTo: %v", err)
		}
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("run %d: front matter not in sorted key order\ngot:\n%s\nwant to contain:\n%s", i, buf.String(), want)
		}
	}
}

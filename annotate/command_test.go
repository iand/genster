package annotate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iand/genster/model"
)

func TestAnnotateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.md")

	original := "Some text [^foo] more text.\n\n[^foo]: A footnote definition.\n"
	if err := os.WriteFile(fpath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	lb := &citationLinkBuilder{citationLinkPattern: "/trees/test/citation/%s/"}

	// Step 1: annotate the file.
	changed, err := processFile(fpath, map[string]*model.GeneralCitation{}, lb, false, 1000)
	if err != nil {
		t.Fatalf("processFile: %v", err)
	}
	if !changed {
		t.Fatal("expected processFile to report a change")
	}

	annotatedBytes, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}
	annotated := string(annotatedBytes)

	// The annotated content must contain the private shortcode markers inside
	// the HTML comment delimiters so Hugo does not carry them into generated HTML.
	if !strings.Contains(annotated, "<!-- "+shortcodeOpen) {
		t.Errorf("annotated content missing shortcode open inside comment:\n%s", annotated)
	}
	if !strings.Contains(annotated, shortcodeClose+" -->") {
		t.Errorf("annotated content missing shortcode close inside comment:\n%s", annotated)
	}

	// The footnote definition must no longer appear as a raw definition line.
	if strings.Contains(annotated, "\n[^foo]:") {
		t.Errorf("annotated content should not contain raw footnote definition:\n%s", annotated)
	}

	// The citations section must also use the private shortcode markers.
	if !strings.Contains(annotated, shortcodeOpen+"begin citations"+shortcodeClose) {
		t.Errorf("annotated content missing private shortcode around begin-citations marker:\n%s", annotated)
	}
	if !strings.Contains(annotated, shortcodeOpen+"end citations"+shortcodeClose) {
		t.Errorf("annotated content missing private shortcode around end-citations marker:\n%s", annotated)
	}

	// Step 2: undo the annotations.
	changed, err = undoFile(fpath, false, 2000)
	if err != nil {
		t.Fatalf("undoFile: %v", err)
	}
	if !changed {
		t.Fatal("expected undoFile to report a change")
	}

	restoredBytes, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}
	restored := string(restoredBytes)

	// The inline footnote reference must be present.
	if !strings.Contains(restored, "[^foo]") {
		t.Errorf("restored content missing footnote reference:\n%s", restored)
	}

	// The footnote definition must be restored at the end.
	if !strings.Contains(restored, "[^foo]: A footnote definition.") {
		t.Errorf("restored content missing footnote definition:\n%s", restored)
	}

	// No shortcode markers should remain in the restored content.
	if strings.Contains(restored, shortcodeOpen) || strings.Contains(restored, shortcodeClose) {
		t.Errorf("restored content should not contain shortcode markers:\n%s", restored)
	}
}

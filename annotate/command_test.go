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
	if err := os.WriteFile(fpath, []byte(original), 0o644); err != nil {
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

	// The footnote definition must no longer appear as a raw definition line.
	if strings.Contains(annotated, "\n[^foo]:") {
		t.Errorf("annotated content should not contain raw footnote definition:\n%s", annotated)
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
}

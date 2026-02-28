package build

import (
	"strings"
	"testing"
	"time"

	"github.com/iand/genster/render/md"
)

// TestParseDocumentRoundTrip verifies that every field type written by
// render/md.Document can be read back correctly by ParseDocument.
func TestParseDocumentRoundTrip(t *testing.T) {
	var doc md.Document

	// String fields — one per category to exercise all code paths.
	doc.ID("I1234")
	doc.Title("John Smith (1850–1920)") // non-ASCII, needs quoting
	doc.Layout("person")
	doc.Summary("A shopkeeper from Cornwall.")
	doc.Category("person")
	doc.Image("/media/john.jpg")
	doc.BasePath("/trees/test/")
	doc.LastUpdated(time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC))
	doc.SetFrontMatterField("gender", "male")
	doc.SetFrontMatterField("era", "1800s")   // contains digit, needs quoting
	doc.SetFrontMatterField("maturity", "mature")
	doc.SetFrontMatterField("trade", "labourer")
	doc.SetFrontMatterField("grampsid", "I1234")
	doc.SetFrontMatterField("slug", "johnsmith")
	doc.SetFrontMatterField("wikitreeformat", "SmithJohn")
	doc.SetFrontMatterField("wikitreeid", "Smith-1")
	doc.SetFrontMatterField("markdownformat", "standard")
	doc.SetFrontMatterField("placetype", "village")
	doc.SetFrontMatterField("buildingkind", "church")
	doc.SetFrontMatterField("month", "January")

	// []string fields
	doc.AddTag("surname:smith")
	doc.AddTag("location:cornwall")
	doc.AddAlias("/old/john-smith/")

	// map[string]string field (sitemap)
	doc.SetSitemapDisable()

	// []map[string]string fields
	doc.AddDiaryLink("Research entry", "/diary/2025/01/15/")
	doc.AddLink("WikiTree profile", "https://wikitree.com/wiki/Smith-1")
	doc.AddDescendant("Jane Smith", "/person/I5678/", "b. 1875")

	content := doc.String()

	fm, body, err := ParseDocument(content)
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}

	// --- string fields ---
	for _, tt := range []struct {
		name string
		got  string
		want string
	}{
		{"ID", fm.ID, "I1234"},
		{"Title", fm.Title, "John Smith (1850–1920)"},
		{"Layout", fm.Layout, "person"},
		{"Summary", fm.Summary, "A shopkeeper from Cornwall."},
		{"Category", fm.Category, "person"},
		{"Image", fm.Image, "/media/john.jpg"},
		{"BasePath", fm.BasePath, "/trees/test/"},
		{"LastMod", fm.LastMod, "2026-02-28T12:00:00Z"},
		{"Gender", fm.Gender, "male"},
		{"Era", fm.Era, "1800s"},
		{"Maturity", fm.Maturity, "mature"},
		{"Trade", fm.Trade, "labourer"},
		{"GrampsID", fm.GrampsID, "I1234"},
		{"Slug", fm.Slug, "johnsmith"},
		{"WikiTreeFormat", fm.WikiTreeFormat, "SmithJohn"},
		{"WikiTreeID", fm.WikiTreeID, "Smith-1"},
		{"MarkdownFormat", fm.MarkdownFormat, "standard"},
		{"PlaceType", fm.PlaceType, "village"},
		{"BuildingKind", fm.BuildingKind, "church"},
		{"Month", fm.Month, "January"},
	} {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	// --- []string: Tags ---
	if len(fm.Tags) != 2 {
		t.Fatalf("Tags: got %d items, want 2", len(fm.Tags))
	}
	if fm.Tags[0] != "surname:smith" {
		t.Errorf("Tags[0]: got %q, want %q", fm.Tags[0], "surname:smith")
	}
	if fm.Tags[1] != "location:cornwall" {
		t.Errorf("Tags[1]: got %q, want %q", fm.Tags[1], "location:cornwall")
	}

	// --- []string: Aliases ---
	if len(fm.Aliases) != 1 {
		t.Fatalf("Aliases: got %d items, want 1", len(fm.Aliases))
	}
	if fm.Aliases[0] != "/old/john-smith/" {
		t.Errorf("Aliases[0]: got %q, want %q", fm.Aliases[0], "/old/john-smith/")
	}

	// --- map[string]string: Sitemap ---
	if fm.Sitemap["disable"] != "1" {
		t.Errorf("Sitemap[disable]: got %q, want %q", fm.Sitemap["disable"], "1")
	}

	// --- []map[string]string: DiaryLinks ---
	if len(fm.DiaryLinks) != 1 {
		t.Fatalf("DiaryLinks: got %d items, want 1", len(fm.DiaryLinks))
	}
	if got := fm.DiaryLinks[0]["title"]; got != "Research entry" {
		t.Errorf("DiaryLinks[0][title]: got %q, want %q", got, "Research entry")
	}
	if got := fm.DiaryLinks[0]["link"]; got != "/diary/2025/01/15/" {
		t.Errorf("DiaryLinks[0][link]: got %q, want %q", got, "/diary/2025/01/15/")
	}

	// --- []map[string]string: Links ---
	if len(fm.Links) != 1 {
		t.Fatalf("Links: got %d items, want 1", len(fm.Links))
	}
	if got := fm.Links[0]["title"]; got != "WikiTree profile" {
		t.Errorf("Links[0][title]: got %q, want %q", got, "WikiTree profile")
	}
	if got := fm.Links[0]["link"]; got != "https://wikitree.com/wiki/Smith-1" {
		t.Errorf("Links[0][link]: got %q, want %q", got, "https://wikitree.com/wiki/Smith-1")
	}

	// --- []map[string]string: Descendants ---
	if len(fm.Descendants) != 1 {
		t.Fatalf("Descendants: got %d items, want 1", len(fm.Descendants))
	}
	if got := fm.Descendants[0]["name"]; got != "Jane Smith" {
		t.Errorf("Descendants[0][name]: got %q, want %q", got, "Jane Smith")
	}
	if got := fm.Descendants[0]["link"]; got != "/person/I5678/" {
		t.Errorf("Descendants[0][link]: got %q, want %q", got, "/person/I5678/")
	}
	if got := fm.Descendants[0]["detail"]; got != "b. 1875" {
		t.Errorf("Descendants[0][detail]: got %q, want %q", got, "b. 1875")
	}

	// Body is the blank line written by Document.WriteTo after the closing ---.
	if !strings.HasPrefix(body, "\n") {
		t.Errorf("body should start with blank line separator, got: %q", body)
	}
}

// TestParseDocumentPagination verifies that numeric-string pagination fields
// (zero-padded like "01") round-trip correctly.
func TestParseDocumentPagination(t *testing.T) {
	var doc md.Document
	doc.Title("People (page 2 of 5)")
	doc.Layout("listpeople")
	doc.SetFrontMatterField(md.MarkdownTagFirstPage, "01")
	doc.SetFrontMatterField(md.MarkdownTagLastPage, "05")
	doc.SetFrontMatterField(md.MarkdownTagPrevPage, "01")
	doc.SetFrontMatterField(md.MarkdownTagNextPage, "03")
	doc.SetSitemapDisable()

	fm, _, err := ParseDocument(doc.String())
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}

	for _, tt := range []struct {
		name string
		got  string
		want string
	}{
		{"First", fm.First, "01"},
		{"Last", fm.Last, "05"},
		{"Prev", fm.Prev, "01"},
		{"Next", fm.Next, "03"},
	} {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	if fm.Sitemap["disable"] != "1" {
		t.Errorf("Sitemap[disable]: got %q, want %q", fm.Sitemap["disable"], "1")
	}
}

// TestParseDocumentNoFrontMatter verifies that plain markdown without a
// front-matter block is returned unchanged as the body.
func TestParseDocumentNoFrontMatter(t *testing.T) {
	content := "# A plain markdown file\n\nWith some content.\n"

	fm, body, err := ParseDocument(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Title != "" {
		t.Errorf("Title: got %q, want empty", fm.Title)
	}
	if body != content {
		t.Errorf("body: got %q, want %q", body, content)
	}
}

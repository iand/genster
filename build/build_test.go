package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile is a helper to create a file and its parent directories in a test.
func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildOutputPath(t *testing.T) {
	b := &Builder{ContentDir: "/content", PubDir: "/pub"}

	for _, tt := range []struct {
		rel  string
		want string
	}{
		// index files stay in place
		{"person/I123/index.md", "/pub/person/I123/index.html"},
		{"_index.md", "/pub/index.html"},
		{"index.md", "/pub/index.html"},
		// non-index .md files get a clean-URL subdirectory
		{"list/people/01.md", "/pub/list/people/01/index.html"},
		{"list/surnames/02.md", "/pub/list/surnames/02/index.html"},
		// non-.md files are copied verbatim
		{"media/photo.jpg", "/pub/media/photo.jpg"},
		{"chart/ancestors.svg", "/pub/chart/ancestors.svg"},
		{"trees/export.ged", "/pub/trees/export.ged"},
	} {
		got := b.outputPath(tt.rel)
		if got != tt.want {
			t.Errorf("outputPath(%q): got %q, want %q", tt.rel, got, tt.want)
		}
	}
}

func TestBuildMarkdownRendered(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "person", "I1", "index.md"),
		"---\nid: I1\ntitle: John Smith\nlayout: person\n---\n\n<p>Hello from the body.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	outPath := filepath.Join(pubDir, "person", "I1", "index.html")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(data)

	if !strings.Contains(out, "John Smith") {
		t.Errorf("output missing title: %s", out)
	}
	if !strings.Contains(out, "Hello from the body.") {
		t.Errorf("output missing body content: %s", out)
	}
	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Errorf("output missing DOCTYPE: %s", out)
	}
}

func TestBuildNonIndexMdGetsCleanURL(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "list", "people", "02.md"),
		"---\ntitle: People page 2\nlayout: listpeople\n---\n\n<p>page content</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Should be at 02/index.html, not 02.html.
	outPath := filepath.Join(pubDir, "list", "people", "02", "index.html")
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("expected output at %s: %v", outPath, err)
	}
}

func TestBuildNonMdFileCopied(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	imgData := []byte("fake jpeg data")
	writeFile(t, filepath.Join(contentDir, "media", "photo.jpg"), string(imgData))

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	outPath := filepath.Join(pubDir, "media", "photo.jpg")
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(got) != string(imgData) {
		t.Errorf("copied file content mismatch: got %q, want %q", got, imgData)
	}
}

func TestBuildIgnoresDotFilesAndDirs(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// A normal page that should be built.
	writeFile(t, filepath.Join(contentDir, "stories", "normal.md"),
		"---\ntitle: Normal Story\n---\n\n<p>content</p>\n")
	// A dot-file at the top level — must be ignored.
	writeFile(t, filepath.Join(contentDir, ".DS_Store"), "mac metadata")
	// A dot-file inside a content directory — must be ignored.
	writeFile(t, filepath.Join(contentDir, "stories", ".draft.md"),
		"---\ntitle: Hidden Draft\n---\n\n<p>secret</p>\n")
	// A dot-directory and a file inside it — directory and all contents must be skipped.
	writeFile(t, filepath.Join(contentDir, ".obsidian", "config.md"),
		"---\ntitle: Obsidian Config\n---\n\n<p>tool config</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Normal page must be built.
	if _, err := os.Stat(filepath.Join(pubDir, "stories", "normal", "index.html")); err != nil {
		t.Errorf("normal page not built: %v", err)
	}
	// Dot-file must not produce output.
	if _, err := os.Stat(filepath.Join(pubDir, ".DS_Store")); err == nil {
		t.Errorf("dot-file .DS_Store should not be copied to pub")
	}
	// Dot-.md file must not produce a page.
	if _, err := os.Stat(filepath.Join(pubDir, "stories", ".draft", "index.html")); err == nil {
		t.Errorf("dot .md file should not produce a page")
	}
	// Dot-directory contents must not produce output.
	if _, err := os.Stat(filepath.Join(pubDir, ".obsidian")); err == nil {
		t.Errorf("dot-directory .obsidian should not appear in pub")
	}
}

func TestBuildPersonSidebar(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// Tree overview page: provides the title that appears in the tree nav.
	writeFile(t, filepath.Join(contentDir, "trees", "test", "index.md"),
		"---\ntitle: Test Tree\nlayout: treeoverview\nbasepath: /trees/test/\n---\n\n<p>overview</p>\n")

	mdContent := `---
id: I1
title: Jane Doe
layout: person
basepath: /trees/test/
tags:
  - doe
  - english
descendants:
  - name: Alice Doe
    link: /trees/test/person/I2/
  - name: Jane Doe
    detail: "b. 1850"
diarylinks:
  - title: Diary 1850
    link: /diary/1850/
links:
  - title: Ancestry
    link: https://ancestry.com/
---

<p>Bio text here.</p>
`
	writeFile(t, filepath.Join(contentDir, "person", "I1", "index.md"), mdContent)

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "person", "I1", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	for _, want := range []string{
		"Jane Doe",
		"Bio text here.",
		`href="/trees/test/"`,
		"Test Tree",
		`href="/trees/test/list/people/"`,
		`href="/trees/test/person/I2/"`, // descendant link
		"b. 1850",                       // descendant detail
		"Diary 1850",
		"/diary/1850/",
		"Ancestry",
		"https://ancestry.com/",
		`href="/tags/doe/"`,
		`href="/tags/english/"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestBuildPaginationLinks(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "list", "people", "02", "index.md"),
		"---\ntitle: People (page 2 of 4)\nlayout: listpeople\nbasepath: /trees/test/\nfirst: \"01\"\nprev: \"\"\nnext: \"03\"\nlast: \"04\"\n---\n\n<p>page 2 content</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "list", "people", "02", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	for _, want := range []string{
		`href="../01/"`, // first page link
		`href="../03/"`, // next page link
		`href="../04/"`, // last page link
		`«« first page`,
		`next page »`,
		`last page »»`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("output missing %q\nHTML:\n%s", want, html)
		}
	}

	// prev should not appear (no prev on page 2)
	if strings.Contains(html, "previous page") {
		t.Errorf("output should not contain previous page link on page 2 of 4")
	}
}

func TestBuildTreeNav(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "trees", "test", "index.md"),
		"---\ntitle: Test Tree\nlayout: treeoverview\nbasepath: /trees/test/\n---\n\n<p>Tree overview.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "trees", "test", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	for _, want := range []string{
		`href="/trees/test/"`,
		`href="/trees/test/list/people/"`,
		`href="/trees/test/list/surnames/"`,
		`href="/trees/test/list/places/"`,
		`href="/trees/test/list/todo/"`,
		`href="/trees/test/list/changes/"`,
		"Test Tree", // tree title shown in nav, not "Overview"
		"Tree overview.",
		"oak-tree.png",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestBuildPlainLayout(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "about", "index.md"),
		"---\ntitle: About\n---\n\n<p>Some plain content.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "about", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("output missing DOCTYPE")
	}
	if !strings.Contains(html, "Some plain content.") {
		t.Errorf("output missing body content")
	}
}

func TestBuildDebugFooter(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "about", "index.md"),
		"---\ntitle: About\nlayout: single\n---\n\n<p>content</p>\n")

	// Without --debug: no debug footer.
	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}
	out, _ := os.ReadFile(filepath.Join(pubDir, "about", "index.html"))
	if strings.Contains(string(out), "debug-footer") {
		t.Errorf("debug footer present without --debug flag")
	}

	// With --debug: debug footer appears and shows the layout.
	b.Debug = true
	if err := b.Build(); err != nil {
		t.Fatalf("Build (debug): %v", err)
	}
	out, _ = os.ReadFile(filepath.Join(pubDir, "about", "index.html"))
	html := string(out)
	if !strings.Contains(html, "debug-footer") {
		t.Errorf("debug footer missing with --debug flag")
	}
	if !strings.Contains(html, "single") {
		t.Errorf("debug footer missing layout name")
	}
}

func TestBuildAliasRedirect(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "person", "I1", "index.md"),
		"---\ntitle: John Smith\nlayout: person\naliases:\n  - /r/I1\n  - /r/smith-john\n---\n\n<p>Bio.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Both alias redirect files should exist.
	for _, alias := range []string{"r/I1", "r/smith-john"} {
		redirectPath := filepath.Join(pubDir, alias, "index.html")
		data, err := os.ReadFile(redirectPath)
		if err != nil {
			t.Fatalf("alias redirect missing at %s: %v", redirectPath, err)
		}
		html := string(data)
		canonical := "/person/I1/"
		if !strings.Contains(html, canonical) {
			t.Errorf("redirect for %s missing canonical URL %q:\n%s", alias, canonical, html)
		}
		if !strings.Contains(html, "http-equiv=\"refresh\"") {
			t.Errorf("redirect for %s missing meta refresh:\n%s", alias, html)
		}
		if !strings.Contains(html, "rel=\"canonical\"") {
			t.Errorf("redirect for %s missing canonical link:\n%s", alias, html)
		}
	}
}

func TestBuildAliasConflictIsSkipped(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// Two pages share the same alias — the second write is silently skipped
	// (first writer wins) and the build succeeds.
	for _, name := range []string{"I1", "I2"} {
		writeFile(t, filepath.Join(contentDir, "person", name, "index.md"),
			"---\ntitle: Person "+name+"\nlayout: person\naliases:\n  - /r/same-alias\n---\n\n<p>Bio.</p>\n")
	}

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("expected build to succeed despite alias conflict, got: %v", err)
	}

	// The redirect file should exist (written by whichever page was processed first).
	if _, err := os.Stat(filepath.Join(pubDir, "r", "same-alias", "index.html")); err != nil {
		t.Errorf("alias redirect not written: %v", err)
	}
}

func TestBuildEmbeddedAssetsWritten(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	for _, want := range []string{
		filepath.Join(pubDir, "css", "main.css"),
		filepath.Join(pubDir, "css", "dimbox.min.css"),
		filepath.Join(pubDir, "js", "dimbox.min.js"),
	} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("expected asset at %s: %v", want, err)
		}
	}
}

func TestBuildExternalAssetsOverride(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()
	assetsDir := t.TempDir()

	// Write a custom CSS file in the external assets dir.
	writeFile(t, filepath.Join(assetsDir, "css", "main.css"), "/* custom */")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir, AssetsDir: assetsDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(pubDir, "css", "main.css"))
	if err != nil {
		t.Fatalf("read main.css: %v", err)
	}
	if string(got) != "/* custom */" {
		t.Errorf("expected custom CSS %q, got %q", "/* custom */", got)
	}
}

func TestBuildStripsHTMLComments(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// Simulate annotate output: a private annotation wrapped in an HTML comment,
	// alongside a plain HTML comment and visible content.
	mdContent := "---\ntitle: Test\nlayout: single\n---\n\n" +
		"<p>Visible content.</p>\n" +
		"<!-- {{< private >}}cite [^foo]: Some Citation{{< /private >}} -->\n" +
		"<!-- plain comment -->\n"

	writeFile(t, filepath.Join(contentDir, "test", "index.md"), mdContent)

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "test", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, "Visible content.") {
		t.Errorf("output missing visible content")
	}
	// The entire comment blocks must be stripped, including their content.
	if strings.Contains(html, "<!--") {
		t.Errorf("output still contains HTML comment: %s", html)
	}
	if strings.Contains(html, "cite [^foo]") {
		t.Errorf("output should not contain annotation text from stripped comment")
	}
	if strings.Contains(html, "plain comment") {
		t.Errorf("output should not contain plain comment text")
	}
}

func TestBuildUnknownLayoutErrors(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "weird", "index.md"),
		"---\ntitle: Weird Page\nlayout: nonexistentlayout\n---\n\n<p>body</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	err := b.Build()
	if err == nil {
		t.Fatal("expected error for unknown layout, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistentlayout") {
		t.Errorf("error should mention the layout name: %v", err)
	}
	if !strings.Contains(err.Error(), "weird") {
		t.Errorf("error should mention the file path: %v", err)
	}
}

func TestBuildDraftStringFormNotPublished(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// draft: "true" (quoted string) — as written by some existing content files.
	writeFile(t, filepath.Join(contentDir, "draft", "index.md"),
		"---\ntitle: Work in Progress\nlayout: single\ndraft: \"true\"\n---\n\n<p>Not ready.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(pubDir, "draft", "index.html")); err == nil {
		t.Error("draft page with string 'true' should not be published")
	}
}

func TestBuildDraftPageNotPublished(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "draft", "index.md"),
		"---\ntitle: Work in Progress\nlayout: single\ndraft: true\n---\n\n<p>Not ready.</p>\n")
	writeFile(t, filepath.Join(contentDir, "published", "index.md"),
		"---\ntitle: Published Page\nlayout: single\n---\n\n<p>Ready.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(pubDir, "draft", "index.html")); err == nil {
		t.Error("draft page should not be published")
	}
	if _, err := os.Stat(filepath.Join(pubDir, "published", "index.html")); err != nil {
		t.Errorf("published page should exist: %v", err)
	}
}

func TestBuildDraftPageNotInSectionListing(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "stories", "_index.md"),
		"---\nlayout: storieshome\n---\n")
	writeFile(t, filepath.Join(contentDir, "stories", "ready", "index.md"),
		"---\ntitle: Ready Story\nlayout: single\n---\n\n<p>content</p>\n")
	writeFile(t, filepath.Join(contentDir, "stories", "wip", "index.md"),
		"---\ntitle: Draft Story\nlayout: single\ndraft: true\n---\n\n<p>not ready</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "stories", "index.html"))
	if err != nil {
		t.Fatalf("read stories/index.html: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, "Ready Story") {
		t.Error("section listing missing published page")
	}
	if strings.Contains(html, "Draft Story") {
		t.Error("section listing should not contain draft page")
	}
}

func TestBuildDraftDirectorySkippedEntirely(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// Draft story directory: has index.md (draft), an image, and a subdirectory.
	writeFile(t, filepath.Join(contentDir, "stories", "wip", "index.md"),
		"---\ntitle: Draft Story\nlayout: single\ndraft: true\n---\n\n<p>not ready</p>\n")
	writeFile(t, filepath.Join(contentDir, "stories", "wip", "photo.jpg"), "imgdata")
	writeFile(t, filepath.Join(contentDir, "stories", "wip", "sub", "index.md"),
		"---\ntitle: Sub Page\nlayout: single\n---\n\n<p>sub</p>\n")

	// Published story directory for contrast.
	writeFile(t, filepath.Join(contentDir, "stories", "ready", "index.md"),
		"---\ntitle: Ready Story\nlayout: single\n---\n\n<p>ready</p>\n")
	writeFile(t, filepath.Join(contentDir, "stories", "ready", "image.png"), "imgdata")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// The draft directory's HTML, image, and subdirectory must all be absent.
	if _, err := os.Stat(filepath.Join(pubDir, "stories", "wip", "index.html")); err == nil {
		t.Error("draft story index.html should not be published")
	}
	if _, err := os.Stat(filepath.Join(pubDir, "stories", "wip", "photo.jpg")); err == nil {
		t.Error("draft story image should not be copied")
	}
	if _, err := os.Stat(filepath.Join(pubDir, "stories", "wip", "sub", "index.html")); err == nil {
		t.Error("draft story subdirectory should not be published")
	}

	// The published story's files must be present.
	if _, err := os.Stat(filepath.Join(pubDir, "stories", "ready", "index.html")); err != nil {
		t.Errorf("published story index.html missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pubDir, "stories", "ready", "image.png")); err != nil {
		t.Errorf("published story image missing: %v", err)
	}
}

func TestStemToTitle(t *testing.T) {
	for _, tt := range []struct {
		stem string
		want string
	}{
		{"2024-02-22", "22 Feb 2024"},
		{"2021-12-01", "1 Dec 2021"},
		{"chambers", "chambers"},
		{"2024", "2024"},
		{"at", "at"},
	} {
		got := stemToTitle(tt.stem)
		if got != tt.want {
			t.Errorf("stemToTitle(%q) = %q, want %q", tt.stem, got, tt.want)
		}
	}
}

func TestCollectChildren(t *testing.T) {
	contentDir := t.TempDir()

	// trees section: two tree subdirectories
	writeFile(t, filepath.Join(contentDir, "trees", "_index.md"),
		"---\nlayout: listtrees\n---\n")
	writeFile(t, filepath.Join(contentDir, "trees", "at", "index.md"),
		"---\ntitle: Alcock Tree\nlayout: treeoverview\n---\n\n<p>content</p>\n")
	writeFile(t, filepath.Join(contentDir, "trees", "cg", "index.md"),
		"---\ntitle: Chambers Tree\nlayout: treeoverview\n---\n\n<p>content</p>\n")

	// diary section: year sub-section with dated leaf entries
	writeFile(t, filepath.Join(contentDir, "diary", "_index.md"),
		"---\nlayout: diaryhome\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "_index.md"),
		"---\ntitle: All diary entries made in 2024\nlayout: diaryentries\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-02-22.md"),
		"<p>entry</p>\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-01-10.md"),
		"<p>entry</p>\n")

	children, sectionTitles, _, _, err := collectChildren(contentDir, false)
	if err != nil {
		t.Fatalf("collectChildren: %v", err)
	}

	// sectionTitles should record the title of each index page
	if got := sectionTitles["trees/at"]; got != "Alcock Tree" {
		t.Errorf("sectionTitles[trees/at] = %q, want %q", got, "Alcock Tree")
	}
	if got := sectionTitles["trees/cg"]; got != "Chambers Tree" {
		t.Errorf("sectionTitles[trees/cg] = %q, want %q", got, "Chambers Tree")
	}

	// trees: children["trees"] contains the two tree index pages
	treeChildren := children["trees"]
	if len(treeChildren) != 2 {
		t.Fatalf("children[trees]: got %d entries, want 2: %v", len(treeChildren), treeChildren)
	}
	// sorted alphabetically by title (no dates)
	if treeChildren[0].Title != "Alcock Tree" || treeChildren[0].URL != "/trees/at/" {
		t.Errorf("children[trees][0]: got %+v, want {Alcock Tree /trees/at/}", treeChildren[0])
	}
	if treeChildren[1].Title != "Chambers Tree" || treeChildren[1].URL != "/trees/cg/" {
		t.Errorf("children[trees][1]: got %+v, want {Chambers Tree /trees/cg/}", treeChildren[1])
	}

	// diary: children["diary"] contains the year section
	diaryChildren := children["diary"]
	if len(diaryChildren) != 1 {
		t.Fatalf("children[diary]: got %d entries, want 1: %v", len(diaryChildren), diaryChildren)
	}
	if diaryChildren[0].Title != "All diary entries made in 2024" || diaryChildren[0].URL != "/diary/2024/" {
		t.Errorf("children[diary][0]: got %+v", diaryChildren[0])
	}

	// diary/2024: two leaf entries sorted by date descending
	yearChildren := children["diary/2024"]
	if len(yearChildren) != 2 {
		t.Fatalf("children[diary/2024]: got %d entries, want 2: %v", len(yearChildren), yearChildren)
	}
	if yearChildren[0].Title != "22 Feb 2024" || yearChildren[0].Date != "2024-02-22" {
		t.Errorf("children[diary/2024][0]: got %+v, want {22 Feb 2024 ... 2024-02-22}", yearChildren[0])
	}
	if yearChildren[1].Title != "10 Jan 2024" || yearChildren[1].Date != "2024-01-10" {
		t.Errorf("children[diary/2024][1]: got %+v, want {10 Jan 2024 ... 2024-01-10}", yearChildren[1])
	}
}

func TestBuildSectionIndexInjectsChildListing(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// trees/_index.md has no body; at/ and cg/ are tree sub-sections
	writeFile(t, filepath.Join(contentDir, "trees", "_index.md"),
		"---\nlayout: listtrees\n---\n")
	writeFile(t, filepath.Join(contentDir, "trees", "at", "index.md"),
		"---\ntitle: Alcock Tree\nlayout: treeoverview\nbasepath: /trees/at/\ntreetitle: Alcock Tree\n---\n\n<p>overview</p>\n")
	writeFile(t, filepath.Join(contentDir, "trees", "cg", "index.md"),
		"---\ntitle: Chambers Tree\nlayout: treeoverview\nbasepath: /trees/cg/\ntreetitle: Chambers Tree\n---\n\n<p>overview</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "trees", "index.html"))
	if err != nil {
		t.Fatalf("read trees/index.html: %v", err)
	}
	html := string(out)

	for _, want := range []string{
		`href="/trees/at/"`,
		"Alcock Tree",
		`href="/trees/cg/"`,
		"Chambers Tree",
		`<ul class="list">`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("trees/index.html missing %q", want)
		}
	}
}

func TestBuildSectionIndexWithBodyUnchanged(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// A section index with a hand-authored body should not have its body replaced.
	// Use a path not covered by layoutRules so the layout: home is respected.
	writeFile(t, filepath.Join(contentDir, "notes", "_index.md"),
		"---\nlayout: home\n---\n\n<p>Hand-authored intro text.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "notes", "index.html"))
	if err != nil {
		t.Fatalf("read diary/index.html: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, "Hand-authored intro text.") {
		t.Errorf("diary/index.html missing hand-authored content")
	}
}

func TestBuildDiaryHomeListing(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "diary", "_index.md"),
		"---\ntitle: Research Diary\nlayout: diaryhome\nsummary: Notes from my research.\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "_index.md"),
		"---\ntitle: 2024\nlayout: diaryentries\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-03-15.md"),
		"---\ntitle: March entry\n---\n<p>content</p>\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2025", "_index.md"),
		"---\ntitle: 2025\nlayout: diaryentries\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2025", "2025-01-10.md"),
		"---\ntitle: January entry\n---\n<p>content</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "diary", "index.html"))
	if err != nil {
		t.Fatalf("read diary/index.html: %v", err)
	}
	html := string(out)

	// Summary must appear.
	if !strings.Contains(html, "Notes from my research.") {
		t.Errorf("diary/index.html missing summary")
	}
	// Both diary entries must appear as links.
	if !strings.Contains(html, "March entry") {
		t.Errorf("diary/index.html missing March entry")
	}
	if !strings.Contains(html, "January entry") {
		t.Errorf("diary/index.html missing January entry")
	}
	// Most recent entry (2025) must appear before older entry (2024).
	janPos := strings.Index(html, "January entry")
	marPos := strings.Index(html, "March entry")
	if janPos == -1 || marPos == -1 {
		t.Fatalf("could not find both entries")
	}
	if janPos > marPos {
		t.Errorf("expected 2025 entry to appear before 2024 entry")
	}
}

func TestBuildDiaryHomeListingLimit(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "diary", "_index.md"),
		"---\ntitle: Research Diary\nlayout: diaryhome\n---\n")
	// 2023: one old entry — should be excluded when 2024 fills the 20-slot limit.
	writeFile(t, filepath.Join(contentDir, "diary", "2023", "_index.md"),
		"---\ntitle: 2023\nlayout: diaryentries\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2023", "2023-12-01.md"),
		"<p>old entry</p>\n")
	// 2024: exactly 20 entries (days 1–20 of January).
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "_index.md"),
		"---\ntitle: 2024\nlayout: diaryentries\n---\n")
	for d := 1; d <= 20; d++ {
		name := fmt.Sprintf("2024-01-%02d.md", d)
		writeFile(t, filepath.Join(contentDir, "diary", "2024", name), "<p>entry</p>\n")
	}

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "diary", "index.html"))
	if err != nil {
		t.Fatalf("read diary/index.html: %v", err)
	}
	html := string(out)

	// The 2023 entry must be excluded (limit of 20 filled by 2024 entries).
	if strings.Contains(html, "1 Dec 2023") {
		t.Errorf("diary/index.html should not contain the excluded 2023 entry")
	}
	// The most recent 2024 entry must appear.
	if !strings.Contains(html, "20 Jan 2024") {
		t.Errorf("diary/index.html missing most recent entry (20 Jan 2024)")
	}
}

func TestBuildDiaryYearSectionDateSorted(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "diary", "2024", "_index.md"),
		"---\ntitle: 2024 Diary\nlayout: diaryentries\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-01-10.md"),
		"<p>January entry.</p>\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-03-15.md"),
		"<p>March entry.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "diary", "2024", "index.html"))
	if err != nil {
		t.Fatalf("read diary/2024/index.html: %v", err)
	}
	html := string(out)

	marchPos := strings.Index(html, "15 Mar 2024")
	janPos := strings.Index(html, "10 Jan 2024")
	if marchPos == -1 || janPos == -1 {
		t.Fatalf("diary/2024/index.html missing expected entries: %s", html)
	}
	if marchPos > janPos {
		t.Errorf("expected March (2024-03-15) to appear before January (2024-01-10) in date-descending order")
	}
}

func TestBuildDiaryYearListingMeta(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "diary", "2024", "_index.md"),
		"---\ntitle: 2024 Diary\nlayout: diaryentries\n---\n")
	// Entry with tags.
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-03-15.md"),
		"---\ntags: [alcock, dunmore]\n---\n\nThis is a diary entry.\n")
	// Entry with no tags, minimal content.
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-01-10.md"),
		"<p>January entry.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "diary", "2024", "index.html"))
	if err != nil {
		t.Fatalf("read diary/2024/index.html: %v", err)
	}
	html := string(out)

	// March entry should show its tags as links.
	if !strings.Contains(html, `href="/tags/alcock/"`) {
		t.Errorf("diary/2024 listing: missing alcock tag link")
	}
	if !strings.Contains(html, `href="/tags/dunmore/"`) {
		t.Errorf("diary/2024 listing: missing dunmore tag link")
	}

	// Entries must remain in date-descending order (March before January).
	marchPos := strings.Index(html, "15 Mar 2024")
	janPos := strings.Index(html, "10 Jan 2024")
	if marchPos == -1 || janPos == -1 {
		t.Fatalf("diary/2024/index.html missing expected entries")
	}
	if marchPos > janPos {
		t.Errorf("expected March to appear before January in date-descending order")
	}
}

func TestBuildStoriesListing(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "stories", "_index.md"),
		"---\nlayout: storieshome\n---\n")
	// Story with summary, tags, and content.
	writeFile(t, filepath.Join(contentDir, "stories", "suffolk.md"),
		"---\ntitle: Suffolk Hinksmans\nsummary: The story of the Hinksman family in Suffolk.\ntags: [hinksman, suffolk]\n---\n\nLots of words about the Hinksman family in Suffolk.\n")
	// Story with no summary, no tags.
	writeFile(t, filepath.Join(contentDir, "stories", "norfolk.md"),
		"---\ntitle: Norfolk Story\n---\n\n<p>A short story.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "stories", "index.html"))
	if err != nil {
		t.Fatalf("read stories/index.html: %v", err)
	}
	html := string(out)

	// Both story titles must appear as links.
	if !strings.Contains(html, "Suffolk Hinksmans") {
		t.Errorf("stories listing: missing Suffolk Hinksmans")
	}
	if !strings.Contains(html, "Norfolk Story") {
		t.Errorf("stories listing: missing Norfolk Story")
	}
	// Summary with word count in parentheses must appear for the Suffolk story.
	if !strings.Contains(html, "The story of the Hinksman family in Suffolk.") {
		t.Errorf("stories listing: missing summary for Suffolk Hinksmans")
	}
	if !strings.Contains(html, "words)") {
		t.Errorf("stories listing: missing word count in parentheses")
	}
	// Tags must appear under "tagged as:".
	if !strings.Contains(html, "tagged as:") {
		t.Errorf("stories listing: missing 'tagged as:' prefix")
	}
	if !strings.Contains(html, `href="/tags/hinksman/"`) {
		t.Errorf("stories listing: missing hinksman tag link")
	}
	if !strings.Contains(html, `href="/tags/suffolk/"`) {
		t.Errorf("stories listing: missing suffolk tag link")
	}
	// Story with no summary must not show a word count.
	norfolkPos := strings.Index(html, "Norfolk Story")
	suffolkPos := strings.Index(html, "Suffolk Hinksmans")
	if norfolkPos == -1 || suffolkPos == -1 {
		t.Fatalf("could not find both story titles")
	}
	// The word count appears in the Suffolk block; Norfolk has no summary so no count.
	// A rough check: "words)" should not appear after the Norfolk title.
	norfolkBlock := html[norfolkPos:]
	nextLI := strings.Index(norfolkBlock[1:], "<li>")
	if nextLI != -1 {
		norfolkBlock = norfolkBlock[:nextLI+1]
	}
	if strings.Contains(norfolkBlock, "words)") {
		t.Errorf("stories listing: word count should not appear for story without summary")
	}
}

func TestBuildDiarySummary(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "diary", "2024", "_index.md"),
		"---\ntitle: 2024 Diary\nlayout: diaryentries\n---\n")
	// Entry with a summary.
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-03-15.md"),
		"---\nsummary: A brief note about March.\n---\n\nSome diary content.\n")
	// Entry without a summary but with tags.
	writeFile(t, filepath.Join(contentDir, "diary", "2024", "2024-01-10.md"),
		"---\ntags: [alcock]\n---\n\nJanuary content.\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "diary", "2024", "index.html"))
	if err != nil {
		t.Fatalf("read diary/2024/index.html: %v", err)
	}
	html := string(out)

	// Summary must appear for the March entry.
	if !strings.Contains(html, "A brief note about March.") {
		t.Errorf("diary listing: missing summary")
	}
	// Tags must appear with "tagged as:" prefix for the January entry.
	if !strings.Contains(html, "tagged as:") {
		t.Errorf("diary listing: missing 'tagged as:' prefix for tagged entry")
	}
	if !strings.Contains(html, `href="/tags/alcock/"`) {
		t.Errorf("diary listing: missing alcock tag link")
	}
}

func TestGroupFromURL(t *testing.T) {
	for _, tt := range []struct {
		url  string
		want string
	}{
		{"/diary/2021/2021-05-17/", "Diary entries"},
		{"/diary/2024/2024-01-01/", "Diary entries"},
		{"/stories/suffolk-hinksmans/", "Stories"},
		{"/trees/at/person/I1234/", "People"},
		{"/trees/cg/person/I9999/", "People"},
		{"/trees/at/place/P42/", "Places"},
		{"/trees/cg/place/P1/", "Places"},
		{"/trees/at/family/F1/", "Other"},
		{"/search/", "Other"},
	} {
		got := groupFromURL(tt.url)
		if got != tt.want {
			t.Errorf("groupFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestBuildTagPages(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// Pages from three different content types all tagged "England".
	writeFile(t, filepath.Join(contentDir, "stories", "suffolk.md"),
		"---\ntitle: Suffolk Story\ntags: [England]\n---\n\n<p>Suffolk content.</p>\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2021", "2021-05-17.md"),
		"---\ntags: [England]\n---\n\n<p>Diary entry.</p>\n")
	writeFile(t, filepath.Join(contentDir, "trees", "at", "person", "I1.md"),
		"---\ntitle: Alice Smith\nlayout: person\ntags: [England]\n---\n\n<p>Person.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "tags", "england", "index.html"))
	if err != nil {
		t.Fatalf("read tags/england/index.html: %v", err)
	}
	html := string(out)

	// All three pages must appear.
	if !strings.Contains(html, "Suffolk Story") {
		t.Errorf("tags/england: missing Suffolk Story")
	}
	if !strings.Contains(html, "Alice Smith") {
		t.Errorf("tags/england: missing Alice Smith")
	}
	if !strings.Contains(html, "17 May 2021") {
		t.Errorf("tags/england: missing diary entry")
	}

	// Groups must appear as headings.
	if !strings.Contains(html, "<h2>People</h2>") {
		t.Errorf("tags/england: missing People heading")
	}
	if !strings.Contains(html, "<h2>Stories</h2>") {
		t.Errorf("tags/england: missing Stories heading")
	}
	if !strings.Contains(html, "<h2>Diary entries</h2>") {
		t.Errorf("tags/england: missing Diary entries heading")
	}

	// People must appear before Stories, Stories before Diary entries.
	peoplePos := strings.Index(html, "<h2>People</h2>")
	storiesPos := strings.Index(html, "<h2>Stories</h2>")
	diaryPos := strings.Index(html, "<h2>Diary entries</h2>")
	if peoplePos > storiesPos {
		t.Errorf("tags/england: People heading should appear before Stories")
	}
	if storiesPos > diaryPos {
		t.Errorf("tags/england: Stories heading should appear before Diary entries")
	}
}

func TestBuildTagAncestorOrdering(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// Two people with the same tag: one a direct ancestor, one not.
	writeFile(t, filepath.Join(contentDir, "trees", "at", "person", "I1.md"),
		"---\ntitle: Alice Smith\nlayout: person\ntags: [Smith]\nancestor: true\n---\n\n<p>Person.</p>\n")
	writeFile(t, filepath.Join(contentDir, "trees", "at", "person", "I2.md"),
		"---\ntitle: Bob Smith\nlayout: person\ntags: [Smith]\n---\n\n<p>Person.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "tags", "smith", "index.html"))
	if err != nil {
		t.Fatalf("read tags/smith/index.html: %v", err)
	}
	html := string(out)

	// Alice (ancestor) must appear before Bob (non-ancestor).
	alicePos := strings.Index(html, "Alice Smith")
	bobPos := strings.Index(html, "Bob Smith")
	if alicePos < 0 {
		t.Fatal("Alice Smith not found in tag page")
	}
	if bobPos < 0 {
		t.Fatal("Bob Smith not found in tag page")
	}
	if alicePos > bobPos {
		t.Errorf("ancestor Alice Smith should appear before non-ancestor Bob Smith")
	}
}

func TestBuildHomeLayoutShowsDefaultOak(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "index.md"),
		"---\ntitle: Family History\nlayout: home\nlastmod: 2024-07-03\n---\n\nWelcome.\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, "default-oak.webp") {
		t.Errorf("homepage missing default-oak.webp feature image; got:\n%s", html)
	}
	if !strings.Contains(html, `class="feature"`) {
		t.Errorf("homepage missing feature image div; got:\n%s", html)
	}
}

func TestBuildTagIndex(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "stories", "suffolk.md"),
		"---\ntitle: Suffolk Story\ntags: [Suffolk, England]\n---\n\n<p>content</p>\n")
	writeFile(t, filepath.Join(contentDir, "stories", "norfolk.md"),
		"---\ntitle: Norfolk Story\ntags: [Norfolk, England]\n---\n\n<p>content</p>\n")
	writeFile(t, filepath.Join(contentDir, "stories", "london.md"),
		"---\ntitle: London Story\ntags: [England]\n---\n\n<p>content</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "tags", "index.html"))
	if err != nil {
		t.Fatalf("read tags/index.html: %v", err)
	}
	html := string(out)

	// All three tags must appear, England with count 3, others with count 1.
	if !strings.Contains(html, "England (3)") {
		t.Errorf("tags/index.html: England should have count 3; got:\n%s", html)
	}
	if !strings.Contains(html, "Norfolk (1)") {
		t.Errorf("tags/index.html: Norfolk should have count 1; got:\n%s", html)
	}
	if !strings.Contains(html, "Suffolk (1)") {
		t.Errorf("tags/index.html: Suffolk should have count 1; got:\n%s", html)
	}

	// England should appear before Norfolk and Suffolk (alphabetical).
	englandPos := strings.Index(html, ">England<")
	norfolkPos := strings.Index(html, ">Norfolk<")
	suffolkPos := strings.Index(html, ">Suffolk<")
	if englandPos > norfolkPos || norfolkPos > suffolkPos {
		t.Errorf("tags/index.html: expected alphabetical order England < Norfolk < Suffolk")
	}
}

// TestBuildDiaryDateTitle covers the two real-world inconsistency cases:
//  1. A leaf diary file with no front-matter title should derive its page title
//     from the YYYY-MM-DD filename stem ("17 May 2021").
//  2. A directory-based diary entry with title: "2021-02-25" in front-matter
//     should have that raw date formatted ("25 Feb 2021") on the page and in
//     the tag index.
func TestBuildDiaryDateTitle(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// Case 1: leaf file, no title.
	writeFile(t, filepath.Join(contentDir, "diary", "2021", "2021-05-17.md"),
		"---\ntags: [alcock]\n---\n\n<p>May entry.</p>\n")

	// Case 2: directory entry with raw date as title.
	writeFile(t, filepath.Join(contentDir, "diary", "2021", "2021-02-25", "index.md"),
		"---\ntitle: \"2021-02-25\"\ntags: [alcock]\n---\n\n<p>Feb entry.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Case 1: page should show formatted date as <h1>.
	out, err := os.ReadFile(filepath.Join(pubDir, "diary", "2021", "2021-05-17", "index.html"))
	if err != nil {
		t.Fatalf("read 2021-05-17/index.html: %v", err)
	}
	if !strings.Contains(string(out), ">17 May 2021</h1>") {
		t.Errorf("2021-05-17 page: expected >17 May 2021</h1>; got:\n%s", out)
	}

	// Case 2: page should show formatted date, not raw string.
	out, err = os.ReadFile(filepath.Join(pubDir, "diary", "2021", "2021-02-25", "index.html"))
	if err != nil {
		t.Fatalf("read 2021-02-25/index.html: %v", err)
	}
	if strings.Contains(string(out), "<h1>2021-02-25</h1>") {
		t.Errorf("2021-02-25 page: raw date should not appear as title")
	}
	if !strings.Contains(string(out), ">25 Feb 2021</h1>") {
		t.Errorf("2021-02-25 page: expected >25 Feb 2021</h1>; got:\n%s", out)
	}

	// Both should appear with formatted titles in the tag index for "alcock".
	tagOut, err := os.ReadFile(filepath.Join(pubDir, "tags", "alcock", "index.html"))
	if err != nil {
		t.Fatalf("read tags/alcock/index.html: %v", err)
	}
	tagHTML := string(tagOut)
	if !strings.Contains(tagHTML, "17 May 2021") {
		t.Errorf("tags/alcock: expected '17 May 2021'; got:\n%s", tagHTML)
	}
	if !strings.Contains(tagHTML, "25 Feb 2021") {
		t.Errorf("tags/alcock: expected '25 Feb 2021'; got:\n%s", tagHTML)
	}
	// The raw date must not appear as link text (it will appear in the URL, which is fine).
	if strings.Contains(tagHTML, ">2021-02-25<") {
		t.Errorf("tags/alcock: raw date '2021-02-25' should not appear as link text")
	}
}

func TestBuildSitemapGenerated(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	// A normal whitelisted page with a lastmod date.
	writeFile(t, filepath.Join(contentDir, "stories", "suffolk.md"),
		"---\ntitle: Suffolk Story\nlastmod: \"2024-03-01\"\n---\n\n<p>content</p>\n")
	// A whitelisted page with sitemap.disable set — opt-out must be honoured.
	writeFile(t, filepath.Join(contentDir, "stories", "hidden.md"),
		"---\ntitle: Hidden\nsitemap:\n  disable: \"1\"\n---\n\n<p>content</p>\n")
	// A non-whitelisted placeholder file (r/ is not in the whitelist).
	writeFile(t, filepath.Join(contentDir, "r", "I0074.md"), "")
	// A tagged page; tag URLs are not whitelisted.
	writeFile(t, filepath.Join(contentDir, "stories", "norfolk.md"),
		"---\ntitle: Norfolk Story\ntags: [England]\n---\n\n<p>content</p>\n")
	// A tree homepage — must appear.
	writeFile(t, filepath.Join(contentDir, "trees", "at", "index.md"),
		"---\ntitle: Alcock Tree\nlayout: treeoverview\nbasepath: /trees/at/\n---\n\n<p>overview</p>\n")
	// A grandchild page under /trees/ — must be excluded.
	writeFile(t, filepath.Join(contentDir, "trees", "at", "person", "I1", "index.md"),
		"---\ntitle: Alice Smith\nlayout: person\nbasepath: /trees/at/\n---\n\n<p>bio</p>\n")

	b := &Builder{
		ContentDir: contentDir,
		PubDir:     pubDir,
		BaseURL:    "https://example.com",
	}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(pubDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("sitemap.xml not created: %v", err)
	}
	xml := string(data)

	// Normal whitelisted page must appear with its lastmod.
	if !strings.Contains(xml, "https://example.com/stories/suffolk/") {
		t.Errorf("sitemap: missing suffolk URL")
	}
	if !strings.Contains(xml, "<lastmod>2024-03-01</lastmod>") {
		t.Errorf("sitemap: missing lastmod for suffolk")
	}

	// Tree homepage must appear.
	if !strings.Contains(xml, "https://example.com/trees/at/") {
		t.Errorf("sitemap: missing tree homepage /trees/at/")
	}

	// Disabled whitelisted page must not appear (sitemap.disable opt-out).
	if strings.Contains(xml, "/stories/hidden/") {
		t.Errorf("sitemap: hidden page with sitemap.disable should be excluded")
	}

	// Non-whitelisted placeholder file must not appear.
	if strings.Contains(xml, "/r/I0074/") {
		t.Errorf("sitemap: r/ placeholder page should be excluded")
	}

	// Tag pages are not whitelisted — must not appear.
	if strings.Contains(xml, "/tags/england/") {
		t.Errorf("sitemap: tag URL /tags/england/ should not be included")
	}
	if strings.Contains(xml, "/tags/") {
		t.Errorf("sitemap: tags index URL should not be included")
	}

	// Tree grandchild pages must not appear.
	if strings.Contains(xml, "/trees/at/person/") {
		t.Errorf("sitemap: grandchild /trees/at/person/ should be excluded")
	}
}

func TestMatchURLPattern(t *testing.T) {
	for _, tt := range []struct {
		pattern string
		url     string
		want    bool
	}{
		// Exact match.
		{"/", "/", true},
		{"/", "/other/", false},
		{"/trees/", "/trees/", true},
		{"/trees/", "/trees/at/", false},

		// Prefix match (**).
		{"/diary/**", "/diary/", true}, // prefix itself included
		{"/diary/**", "/diary/2021/", true},
		{"/diary/**", "/diary/2021/2021-05-17/", true},
		{"/diary/**", "/diaryextra/", false}, // must not match partial segments
		{"/stories/**", "/stories/foo/bar/", true},

		// Single-segment wildcard (*).
		{"/trees/*/", "/trees/at/", true},
		{"/trees/*/", "/trees/cg/", true},
		{"/trees/*/", "/trees/", false},           // * must match a non-empty segment
		{"/trees/*/", "/trees/at/person/", false}, // too many segments
		{"/trees/*/*", "/trees/at/", false},
		{"/trees/*/*", "/trees/at/foo", true},
	} {
		got := matchURLPattern(tt.pattern, tt.url)
		if got != tt.want {
			t.Errorf("matchSitemapRule(%q, %q) = %v, want %v", tt.pattern, tt.url, got, tt.want)
		}
	}
}

func TestSitemapIncluded(t *testing.T) {
	for _, tt := range []struct {
		url  string
		want bool
	}{
		// Homepage.
		{"/", true},
		// Diary section index and all entries under it.
		{"/diary/", true},
		{"/diary/2021/", true},
		{"/diary/2021/2021-05-17/", true},
		// Stories section index and all entries.
		{"/stories/", true},
		{"/stories/suffolk-hinksmans/", true},
		// Trees section index.
		{"/trees/", true},
		// Tree homepages (one level deep).
		{"/trees/at/", true},
		{"/trees/cg/", true},
		// Tree grandchildren — excluded.
		{"/trees/at/person/", false},
		{"/trees/at/person/I1234/", false},
		{"/trees/at/place/P42/", false},
		{"/trees/at/list/people/", false},
		// Tags — excluded.
		{"/tags/", false},
		{"/tags/england/", false},
		// Short-links — excluded.
		{"/r/I0074/", false},
		// Search — excluded.
		{"/search/", false},
	} {
		got := sitemapIncluded(tt.url)
		if got != tt.want {
			t.Errorf("sitemapIncluded(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestBuildSitemapOmittedWithoutBaseURL(t *testing.T) {
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "stories", "suffolk.md"),
		"---\ntitle: Suffolk Story\n---\n\n<p>content</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir} // no BaseURL
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(pubDir, "sitemap.xml")); err == nil {
		t.Errorf("sitemap.xml should not be created when BaseURL is empty")
	}
}

func TestBuildFeatureImagePersonGeneric(t *testing.T) {
	// A person page with no image: front matter but a matching silhouette in
	// images/ should render the silhouette with the representative-image caption.
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	touch(t, filepath.Join(contentDir, "images", "person-male-1800s.webp"))
	writeFile(t, filepath.Join(contentDir, "person", "I1", "index.md"),
		"---\nlayout: person\ncategory: person\ngender: male\nera: 1800s\n---\n\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "person", "I1", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, "/images/person-male-1800s.webp") {
		t.Errorf("expected person-male-1800s.webp in output; got:\n%s", html)
	}
	if !strings.Contains(html, "representative image") {
		t.Errorf("expected representative-image caption in output; got:\n%s", html)
	}
}

func TestBuildFeatureImagePersonRealPhoto(t *testing.T) {
	// A person page with image: set (real photo from gen) should render that
	// photo with no representative-image caption.
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "person", "I1", "index.md"),
		"---\nlayout: person\ncategory: person\ngender: male\nera: 1800s\nimage: /trees/abc/person/I1/photo.jpg\n---\n\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "person", "I1", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	if !strings.Contains(html, "/trees/abc/person/I1/photo.jpg") {
		t.Errorf("expected real photo URL in output; got:\n%s", html)
	}
	if strings.Contains(html, "representative image") {
		t.Errorf("expected no representative-image caption for real photo; got:\n%s", html)
	}
}

func TestBuildFeatureImageNoMatch(t *testing.T) {
	// A person page with no image and no matching silhouette files renders no
	// <img> tag at all (no broken image).
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "person", "I1", "index.md"),
		"---\nlayout: person\ncategory: person\ngender: male\nera: 1800s\n---\n\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(pubDir, "person", "I1", "index.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	html := string(out)

	if strings.Contains(html, `class="feature"`) {
		t.Errorf("expected no feature image when no files match; got:\n%s", html)
	}
}

func TestBuildDiaryNavigation(t *testing.T) {
	// Three entries across two years.  Chronological order: A < B < C.
	// A should have only a next link (to B).
	// B should have both prev (A) and next (C).
	// C should have only a prev link (to B).
	// Year index pages should have no nav at all.
	contentDir := t.TempDir()
	pubDir := t.TempDir()

	writeFile(t, filepath.Join(contentDir, "diary", "2020", "2020-06-01.md"),
		"<p>Entry A.</p>\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2021", "_index.md"),
		"---\ntitle: 2021\n---\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2021", "2021-02-10", "index.md"),
		"<p>Entry B.</p>\n")
	writeFile(t, filepath.Join(contentDir, "diary", "2021", "2021-05-17.md"),
		"<p>Entry C.</p>\n")

	b := &Builder{ContentDir: contentDir, PubDir: pubDir}
	if err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	readHTML := func(path string) string {
		t.Helper()
		out, err := os.ReadFile(filepath.Join(pubDir, path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(out)
	}

	// Entry A: next only.
	a := readHTML(filepath.Join("diary", "2020", "2020-06-01", "index.html"))
	if strings.Contains(a, "previous entry") {
		t.Errorf("entry A: should have no previous link; got:\n%s", a)
	}
	if !strings.Contains(a, `href="/diary/2021/2021-02-10/"`) {
		t.Errorf("entry A: expected next link to /diary/2021/2021-02-10/; got:\n%s", a)
	}

	// Entry B: both prev and next.
	b2 := readHTML(filepath.Join("diary", "2021", "2021-02-10", "index.html"))
	if !strings.Contains(b2, `href="/diary/2020/2020-06-01/"`) {
		t.Errorf("entry B: expected prev link to /diary/2020/2020-06-01/; got:\n%s", b2)
	}
	if !strings.Contains(b2, `href="/diary/2021/2021-05-17/"`) {
		t.Errorf("entry B: expected next link to /diary/2021/2021-05-17/; got:\n%s", b2)
	}

	// Entry C: prev only.
	c := readHTML(filepath.Join("diary", "2021", "2021-05-17", "index.html"))
	if !strings.Contains(c, `href="/diary/2021/2021-02-10/"`) {
		t.Errorf("entry C: expected prev link to /diary/2021/2021-02-10/; got:\n%s", c)
	}
	if strings.Contains(c, "next entry") {
		t.Errorf("entry C: should have no next link; got:\n%s", c)
	}

	// Year index pages should not have nav links.
	y := readHTML(filepath.Join("diary", "2021", "index.html"))
	if strings.Contains(y, `class="pages"`) {
		t.Errorf("diary/2021 year index: should have no nav; got:\n%s", y)
	}
}

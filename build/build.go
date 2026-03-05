package build

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

// renderer converts markdown body content to HTML, configured to match Hugo's
// goldmark defaults. WithUnsafe is required because genster content files
// embed raw HTML blocks directly (Hugo sets unsafe:false but uses render hooks
// instead; we don't have that layer).
var renderer = goldmark.New(
	goldmark.WithRendererOptions(
		goldhtml.WithUnsafe(),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
		parser.WithAttribute(),
	),
	goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		extension.DefinitionList,
		extension.Footnote,
		extension.Typographer,
	),
)

// htmlCommentRE matches HTML comments, including multi-line ones.
// Used to strip comments from rendered output so that annotation markers
// (<!-- {{< private >}} ... -->) and any other comments do not reach the browser.
var htmlCommentRE = regexp.MustCompile(`(?s)<!--.*?-->`)

// TreeData holds site-level metadata for the genealogy tree a page belongs to.
// It is populated from the tree's section index page and is available in all
// templates as {{.Tree.Title}}, {{.Tree.BasePath}}, etc., without needing to
// repeat the information in every page's front-matter.
type TreeData struct {
	Title    string
	BasePath string
}

// NavEntry is a link to an adjacent page used for previous/next navigation.
// A zero-value NavEntry (empty URL) means no adjacent page exists.
type NavEntry struct {
	URL   string
	Title string
}

// PageData is passed to each page template during rendering. Embedding
// FrontMatter lets templates access fields like {{.Title}}, {{.Layout}}, etc.
// directly alongside {{.Body}} and {{.Tree}}.
type PageData struct {
	FrontMatter
	Body       template.HTML
	Tree       TreeData
	Section    string      // human-readable section name inferred from path (e.g. "Stories", "Research Diary")
	PrevEntry  NavEntry    // previous page in sequence (e.g. previous diary entry); zero if none
	NextEntry  NavEntry    // next page in sequence (e.g. next diary entry); zero if none
	Children   []childPage // non-nil when page uses diaryentries or storieshome layout
	DiaryYears []string    // descending list of diary years for sidebar nav (diaryhome and diaryentries only)
	Debug      bool        // true when the build was invoked with --debug
	PageLayout string      // resolved layout name, available to debug footer
}

// Builder walks a content directory and renders each file into a pub directory.
type Builder struct {
	ContentDir string
	PubDir     string
	// AssetsDir, if non-empty, is a directory of static assets (css/, js/)
	// to copy into PubDir. When empty the assets embedded in the binary are used.
	AssetsDir string
	// Debug, when true, adds a debug footer to every rendered page.
	Debug bool
	// BaseURL, if non-empty, is the scheme+host used to build absolute <loc>
	// URLs in sitemap.xml (e.g. "https://example.com"). When empty, no
	// sitemap.xml is written.
	BaseURL string
	// IncludeDrafts, when true, publishes pages with draft: true in their
	// front-matter instead of skipping them.
	IncludeDrafts bool

	// sitemapEntries accumulates pages for sitemap.xml during the build.
	sitemapEntries []sitemapEntry

	// diaryNav maps each diary entry URL to its [prev, next] NavEntry pair.
	// Built during the first pass (collectChildren) and consulted in renderMarkdown.
	diaryNav map[string][2]NavEntry

	// templates is the per-build template set, created at the start of Build()
	// with image-selection functions that close over ContentDir.
	templates *template.Template
}

// Build walks ContentDir and processes every file into PubDir. Markdown files
// are parsed, rendered through goldmark, and written through a layout template.
// All other files are copied verbatim. Static assets (CSS, JS) are written
// from the embedded binary assets or from AssetsDir if set.
//
// Build uses a two-pass strategy: the first pass collects child pages for
// every section so that section index files with empty bodies can have a
// generated child listing injected before rendering.
func (b *Builder) Build() error {
	if err := writeAssets(b.PubDir, b.AssetsDir); err != nil {
		return fmt.Errorf("write assets: %w", err)
	}

	b.templates = buildSiteTemplates(filepath.Join(b.ContentDir, "images"))

	children, sectionTitles, tagIndex, draftDirs, err := collectChildren(b.ContentDir, b.IncludeDrafts)
	if err != nil {
		return fmt.Errorf("collect children: %w", err)
	}
	b.diaryNav = buildDiaryNav(children)

	if err := filepath.WalkDir(b.ContentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if len(draftDirs) > 0 {
				rel, err := filepath.Rel(b.ContentDir, path)
				if err != nil {
					return err
				}
				if draftDirs[filepath.ToSlash(rel)] {
					return filepath.SkipDir
				}
			}
			return nil
		}

		rel, err := filepath.Rel(b.ContentDir, path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(d.Name(), ".md") {
			return b.renderMarkdown(path, rel, children, sectionTitles)
		}
		return b.copyFile(path, rel)
	}); err != nil {
		return err
	}

	if err := b.writeTags(tagIndex); err != nil {
		return fmt.Errorf("write tags: %w", err)
	}

	if b.BaseURL != "" {
		if err := writeSitemap(b.PubDir, b.BaseURL, b.sitemapEntries); err != nil {
			return fmt.Errorf("write sitemap: %w", err)
		}
	}

	return nil
}

func (b *Builder) renderMarkdown(srcPath, rel string, children map[string][]childPage, sectionTitles map[string]string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}

	fm, body, err := ParseDocument(string(data))
	if err != nil {
		return fmt.Errorf("parse %s: %w", srcPath, err)
	}

	if fm.LastMod == "" {
		if info, err := os.Stat(srcPath); err == nil {
			fm.LastMod = info.ModTime().UTC().Format("2006-01-02T15:04:05Z")
		}
	}

	if bool(fm.Draft) && !b.IncludeDrafts {
		return nil
	}

	// Normalise basepath to always have a trailing slash so template links
	// like {{.BasePath}}list/people/ produce correct URLs regardless of how
	// the content was generated.
	if fm.BasePath != "" && !strings.HasSuffix(fm.BasePath, "/") {
		fm.BasePath += "/"
	}

	// For section index files with no hand-authored body, inject a generated
	// child-page listing so the rendered page is not empty.
	stem := strings.TrimSuffix(filepath.Base(rel), ".md")

	// Normalise date-like titles: "2021-02-25" → "25 Feb 2021".
	fm.Title = stemToTitle(fm.Title)
	// For diary pages with no explicit title, derive it from the filename or
	// directory stem.  This only applies to the diary section because other
	// date-named stems (if they ever exist) should not have titles invented for them.
	if fm.Title == "" && strings.HasPrefix(filepath.ToSlash(rel), "diary/") {
		nameStem := stem
		if nameStem == "_index" || nameStem == "index" {
			nameStem = filepath.Base(filepath.Dir(rel))
		}
		if formatted := stemToTitle(nameStem); formatted != nameStem {
			fm.Title = formatted
		}
	}

	relSlash := filepath.ToSlash(rel)
	relDir := filepath.ToSlash(filepath.Dir(rel))

	// Compute the canonical URL early so it can drive layout resolution.
	var pageURL string
	if stem == "_index" || stem == "index" {
		if relDir == "." {
			pageURL = "/"
		} else {
			pageURL = "/" + relDir + "/"
		}
	} else {
		pageURL = "/" + relDir + "/" + stem + "/"
	}

	layout := resolveLayout(pageURL, fm.Layout)
	if layout == "" {
		logging.Warn("layout not resolved", "url", pageURL)
	}
	var listingChildren []childPage
	var diaryYears []string
	if layout == "diaryhome" {
		listingChildren = recentDiaryEntries(children, 20)
		diaryYears = collectDiaryYears(children)
	} else if (stem == "_index" || stem == "index") && strings.TrimSpace(body) == "" {
		switch layout {
		case "diaryentries":
			listingChildren = children[relDir]
			diaryYears = collectDiaryYears(children)
		case "storieshome":
			listingChildren = children[relDir]
		default:
			if listing := generateSectionListing(children[relDir]); listing != "" {
				body = string(listing)
			}
		}
	}

	// Warn about Hugo figure shortcodes in hand-authored sections. Generated
	// content (person pages, lists, etc.) never contains shortcodes, so the
	// check is limited to /diary/ and /stories/ to avoid noise.
	if strings.HasPrefix(relSlash, "diary/") || strings.HasPrefix(relSlash, "stories/") {
		if strings.Contains(body, "{{< figure") {
			logging.Warn("Hugo figure shortcode found — replace with plain HTML <figure>", "file", srcPath)
		}
	}

	var buf bytes.Buffer
	if err := renderer.Convert([]byte(body), &buf); err != nil {
		return fmt.Errorf("render markdown %s: %w", srcPath, err)
	}
	rendered := htmlCommentRE.ReplaceAll(buf.Bytes(), nil)

	tmpl, err := selectTemplate(b.templates, layout, srcPath)
	if err != nil {
		return err
	}

	outPath := b.outputPath(rel)

	// Populate tree-level data. BasePath comes from the normalised front-matter
	// field; Title is looked up from the tree's own section index so it doesn't
	// need to be repeated in every page's front-matter.
	var tree TreeData
	if fm.BasePath != "" {
		tree.BasePath = fm.BasePath
		tree.Title = sectionTitles[strings.Trim(fm.BasePath, "/")]
	}

	var section string
	switch {
	case strings.HasPrefix(relSlash, "stories/"):
		section = "Stories"
	case strings.HasPrefix(relSlash, "diary/"):
		// Exclude diary year index pages (diary/YYYY/_index.md, diary/YYYY/index.md).
		// Those are listing pages, not individual entries.
		parts := strings.Split(relDir, "/")
		if !((stem == "_index" || stem == "index") && len(parts) == 2) {
			section = "Research Diary"
		}
	}

	var prevEntry, nextEntry NavEntry
	if b.diaryNav != nil {
		if pair, ok := b.diaryNav[pageURL]; ok {
			prevEntry = pair[0]
			nextEntry = pair[1]
		}
	}

	if err := writePageFile(tmpl, outPath, PageData{FrontMatter: fm, Body: template.HTML(rendered), Tree: tree, Section: section, PrevEntry: prevEntry, NextEntry: nextEntry, Children: listingChildren, DiaryYears: diaryYears, Debug: b.Debug, PageLayout: layout}); err != nil {
		return fmt.Errorf("render %s: %w", srcPath, err)
	}

	if len(fm.Aliases) > 0 {
		if err := b.writeAliases(fm.Aliases, b.canonicalURL(outPath)); err != nil {
			return fmt.Errorf("aliases for %s: %w", srcPath, err)
		}
	}

	// Add to sitemap only for whitelisted URLs (homepage, /diary/, /stories/,
	// /trees/ and tree homepages). sitemap.disable provides an opt-out for
	// whitelisted pages that should still be excluded (e.g. paginated sub-pages).
	canonURL := b.canonicalURL(outPath)
	if sitemapIncluded(canonURL) && (fm.Sitemap == nil || fm.Sitemap["disable"] == "") {
		b.sitemapEntries = append(b.sitemapEntries, sitemapEntry{
			URL:     canonURL,
			LastMod: fm.LastMod,
		})
	}

	return nil
}

func (b *Builder) copyFile(srcPath, rel string) error {
	outPath := filepath.Join(b.PubDir, rel)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", outPath, err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcPath, err)
	}
	defer src.Close()

	dst, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", srcPath, outPath, err)
	}
	return nil
}

// outputPath converts a content-relative path to its pub destination.
//
//   - index.md and _index.md  →  {dir}/index.html   (same directory)
//   - other .md files          →  {dir}/{stem}/index.html  (clean URL)
//   - everything else          →  {dir}/{file}  (verbatim)
func (b *Builder) outputPath(rel string) string {
	dir := filepath.Dir(rel)
	base := filepath.Base(rel)

	if strings.HasSuffix(base, ".md") {
		stem := strings.TrimSuffix(base, ".md")
		if stem == "index" || stem == "_index" {
			return filepath.Join(b.PubDir, dir, "index.html")
		}
		return filepath.Join(b.PubDir, dir, stem, "index.html")
	}

	return filepath.Join(b.PubDir, rel)
}

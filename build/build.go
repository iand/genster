package build

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// TreeData holds site-level metadata for the genealogy tree a page belongs to.
// It is populated from the tree's section index page and is available in all
// templates as {{.Tree.Title}}, {{.Tree.BasePath}}, etc., without needing to
// repeat the information in every page's front-matter.
type TreeData struct {
	Title    string
	BasePath string
}

// PageData is passed to each page template during rendering. Embedding
// FrontMatter lets templates access fields like {{.Title}}, {{.Layout}}, etc.
// directly alongside {{.Body}} and {{.Tree}}.
type PageData struct {
	FrontMatter
	Body template.HTML
	Tree TreeData
}

// Builder walks a content directory and renders each file into a pub directory.
type Builder struct {
	ContentDir string
	PubDir     string
	// AssetsDir, if non-empty, is a directory of static assets (css/, js/)
	// to copy into PubDir. When empty the assets embedded in the binary are used.
	AssetsDir string
	// BaseURL, if non-empty, is the scheme+host used to build absolute <loc>
	// URLs in sitemap.xml (e.g. "https://example.com"). When empty, no
	// sitemap.xml is written.
	BaseURL string
	// IncludeDrafts, when true, publishes pages with draft: true in their
	// front-matter instead of skipping them.
	IncludeDrafts bool

	// sitemapEntries accumulates pages for sitemap.xml during the build.
	sitemapEntries []sitemapEntry
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

	children, sectionTitles, tagIndex, err := collectChildren(b.ContentDir, b.IncludeDrafts)
	if err != nil {
		return fmt.Errorf("collect children: %w", err)
	}

	if err := filepath.WalkDir(b.ContentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
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

	if (stem == "_index" || stem == "index") && strings.TrimSpace(body) == "" {
		dir := filepath.ToSlash(filepath.Dir(rel))
		var listing template.HTML
		// Diary year index pages (diary/YYYY/) get a richer listing with word
		// counts and tags; all other section index pages get the plain listing.
		parts := strings.Split(dir, "/")
		if len(parts) == 2 && parts[0] == "diary" {
			listing = generateDiaryListing(children[dir])
		} else {
			listing = generateSectionListing(children[dir])
		}
		if listing != "" {
			body = string(listing)
		}
	}

	var buf bytes.Buffer
	if err := renderer.Convert([]byte(stripPrivateShortcodes(body)), &buf); err != nil {
		return fmt.Errorf("render markdown %s: %w", srcPath, err)
	}

	tmpl, err := selectTemplate(fm.Layout, srcPath)
	if err != nil {
		return err
	}

	outPath := b.outputPath(rel)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", outPath, err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()

	// Populate tree-level data. BasePath comes from the normalised front-matter
	// field; Title is looked up from the tree's own section index so it doesn't
	// need to be repeated in every page's front-matter.
	var tree TreeData
	if fm.BasePath != "" {
		tree.BasePath = fm.BasePath
		tree.Title = sectionTitles[strings.Trim(fm.BasePath, "/")]
	}

	if err := tmpl.Execute(f, PageData{FrontMatter: fm, Body: template.HTML(buf.String()), Tree: tree}); err != nil {
		return fmt.Errorf("execute template for %s: %w", srcPath, err)
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

// stripPrivateShortcodes removes Hugo {{< private >}} shortcode markers from
// body text before rendering. These markers are inserted by the annotate
// command to prevent Hugo from including comment content in output; without
// Hugo in the pipeline they would appear literally in HTML comments.
// Stripping is global because {{< private >}} is a genster-specific construct
// that never appears in hand-authored content.
func stripPrivateShortcodes(s string) string {
	s = strings.ReplaceAll(s, "{{< private >}}", "")
	s = strings.ReplaceAll(s, "{{< /private >}}", "")
	return s
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

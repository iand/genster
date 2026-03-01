package build

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// childPage holds the metadata needed to render one entry in a section listing.
type childPage struct {
	Title     string
	URL       string
	Date      string   // YYYY-MM-DD for chronological sorting; empty when unknown
	Tags      []string // front-matter tags; populated for diary entries
	WordCount int      // rough word count of the body; populated for diary entries
}

// collectChildren walks contentDir and returns:
//   - children: a map from content-relative, slash-separated section directory
//     to the list of its immediate child pages.
//   - sectionTitles: a map from content-relative, slash-separated section
//     directory to the title of its index page, used to look up tree titles.
//
// Section index files (_index.md / index.md) are registered as children of
// their parent section; leaf .md files are registered as children of their
// own directory.  The root _index.md is skipped (it has no parent section).
func collectChildren(contentDir string, includeDrafts bool) (children map[string][]childPage, sectionTitles map[string]string, tagIndex map[string][]pageRef, err error) {
	children = make(map[string][]childPage)
	sectionTitles = make(map[string]string)
	tagIndex = make(map[string][]pageRef)

	walkErr := filepath.WalkDir(contentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		rel, err := filepath.Rel(contentDir, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		dir := filepath.ToSlash(filepath.Dir(relSlash))
		stem := strings.TrimSuffix(d.Name(), ".md")

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		fm, body, err := ParseDocument(string(data))
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		if bool(fm.Draft) && !includeDrafts {
			return nil
		}

		// Compute the canonical URL for this file for use in tag listings.
		var pageURL string
		if stem == "_index" || stem == "index" {
			if dir == "." {
				pageURL = "/"
			} else {
				pageURL = "/" + dir + "/"
			}
		} else {
			if dir == "." {
				pageURL = "/" + stem + "/"
			} else {
				pageURL = "/" + dir + "/" + stem + "/"
			}
		}

		// Collect tags: any page with tags contributes to the tag index.
		// stemToTitle normalises raw YYYY-MM-DD title strings to formatted dates.
		if len(fm.Tags) > 0 {
			pageTitle := stemToTitle(fm.Title)
			if pageTitle == "" {
				if stem == "_index" || stem == "index" {
					pageTitle = stemToTitle(filepath.Base(dir))
				} else {
					pageTitle = stemToTitle(stem)
				}
			}
			ref := pageRef{Title: pageTitle, URL: pageURL}
			for _, tag := range fm.Tags {
				tagIndex[tag] = append(tagIndex[tag], ref)
			}
		}

		var cp childPage

		if stem == "_index" || stem == "index" {
			// Record the title for this directory so tree pages can look it up.
			if fm.Title != "" {
				sectionTitles[dir] = fm.Title
			}

			// This file is the index for its directory; it is a child of
			// the parent section.  Skip the root (no parent to register under).
			if dir == "." {
				return nil
			}
			parentDir := filepath.ToSlash(filepath.Dir(dir))
			dirBase := filepath.Base(dir)
			cp.Title = stemToTitle(fm.Title)
			if cp.Title == "" {
				cp.Title = stemToTitle(dirBase)
			}
			cp.URL = pageURL
			cp.Date = dateFromStem(dirBase)
			cp.Tags = fm.Tags
			cp.WordCount = countWords(body)
			children[parentDir] = append(children[parentDir], cp)
		} else {
			// Leaf file: child of its own directory.
			cp.Title = stemToTitle(fm.Title)
			if cp.Title == "" {
				cp.Title = stemToTitle(stem)
			}
			cp.URL = pageURL
			cp.Date = dateFromStem(stem)
			cp.Tags = fm.Tags
			cp.WordCount = countWords(body)
			children[dir] = append(children[dir], cp)
		}

		return nil
	})
	if walkErr != nil {
		return nil, nil, nil, walkErr
	}

	// Sort each child list: date descending, then title ascending.
	for key := range children {
		slices.SortFunc(children[key], func(a, b childPage) int {
			if a.Date != b.Date {
				return strings.Compare(b.Date, a.Date) // reversed for descending
			}
			return strings.Compare(a.Title, b.Title)
		})
	}

	return children, sectionTitles, tagIndex, nil
}

// countWords returns a rough word count for a body string by stripping HTML
// tags and splitting on whitespace.
func countWords(body string) int {
	var inTag bool
	var sb strings.Builder
	for _, r := range body {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			sb.WriteByte(' ')
		case !inTag:
			sb.WriteRune(r)
		}
	}
	return len(strings.Fields(sb.String()))
}

// stemToTitle converts a file or directory stem into a human-readable title.
// Stems in YYYY-MM-DD format become "2 Jan 2006"; anything else is returned as-is.
func stemToTitle(stem string) string {
	t, err := time.Parse("2006-01-02", stem)
	if err == nil {
		return t.Format("2 Jan 2006")
	}
	return stem
}

// dateFromStem returns the stem unchanged if it parses as YYYY-MM-DD,
// or an empty string otherwise.
func dateFromStem(stem string) string {
	if _, err := time.Parse("2006-01-02", stem); err == nil {
		return stem
	}
	return ""
}

// generateSectionListing returns an HTML unordered list of the given child
// pages.  The caller is responsible for sorting before calling.
// Returns empty HTML when pages is empty.
func generateSectionListing(pages []childPage) template.HTML {
	if len(pages) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`<ul class="list">`)
	for _, p := range pages {
		sb.WriteString("\n  <li><a href=\"")
		sb.WriteString(p.URL)
		sb.WriteString("\">")
		sb.WriteString(template.HTMLEscapeString(p.Title))
		sb.WriteString("</a></li>")
	}
	sb.WriteString("\n</ul>\n")
	return template.HTML(sb.String())
}

// generateDiaryListing returns an HTML list for a diary year index page.
// Each entry shows a date link, a rough word count, and the entry's tags.
func generateDiaryListing(pages []childPage) template.HTML {
	if len(pages) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`<ul class="list">`)
	for _, p := range pages {
		sb.WriteString("\n  <li>")
		sb.WriteString(`<a href="`)
		sb.WriteString(p.URL)
		sb.WriteString(`">`)
		sb.WriteString(template.HTMLEscapeString(p.Title))
		sb.WriteString("</a>")
		hasMeta := p.WordCount > 0 || len(p.Tags) > 0
		if hasMeta {
			sb.WriteString(`<hr class="hr-list"><span class="diary-meta">`)
			if p.WordCount > 0 {
				fmt.Fprintf(&sb, "~%d words", p.WordCount)
			}
			for i, tag := range p.Tags {
				if i > 0 || p.WordCount > 0 {
					sb.WriteString(" · ")
				}
				sb.WriteString(`<a href="/tags/`)
				sb.WriteString(urlize(tag))
				sb.WriteString(`/">`)
				sb.WriteString(template.HTMLEscapeString(tag))
				sb.WriteString("</a>")
			}
			sb.WriteString("</span>")
		}
		sb.WriteString("</li>")
	}
	sb.WriteString("\n</ul>\n")
	return template.HTML(sb.String())
}

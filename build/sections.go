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
	Summary   string   // optional description shown beneath the title in listings
	Tags      []string // front-matter tags; populated for diary and story entries
	WordCount int      // rough word count of the body; populated for diary and story entries
	FM        FrontMatter // complete front-matter for the child page
}

// collectChildren walks contentDir and returns:
//   - children: a map from content-relative, slash-separated section directory
//     to the list of its immediate child pages.
//   - sectionTitles: a map from content-relative, slash-separated section
//     directory to the title of its index page, used to look up tree titles.
//   - draftDirs: content-relative slash-separated directory paths whose index
//     file has draft:true (only populated when includeDrafts is false). Used
//     by Build to skip the entire directory, including images.
//
// Section index files (_index.md / index.md) are registered as children of
// their parent section; leaf .md files are registered as children of their
// own directory.  The root _index.md is skipped (it has no parent section).
func collectChildren(contentDir string, includeDrafts bool) (children map[string][]childPage, sectionTitles map[string]string, tagIndex map[string][]pageRef, draftDirs map[string]bool, err error) {
	children = make(map[string][]childPage)
	sectionTitles = make(map[string]string)
	tagIndex = make(map[string][]pageRef)
	draftDirs = make(map[string]bool)

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
			// If this is a directory index, mark the whole directory as a draft
			// so that Build can skip it entirely (including images etc.).
			if stem == "_index" || stem == "index" {
				draftDirs[dir] = true
			}
			return nil
		}

		if bool(fm.Hide) {
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
			ref := pageRef{Title: pageTitle, URL: pageURL, Summary: fm.Summary, Ancestor: bool(fm.Ancestor)}
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
			cp.Summary = fm.Summary

			cp.Tags = fm.Tags
			cp.WordCount = countWords(body)
			cp.FM = fm
			children[parentDir] = append(children[parentDir], cp)
		} else {
			// Leaf file: child of its own directory.
			cp.Title = stemToTitle(fm.Title)
			if cp.Title == "" {
				cp.Title = stemToTitle(stem)
			}
			cp.URL = pageURL
			cp.Date = dateFromStem(stem)
			cp.Summary = fm.Summary

			cp.Tags = fm.Tags
			cp.WordCount = countWords(body)
			cp.FM = fm
			children[dir] = append(children[dir], cp)
		}

		return nil
	})
	if walkErr != nil {
		return nil, nil, nil, nil, walkErr
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

	return children, sectionTitles, tagIndex, draftDirs, nil
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

// buildDiaryNav collects all diary leaf entries (children of diary/YYYY
// directories), sorts them chronologically ascending, and returns a map from
// each entry's canonical URL to its [prev, next] NavEntry pair.  Entries with
// no date sort last.  A zero NavEntry means no adjacent entry exists.
func buildDiaryNav(children map[string][]childPage) map[string][2]NavEntry {
	var entries []childPage
	for key, pages := range children {
		parts := strings.Split(key, "/")
		if len(parts) == 2 && parts[0] == "diary" {
			entries = append(entries, pages...)
		}
	}
	if len(entries) == 0 {
		return nil
	}

	// Sort ascending by date so prev=older, next=newer.
	slices.SortFunc(entries, func(a, b childPage) int {
		if a.Date == b.Date {
			return strings.Compare(a.Title, b.Title)
		}
		if a.Date == "" {
			return 1
		}
		if b.Date == "" {
			return -1
		}
		return strings.Compare(a.Date, b.Date)
	})

	nav := make(map[string][2]NavEntry, len(entries))
	for i, e := range entries {
		var prev, next NavEntry
		if i > 0 {
			prev = NavEntry{URL: entries[i-1].URL, Title: entries[i-1].Title}
		}
		if i < len(entries)-1 {
			next = NavEntry{URL: entries[i+1].URL, Title: entries[i+1].Title}
		}
		nav[e.URL] = [2]NavEntry{prev, next}
	}
	return nav
}

// recentDiaryEntries collects all diary leaf entries across all diary year
// directories, sorts them by date descending (most recent first), and returns
// at most n entries. Entries with no date sort last.
func recentDiaryEntries(children map[string][]childPage, n int) []childPage {
	var entries []childPage
	for key, pages := range children {
		parts := strings.Split(key, "/")
		if len(parts) == 2 && parts[0] == "diary" {
			entries = append(entries, pages...)
		}
	}
	slices.SortFunc(entries, func(a, b childPage) int {
		if a.Date == b.Date {
			return strings.Compare(b.Title, a.Title)
		}
		if a.Date == "" {
			return 1
		}
		if b.Date == "" {
			return -1
		}
		return strings.Compare(b.Date, a.Date) // descending
	})
	if n > 0 && len(entries) > n {
		entries = entries[:n]
	}
	return entries
}

// collectDiaryYears returns the diary years present in children, sorted
// descending (most recent first), for use in sidebar navigation.
func collectDiaryYears(children map[string][]childPage) []string {
	var years []string
	for key := range children {
		parts := strings.Split(key, "/")
		if len(parts) == 2 && parts[0] == "diary" {
			years = append(years, parts[1])
		}
	}
	slices.Sort(years)
	slices.Reverse(years)
	return years
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

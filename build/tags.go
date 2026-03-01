package build

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// pageRef identifies a page for use in tag index listings.
type pageRef struct {
	Title string
	URL   string
}

// writeTags writes pub/tags/index.html and pub/tags/{slug}/index.html for
// every tag collected during the first pass.  Tags are sorted alphabetically;
// pages within each tag are sorted by title.
func (b *Builder) writeTags(tagIndex map[string][]pageRef) error {
	if len(tagIndex) == 0 {
		return nil
	}

	tagTmpl := siteTemplates.Lookup("single")
	if tagTmpl == nil {
		return fmt.Errorf("single template not found")
	}
	indexTmpl := siteTemplates.Lookup("tagsindex")
	if indexTmpl == nil {
		return fmt.Errorf("tagsindex template not found")
	}

	// Sorted tag list for the index page.
	tags := make([]string, 0, len(tagIndex))
	for tag := range tagIndex {
		tags = append(tags, tag)
	}
	slices.Sort(tags)

	// Write each individual tag page.
	for _, tag := range tags {
		pages := tagIndex[tag]
		slices.SortFunc(pages, func(a, b pageRef) int {
			return strings.Compare(a.Title, b.Title)
		})

		outPath := filepath.Join(b.PubDir, "tags", urlize(tag), "index.html")
		if err := writePageFile(tagTmpl, outPath, PageData{
			FrontMatter: FrontMatter{Title: "Pages tagged \"" + tag + "\""},
			Body:        tagPageBody(pages),
		}); err != nil {
			return fmt.Errorf("tag page %q: %w", tag, err)
		}
	}

	// Write the tags index using the dedicated tagsindex template which
	// includes sidebar text explaining how tags work.
	if err := writePageFile(indexTmpl, filepath.Join(b.PubDir, "tags", "index.html"), PageData{
		FrontMatter: FrontMatter{Title: "Tags"},
		Body:        tagIndexBody(tags, tagIndex),
	}); err != nil {
		return fmt.Errorf("tags index: %w", err)
	}

	return nil
}

// writePageFile creates outPath (and any parent directories) and executes tmpl
// into it.
func writePageFile(tmpl *template.Template, outPath string, data PageData) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outPath, err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute %s: %w", outPath, err)
	}
	return nil
}

// tagGroups defines the display order for content-type groups on tag pages.
var tagGroups = []string{"People", "Places", "Stories", "Diary entries", "Other"}

// groupFromURL returns the content category for a page based on its URL.
func groupFromURL(url string) string {
	switch {
	case strings.HasPrefix(url, "/diary/"):
		return "Diary entries"
	case strings.HasPrefix(url, "/stories/"):
		return "Stories"
	case strings.Contains(url, "/person/"):
		return "People"
	case strings.Contains(url, "/place/"):
		return "Places"
	default:
		return "Other"
	}
}

// tagPageBody generates the HTML listing for a single tag's page, with pages
// grouped by content type under headings.
func tagPageBody(pages []pageRef) template.HTML {
	// Bucket pages into groups.
	grouped := make(map[string][]pageRef, len(tagGroups))
	for _, p := range pages {
		g := groupFromURL(p.URL)
		grouped[g] = append(grouped[g], p)
	}

	var sb strings.Builder
	for _, group := range tagGroups {
		ps := grouped[group]
		if len(ps) == 0 {
			continue
		}
		sb.WriteString("<h2>")
		sb.WriteString(group)
		sb.WriteString("</h2>\n<ul>\n")
		for _, p := range ps {
			sb.WriteString("  <li><a href=\"")
			sb.WriteString(p.URL)
			sb.WriteString("\">")
			sb.WriteString(template.HTMLEscapeString(p.Title))
			sb.WriteString("</a></li>\n")
		}
		sb.WriteString("</ul>\n")
	}
	return template.HTML(sb.String())
}

// tagIndexBody generates the HTML listing for the /tags/ index page.
func tagIndexBody(tags []string, tagIndex map[string][]pageRef) template.HTML {
	var sb strings.Builder
	sb.WriteString("<ul class=\"tag-list\">\n")
	for _, tag := range tags {
		sb.WriteString("  <li><a href=\"/tags/")
		sb.WriteString(urlize(tag))
		sb.WriteString("/\">")
		sb.WriteString(template.HTMLEscapeString(tag))
		fmt.Fprintf(&sb, " (%d)", len(tagIndex[tag]))
		sb.WriteString("</a></li>\n")
	}
	sb.WriteString("</ul>\n")
	return template.HTML(sb.String())
}

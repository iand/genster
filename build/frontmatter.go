// Package build implements the static site build step that renders markdown
// content files into complete HTML pages.
package build

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontMatter holds the metadata written by genster into the YAML front-matter
// block (delimited by ---) of each markdown content file.
type FrontMatter struct {
	// Core identification
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Layout string `yaml:"layout"`

	// Page description
	Summary  string `yaml:"summary"`
	Category string `yaml:"category"`
	Image    string `yaml:"image"`
	BasePath string `yaml:"basepath"`
	LastMod  string `yaml:"lastmod"`

	// Taxonomy
	Tags    []string `yaml:"tags"`
	Aliases []string `yaml:"aliases"`

	// Pagination (list pages only)
	First string `yaml:"first"`
	Last  string `yaml:"last"`
	Next  string `yaml:"next"`
	Prev  string `yaml:"prev"`

	// Person-specific
	Gender         string `yaml:"gender"`
	Era            string `yaml:"era"`
	Maturity       string `yaml:"maturity"`
	Trade          string `yaml:"trade"`
	GrampsID       string `yaml:"grampsid"`
	Slug           string `yaml:"slug"`
	WikiTreeFormat string `yaml:"wikitreeformat"`
	WikiTreeID     string `yaml:"wikitreeid"`
	MarkdownFormat string `yaml:"markdownformat"`

	// Place-specific
	PlaceType    string `yaml:"placetype"`
	BuildingKind string `yaml:"buildingkind"`

	// Calendar-specific
	Month string `yaml:"month"`

	// Sidebar link lists (person pages)
	DiaryLinks  []map[string]string `yaml:"diarylinks"`
	Links       []map[string]string `yaml:"links"`
	Descendants []map[string]string `yaml:"descendants"`

	// Sitemap control: {disable: "1"} suppresses the page from sitemap.xml
	Sitemap map[string]string `yaml:"sitemap"`
}

// ParseDocument splits a genster-generated markdown file into its YAML
// front-matter and body text. The front-matter block must be delimited by
// lines containing only "---". If no front-matter block is present, fm is
// zero-valued and body equals the full input.
func ParseDocument(content string) (fm FrontMatter, body string, err error) {
	const delim = "---\n"

	if !strings.HasPrefix(content, delim) {
		return fm, content, nil
	}

	rest := content[len(delim):]

	// The closing delimiter must be at the start of a line, so look for \n---\n.
	end := strings.Index(rest, "\n"+delim)
	if end == -1 {
		return fm, content, fmt.Errorf("front-matter: missing closing ---")
	}

	yamlBlock := rest[:end]
	body = rest[end+1+len(delim):]

	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return fm, body, fmt.Errorf("front-matter: %w", err)
	}

	return fm, body, nil
}

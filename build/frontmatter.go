// Package build implements the static site build step that renders markdown
// content files into complete HTML pages.
package build

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// flexStrings is a string slice that unmarshals from either a single YAML
// string or a YAML sequence, so front-matter fields like "people" work
// whether the author writes one value or a list.
type flexStrings []string

func (f *flexStrings) UnmarshalYAML(value *yaml.Node) error {
	switch value.Tag {
	case "!!str":
		*f = []string{value.Value}
	case "!!seq":
		var v []string
		if err := value.Decode(&v); err != nil {
			return err
		}
		*f = v
	default:
		return fmt.Errorf("cannot unmarshal %s into []string", value.Tag)
	}
	return nil
}

// flexBool is a bool that unmarshals from either a YAML boolean (draft: true)
// or a quoted string (draft: "true"), as both appear in existing content files.
type flexBool bool

func (b *flexBool) UnmarshalYAML(value *yaml.Node) error {
	switch value.Tag {
	case "!!bool":
		var v bool
		if err := value.Decode(&v); err != nil {
			return err
		}
		*b = flexBool(v)
	case "!!str":
		s := strings.ToLower(strings.TrimSpace(value.Value))
		*b = flexBool(s == "true" || s == "yes" || s == "1")
	default:
		return fmt.Errorf("cannot unmarshal %s into bool", value.Tag)
	}
	return nil
}

// FrontMatter holds the metadata written by genster into the YAML front-matter
// block (delimited by ---) of each markdown content file.
type FrontMatter struct {
	// Core identification
	ID      string   `yaml:"id"`
	Title   string   `yaml:"title"`
	Layout  string   `yaml:"layout"`
	Draft   flexBool `yaml:"draft"`
	Private flexBool `yaml:"private"`
	Hide    flexBool `yaml:"hide"`

	// Page description
	Summary   string `yaml:"summary"`
	Category  string `yaml:"category"`
	Image     string `yaml:"image"`
	BasePath  string `yaml:"basepath"`
	TreeTitle string `yaml:"treetitle"`
	LastMod   string `yaml:"lastmod"`
	Converted bool   `yaml:"converted"`

	// Taxonomy
	Tags    []string `yaml:"tags"`
	Aliases []string `yaml:"aliases"`

	// Pagination (list pages only)
	First string `yaml:"first"`
	Last  string `yaml:"last"`
	Next  string `yaml:"next"`
	Prev  string `yaml:"prev"`

	// Person-specific
	Gender         string   `yaml:"gender"`
	Era            string   `yaml:"era"`
	Maturity       string   `yaml:"maturity"`
	Trade          string   `yaml:"trade"`
	GrampsID       string   `yaml:"grampsid"`
	Slug           string   `yaml:"slug"`
	WikiTreeFormat string   `yaml:"wikitreeformat"`
	WikiTreeID     string   `yaml:"wikitreeid"`
	MarkdownFormat string   `yaml:"markdownformat"`
	Ancestor       flexBool `yaml:"ancestor"`

	// Place-specific
	PlaceType    string `yaml:"placetype"`
	BuildingKind string `yaml:"buildingkind"`

	// Calendar-specific
	Month string `yaml:"month"`

	// Question/story-specific
	People     flexStrings         `yaml:"people"`
	Author     string              `yaml:"author"`
	Started    string              `yaml:"started"`
	Updated    string              `yaml:"updated"`
	Status     string              `yaml:"status"`
	Ai         string              `yaml:"ai"`
	StoryParts []map[string]string `yaml:"storyparts"`

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

package build

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"
)

//go:embed templates
var templateFS embed.FS

// urlize converts a string to a URL-safe slug: lowercase, spaces → hyphens.
// It is used both in templates (via templateFuncs) and directly in Go code
// wherever tag slugs are needed, keeping the two in sync.
func urlize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// templateFuncs provides helper functions available to all page templates.
var templateFuncs = template.FuncMap{
	"urlize": urlize,
	// ukdate formats a YYYY-MM-DD date string as "2 January 2006".
	// Returns the original string unchanged if it cannot be parsed.
	"ukdate": func(s string) string {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return s
		}
		return t.Format("2 January 2006")
	},
}

// siteTemplates is the parsed template set containing all named layout
// templates and the shared partials they reference (head, footer, etc.).
var siteTemplates = template.Must(
	template.New("").Funcs(templateFuncs).ParseFS(templateFS, "templates/*.html"),
)

// knownLayouts is the set of layout values that genster and manual content
// files may use. An unrecognised layout causes an error at build time so
// misconfigured files surface immediately rather than silently rendering blank.
var knownLayouts = map[string]bool{
	// Generated tree pages
	"person":          true,
	"family":          true,
	"place":           true,
	"citation":        true,
	"source":          true,
	"listpeople":      true,
	"listsurnames":    true,
	"listplaces":      true,
	"listsources":     true,
	"listanomalies":   true,
	"listinferences":  true,
	"listtodo":        true,
	"listfamilies":    true,
	"listfamilylines": true,
	"chartancestors":  true,
	"treeoverview":    true,
	"listtrees":       true,
	"calendar":        true,
	// Manual content
	"home":   true,
	"search": true,
	"single": true,
	"list":   true,
	"diary":  true,
	// No layout specified (plain content files)
	"": true,
}

// selectTemplate returns the named template to render the given layout. An
// unrecognised layout returns an error that identifies both the source file
// and the bad layout value.
func selectTemplate(layout, srcPath string) (*template.Template, error) {
	if !knownLayouts[layout] {
		return nil, fmt.Errorf("unknown layout %q in %s", layout, srcPath)
	}

	// Map the empty layout to the "plain" template.
	name := layout
	if name == "" {
		name = "plain"
	}

	tmpl := siteTemplates.Lookup(name)
	if tmpl == nil {
		// All knownLayouts should have matching templates; this indicates a
		// programming error (template file missing or define block misspelled).
		return nil, fmt.Errorf("no template found for layout %q in %s", layout, srcPath)
	}
	return tmpl, nil
}

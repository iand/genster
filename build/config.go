package build

import "github.com/iand/genster/layout"

// config.go contains the site-wide configuration that an operator might want
// to adjust: which URL patterns map to which layout templates, and which URLs
// appear in sitemap.xml.
//
// Both sets of rules use the same URL pattern syntax (gitignore-inspired,
// root-anchored):
//
//	/exact/path/   Exact match.
//	/prefix/**     Prefix match including the prefix URL itself.
//	/path/*/rest/  Single-segment wildcard (* matches one non-empty segment).

// layoutRules maps URL patterns to layout template names. Rules are checked in
// order; the first match wins and overrides the layout field in front-matter.
// Pages not covered by any rule use their front-matter layout (or "plain" if
// none is set).
var layoutRules = []struct {
	Pattern string
	Layout  layout.PageLayout
}{
	// Site-wide hand-authored pages.
	{"/search/", "search"},

	// Diary section.
	{"/diary/", "diaryhome"},      // diary homepage with recent entries
	{"/diary/*/", "diaryentries"}, // per-year listing (e.g. /diary/2024/)
	{"/diary/*/*/", "diary"},      // diary entry

	// Stories section.
	{"/stories/", "storieshome"},
	{"/stories/**", "story"},

	// Questions section.
	{"/questions/", "questionshome"},
	{"/questions/**", "question"},

	// Trees section index.
	{"/trees/", "listtrees"},
	{"/trees/*/", "treeoverview"},

	// Tree entity pages — one per person, family, place, citation, or source.
	{"/trees/*/person/*/", layout.PageLayoutPerson},
	{"/trees/*/family/*/", layout.PageLayoutFamily},
	{"/trees/*/place/*/", "place"},
	{"/trees/*/citation/*/", "citation"},
	{"/trees/*/source/*/", "source"},

	// Tree calendar pages — one per month (e.g. /trees/cg/calendar/january/).
	{"/trees/*/calendar/*/", "calendar"},

	// Tree chart pages.
	{"/trees/*/chart/ancestors/*/", "chartancestors"},
	{"/trees/*/chart/trees/*/", "charttrees"},

	// Tree list pages.  Each type has two patterns because some types have
	// both a top-level index page and paginated sub-pages; others have only
	// one form.  Having both patterns is harmless for types that use only one.
	{"/trees/*/list/people/*/", "listpeople"},
	{"/trees/*/list/surnames/*/", "listsurnames"},
	{"/trees/*/list/places/*/", "listplaces"},
	{"/trees/*/list/sources/*/", "listsources"},
	{"/trees/*/list/anomalies/", "listanomalies"},
	{"/trees/*/list/anomalies/*/", "listanomalies"},
	{"/trees/*/list/inferences/", "listinferences"},
	{"/trees/*/list/inferences/*/", "listinferences"},
	{"/trees/*/list/todo/*/", "listtodo"},
	{"/trees/*/list/families/*/", "listfamilies"},
	{"/trees/*/list/familylines/", "listfamilylines"},
	{"/trees/*/list/familylines/*/", "listfamilylines"},
	{"/trees/*/list/changes/", "listchanges"},
	{"/trees/*/list/changes/*/", "listchanges"},
}

// resolveLayout returns the effective layout for a page. layoutRules are
// consulted first; fmLayout (the value of the layout field in front-matter) is
// used as a fallback when no rule matches.
func resolveLayout(pageURL, fmLayout string) string {
	for _, r := range layoutRules {
		if matchURLPattern(r.Pattern, pageURL) {
			return string(r.Layout)
		}
	}
	return fmLayout
}

// sitemapRules is the inclusion allow-list for sitemap.xml. Every URL is
// excluded by default; a URL is included only when it matches at least one
// rule here.
//
// Precedence (highest to lowest):
//
//  1. Front-matter sitemap.disable — if set on a page, that page is excluded
//     regardless of sitemapRules.
//  2. sitemapRules — the first matching rule wins. If no rule matches, the
//     URL is excluded.
var sitemapRules = []string{
	"/",             // homepage only
	"/diary/**",     // diary section index and all diary entries
	"/stories/**",   // stories section index and all stories
	"/questions/**", // stories section index and all stories
	"/trees/",       // trees section index
	"/trees/*/",     // tree homepages (one level deep; grandchildren excluded)
}

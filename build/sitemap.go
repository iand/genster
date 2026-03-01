package build

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// sitemapEntry is one URL entry to be written to sitemap.xml.
type sitemapEntry struct {
	URL     string // site-root-relative URL, e.g. "/diary/2021/2021-05-17/"
	LastMod string // YYYY-MM-DD; empty when unknown
}

// sitemapRules is the inclusion allow-list for sitemap.xml.  Every URL is
// excluded by default; a URL is included only when it matches at least one
// rule here.
//
// Precedence (highest to lowest):
//
//  1. Front-matter sitemap.disable — if set on a page, that page is excluded
//     regardless of sitemapRules.
//  2. sitemapRules — the first matching rule wins.  If no rule matches, the
//     URL is excluded.
//
// Rule syntax (gitignore-inspired; all rules are root-anchored):
//
//	/exact/path/    Exact match.  Only that specific URL is included.
//
//	/prefix/**      Prefix match.  The prefix URL itself and every URL beneath
//	                it are included.  "**" at the end matches zero or more
//	                additional path segments, so /diary/** includes /diary/,
//	                /diary/2021/, and /diary/2021/2021-05-17/.
//
//	/path/*/rest/   Single-segment wildcard.  "*" matches exactly one non-empty
//	                path segment and does not cross a slash.  For example,
//	                /trees/*/ includes /trees/at/ and /trees/cg/ but not
//	                /trees/ itself or /trees/at/person/.
var sitemapRules = []string{
	"/",          // homepage only
	"/diary/**",  // diary section index and all diary entries
	"/stories/**", // stories section index and all stories
	"/trees/",    // trees section index
	"/trees/*/",  // tree homepages (one level deep; grandchildren excluded)
}

// sitemapIncluded reports whether the site-root-relative url should appear in
// sitemap.xml according to sitemapRules.  It does not consult front-matter;
// callers must apply the sitemap.disable override separately.
func sitemapIncluded(url string) bool {
	for _, rule := range sitemapRules {
		if matchSitemapRule(rule, url) {
			return true
		}
	}
	return false
}

// matchSitemapRule reports whether url matches the single pattern rule.
func matchSitemapRule(pattern, url string) bool {
	if !strings.Contains(pattern, "*") {
		// No wildcard: exact match only.
		return pattern == url
	}
	// "/**" suffix: prefix match — the prefix URL itself and everything beneath.
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "**")
		return strings.HasPrefix(url, prefix)
	}
	// Single-segment "*": split by "/" and match segment by segment.
	// Pattern and URL must produce the same number of segments.
	patParts := strings.Split(pattern, "/")
	urlParts := strings.Split(url, "/")
	if len(patParts) != len(urlParts) {
		return false
	}
	for i, p := range patParts {
		if p == "*" {
			if urlParts[i] == "" {
				return false // "*" must match a non-empty segment
			}
			continue
		}
		if p != urlParts[i] {
			return false
		}
	}
	return true
}

// writeSitemap writes pub/sitemap.xml from the collected entries.
// baseURL must be the scheme+host of the site (e.g. "https://example.com").
// Entries with empty URLs are skipped.
func writeSitemap(pubDir, baseURL string, entries []sitemapEntry) error {
	baseURL = strings.TrimRight(baseURL, "/")

	type urlElement struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	}
	type urlSet struct {
		XMLName xml.Name     `xml:"urlset"`
		XMLNS   string       `xml:"xmlns,attr"`
		URLs    []urlElement `xml:"url"`
	}

	us := urlSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, e := range entries {
		if e.URL == "" {
			continue
		}
		us.URLs = append(us.URLs, urlElement{
			Loc:     baseURL + e.URL,
			LastMod: e.LastMod,
		})
	}

	data, err := xml.MarshalIndent(us, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sitemap: %w", err)
	}

	outPath := filepath.Join(pubDir, "sitemap.xml")
	if err := os.WriteFile(outPath, append([]byte(xml.Header), data...), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

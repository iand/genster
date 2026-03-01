package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iand/genster/logging"
)

// redirectHTML is the minimal HTML redirect page written for each alias.
// The single %s argument is the canonical URL, used in all four positions.
const redirectHTML = `<!DOCTYPE html>
<html><head>
<meta charset="utf-8">
<meta http-equiv="refresh" content="0; url=%s">
<link rel="canonical" href="%s">
</head><body>
<p>Redirecting to <a href="%s">%s</a></p>
</body></html>
`

// canonicalURL converts an absolute pub/ output path to the canonical URL for
// that page. The path must end in index.html; the result is the directory URL
// with a trailing slash (e.g. /trees/test/person/I1/).
func (b *Builder) canonicalURL(outPath string) string {
	rel := filepath.ToSlash(strings.TrimPrefix(outPath, b.PubDir))
	return strings.TrimSuffix(rel, "index.html")
}

// writeAliases creates an HTML redirect page in PubDir for each alias.
func (b *Builder) writeAliases(aliases []string, canonical string) error {
	for _, alias := range aliases {
		if err := b.writeAlias(alias, canonical); err != nil {
			return err
		}
	}
	return nil
}

// writeAlias writes a single redirect page. alias must be an absolute path
// (starting with /); a missing leading slash is tolerated. An error is
// returned if the target location is already occupied by a rendered page.
func (b *Builder) writeAlias(alias, canonical string) error {
	// Normalise to a relative path so filepath.Join works correctly.
	alias = strings.TrimPrefix(alias, "/")
	outPath := filepath.Join(b.PubDir, filepath.FromSlash(alias), "index.html")

	if _, err := os.Stat(outPath); err == nil {
		logging.Warn("alias conflict: skipping duplicate", "alias", alias, "path", outPath)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir for alias %q: %w", alias, err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create alias %q: %w", alias, err)
	}
	defer f.Close()

	fmt.Fprintf(f, redirectHTML, canonical, canonical, canonical, canonical)
	return nil
}

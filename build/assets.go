package build

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// embeddedAssets holds the CSS and JS files bundled into the binary.
// Images (stock fallback photos for feature image selection) are not embedded
// here because of their size; they are handled separately by gen-b81d via the
// --assets flag.
//
//go:embed assets/css assets/js assets/images
var embeddedAssets embed.FS

// writeAssets copies static CSS and JS files into pubDir. If assetsDir is
// non-empty it is used as the source (must contain css/ and js/ subdirectories
// mirroring the embedded layout); otherwise the files embedded in the binary
// are written.
func writeAssets(pubDir, assetsDir string) error {
	var src fs.FS
	if assetsDir != "" {
		src = os.DirFS(assetsDir)
	} else {
		sub, err := fs.Sub(embeddedAssets, "assets")
		if err != nil {
			return fmt.Errorf("assets sub-fs: %w", err)
		}
		src = sub
	}
	return copyFS(pubDir, src)
}

// copyFS walks src and writes every file into dstDir, preserving the relative
// path structure.
func copyFS(dstDir string, src fs.FS) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(src, path)
		if err != nil {
			return fmt.Errorf("read asset %s: %w", path, err)
		}
		dst := filepath.Join(dstDir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", dst, err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("write asset %s: %w", dst, err)
		}
		return nil
	})
}

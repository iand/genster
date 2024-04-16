package site

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CreateFile(fname string) (*os.File, error) {
	path := filepath.Dir(fname)

	if err := os.MkdirAll(path, 0o777); err != nil {
		return nil, fmt.Errorf("create path: %w", err)
	}

	f, err := os.Create(fname)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return f, nil
}

func CopyFile(dst, src string) error {
	if dst == src {
		return fmt.Errorf("can't copy onto same file")
	}

	fsrc, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer fsrc.Close()

	fdst, err := CreateFile(dst)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer fdst.Close()

	if _, err := io.Copy(fdst, fsrc); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}

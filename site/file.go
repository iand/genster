package site

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateFile(fname string) (*os.File, error) {
	path := filepath.Dir(fname)

	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, fmt.Errorf("create path: %w", err)
	}

	f, err := os.Create(fname)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return f, nil
}

package archive

import (
	"os"
	"path/filepath"
)

// ignoredRootEntries are skipped when detecting a single top-level directory in archives.
var ignoredRootEntries = map[string]struct{}{
	".DS_Store":     {},
	"Thumbs.db":     {},
	"desktop.ini":   {},
	"__MACOSX":      {},
}

// StripSingleRootDir returns the sole child directory when root contains exactly one
// top-level directory and no other meaningful entries (files or other dirs).
// Otherwise it returns root unchanged.
func StripSingleRootDir(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}

	var soleDir string
	dirCount := 0
	for _, e := range entries {
		name := e.Name()
		if _, skip := ignoredRootEntries[name]; skip {
			continue
		}
		if e.IsDir() {
			dirCount++
			soleDir = name
			continue
		}
		// Any top-level file => do not strip.
		return root, nil
	}
	if dirCount != 1 || soleDir == "" {
		return root, nil
	}
	return filepath.Join(root, soleDir), nil
}

package hub

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// ValidatePackage checks that an extracted package matches its declared category.
func ValidatePackage(category, rootDir string) error {
	category, err := NormalizeCategory(category)
	if err != nil {
		return err
	}
	switch category {
	case CategoryPicoClaw:
		// No required manifest files for picoclaw packages.
	case CategoryOpenClaw:
		if !treeContainsFile(rootDir, "AGENT.md") {
			return fmt.Errorf("openclaw package must contain AGENT.md")
		}
	default:
		if !treeContainsFile(rootDir, "AGENT.md") {
			return fmt.Errorf("package must contain AGENT.md")
		}
	}
	return nil
}

func treeContainsFile(root, baseName string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == baseName {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateIdentifier checks that agentName/version is non-empty and path-safe.
func ValidateIdentifier(id string) error {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return fmt.Errorf("identifier is required and must be a non-empty string")
	}
	if strings.ContainsAny(trimmed, "/\\") || strings.Contains(trimmed, "..") {
		return fmt.Errorf("identifier must not contain path separators or '..'")
	}
	return nil
}

// SafePath validates that rel stays inside root without checking existence.
func SafePath(root, rel string) (string, error) {
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "/"))
	if rel == "" || rel == "." {
		return "", fmt.Errorf("file path is required")
	}
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("invalid path")
	}
	full := filepath.Join(root, filepath.FromSlash(rel))
	rootClean := filepath.Clean(root)
	fullClean := filepath.Clean(full)
	if fullClean != rootClean && !strings.HasPrefix(fullClean, rootClean+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	relOut, err := filepath.Rel(rootClean, fullClean)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(relOut), nil
}

// SafeRelativePath ensures a requested file path stays inside root and exists as a file.
func SafeRelativePath(root, rel string) (string, error) {
	safe, err := SafePath(root, rel)
	if err != nil {
		return "", err
	}
	fullClean := filepath.Clean(filepath.Join(root, filepath.FromSlash(safe)))
	info, err := os.Stat(fullClean)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}
	return safe, nil
}

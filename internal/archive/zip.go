package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const maxZipEntrySize = 200 << 20 // align with hub upload limit

// ExtractZipFile extracts a ZIP archive to targetDir with path traversal protection.
func ExtractZipFile(zipPath, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("invalid ZIP: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	targetClean := filepath.Clean(targetDir)
	for _, f := range reader.File {
		rel, isDir, err := normalizeZipEntry(f.Name, f.FileInfo().IsDir(), f.UncompressedSize64)
		if err != nil {
			return err
		}
		if rel == "" {
			continue
		}

		destPath := filepath.Join(targetDir, filepath.FromSlash(rel))
		destClean := filepath.Clean(destPath)
		if destClean != targetClean && !strings.HasPrefix(destClean, targetClean+string(filepath.Separator)) {
			return fmt.Errorf("zip entry escapes target dir: %q", f.Name)
		}

		mode := f.Mode()
		if mode&os.ModeSymlink != 0 {
			return fmt.Errorf("zip contains symlink %q; symlinks are not allowed", f.Name)
		}

		if isDir {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		if err := extractZipEntry(f, destPath); err != nil {
			return err
		}
	}
	return nil
}

// normalizeZipEntry converts a ZIP entry name to a safe relative path using slash separators.
func normalizeZipEntry(name string, modeIsDir bool, uncompressedSize uint64) (rel string, isDir bool, err error) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "./")
	isDir = classifyZipEntryIsDir(name, modeIsDir, uncompressedSize)
	name = strings.TrimSuffix(name, "/")
	name = path.Clean(name)
	if name == "." || name == "" {
		return "", isDir, nil
	}
	if strings.HasPrefix(name, "..") || path.IsAbs(name) {
		return "", false, fmt.Errorf("zip entry has unsafe path: %q", name)
	}
	return name, isDir, nil
}

// classifyZipEntryIsDir decides whether a ZIP entry is a directory.
// Some archivers (notably on Windows) set the directory bit on file entries; entries
// with payload must always be treated as files.
func classifyZipEntryIsDir(name string, modeIsDir bool, uncompressedSize uint64) bool {
	if strings.HasSuffix(name, "/") {
		return true
	}
	if uncompressedSize > 0 {
		return false
	}
	if !modeIsDir {
		return false
	}
	base := path.Base(name)
	if base == "" || base == "." {
		return true
	}
	// Zero-byte entries with a dotted basename are almost always files.
	if strings.Contains(base, ".") {
		return false
	}
	return true
}

func extractZipEntry(f *zip.File, destPath string) error {
	if f.UncompressedSize64 > maxZipEntrySize {
		return fmt.Errorf("zip entry %q is too large (%d bytes)", f.Name, f.UncompressedSize64)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %q: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file %q: %w", destPath, err)
	}

	written, err := io.Copy(out, io.LimitReader(rc, maxZipEntrySize+1))
	if err != nil {
		_ = out.Close()
		_ = os.Remove(destPath)
		return fmt.Errorf("extract %q: %w", f.Name, err)
	}
	if written > maxZipEntrySize {
		_ = out.Close()
		_ = os.Remove(destPath)
		return fmt.Errorf("zip entry %q exceeds max size (%d bytes)", f.Name, written)
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("close file %q: %w", destPath, err)
	}
	return nil
}

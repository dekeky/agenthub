package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeRelativePath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "readme.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SafeRelativePath(root, "../etc/passwd"); err == nil {
		t.Fatal("expected traversal to fail")
	}
	got, err := SafeRelativePath(root, "readme.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got != "readme.txt" {
		t.Fatalf("got %q", got)
	}
}

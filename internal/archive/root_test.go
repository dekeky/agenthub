package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStripSingleRootDir(t *testing.T) {
	root := t.TempDir()

	t.Run("no strip for flat files", func(t *testing.T) {
		dir := filepath.Join(root, "flat")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := StripSingleRootDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != dir {
			t.Fatalf("got %q want %q", got, dir)
		}
	})

	t.Run("strip single wrapper dir", func(t *testing.T) {
		dir := filepath.Join(root, "wrapped")
		wrapper := filepath.Join(dir, "伯克希尔金融集团")
		if err := os.MkdirAll(filepath.Join(wrapper, "workspace"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(wrapper, "workspace", "AGENT.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := StripSingleRootDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != wrapper {
			t.Fatalf("got %q want %q", got, wrapper)
		}
	})

	t.Run("ignore macOS metadata", func(t *testing.T) {
		dir := filepath.Join(root, "mac")
		wrapper := filepath.Join(dir, "pkg")
		if err := os.MkdirAll(filepath.Join(wrapper, "data"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, "__MACOSX"), 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := StripSingleRootDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != wrapper {
			t.Fatalf("got %q want %q", got, wrapper)
		}
	})

	t.Run("no strip when multiple top-level dirs", func(t *testing.T) {
		dir := filepath.Join(root, "multi")
		if err := os.MkdirAll(filepath.Join(dir, "a"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, "b"), 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := StripSingleRootDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if got != dir {
			t.Fatalf("got %q want %q", got, dir)
		}
	})
}

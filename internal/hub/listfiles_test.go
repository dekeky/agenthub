package hub

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFilesExactVersionContents(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}

	agentRoot := filepath.Join(root, agentsDirName, "exact")
	ver := filepath.Join(agentRoot, "versions", "1.0.0")
	if err := os.MkdirAll(agentRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentRoot, metaFileName), []byte(`{"agentName":"exact","latestVersion":"1.0.0","versions":["1.0.0"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ver, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ver, "empty-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ver, "README.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ver, "scripts", "run.sh"), []byte("sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ver, "meta.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := store.listFiles("exact", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	var files, dirs int
	paths := make(map[string]bool)
	for _, e := range entries {
		if paths[e.Path] {
			t.Fatalf("duplicate path %q", e.Path)
		}
		paths[e.Path] = true
		if e.Dir {
			dirs++
		} else {
			files++
		}
	}

	if files != 2 {
		t.Fatalf("want 2 files (README.md, scripts/run.sh), got %d: %+v", files, entries)
	}
	if dirs != 1 {
		t.Fatalf("want 1 empty dir (empty-dir), got %d: %+v", dirs, entries)
	}
	if paths["meta.json"] {
		t.Fatal("meta.json must be excluded")
	}
	if paths["scripts"] {
		t.Fatal("non-empty dir scripts must not be a separate entry")
	}
	if !paths["empty-dir"] {
		t.Fatal("empty-dir must be listed")
	}
}

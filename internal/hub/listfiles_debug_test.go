package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestListFilesRealStorageAgents(t *testing.T) {
	root := filepath.Join("..", "..", "storage")
	if _, err := os.Stat(root); err != nil {
		t.Skip("storage not present")
	}
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"agent-upload", "test", "demo-weather"} {
		d, err := store.GetAgent(name, "")
		if err != nil {
			t.Logf("%s: %v", name, err)
			continue
		}
		files, dirs := 0, 0
		for _, f := range d.Files {
			if f.Dir {
				dirs++
			} else {
				files++
			}
		}
		t.Logf("%s: %d files, %d dirs", name, files, dirs)
		if files == 0 && dirs > 0 {
			t.Errorf("%s: dirs only without files — sample: %v", name, sampleEntries(d.Files, 5))
		}
		if name == "agent-upload" && files == 0 && dirs == 0 {
			t.Logf("%s: empty listing (re-upload package to restore files)", name)
		}
	}
}

func sampleEntries(entries []FileEntry, n int) []string {
	out := make([]string, 0, n)
	for _, e := range entries {
		if len(out) >= n {
			break
		}
		out = append(out, fmt.Sprintf("%q dir=%v", e.Path, e.Dir))
	}
	return out
}

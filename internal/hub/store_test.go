package hub



import (

	"archive/zip"

	"bytes"

	"os"

	"path/filepath"

	"testing"

)



func TestStoreSaveAndList(t *testing.T) {

	root := t.TempDir()

	store, err := NewStore(root)

	if err != nil {

		t.Fatal(err)

	}



	pkg := filepath.Join(t.TempDir(), "pkg")

	if err := os.MkdirAll(filepath.Join(pkg, "scripts"), 0o755); err != nil {

		t.Fatal(err)

	}

	if err := os.MkdirAll(filepath.Join(pkg, "empty-dir"), 0o755); err != nil {

		t.Fatal(err)

	}

	if err := os.WriteFile(filepath.Join(pkg, "README.md"), []byte("# Weather\n"), 0o644); err != nil {

		t.Fatal(err)

	}

	if err := os.WriteFile(filepath.Join(pkg, "scripts", "run.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {

		t.Fatal(err)

	}



	meta, err := store.SavePackage("weather", "1.0.0", CategoryPicoClaw, pkg)

	if err != nil {

		t.Fatalf("SavePackage: %v", err)

	}

	if meta.AgentName != "weather" || meta.LatestVersion != "1.0.0" {

		t.Fatalf("unexpected meta: %+v", meta)

	}



	list, err := store.ListAgents()

	if err != nil {

		t.Fatal(err)

	}

	if len(list) != 1 || list[0].AgentName != "weather" {

		t.Fatalf("unexpected list: %+v", list)

	}



	detail, err := store.GetAgent("weather", "")

	if err != nil {

		t.Fatal(err)

	}

	if len(detail.Files) < 3 {

		t.Fatalf("expected files and dirs, got %+v", detail.Files)

	}

	var hasEmptyDir, hasReadme, hasRunSh bool

	for _, f := range detail.Files {

		switch f.Path {

		case "scripts":

			t.Fatalf("non-empty dir scripts must not be listed separately, got %+v", detail.Files)

		case "empty-dir":

			if !f.Dir {

				t.Fatalf("empty-dir should be a directory entry, got %+v", f)

			}

			hasEmptyDir = true

		case "README.md":

			if f.Dir {

				t.Fatalf("README.md should be a file entry, got %+v", f)

			}

			hasReadme = true

		case "scripts/run.sh":

			if f.Dir {

				t.Fatalf("scripts/run.sh should be a file entry, got %+v", f)

			}

			hasRunSh = true

		}

	}

	if !hasEmptyDir || !hasReadme || !hasRunSh {

		t.Fatalf("missing expected entries, got %+v", detail.Files)

	}



	zipPath, cleanup, err := store.OpenVersionZip("weather", "1.0.0")

	if err != nil {

		t.Fatal(err)

	}

	defer cleanup()

	if _, err := os.Stat(zipPath); err != nil {

		t.Fatal(err)

	}

}



func TestServiceUploadZip(t *testing.T) {

	root := t.TempDir()

	store, err := NewStore(root)

	if err != nil {

		t.Fatal(err)

	}

	svc := NewService(store)



	var buf bytes.Buffer

	zw := zip.NewWriter(&buf)

	w, err := zw.Create("data/config.json")

	if err != nil {

		t.Fatal(err)

	}

	if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {

		t.Fatal(err)

	}

	if err := zw.Close(); err != nil {

		t.Fatal(err)

	}



	extract := filepath.Join(t.TempDir(), "extract")

	if err := os.MkdirAll(filepath.Join(extract, "data"), 0o755); err != nil {

		t.Fatal(err)

	}

	if err := os.WriteFile(filepath.Join(extract, "data", "config.json"), []byte(`{"ok":true}`), 0o644); err != nil {

		t.Fatal(err)

	}

	meta, err := store.SavePackage("zip-agent", "", CategoryPicoClaw, extract)

	if err != nil {

		t.Fatal(err)

	}

	if meta.AgentName != "zip-agent" {

		t.Fatalf("unexpected agentName: %s", meta.AgentName)

	}

	_ = svc

}



func TestSavePackageAnyDirectory(t *testing.T) {

	root := t.TempDir()

	store, err := NewStore(root)

	if err != nil {

		t.Fatal(err)

	}



	src := filepath.Join(t.TempDir(), "src")

	if err := os.MkdirAll(src, 0o755); err != nil {

		t.Fatal(err)

	}

	if err := os.WriteFile(filepath.Join(src, "note.txt"), []byte("hello"), 0o644); err != nil {

		t.Fatal(err)

	}



	meta, err := store.SavePackage("plain-dir", "1.0.0", CategoryOpenClaw, src)

	if err != nil {

		t.Fatalf("SavePackage: %v", err)

	}

	if meta.DisplayName != "plain-dir" {

		t.Fatalf("unexpected displayName: %s", meta.DisplayName)

	}



	content, err := store.GetFile("plain-dir", "1.0.0", "note.txt")

	if err != nil {

		t.Fatal(err)

	}

	if string(content) != "hello" {

		t.Fatalf("unexpected content: %q", content)

	}

}


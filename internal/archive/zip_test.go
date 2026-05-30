package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZipFileWithBackslashPaths(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "pkg.zip")
	target := filepath.Join(tmp, "out")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	write := func(name, content string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatal(err)
		}
	}
	write(`root\workspace\SKILL.md`, "# skill")
	write(`root\workspace\scripts\`, "")
	write(`root\workspace\scripts\run.sh`, "#!/bin/sh")
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := ExtractZipFile(zipPath, target); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		"root/workspace/SKILL.md",
		"root/workspace/scripts/run.sh",
	} {
		p := filepath.Join(target, filepath.FromSlash(rel))
		if _, err := os.Stat(p); err != nil {
			_ = filepath.WalkDir(target, func(path string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					t.Logf("extracted file: %s", path)
				}
				return nil
			})
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
}

func TestExtractZipFileSingleFile(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "pkg.zip")
	target := filepath.Join(tmp, "out")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(w, "hi"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := ExtractZipFile(zipPath, target); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(target, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hi" {
		t.Fatalf("got %q", data)
	}
}

func TestExtractZipFileForwardSlashPaths(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "pkg.zip")
	target := filepath.Join(tmp, "out")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("workspace/AGENT.md")
	_, _ = io.WriteString(w, "# agent")
	_ = zw.Close()
	_ = f.Close()

	if err := ExtractZipFile(zipPath, target); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "workspace", "AGENT.md")); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeZipEntry(t *testing.T) {
	rel, isDir, err := normalizeZipEntry(`.\root\workspace\`, true, 0)
	if err != nil || !isDir || rel != "root/workspace" {
		t.Fatalf("got %q dir=%v err=%v", rel, isDir, err)
	}
}

func TestExtractZipFileMislabeledFileAsDir(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "pkg.zip")
	target := filepath.Join(tmp, "out")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	hdr := &zip.FileHeader{
		Name:   "wrapper/workspace/SKILL.md",
		Method: zip.Store,
	}
	hdr.SetMode(0o755 | os.ModeDir)
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(w, "# skill"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := ExtractZipFile(zipPath, target); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(target, "wrapper", "workspace", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# skill" {
		t.Fatalf("got %q", data)
	}
}

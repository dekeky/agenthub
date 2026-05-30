package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServerDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agenthub-server.toml")
	content := `
addr = ":9090"
storage_dir = "data"
upload_token = "secret"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadServer(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Addr != ":9090" {
		t.Fatalf("addr = %q", s.Addr)
	}
	wantStorage := filepath.Join(dir, "data")
	if s.StorageDir != wantStorage {
		t.Fatalf("storage_dir = %q, want %q", s.StorageDir, wantStorage)
	}
	if s.UploadToken != "secret" {
		t.Fatalf("upload_token = %q", s.UploadToken)
	}
}

func TestLoadCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agenthub-cli.toml")
	content := `
url = "http://127.0.0.1:8080"
upload_token = "cli-token"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := LoadCLI(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.URL != "http://127.0.0.1:8080" {
		t.Fatalf("url = %q", c.URL)
	}
	if c.UploadToken != "cli-token" {
		t.Fatalf("upload_token = %q", c.UploadToken)
	}
}

func TestResolvePathMissing(t *testing.T) {
	dir := t.TempDir()
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	_, err := resolvePath("", DefaultServerFilename, "AGENTHUB_SERVER_CONFIG")
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

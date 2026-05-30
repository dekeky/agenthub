package main

import (
	"os"
	"path/filepath"
	"testing"
)

const testConfig = `
url = "http://localhost:8080"
upload_token = "test-token"
`

func writeTestConfig(t *testing.T) (cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "agenthub-cli.toml")
	if err := os.WriteFile(path, []byte(testConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() { _ = os.Chdir(oldWD) }
}

func TestRunHelp(t *testing.T) {
	if code := run([]string{"help"}); code != 0 {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	cleanup := writeTestConfig(t)
	defer cleanup()

	if code := run([]string{"nope"}); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

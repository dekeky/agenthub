package hub

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePackagePicoClaw(t *testing.T) {
	root := t.TempDir()
	if err := ValidatePackage(CategoryPicoClaw, root); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePackageOpenClaw(t *testing.T) {
	root := t.TempDir()
	if err := ValidatePackage(CategoryOpenClaw, root); err == nil {
		t.Fatal("expected error without AGENT.md")
	}
	if err := os.WriteFile(filepath.Join(root, "AGENT.md"), []byte("# agent"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePackage(CategoryOpenClaw, root); err != nil {
		t.Fatal(err)
	}
}

func TestServiceListFilterCategory(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(store)

	pico := filepath.Join(t.TempDir(), "pico")
	open := filepath.Join(t.TempDir(), "open")
	for _, spec := range []struct {
		dir, name, cat, marker string
	}{
		{pico, "pico-agent", CategoryPicoClaw, "README.md"},
		{open, "open-agent", CategoryOpenClaw, "AGENT.md"},
	} {
		if err := os.MkdirAll(spec.dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(spec.dir, spec.marker), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := store.SavePackage(spec.name, "1.0.0", spec.cat, spec.dir); err != nil {
			t.Fatal(err)
		}
	}

	all, err := svc.List("")
	if err != nil || len(all) != 2 {
		t.Fatalf("list all: %d agents, err=%v", len(all), err)
	}
	picoOnly, err := svc.List(CategoryPicoClaw)
	if err != nil || len(picoOnly) != 1 || picoOnly[0].AgentName != "pico-agent" {
		t.Fatalf("pico filter: %+v, err=%v", picoOnly, err)
	}
	if _, err := svc.List("../bad"); err == nil {
		t.Fatal("expected invalid category error")
	}
}

func TestIsInvalidCategory(t *testing.T) {
	_, err := NormalizeCategory("../x")
	if !IsInvalidCategory(err) {
		t.Fatalf("expected invalid category, got %v", err)
	}
}

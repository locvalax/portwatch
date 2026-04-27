package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "profiles-*.yaml")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

func TestLoadFile_ParsesProfiles(t *testing.T) {
	path := writeTemp(t, `
profiles:
  - name: web
    hosts: ["192.168.1.1"]
    ports: "80,443"
    interval: 1m
  - name: db
    hosts: ["192.168.1.2"]
    ports: "5432"
`)
	r := New()
	if err := LoadFile(path, r); err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if _, ok := r.Get("web"); !ok {
		t.Error("expected profile 'web'")
	}
	if _, ok := r.Get("db"); !ok {
		t.Error("expected profile 'db'")
	}
}

func TestLoadFile_MissingFile(t *testing.T) {
	r := New()
	err := LoadFile(filepath.Join(t.TempDir(), "no-such.yaml"), r)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFile_InvalidYAML(t *testing.T) {
	path := writeTemp(t, `:::invalid yaml:::`)
	r := New()
	if err := LoadFile(path, r); err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadFile_EmptyProfiles(t *testing.T) {
	path := writeTemp(t, `profiles: []`)
	r := New()
	if err := LoadFile(path, r); err == nil {
		t.Fatal("expected error for empty profiles list")
	}
}

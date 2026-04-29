package remap_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/remap"
)

func TestLookup_WellKnownPort(t *testing.T) {
	r := remap.New()
	if got := r.Lookup(22); got != "ssh" {
		t.Fatalf("expected ssh, got %s", got)
	}
}

func TestLookup_UnknownPort_ReturnsFallback(t *testing.T) {
	r := remap.New()
	if got := r.Lookup(9999); got != "9999" {
		t.Fatalf("expected \"9999\", got %s", got)
	}
}

func TestRegister_AddsEntry(t *testing.T) {
	r := remap.New()
	err := r.Register(remap.Entry{Port: 9000, ServiceName: "custom", Description: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := r.Lookup(9000); got != "custom" {
		t.Fatalf("expected custom, got %s", got)
	}
}

func TestRegister_InvalidPort_ReturnsError(t *testing.T) {
	r := remap.New()
	if err := r.Register(remap.Entry{Port: 0, ServiceName: "bad"}); err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestRegister_EmptyName_ReturnsError(t *testing.T) {
	r := remap.New()
	if err := r.Register(remap.Entry{Port: 80, ServiceName: ""}); err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	r := remap.New()
	r.Delete(22)
	if got := r.Lookup(22); got != "22" {
		t.Fatalf("expected fallback \"22\", got %s", got)
	}
}

func TestLookupEntry_Found(t *testing.T) {
	r := remap.New()
	e, ok := r.LookupEntry(443)
	if !ok {
		t.Fatal("expected entry for port 443")
	}
	if e.ServiceName != "https" {
		t.Fatalf("expected https, got %s", e.ServiceName)
	}
}

func TestLoadFile_ParsesMappings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mappings.yaml")
	content := "mappings:\n  - port: 8443\n    service: https-alt\n    description: HTTPS alternate\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	r := remap.New()
	if err := remap.LoadFile(path, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := r.Lookup(8443); got != "https-alt" {
		t.Fatalf("expected https-alt, got %s", got)
	}
}

func TestLoadFile_MissingFile_ReturnsError(t *testing.T) {
	r := remap.New()
	if err := remap.LoadFile("/nonexistent/path.yaml", r); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFile_InvalidYAML_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("::not yaml::"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := remap.New()
	if err := remap.LoadFile(path, r); err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

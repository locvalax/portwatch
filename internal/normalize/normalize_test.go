package normalize_test

import (
	"testing"

	"github.com/user/portwatch/internal/normalize"
)

func TestDefaultOptions(t *testing.T) {
	opts := normalize.DefaultOptions()
	if !opts.Lowercase {
		t.Error("expected Lowercase to be true by default")
	}
	if opts.StripPort {
		t.Error("expected StripPort to be false by default")
	}
}

func TestHost_Lowercase(t *testing.T) {
	n := normalize.New(normalize.DefaultOptions())
	got, err := n.Host("EXAMPLE.COM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "example.com" {
		t.Errorf("got %q, want %q", got, "example.com")
	}
}

func TestHost_IPv6_BracketsStripped(t *testing.T) {
	n := normalize.New(normalize.DefaultOptions())
	got, err := n.Host("[::1]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "::1" {
		t.Errorf("got %q, want %q", got, "::1")
	}
}

func TestHost_StripPort(t *testing.T) {
	opts := normalize.DefaultOptions()
	opts.StripPort = true
	n := normalize.New(opts)

	got, err := n.Host("example.com:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "example.com" {
		t.Errorf("got %q, want %q", got, "example.com")
	}
}

func TestHost_StripPort_IPv6WithPort(t *testing.T) {
	opts := normalize.DefaultOptions()
	opts.StripPort = true
	n := normalize.New(opts)

	got, err := n.Host("[::1]:9000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "::1" {
		t.Errorf("got %q, want %q", got, "::1")
	}
}

func TestHost_NoPort_StripPortNoOp(t *testing.T) {
	opts := normalize.DefaultOptions()
	opts.StripPort = true
	n := normalize.New(opts)

	got, err := n.Host("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "example.com" {
		t.Errorf("got %q, want %q", got, "example.com")
	}
}

func TestHost_EmptyString_ReturnsError(t *testing.T) {
	n := normalize.New(normalize.DefaultOptions())
	_, err := n.Host("")
	if err == nil {
		t.Error("expected error for empty host, got nil")
	}
}

func TestHosts_NormalisesAll(t *testing.T) {
	n := normalize.New(normalize.DefaultOptions())
	raw := []string{"HOST-A", "HOST-B", "HOST-C"}
	got, err := n.Hosts(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"host-a", "host-b", "host-c"}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("index %d: got %q, want %q", i, g, want[i])
		}
	}
}

func TestHosts_EmptyEntry_ReturnsError(t *testing.T) {
	n := normalize.New(normalize.DefaultOptions())
	_, err := n.Hosts([]string{"valid.host", ""})
	if err == nil {
		t.Error("expected error for slice containing empty host, got nil")
	}
}

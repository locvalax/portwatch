package redact_test

import (
	"strings"
	"testing"

	"github.com/sgreben/portwatch/internal/redact"
)

func TestHost_HashMode_IsStable(t *testing.T) {
	r := redact.New(redact.DefaultOptions())
	a := r.Host("192.168.1.1")
	b := r.Host("192.168.1.1")
	if a != b {
		t.Fatalf("expected stable hash, got %q and %q", a, b)
	}
}

func TestHost_HashMode_DifferentHosts_DifferentValues(t *testing.T) {
	r := redact.New(redact.DefaultOptions())
	a := r.Host("192.168.1.1")
	b := r.Host("10.0.0.1")
	if a == b {
		t.Fatal("expected different hashes for different hosts")
	}
}

func TestHost_HashMode_HasPrefix(t *testing.T) {
	r := redact.New(redact.DefaultOptions())
	v := r.Host("example.com")
	if !strings.HasPrefix(v, "host-") {
		t.Fatalf("expected host- prefix, got %q", v)
	}
}

func TestHost_HashMode_LengthRespected(t *testing.T) {
	opts := redact.DefaultOptions()
	opts.Length = 4
	r := redact.New(opts)
	v := r.Host("example.com")
	// "host-" (5) + 4 hex chars = 9
	if len(v) != 9 {
		t.Fatalf("expected length 9, got %d (%q)", len(v), v)
	}
}

func TestHost_MaskMode_IP(t *testing.T) {
	opts := redact.DefaultOptions()
	opts.Mode = redact.ModeMask
	r := redact.New(opts)
	v := r.Host("10.0.0.5")
	if v != "<ip-redacted>" {
		t.Fatalf("expected <ip-redacted>, got %q", v)
	}
}

func TestHost_MaskMode_Hostname(t *testing.T) {
	opts := redact.DefaultOptions()
	opts.Mode = redact.ModeMask
	r := redact.New(opts)
	v := r.Host("db.internal")
	if v != "<redacted>" {
		t.Fatalf("expected <redacted>, got %q", v)
	}
}

func TestFlush_ClearsCache(t *testing.T) {
	opts := redact.DefaultOptions()
	opts.Salt = "before"
	r := redact.New(opts)
	a := r.Host("host.example")
	r.Flush()
	opts.Salt = "after"
	r2 := redact.New(opts)
	b := r2.Host("host.example")
	if a == b {
		t.Fatal("expected different values after salt change")
	}
}

func TestDefaultOptions_HashLength(t *testing.T) {
	opts := redact.DefaultOptions()
	if opts.Length != 8 {
		t.Fatalf("expected default length 8, got %d", opts.Length)
	}
}

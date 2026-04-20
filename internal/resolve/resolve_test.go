package resolve_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/resolve"
)

func TestNew_DefaultTimeout(t *testing.T) {
	r := resolve.New(0)
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
}

func TestResolve_Localhost(t *testing.T) {
	r := resolve.New(5 * time.Second)
	res, err := r.Resolve("localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Host != "localhost" {
		t.Errorf("expected host localhost, got %q", res.Host)
	}
	if len(res.Addresses) == 0 {
		t.Error("expected at least one address")
	}
	if res.ResolvedAt.IsZero() {
		t.Error("expected non-zero ResolvedAt")
	}
}

func TestResolve_InvalidHost(t *testing.T) {
	r := resolve.New(3 * time.Second)
	_, err := r.Resolve("this.host.does.not.exist.invalid")
	if err == nil {
		t.Fatal("expected error for invalid host")
	}
}

func TestResolveAll_MixedHosts(t *testing.T) {
	r := resolve.New(5 * time.Second)
	hosts := []string{"localhost", "this.host.does.not.exist.invalid"}
	results, errs := r.ResolveAll(hosts)

	if _, ok := results["localhost"]; !ok {
		t.Error("expected result for localhost")
	}
	if _, ok := errs["this.host.does.not.exist.invalid"]; !ok {
		t.Error("expected error for invalid host")
	}
}

func TestResolveAll_Empty(t *testing.T) {
	r := resolve.New(5 * time.Second)
	results, errs := r.ResolveAll([]string{})
	if len(results) != 0 || len(errs) != 0 {
		t.Error("expected empty maps for empty input")
	}
}

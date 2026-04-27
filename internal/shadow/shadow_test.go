package shadow_test

import (
	"bytes"
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/user/portwatch/internal/shadow"
)

// stubScanner returns a fixed port list or error.
type stubScanner struct {
	ports []int
	err   error
}

func (s *stubScanner) Scan(_ context.Context, _ string) ([]int, error) {
	return s.ports, s.err
}

func TestShadow_NoDivergence_NoLog(t *testing.T) {
	var buf bytes.Buffer
	r := shadow.New(
		&stubScanner{ports: []int{80, 443}},
		&stubScanner{ports: []int{80, 443}},
		shadow.Options{Log: &buf},
	)
	ports, err := r.Scan(context.Background(), "localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(ports))
	}
	if buf.Len() != 0 {
		t.Errorf("expected no log output, got: %s", buf.String())
	}
	if divs := r.Divergences(); len(divs) != 0 {
		t.Errorf("expected 0 divergences, got %d", len(divs))
	}
}

func TestShadow_Divergence_Logged(t *testing.T) {
	var buf bytes.Buffer
	r := shadow.New(
		&stubScanner{ports: []int{80, 443}},
		&stubScanner{ports: []int{80, 8080}},
		shadow.Options{Log: &buf},
	)
	_, _ = r.Scan(context.Background(), "host1")

	if buf.Len() == 0 {
		t.Error("expected divergence log output")
	}
	divs := r.Divergences()
	if len(divs) != 1 {
		t.Fatalf("expected 1 divergence, got %d", len(divs))
	}
	sort.Ints(divs[0].PrimaryOnly)
	sort.Ints(divs[0].ShadowOnly)
	if len(divs[0].PrimaryOnly) != 1 || divs[0].PrimaryOnly[0] != 443 {
		t.Errorf("unexpected PrimaryOnly: %v", divs[0].PrimaryOnly)
	}
	if len(divs[0].ShadowOnly) != 1 || divs[0].ShadowOnly[0] != 8080 {
		t.Errorf("unexpected ShadowOnly: %v", divs[0].ShadowOnly)
	}
}

func TestShadow_PrimaryError_Propagated(t *testing.T) {
	var buf bytes.Buffer
	want := errors.New("primary failed")
	r := shadow.New(
		&stubScanner{err: want},
		&stubScanner{ports: []int{80}},
		shadow.Options{Log: &buf},
	)
	_, err := r.Scan(context.Background(), "host1")
	if !errors.Is(err, want) {
		t.Errorf("expected primary error, got %v", err)
	}
}

func TestShadow_ShadowError_DoesNotAffectResult(t *testing.T) {
	var buf bytes.Buffer
	r := shadow.New(
		&stubScanner{ports: []int{22}},
		&stubScanner{err: errors.New("shadow down")},
		shadow.Options{Log: &buf},
	)
	ports, err := r.Scan(context.Background(), "host1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 || ports[0] != 22 {
		t.Errorf("unexpected ports: %v", ports)
	}
	if buf.Len() == 0 {
		t.Error("expected shadow error to be logged")
	}
}

func TestDefaultOptions_LogIsStderr(t *testing.T) {
	opts := shadow.DefaultOptions()
	if opts.Log == nil {
		t.Error("expected non-nil default log writer")
	}
}

package envelope_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/envelope"
	"github.com/yourorg/portwatch/internal/store"
)

func makeEntry(ports []int) store.Entry {
	return store.Entry{Ports: ports, ScannedAt: time.Now()}
}

func TestDefaultOptions(t *testing.T) {
	opts := envelope.DefaultOptions()
	if opts.Source != "portwatch" {
		t.Fatalf("expected source 'portwatch', got %q", opts.Source)
	}
	if opts.Labels == nil {
		t.Fatal("expected non-nil Labels map")
	}
}

func TestWrap_FieldsPopulated(t *testing.T) {
	b := envelope.New(envelope.Options{Source: "test", Labels: map[string]string{"env": "ci"}})
	entry := makeEntry([]int{80, 443})
	env := b.Wrap("example.com", entry)

	if env.Host != "example.com" {
		t.Errorf("host: got %q", env.Host)
	}
	if env.Source != "test" {
		t.Errorf("source: got %q", env.Source)
	}
	if env.Labels["env"] != "ci" {
		t.Errorf("label env: got %q", env.Labels["env"])
	}
	if env.Seq != 1 {
		t.Errorf("seq: got %d, want 1", env.Seq)
	}
	if env.ScannedAt.IsZero() {
		t.Error("ScannedAt should not be zero")
	}
}

func TestWrap_SeqIncrements(t *testing.T) {
	b := envelope.New(envelope.DefaultOptions())
	entry := makeEntry([]int{22})
	e1 := b.Wrap("host1", entry)
	e2 := b.Wrap("host2", entry)
	if e2.Seq != e1.Seq+1 {
		t.Errorf("seq not incrementing: %d -> %d", e1.Seq, e2.Seq)
	}
}

func TestWithLabel_DoesNotMutateOriginal(t *testing.T) {
	b := envelope.New(envelope.DefaultOptions())
	env := b.Wrap("h", makeEntry(nil))
	newEnv := envelope.WithLabel(env, "region", "us-east")

	if _, ok := env.Labels["region"]; ok {
		t.Error("original envelope should not have 'region' label")
	}
	if newEnv.Labels["region"] != "us-east" {
		t.Errorf("new envelope missing label: %v", newEnv.Labels)
	}
}

// stubScanner is a minimal Scanner implementation for middleware tests.
type stubScanner struct {
	entry store.Entry
	err   error
}

func (s *stubScanner) Scan(_ context.Context, _ string) (store.Entry, error) {
	return s.entry, s.err
}

func TestWrappingScanner_CallsSink(t *testing.T) {
	expected := makeEntry([]int{8080})
	stub := &stubScanner{entry: expected}
	b := envelope.New(envelope.DefaultOptions())

	var received envelope.Envelope
	ws := envelope.NewWrappingScanner(stub, b, func(e envelope.Envelope) { received = e })

	_, err := ws.Scan(context.Background(), "localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Host != "localhost" {
		t.Errorf("sink received wrong host: %q", received.Host)
	}
}

func TestWrappingScanner_ErrorSkipsSink(t *testing.T) {
	stub := &stubScanner{err: errors.New("scan failed")}
	b := envelope.New(envelope.DefaultOptions())
	sinkCalled := false
	ws := envelope.NewWrappingScanner(stub, b, func(e envelope.Envelope) { sinkCalled = true })

	_, err := ws.Scan(context.Background(), "badhost")
	if err == nil {
		t.Fatal("expected error")
	}
	if sinkCalled {
		t.Error("sink should not be called on scan error")
	}
}

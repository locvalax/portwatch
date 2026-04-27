package cache

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

// stubScanner counts how many times Scan is called.
type stubScanner struct {
	calls int
	ports []uint16
	err   error
}

func (s *stubScanner) Scan(_ context.Context, _ string, _ scanner.Options) ([]uint16, error) {
	s.calls++
	return s.ports, s.err
}

func startTCPServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return ln.Addr().String()
}

func TestCachedScanner_HitSkipsInner(t *testing.T) {
	stub := &stubScanner{ports: []uint16{80}}
	c := New(DefaultOptions())
	cs := NewCachedScanner(stub, c)
	ctx := context.Background()
	opts := scanner.DefaultOptions()

	// First call — populates cache.
	if _, err := cs.Scan(ctx, "host1", opts); err != nil {
		t.Fatalf("first scan: %v", err)
	}
	// Second call — should use cache.
	if _, err := cs.Scan(ctx, "host1", opts); err != nil {
		t.Fatalf("second scan: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected inner called once, got %d", stub.calls)
	}
}

func TestCachedScanner_MissCallsInner(t *testing.T) {
	stub := &stubScanner{ports: []uint16{443}}
	c := New(DefaultOptions())
	cs := NewCachedScanner(stub, c)
	ctx := context.Background()
	opts := scanner.DefaultOptions()

	if _, err := cs.Scan(ctx, "host2", opts); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", stub.calls)
	}
}

func TestCachedScanner_InnerError_NotCached(t *testing.T) {
	stub := &stubScanner{err: errors.New("timeout")}
	c := New(DefaultOptions())
	cs := NewCachedScanner(stub, c)
	ctx := context.Background()
	opts := scanner.DefaultOptions()

	if _, err := cs.Scan(ctx, "host3", opts); err == nil {
		t.Fatal("expected error from inner scanner")
	}
	// Entry must not have been cached.
	if _, ok := c.Get("host3"); ok {
		t.Fatal("error result should not be cached")
	}
}

package batch_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/batch"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
)

// stubScanner returns preset entries or errors keyed by host.
type stubScanner struct {
	entries map[string]store.Entry
	errs    map[string]error
}

func (s *stubScanner) Scan(_ context.Context, host string, _ scanner.Options) (store.Entry, error) {
	if err, ok := s.errs[host]; ok {
		return store.Entry{}, err
	}
	if e, ok := s.entries[host]; ok {
		return e, nil
	}
	return store.Entry{Host: host}, nil
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

func TestRun_AllSucceed(t *testing.T) {
	hosts := []string{"h1", "h2", "h3"}
	sc := &stubScanner{
		entries: map[string]store.Entry{
			"h1": {Host: "h1", Ports: []int{80}},
			"h2": {Host: "h2", Ports: []int{443}},
			"h3": {Host: "h3", Ports: []int{22}},
		},
	}

	results := batch.Run(context.Background(), hosts, sc, batch.DefaultOptions())

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d]: unexpected error: %v", i, r.Err)
		}
		if r.Host != hosts[i] {
			t.Errorf("result[%d]: host = %q, want %q", i, r.Host, hosts[i])
		}
	}
}

func TestRun_PartialErrors(t *testing.T) {
	hosts := []string{"ok", "bad"}
	sc := &stubScanner{
		entries: map[string]store.Entry{"ok": {Host: "ok", Ports: []int{80}}},
		errs:    map[string]error{"bad": errors.New("refused")},
	}

	results := batch.Run(context.Background(), hosts, sc, batch.DefaultOptions())

	if results[0].Err != nil {
		t.Errorf("expected no error for 'ok', got %v", results[0].Err)
	}
	if results[1].Err == nil {
		t.Error("expected error for 'bad'")
	}
}

func TestRun_WorkerConcurrency(t *testing.T) {
	const n = 10
	hosts := make([]string, n)
	for i := range hosts {
		hosts[i] = fmt.Sprintf("host%d", i)
	}

	var mu sync.Mutex
	var peak, current int
	sc := batch.ScanFunc(func(_ context.Context, host string, _ scanner.Options) (store.Entry, error) {
		mu.Lock()
		current++
		if current > peak {
			peak = current
		}
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		current--
		mu.Unlock()
		return store.Entry{Host: host}, nil
	})

	opts := batch.Options{Workers: 3, ScanOptions: scanner.DefaultOptions()}
	batch.Run(context.Background(), hosts, sc, opts)

	if peak > 3 {
		t.Errorf("peak concurrency %d exceeded worker limit 3", peak)
	}
}

func TestRun_EmptyHosts(t *testing.T) {
	sc := &stubScanner{}
	results := batch.Run(context.Background(), nil, sc, batch.DefaultOptions())
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := batch.DefaultOptions()
	if opts.Workers != 5 {
		t.Errorf("Workers = %d, want 5", opts.Workers)
	}
}

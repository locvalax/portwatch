package watch_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
	"github.com/user/portwatch/internal/watch"
)

func startTCPServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	return ln.Addr().String()
}

func tempStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	st, err := store.New(filepath.Join(dir, "data.json"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return st
}

func TestWatcher_Once_ScansAndStores(t *testing.T) {
	addr := startTCPServer(t)
	sc := scanner.New(scanner.DefaultOptions())
	st := tempStore(t)
	al := alert.New(alert.Options{Writer: os.Stdout})

	opts := watch.Options{
		Hosts:    []string{addr},
		Interval: time.Second,
		Once:     true,
	}
	w := watch.New(opts, sc, st, al)

	ctx := context.Background()
	if err := w.Run(ctx); err != nil {
		t.Fatalf("Run: %v", err)
	}

	entry, err := st.Latest(addr)
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if len(entry.Ports) == 0 {
		t.Error("expected at least one open port")
	}
}

func TestWatcher_CancelStops(t *testing.T) {
	addr := startTCPServer(t)
	sc := scanner.New(scanner.DefaultOptions())
	st := tempStore(t)
	al := alert.New(alert.Options{Writer: os.Stdout})

	opts := watch.Options{
		Hosts:    []string{addr},
		Interval: 50 * time.Millisecond,
		Once:     false,
	}
	w := watch.New(opts, sc, st, al)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	start := time.Now()
	if err := w.Run(ctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("Run took too long: %v", elapsed)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := watch.DefaultOptions()
	if opts.Interval != 60*time.Second {
		t.Errorf("expected 60s interval, got %v", opts.Interval)
	}
	if opts.Once {
		t.Error("expected Once=false by default")
	}
}

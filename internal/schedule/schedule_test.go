package schedule_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/schedule"
	"github.com/user/portwatch/internal/store"
)

func tempStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := store.New(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return s
}

func TestRunner_ScansAndStores(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	s := tempStore(t)
	a, _ := alert.New(os.Stdout)
	opts := scanner.Options{Ports: []int{port}, Timeout: time.Second}
	r := schedule.New([]string{"127.0.0.1"}, 50*time.Millisecond, s, a, opts)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	r.Run(ctx)

	ports, err := s.Latest("127.0.0.1")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if len(ports) == 0 {
		t.Fatal("expected at least one open port")
	}
}

func TestRunner_CancelStops(t *testing.T) {
	s := tempStore(t)
	a, _ := alert.New(os.Stdout)
	opts := scanner.DefaultOptions()
	r := schedule.New([]string{"127.0.0.1"}, time.Hour, s, a, opts)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}

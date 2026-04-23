package timeout_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/timeout"
)

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

func TestNew_InvalidDeadline_ReturnsError(t *testing.T) {
	_, err := timeout.New(timeout.Options{Deadline: 0})
	if err == nil {
		t.Fatal("expected error for zero deadline")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := timeout.DefaultOptions()
	if opts.Deadline <= 0 {
		t.Fatalf("expected positive deadline, got %s", opts.Deadline)
	}
}

func TestTimedScanner_CompletesWithinDeadline(t *testing.T) {
	g, err := timeout.New(timeout.Options{Deadline: 2 * time.Second})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := timeout.NewTimedScanner(g, scanner.Scan)
	addr := startTCPServer(t)
	host, _, _ := net.SplitHostPort(addr)

	_, err = ts.Scan(context.Background(), host, scanner.DefaultOptions())
	// may or may not find ports; we only care that no timeout fired
	if timeout.IsTimeout(err) {
		t.Fatalf("unexpected timeout: %v", err)
	}
}

func TestTimedScanner_ExceedsDeadline_ReturnsTimeout(t *testing.T) {
	g, err := timeout.New(timeout.Options{Deadline: 1 * time.Millisecond})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	slow := func(ctx context.Context, _ string, _ scanner.Options) ([]int, error) {
		select {
		case <-time.After(500 * time.Millisecond):
			return []int{80}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	ts := timeout.NewTimedScanner(g, slow)
	_, err = ts.Scan(context.Background(), "127.0.0.1", scanner.DefaultOptions())
	if !timeout.IsTimeout(err) {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestIsTimeout_WrapsDeadlineExceeded(t *testing.T) {
	if !timeout.IsTimeout(context.DeadlineExceeded) {
		t.Fatal("expected DeadlineExceeded to be recognised as timeout")
	}
	if timeout.IsTimeout(errors.New("other")) {
		t.Fatal("non-timeout error should not match")
	}
}

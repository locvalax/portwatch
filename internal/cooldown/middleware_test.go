package cooldown_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/cooldown"
	"github.com/user/portwatch/internal/scanner"
)

func startTCPServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start TCP server: %v", err)
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
	return ln.Addr().(*net.TCPAddr).Addr().String()
}

func TestCooledScanner_FirstScan_Succeeds(t *testing.T) {
	host := startTCPServer(t)

	cd := cooldown.New(cooldown.DefaultOptions())
	base := scanner.DefaultOptions()
	guarded := cooldown.NewCooledScanner(cd, base)

	ctx := context.Background()
	_, err := guarded.Scan(ctx, host)
	if err != nil {
		t.Fatalf("expected first scan to succeed, got: %v", err)
	}
}

func TestCooledScanner_SecondScan_WithinWindow_Blocked(t *testing.T) {
	host := startTCPServer(t)

	opts := cooldown.DefaultOptions()
	opts.Window = 5 * time.Second // long window so second call is blocked
	cd := cooldown.New(opts)
	base := scanner.DefaultOptions()
	guarded := cooldown.NewCooledScanner(cd, base)

	ctx := context.Background()
	// First scan — should succeed
	_, err := guarded.Scan(ctx, host)
	if err != nil {
		t.Fatalf("expected first scan to succeed, got: %v", err)
	}

	// Second scan within cooldown window — should be blocked
	_, err = guarded.Scan(ctx, host)
	if err == nil {
		t.Fatal("expected second scan to be blocked by cooldown, got nil error")
	}
}

func TestCooledScanner_AfterWindowExpires_Succeeds(t *testing.T) {
	host := startTCPServer(t)

	opts := cooldown.DefaultOptions()
	opts.Window = 50 * time.Millisecond
	cd := cooldown.New(opts)
	base := scanner.DefaultOptions()
	guarded := cooldown.NewCooledScanner(cd, base)

	ctx := context.Background()
	// First scan
	_, err := guarded.Scan(ctx, host)
	if err != nil {
		t.Fatalf("expected first scan to succeed, got: %v", err)
	}

	// Wait for cooldown to expire
	time.Sleep(100 * time.Millisecond)

	// Second scan after window — should succeed
	_, err = guarded.Scan(ctx, host)
	if err != nil {
		t.Fatalf("expected scan after cooldown to succeed, got: %v", err)
	}
}

func TestCooledScanner_DifferentHosts_Independent(t *testing.T) {
	host1 := startTCPServer(t)
	host2 := startTCPServer(t)

	opts := cooldown.DefaultOptions()
	opts.Window = 5 * time.Second
	cd := cooldown.New(opts)
	base := scanner.DefaultOptions()
	guarded := cooldown.NewCooledScanner(cd, base)

	ctx := context.Background()

	// Scan host1 — succeeds
	_, err := guarded.Scan(ctx, host1)
	if err != nil {
		t.Fatalf("host1 first scan failed: %v", err)
	}

	// Scan host2 — should also succeed (different cooldown bucket)
	_, err = guarded.Scan(ctx, host2)
	if err != nil {
		t.Fatalf("host2 first scan failed: %v", err)
	}

	// Scan host1 again — should be blocked
	_, err = guarded.Scan(ctx, host1)
	if err == nil {
		t.Fatal("expected host1 second scan to be blocked, got nil error")
	}
}

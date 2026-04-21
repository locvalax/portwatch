package metrics

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/user/portwatch/internal/scanner"
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

func TestInstrumented_SuccessRecordsAlerts(t *testing.T) {
	c := New()
	addr := startTCPServer(t)

	is := NewInstrumentedScanner(c, scanner.Scan)
	_, err := is.Scan(context.Background(), addr, scanner.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := c.Snapshot()
	if snap[addr].Scans != 1 {
		t.Fatalf("expected 1 scan, got %d", snap[addr].Scans)
	}
	if snap[addr].Alerts != 1 {
		t.Fatalf("expected 1 alert, got %d", snap[addr].Alerts)
	}
}

func TestInstrumented_ErrorRecordsError(t *testing.T) {
	c := New()
	failing := func(_ context.Context, host string, _ scanner.Options) ([]int, error) {
		return nil, errors.New("scan failed")
	}

	is := NewInstrumentedScanner(c, failing)
	_, err := is.Scan(context.Background(), "bad-host", scanner.DefaultOptions())
	if err == nil {
		t.Fatal("expected error")
	}

	snap := c.Snapshot()
	if snap["bad-host"].Errors != 1 {
		t.Fatalf("expected 1 error, got %d", snap["bad-host"].Errors)
	}
	if snap["bad-host"].Alerts != 0 {
		t.Fatalf("expected 0 alerts, got %d", snap["bad-host"].Alerts)
	}
}

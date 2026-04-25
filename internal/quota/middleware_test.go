package quota

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func startTCPServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
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

func TestGuardedScanner_FirstScan_Succeeds(t *testing.T) {
	addr := startTCPServer(t)
	host, portStr, _ := net.SplitHostPort(addr)
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	opts := Options{MaxScans: 5, Window: time.Hour}
	l := New(opts)

	base := scanner.DefaultOptions()
	inner, _ := scanner.New(base)
	gs := NewGuardedScanner(inner, l)

	_, err := gs.Scan(context.Background(), host, []int{port})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestGuardedScanner_QuotaExceeded_ReturnsError(t *testing.T) {
	addr := startTCPServer(t)
	host, portStr, _ := net.SplitHostPort(addr)
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	opts := Options{MaxScans: 1, Window: time.Hour}
	l := New(opts)

	base := scanner.DefaultOptions()
	inner, _ := scanner.New(base)
	gs := NewGuardedScanner(inner, l)

	gs.Scan(context.Background(), host, []int{port})
	_, err := gs.Scan(context.Background(), host, []int{port})
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

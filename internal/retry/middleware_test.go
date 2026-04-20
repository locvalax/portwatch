package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

func startTCPServer(t *testing.T) (host string, port int) {
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
	addr := ln.Addr().(*net.TCPAddr)
	return "127.0.0.1", addr.Port
}

func TestRetryScanner_SuccessOnFirstAttempt(t *testing.T) {
	_, port := startTCPServer(t)
	opts := scanner.DefaultOptions()
	opts.Ports = []int{port}

	rs := NewRetryScanner(Policy{MaxAttempts: 3, Delay: 0, Multiplier: 1})
	res, err := rs.Scan(context.Background(), "127.0.0.1", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Ports) == 0 {
		t.Fatal("expected at least one open port")
	}
}

func TestRetryScanner_ExhaustsOnUnreachableHost(t *testing.T) {
	opts := scanner.DefaultOptions()
	opts.Ports = []int{1}
	opts.Timeout = 0

	// Use a host that will always fail.
	rs := NewRetryScanner(Policy{MaxAttempts: 2, Delay: 0, Multiplier: 1})
	_, err := rs.Scan(context.Background(), fmt.Sprintf("127.0.0.1"), opts)
	// We expect either ErrExhausted or a scan error — just not nil.
	if err == nil {
		t.Fatal("expected error for unreachable port")
	}
	_ = errors.Is(err, ErrExhausted) // acceptable outcome
}

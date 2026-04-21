package healthcheck_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/healthcheck"
)

func startTCPServer(t *testing.T) (string, int) {
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
	return addr.IP.String(), addr.Port
}

func TestProbe_ReachableHost(t *testing.T) {
	host, port := startTCPServer(t)
	opts := healthcheck.Options{Timeout: 2 * time.Second, ProbePort: port}
	c := healthcheck.New(opts)

	status := c.Probe(context.Background(), host)
	if !status.Reachable {
		t.Fatalf("expected reachable, got err: %v", status.Err)
	}
	if status.Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestProbe_UnreachableHost(t *testing.T) {
	opts := healthcheck.Options{Timeout: 200 * time.Millisecond, ProbePort: 1}
	c := healthcheck.New(opts)

	status := c.Probe(context.Background(), "192.0.2.1") // TEST-NET, unroutable
	if status.Reachable {
		t.Fatal("expected unreachable")
	}
	if status.Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestProbeAll_MixedHosts(t *testing.T) {
	host, port := startTCPServer(t)
	opts := healthcheck.Options{Timeout: 500 * time.Millisecond, ProbePort: port}
	c := healthcheck.New(opts)

	statuses := c.ProbeAll(context.Background(), []string{host, "192.0.2.1"})
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if !statuses[0].Reachable {
		t.Errorf("expected first host reachable")
	}
	if statuses[1].Reachable {
		t.Errorf("expected second host unreachable")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := healthcheck.DefaultOptions()
	if opts.Timeout <= 0 {
		t.Error("expected positive default timeout")
	}
	if opts.ProbePort <= 0 {
		t.Error("expected positive default probe port")
	}
}

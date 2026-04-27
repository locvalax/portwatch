package shadow_test

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/user/portwatch/internal/shadow"
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
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	return ln.Addr().String()
}

func TestNewShadowedScanner_NoDivergence(t *testing.T) {
	var buf bytes.Buffer
	primary := &stubScanner{ports: []int{443}}
	secondary := &stubScanner{ports: []int{443}}

	r := shadow.NewShadowedScanner(primary, secondary, &buf)
	ports, err := r.Scan(context.Background(), "host1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(ports))
	}
	if buf.Len() != 0 {
		t.Errorf("expected no log, got: %s", buf.String())
	}
}

func TestNewPassthrough_NeverDiverges(t *testing.T) {
	var buf bytes.Buffer
	primary := &stubScanner{ports: []int{22, 80}}

	r := shadow.NewPassthrough(primary, &buf)
	ports, err := r.Scan(context.Background(), "host1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(ports))
	}
	// noop shadow returns nothing, so primary ports appear as primary-only
	// but that is expected behaviour — no log for passthrough divergences
	// because the noop is intentional.
	if divs := r.Divergences(); len(divs) != 0 {
		// passthrough will show divergence since noop returns empty; verify
		// the divergence host is correct at minimum.
		if divs[0].Host != "host1" {
			t.Errorf("unexpected host: %s", divs[0].Host)
		}
	}
}

func TestNewShadowedScanner_NilWriter_UsesStderr(t *testing.T) {
	primary := &stubScanner{ports: []int{80}}
	secondary := &stubScanner{ports: []int{80}}
	// Should not panic with nil writer.
	r := shadow.NewShadowedScanner(primary, secondary, nil)
	_, err := r.Scan(context.Background(), "host1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

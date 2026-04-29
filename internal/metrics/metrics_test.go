package metrics

import (
	"testing"
	"time"
)

func TestRecordScan_IncrementsCount(t *testing.T) {
	c := New()
	c.RecordScan("host-a")
	c.RecordScan("host-a")
	snap := c.Snapshot()
	if snap["host-a"].Scans != 2 {
		t.Fatalf("expected 2 scans, got %d", snap["host-a"].Scans)
	}
}

func TestRecordAlert_IncrementsCount(t *testing.T) {
	c := New()
	c.RecordScan("host-b")
	c.RecordAlert("host-b")
	snap := c.Snapshot()
	if snap["host-b"].Alerts != 1 {
		t.Fatalf("expected 1 alert, got %d", snap["host-b"].Alerts)
	}
}

func TestRecordError_IncrementsCount(t *testing.T) {
	c := New()
	c.RecordScan("host-c")
	c.RecordError("host-c")
	c.RecordError("host-c")
	snap := c.Snapshot()
	if snap["host-c"].Errors != 2 {
		t.Fatalf("expected 2 errors, got %d", snap["host-c"].Errors)
	}
}

func TestLastScan_IsRecent(t *testing.T) {
	c := New()
	before := time.Now()
	c.RecordScan("host-d")
	after := time.Now()
	snap := c.Snapshot()
	ls := snap["host-d"].LastScan
	if ls.Before(before) || ls.After(after) {
		t.Fatalf("LastScan %v not between %v and %v", ls, before, after)
	}
}

func TestReset_ClearsHost(t *testing.T) {
	c := New()
	c.RecordScan("host-e")
	c.RecordAlert("host-e")
	c.Reset("host-e")
	snap := c.Snapshot()
	if _, ok := snap["host-e"]; ok {
		t.Fatal("expected host-e to be absent after reset")
	}
}

func TestDifferentHosts_Independent(t *testing.T) {
	c := New()
	c.RecordScan("x")
	c.RecordScan("x")
	c.RecordScan("y")
	snap := c.Snapshot()
	if snap["x"].Scans != 2 {
		t.Fatalf("x: expected 2, got %d", snap["x"].Scans)
	}
	if snap["y"].Scans != 1 {
		t.Fatalf("y: expected 1, got %d", snap["y"].Scans)
	}
}

func TestSnapshot_IsIndependentCopy(t *testing.T) {
	c := New()
	c.RecordScan("host-f")
	snap1 := c.Snapshot()
	c.RecordScan("host-f")
	snap2 := c.Snapshot()
	if snap1["host-f"].Scans != 1 {
		t.Fatalf("snap1: expected 1 scan, got %d", snap1["host-f"].Scans)
	}
	if snap2["host-f"].Scans != 2 {
		t.Fatalf("snap2: expected 2 scans, got %d", snap2["host-f"].Scans)
	}
}

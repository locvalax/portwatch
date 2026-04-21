package dedupe

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/store"
)

func entry(host string, ports []int) store.Entry {
	return store.Entry{Host: host, Ports: ports, ScannedAt: time.Now()}
}

func TestIsDuplicate_FirstCall_NotDuplicate(t *testing.T) {
	c := New()
	if c.IsDuplicate(entry("host-a", []int{80, 443})) {
		t.Fatal("first call should never be a duplicate")
	}
}

func TestIsDuplicate_SamePorts_IsDuplicate(t *testing.T) {
	c := New()
	c.IsDuplicate(entry("host-a", []int{80, 443}))
	if !c.IsDuplicate(entry("host-a", []int{80, 443})) {
		t.Fatal("identical port set should be detected as duplicate")
	}
}

func TestIsDuplicate_DifferentPorts_NotDuplicate(t *testing.T) {
	c := New()
	c.IsDuplicate(entry("host-a", []int{80, 443}))
	if c.IsDuplicate(entry("host-a", []int{80, 443, 8080})) {
		t.Fatal("changed port set should not be a duplicate")
	}
}

func TestIsDuplicate_DifferentHosts_Independent(t *testing.T) {
	c := New()
	c.IsDuplicate(entry("host-a", []int{80}))
	c.IsDuplicate(entry("host-b", []int{80}))

	// second call per host — both should now be duplicates
	if !c.IsDuplicate(entry("host-a", []int{80})) {
		t.Error("host-a: expected duplicate")
	}
	if !c.IsDuplicate(entry("host-b", []int{80})) {
		t.Error("host-b: expected duplicate")
	}
}

func TestReset_ForcesFreshEntry(t *testing.T) {
	c := New()
	c.IsDuplicate(entry("host-a", []int{22, 80}))
	c.Reset("host-a")
	if c.IsDuplicate(entry("host-a", []int{22, 80})) {
		t.Fatal("after Reset the same ports should not be a duplicate")
	}
}

func TestFlush_ClearsAll(t *testing.T) {
	c := New()
	c.IsDuplicate(entry("host-a", []int{80}))
	c.IsDuplicate(entry("host-b", []int{443}))
	c.Flush()

	if c.IsDuplicate(entry("host-a", []int{80})) {
		t.Error("host-a: expected non-duplicate after Flush")
	}
	if c.IsDuplicate(entry("host-b", []int{443})) {
		t.Error("host-b: expected non-duplicate after Flush")
	}
}

func TestIsDuplicate_EmptyPorts(t *testing.T) {
	c := New()
	c.IsDuplicate(entry("host-a", []int{}))
	if !c.IsDuplicate(entry("host-a", []int{})) {
		t.Fatal("empty port sets should be treated as duplicates")
	}
}

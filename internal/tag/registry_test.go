package tag

import (
	"sort"
	"testing"
)

func TestRegistry_SetAndGet(t *testing.T) {
	r := NewRegistry()
	s, _ := New([]string{"env=prod"})
	r.Set("host-a", s)

	got, ok := r.Get("host-a")
	if !ok {
		t.Fatal("expected host-a to be registered")
	}
	if v, _ := got.Get("env"); v != "prod" {
		t.Errorf("expected env=prod, got %q", v)
	}
}

func TestRegistry_Get_Unknown(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("unknown")
	if ok {
		t.Error("expected false for unknown host")
	}
}

func TestRegistry_Delete(t *testing.T) {
	r := NewRegistry()
	s, _ := New([]string{"env=prod"})
	r.Set("host-a", s)
	r.Delete("host-a")
	_, ok := r.Get("host-a")
	if ok {
		t.Error("expected host-a to be deleted")
	}
}

func TestRegistry_Filter_MatchesSubset(t *testing.T) {
	r := NewRegistry()
	prod, _ := New([]string{"env=prod", "region=us"})
	staging, _ := New([]string{"env=staging", "region=us"})
	r.Set("host-prod", prod)
	r.Set("host-staging", staging)

	filter, _ := New([]string{"env=prod"})
	matched := r.Filter(filter)
	if len(matched) != 1 || matched[0] != "host-prod" {
		t.Errorf("expected [host-prod], got %v", matched)
	}
}

func TestRegistry_Filter_EmptyFilter_ReturnsAll(t *testing.T) {
	r := NewRegistry()
	s1, _ := New([]string{"env=prod"})
	s2, _ := New([]string{"env=staging"})
	r.Set("host-a", s1)
	r.Set("host-b", s2)

	matched := r.Filter(Set{})
	sort.Strings(matched)
	if len(matched) != 2 {
		t.Errorf("expected 2 hosts, got %v", matched)
	}
}

func TestRegistry_Hosts(t *testing.T) {
	r := NewRegistry()
	s, _ := New([]string{"env=prod"})
	r.Set("host-a", s)
	r.Set("host-b", s)

	hosts := r.Hosts()
	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}

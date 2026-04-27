package profile

import (
	"testing"
	"time"
)

func TestRegistry_SetAndGet(t *testing.T) {
	r := New()
	p := &Profile{Name: "prod", Hosts: []string{"10.0.0.1"}, Interval: 30 * time.Second}
	if err := r.Set(p); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := r.Get("prod")
	if !ok {
		t.Fatal("expected profile to exist")
	}
	if got.Name != "prod" {
		t.Errorf("name: got %q, want %q", got.Name, "prod")
	}
}

func TestRegistry_Get_Unknown(t *testing.T) {
	r := New()
	_, ok := r.Get("missing")
	if ok {
		t.Fatal("expected profile to be absent")
	}
}

func TestRegistry_Set_NilReturnsError(t *testing.T) {
	r := New()
	if err := r.Set(nil); err == nil {
		t.Fatal("expected error for nil profile")
	}
}

func TestRegistry_Set_EmptyNameReturnsError(t *testing.T) {
	r := New()
	if err := r.Set(&Profile{}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegistry_Delete(t *testing.T) {
	r := New()
	_ = r.Set(&Profile{Name: "tmp"})
	r.Delete("tmp")
	_, ok := r.Get("tmp")
	if ok {
		t.Fatal("expected profile to be deleted")
	}
}

func TestRegistry_All(t *testing.T) {
	r := New()
	_ = r.Set(&Profile{Name: "a"})
	_ = r.Set(&Profile{Name: "b"})
	if got := len(r.All()); got != 2 {
		t.Errorf("All: got %d profiles, want 2", got)
	}
}

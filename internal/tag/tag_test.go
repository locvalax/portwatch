package tag

import (
	"testing"
)

func TestNew_ValidEntries(t *testing.T) {
	s, err := New([]string{"env=prod", "region=us-east-1", "critical"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := s.Get("env"); !ok || v != "prod" {
		t.Errorf("expected env=prod, got %q", v)
	}
	if v, ok := s.Get("critical"); !ok || v != "" {
		t.Errorf("expected critical with empty value, got %q", v)
	}
}

func TestNew_InvalidEntry(t *testing.T) {
	_, err := New([]string{"="})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSorted_Order(t *testing.T) {
	s, _ := New([]string{"z=last", "a=first", "m=mid"})
	tags := s.Sorted()
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	if tags[0].Key != "a" || tags[1].Key != "m" || tags[2].Key != "z" {
		t.Errorf("unexpected order: %v", tags)
	}
}

func TestMatch_AllPresent(t *testing.T) {
	s, _ := New([]string{"env=prod", "region=us-east-1"})
	filter, _ := New([]string{"env=prod"})
	if !s.Match(filter) {
		t.Error("expected match")
	}
}

func TestMatch_MissingKey(t *testing.T) {
	s, _ := New([]string{"env=prod"})
	filter, _ := New([]string{"env=prod", "region=us-east-1"})
	if s.Match(filter) {
		t.Error("expected no match")
	}
}

func TestMatch_WrongValue(t *testing.T) {
	s, _ := New([]string{"env=staging"})
	filter, _ := New([]string{"env=prod"})
	if s.Match(filter) {
		t.Error("expected no match for wrong value")
	}
}

func TestTagString(t *testing.T) {
	tag := Tag{Key: "env", Value: "prod"}
	if tag.String() != "env=prod" {
		t.Errorf("unexpected string: %s", tag.String())
	}
}

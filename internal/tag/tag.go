package tag

import (
	"fmt"
	"sort"
	"strings"
)

// Tag represents a key-value label attached to a host.
type Tag struct {
	Key   string
	Value string
}

// String returns the canonical "key=value" representation.
func (t Tag) String() string {
	return fmt.Sprintf("%s=%s", t.Key, t.Value)
}

// Set holds a collection of tags for a single host.
type Set map[string]string

// New creates a Set from a slice of "key=value" strings.
// Entries that do not contain "=" are stored with an empty value.
func New(raw []string) (Set, error) {
	s := make(Set, len(raw))
	for _, r := range raw {
		parts := strings.SplitN(r, "=", 2)
		if len(parts) == 0 || parts[0] == "" {
			return nil, fmt.Errorf("tag: invalid entry %q", r)
		}
		key := strings.TrimSpace(parts[0])
		val := ""
		if len(parts) == 2 {
			val = strings.TrimSpace(parts[1])
		}
		s[key] = val
	}
	return s, nil
}

// Get returns the value for key and whether the key exists.
func (s Set) Get(key string) (string, bool) {
	v, ok := s[key]
	return v, ok
}

// Sorted returns all tags as a sorted slice of Tag structs.
func (s Set) Sorted() []Tag {
	tags := make([]Tag, 0, len(s))
	for k, v := range s {
		tags = append(tags, Tag{Key: k, Value: v})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})
	return tags
}

// Match reports whether the set contains all tags in the filter set.
func (s Set) Match(filter Set) bool {
	for k, v := range filter {
		got, ok := s[k]
		if !ok || got != v {
			return false
		}
	}
	return true
}

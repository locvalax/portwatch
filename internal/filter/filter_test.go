package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApply_NoRules_ReturnsAll(t *testing.T) {
	f := New(Rule{})
	result := f.Apply([]int{22, 80, 443, 8080})
	assert.Equal(t, []int{22, 80, 443, 8080}, result)
}

func TestApply_ExcludePorts(t *testing.T) {
	f := New(Rule{ExcludePorts: []int{22, 8080}})
	result := f.Apply([]int{22, 80, 443, 8080})
	assert.Equal(t, []int{80, 443}, result)
}

func TestApply_IncludePorts(t *testing.T) {
	f := New(Rule{IncludePorts: []int{80, 443}})
	result := f.Apply([]int{22, 80, 443, 8080})
	assert.Equal(t, []int{80, 443}, result)
}

func TestApply_PortRange(t *testing.T) {
	f := New(Rule{MinPort: 80, MaxPort: 1000})
	result := f.Apply([]int{22, 80, 443, 8080})
	assert.Equal(t, []int{80, 443}, result)
}

func TestApply_Empty(t *testing.T) {
	f := New(Rule{})
	result := f.Apply([]int{})
	assert.Nil(t, result)
}

func TestParsePorts_Valid(t *testing.T) {
	result := ParsePorts([]string{"22", "80", "443"})
	assert.Equal(t, []int{22, 80, 443}, result)
}

func TestParsePorts_SkipsInvalid(t *testing.T) {
	result := ParsePorts([]string{"22", "abc", "-1", "99999", "80"})
	assert.Equal(t, []int{22, 80}, result)
}

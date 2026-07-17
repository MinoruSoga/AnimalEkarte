package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsValidResource guards BUG-146: every entry in AllResources must validate,
// and an arbitrary/unregistered resource name must not.
func TestIsValidResource(t *testing.T) {
	for _, r := range AllResources {
		t.Run("registered resource "+string(r), func(t *testing.T) {
			assert.True(t, IsValidResource(string(r)))
		})
	}

	tests := []struct {
		name string
		in   string
	}{
		{"unregistered resource", "not-a-real-resource"},
		{"empty resource", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, IsValidResource(tt.in))
		})
	}
}

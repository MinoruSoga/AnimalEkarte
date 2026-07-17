package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidManualCategory(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"screens is valid", string(ManualCategoryScreens), true},
		{"workflows is valid", string(ManualCategoryWorkflows), true},
		{"unknown category is invalid", "unknown", false},
		{"empty category is invalid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidManualCategory(tt.in))
		})
	}
}

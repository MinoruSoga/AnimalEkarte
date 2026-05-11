package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDormantThresholds_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   DormantThresholds
		want DormantThresholds
	}{
		{
			name: "all zero filled with defaults",
			in:   DormantThresholds{Stage180: 0, Stage210: 0, Stage240: 0, Stage365: 0},
			want: DormantThresholds{Stage180: 180, Stage210: 210, Stage240: 240, Stage365: 365},
		},
		{
			name: "all custom preserved",
			in:   DormantThresholds{Stage180: 175, Stage210: 200, Stage240: 230, Stage365: 350},
			want: DormantThresholds{Stage180: 175, Stage210: 200, Stage240: 230, Stage365: 350},
		},
		{
			name: "partial zero filled selectively",
			in:   DormantThresholds{Stage180: 175, Stage210: 0, Stage240: 230, Stage365: 0},
			want: DormantThresholds{Stage180: 175, Stage210: 210, Stage240: 230, Stage365: 365},
		},
		{
			name: "negative treated as missing",
			in:   DormantThresholds{Stage180: -5, Stage210: -1, Stage240: 0, Stage365: 0},
			want: DormantThresholds{Stage180: 180, Stage210: 210, Stage240: 240, Stage365: 365},
		},
		{
			name: "value of 1 preserved (boundary above zero)",
			in:   DormantThresholds{Stage180: 1, Stage210: 1, Stage240: 1, Stage365: 1},
			want: DormantThresholds{Stage180: 1, Stage210: 1, Stage240: 1, Stage365: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.WithDefaults())
		})
	}
}

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthPreventionThresholds_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   HealthPreventionThresholds
		want HealthPreventionThresholds
	}{
		{
			name: "all zero filled with defaults",
			in:   HealthPreventionThresholds{},
			want: HealthPreventionThresholds{
				LookbackDays:    DefaultHealthPreventionLookbackDays,
				VaccineDeadline: DefaultVaccineDeadlineDays,
			},
		},
		{
			name: "all custom preserved",
			in:   HealthPreventionThresholds{LookbackDays: 400, VaccineDeadline: 90},
			want: HealthPreventionThresholds{LookbackDays: 400, VaccineDeadline: 90},
		},
		{
			name: "partial zero filled selectively",
			in:   HealthPreventionThresholds{LookbackDays: 400, VaccineDeadline: 0},
			want: HealthPreventionThresholds{LookbackDays: 400, VaccineDeadline: DefaultVaccineDeadlineDays},
		},
		{
			name: "negative treated as missing",
			in:   HealthPreventionThresholds{LookbackDays: -1, VaccineDeadline: -1},
			want: HealthPreventionThresholds{
				LookbackDays:    DefaultHealthPreventionLookbackDays,
				VaccineDeadline: DefaultVaccineDeadlineDays,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.WithDefaults())
		})
	}
}

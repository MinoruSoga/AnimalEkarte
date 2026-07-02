package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCPMV1Thresholds_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   CPMV1Thresholds
		want CPMV1Thresholds
	}{
		{
			name: "all zero filled with defaults",
			in:   CPMV1Thresholds{},
			want: CPMV1Thresholds{
				DormantDays:      DefaultCPMV1DormantDays,
				NoahDays:         DefaultCPMV1NoahDays,
				NoahAnnualVisits: DefaultCPMV1NoahAnnualVisits,
				NoahLTV:          DefaultCPMV1NoahLTV,
				CoreDays:         DefaultCPMV1CoreDays,
				CoreAnnualVisits: DefaultCPMV1CoreAnnualVisits,
				CoreLTV:          DefaultCPMV1CoreLTV,
				SpotMinAmount:    DefaultCPMV1SpotMinAmount,
				SpotInactiveDays: DefaultCPMV1SpotInactiveDays,
				GrowingMaxDays:   DefaultCPMV1GrowingMaxDays,
				GrowingMinVisits: DefaultCPMV1GrowingMinVisits,
				GrowingMaxVisits: DefaultCPMV1GrowingMaxVisits,
				LTVBreakLow:      DefaultCPMV1LTVBreakLow,
			},
		},
		{
			name: "all custom preserved",
			in: CPMV1Thresholds{
				DormantDays: 300, NoahDays: 400, NoahAnnualVisits: 5, NoahLTV: 100_000,
				CoreDays: 200, CoreAnnualVisits: 3, CoreLTV: 60_000,
				SpotMinAmount: 40_000, SpotInactiveDays: 120,
				GrowingMaxDays: 100, GrowingMinVisits: 3, GrowingMaxVisits: 5,
				LTVBreakLow: 25_000,
			},
			want: CPMV1Thresholds{
				DormantDays: 300, NoahDays: 400, NoahAnnualVisits: 5, NoahLTV: 100_000,
				CoreDays: 200, CoreAnnualVisits: 3, CoreLTV: 60_000,
				SpotMinAmount: 40_000, SpotInactiveDays: 120,
				GrowingMaxDays: 100, GrowingMinVisits: 3, GrowingMaxVisits: 5,
				LTVBreakLow: 25_000,
			},
		},
		{
			name: "negative treated as missing",
			in:   CPMV1Thresholds{DormantDays: -1, NoahLTV: -100},
			want: CPMV1Thresholds{
				DormantDays:      DefaultCPMV1DormantDays,
				NoahDays:         DefaultCPMV1NoahDays,
				NoahAnnualVisits: DefaultCPMV1NoahAnnualVisits,
				NoahLTV:          DefaultCPMV1NoahLTV,
				CoreDays:         DefaultCPMV1CoreDays,
				CoreAnnualVisits: DefaultCPMV1CoreAnnualVisits,
				CoreLTV:          DefaultCPMV1CoreLTV,
				SpotMinAmount:    DefaultCPMV1SpotMinAmount,
				SpotInactiveDays: DefaultCPMV1SpotInactiveDays,
				GrowingMaxDays:   DefaultCPMV1GrowingMaxDays,
				GrowingMinVisits: DefaultCPMV1GrowingMinVisits,
				GrowingMaxVisits: DefaultCPMV1GrowingMaxVisits,
				LTVBreakLow:      DefaultCPMV1LTVBreakLow,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.WithDefaults())
		})
	}
}

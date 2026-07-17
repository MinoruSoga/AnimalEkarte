package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCPMV2Thresholds_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   CPMV2Thresholds
		want CPMV2Thresholds
	}{
		{
			name: "all zero filled with defaults",
			in:   CPMV2Thresholds{},
			want: CPMV2Thresholds{
				Coming: DefaultCPMV2ComingThreshold,
				Good:   DefaultCPMV2GoodThreshold,
				Family: DefaultCPMV2FamilyThreshold,
				Noah:   DefaultCPMV2NoahThreshold,
			},
		},
		{
			name: "all custom preserved",
			in:   CPMV2Thresholds{Coming: 3, Good: 5, Family: 10, Noah: 15},
			want: CPMV2Thresholds{Coming: 3, Good: 5, Family: 10, Noah: 15},
		},
		{
			name: "partial zero filled selectively",
			in:   CPMV2Thresholds{Coming: 3, Good: 0, Family: 10, Noah: 0},
			want: CPMV2Thresholds{Coming: 3, Good: DefaultCPMV2GoodThreshold, Family: 10, Noah: DefaultCPMV2NoahThreshold},
		},
		{
			name: "negative treated as missing",
			in:   CPMV2Thresholds{Coming: -1, Good: -1, Family: -1, Noah: -1},
			want: CPMV2Thresholds{
				Coming: DefaultCPMV2ComingThreshold,
				Good:   DefaultCPMV2GoodThreshold,
				Family: DefaultCPMV2FamilyThreshold,
				Noah:   DefaultCPMV2NoahThreshold,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.WithDefaults())
		})
	}
}

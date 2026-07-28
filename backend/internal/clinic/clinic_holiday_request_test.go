package clinic

import (
	"net/url"
	"testing"
)

func TestNewListClinicHolidaysQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   string
	}{
		{
			name:   "returns year_month when present",
			values: url.Values{"year_month": []string{"2026-05"}},
			want:   "2026-05",
		},
		{
			name:   "returns empty string when year_month is absent",
			values: url.Values{},
			want:   "",
		},
		{
			name:   "returns empty string for nil values",
			values: nil,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewListClinicHolidaysQuery(tt.values)
			if got.YearMonth != tt.want {
				t.Fatalf("YearMonth = %q, want %q", got.YearMonth, tt.want)
			}
		})
	}
}

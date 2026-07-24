package clinic

import (
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToClinicSettingsResponse(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		in   *model.ClinicSettings
	}{
		{
			name: "converts full settings",
			in: &model.ClinicSettings{
				ClinicID:            1,
				ClosingAmPmBoundary: "12:00",
				ClosingWeekdayEnd:   "18:30",
				ClosingSundayEnd:    "17:30",
				ClosedWeekdays:      pq.Int64Array{0, 6},
				CreatedAt:           now,
				UpdatedAt:           now,
			},
		},
		{
			name: "converts zero-value settings",
			in:   &model.ClinicSettings{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToClinicSettingsResponse(tt.in)
			if got.ClinicID != tt.in.ClinicID {
				t.Errorf("ClinicID = %d, want %d", got.ClinicID, tt.in.ClinicID)
			}
			if got.ClosingAmPmBoundary != tt.in.ClosingAmPmBoundary {
				t.Errorf("ClosingAmPmBoundary = %q, want %q", got.ClosingAmPmBoundary, tt.in.ClosingAmPmBoundary)
			}
			if got.ClosingWeekdayEnd != tt.in.ClosingWeekdayEnd {
				t.Errorf("ClosingWeekdayEnd = %q, want %q", got.ClosingWeekdayEnd, tt.in.ClosingWeekdayEnd)
			}
			if got.ClosingSundayEnd != tt.in.ClosingSundayEnd {
				t.Errorf("ClosingSundayEnd = %q, want %q", got.ClosingSundayEnd, tt.in.ClosingSundayEnd)
			}
			if len(got.ClosedWeekdays) != len(tt.in.ClosedWeekdays) {
				t.Errorf("len(ClosedWeekdays) = %d, want %d", len(got.ClosedWeekdays), len(tt.in.ClosedWeekdays))
			}
		})
	}
}

func TestToClosingSpecialPeriodResponse(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	tests := []struct {
		name string
		in   *model.ClosingSpecialPeriod
		want ClosingSpecialPeriodResponse
	}{
		{
			name: "converts special period with dates",
			in: &model.ClosingSpecialPeriod{
				ID:           1,
				ClinicID:     2,
				StartDate:    start,
				EndDate:      end,
				AmPmBoundary: "12:00",
				PmEnd:        "17:00",
				Note:         "gw",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			want: ClosingSpecialPeriodResponse{
				ID:           1,
				ClinicID:     2,
				StartDate:    start.In(time.Local).Format("2006-01-02"),
				EndDate:      end.In(time.Local).Format("2006-01-02"),
				AmPmBoundary: "12:00",
				PmEnd:        "17:00",
				Note:         "gw",
			},
		},
		{
			name: "converts zero-value special period",
			in:   &model.ClosingSpecialPeriod{},
			want: ClosingSpecialPeriodResponse{
				StartDate: time.Time{}.In(time.Local).Format("2006-01-02"),
				EndDate:   time.Time{}.In(time.Local).Format("2006-01-02"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToClosingSpecialPeriodResponse(tt.in)
			if got.ID != tt.want.ID {
				t.Errorf("ID = %d, want %d", got.ID, tt.want.ID)
			}
			if got.ClinicID != tt.want.ClinicID {
				t.Errorf("ClinicID = %d, want %d", got.ClinicID, tt.want.ClinicID)
			}
			if got.StartDate != tt.want.StartDate {
				t.Errorf("StartDate = %q, want %q", got.StartDate, tt.want.StartDate)
			}
			if got.EndDate != tt.want.EndDate {
				t.Errorf("EndDate = %q, want %q", got.EndDate, tt.want.EndDate)
			}
			if got.AmPmBoundary != tt.want.AmPmBoundary {
				t.Errorf("AmPmBoundary = %q, want %q", got.AmPmBoundary, tt.want.AmPmBoundary)
			}
			if got.PmEnd != tt.want.PmEnd {
				t.Errorf("PmEnd = %q, want %q", got.PmEnd, tt.want.PmEnd)
			}
			if got.Note != tt.want.Note {
				t.Errorf("Note = %q, want %q", got.Note, tt.want.Note)
			}
		})
	}
}

func TestToClosingSettingsFullResponse(t *testing.T) {
	tests := []struct {
		name       string
		settings   *model.ClinicSettings
		periods    []model.ClosingSpecialPeriod
		wantCount  int
		wantClinic uint64
	}{
		{
			name:       "combines settings and special periods",
			settings:   &model.ClinicSettings{ClinicID: 1, ClosingAmPmBoundary: "12:00"},
			periods:    []model.ClosingSpecialPeriod{{ID: 1, ClinicID: 1}, {ID: 2, ClinicID: 1}},
			wantCount:  2,
			wantClinic: 1,
		},
		{
			name:       "handles nil special periods slice",
			settings:   &model.ClinicSettings{ClinicID: 1},
			periods:    nil,
			wantCount:  0,
			wantClinic: 1,
		},
		{
			name:       "handles empty special periods slice",
			settings:   &model.ClinicSettings{ClinicID: 3},
			periods:    []model.ClosingSpecialPeriod{},
			wantCount:  0,
			wantClinic: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToClosingSettingsFullResponse(tt.settings, tt.periods)
			if got.Settings.ClinicID != tt.wantClinic {
				t.Errorf("Settings.ClinicID = %d, want %d", got.Settings.ClinicID, tt.wantClinic)
			}
			if len(got.SpecialPeriods) != tt.wantCount {
				t.Errorf("len(SpecialPeriods) = %d, want %d", len(got.SpecialPeriods), tt.wantCount)
			}
		})
	}
}

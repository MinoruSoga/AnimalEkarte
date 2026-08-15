package clinic

import (
	"time"

	"github.com/lib/pq"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// clinicSettingsResponse は ClinicSettings の HTTP レスポンス型
type ClinicSettingsResponse struct {
	ClinicID            uint64        `json:"clinic_id"`
	ClosingAmPmBoundary string        `json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   string        `json:"closing_weekday_end"`
	ClosingSundayEnd    string        `json:"closing_sunday_end"`
	ClosedWeekdays      pq.Int64Array `json:"closed_weekdays"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

func ToClinicSettingsResponse(s *model.ClinicSettings) ClinicSettingsResponse {
	return ClinicSettingsResponse{
		ClinicID:            s.ClinicID,
		ClosingAmPmBoundary: s.ClosingAmPmBoundary,
		ClosingWeekdayEnd:   s.ClosingWeekdayEnd,
		ClosingSundayEnd:    s.ClosingSundayEnd,
		ClosedWeekdays:      s.ClosedWeekdays,
		CreatedAt:           httpapi.LocalTime(s.CreatedAt),
		UpdatedAt:           httpapi.LocalTime(s.UpdatedAt),
	}
}

// closingSpecialPeriodResponse は ClosingSpecialPeriod の HTTP レスポンス型
type ClosingSpecialPeriodResponse struct {
	ID           uint64    `json:"id"`
	ClinicID     uint64    `json:"clinic_id"`
	StartDate    string    `json:"start_date"`
	EndDate      string    `json:"end_date"`
	AmPmBoundary string    `json:"am_pm_boundary"`
	PmEnd        string    `json:"pm_end"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// closingSettingsFullResponse は Get エンドポイントの HTTP レスポンス型（設定 + 特別期間）
type ClosingSettingsFullResponse struct {
	Settings       ClinicSettingsResponse         `json:"settings"`
	SpecialPeriods []ClosingSpecialPeriodResponse `json:"special_periods"`
}

func ToClosingSettingsFullResponse(s *model.ClinicSettings, periods []model.ClosingSpecialPeriod) ClosingSettingsFullResponse {
	sp := httpapi.MapSlice(periods, ToClosingSpecialPeriodResponse)
	return ClosingSettingsFullResponse{
		Settings:       ToClinicSettingsResponse(s),
		SpecialPeriods: sp,
	}
}

func ToClosingSpecialPeriodResponse(p *model.ClosingSpecialPeriod) ClosingSpecialPeriodResponse {
	return ClosingSpecialPeriodResponse{
		ID:           p.ID,
		ClinicID:     p.ClinicID,
		StartDate:    p.StartDate.In(time.Local).Format(time.DateOnly),
		EndDate:      p.EndDate.In(time.Local).Format(time.DateOnly),
		AmPmBoundary: p.AmPmBoundary,
		PmEnd:        p.PmEnd,
		Note:         p.Note,
		CreatedAt:    p.CreatedAt.In(time.Local),
		UpdatedAt:    p.UpdatedAt.In(time.Local),
	}
}

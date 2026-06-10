package handler

import (
	"time"

	"github.com/lib/pq"

	"github.com/animal-ekarte/backend/internal/model"
)

// clinicSettingsResponse は ClinicSettings の HTTP レスポンス型
type clinicSettingsResponse struct {
	ClinicID            uint64        `json:"clinic_id"`
	ClosingAmPmBoundary string        `json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   string        `json:"closing_weekday_end"`
	ClosingSundayEnd    string        `json:"closing_sunday_end"`
	ClosedWeekdays      pq.Int64Array `json:"closed_weekdays"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

func toClinicSettingsResponse(s *model.ClinicSettings) clinicSettingsResponse {
	return clinicSettingsResponse{
		ClinicID:            s.ClinicID,
		ClosingAmPmBoundary: s.ClosingAmPmBoundary,
		ClosingWeekdayEnd:   s.ClosingWeekdayEnd,
		ClosingSundayEnd:    s.ClosingSundayEnd,
		ClosedWeekdays:      s.ClosedWeekdays,
		CreatedAt:           localTime(s.CreatedAt),
		UpdatedAt:           localTime(s.UpdatedAt),
	}
}

// closingSpecialPeriodResponse は ClosingSpecialPeriod の HTTP レスポンス型
type closingSpecialPeriodResponse struct {
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
type closingSettingsFullResponse struct {
	Settings       clinicSettingsResponse         `json:"settings"`
	SpecialPeriods []closingSpecialPeriodResponse `json:"special_periods"`
}

func toClosingSettingsFullResponse(s *model.ClinicSettings, periods []model.ClosingSpecialPeriod) closingSettingsFullResponse {
	sp := make([]closingSpecialPeriodResponse, 0, len(periods))
	for i := range periods {
		sp = append(sp, toClosingSpecialPeriodResponse(&periods[i]))
	}
	return closingSettingsFullResponse{
		Settings:       toClinicSettingsResponse(s),
		SpecialPeriods: sp,
	}
}

func toClosingSpecialPeriodResponse(p *model.ClosingSpecialPeriod) closingSpecialPeriodResponse {
	return closingSpecialPeriodResponse{
		ID:           p.ID,
		ClinicID:     p.ClinicID,
		StartDate:    p.StartDate.In(time.Local).Format("2006-01-02"),
		EndDate:      p.EndDate.In(time.Local).Format("2006-01-02"),
		AmPmBoundary: p.AmPmBoundary,
		PmEnd:        p.PmEnd,
		Note:         p.Note,
		CreatedAt:    p.CreatedAt.In(time.Local),
		UpdatedAt:    p.UpdatedAt.In(time.Local),
	}
}

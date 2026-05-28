package handler

import "github.com/animal-ekarte/backend/internal/service"

type updateClinicSettingsRequest struct {
	ClosingAmPmBoundary *string `json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   *string `json:"closing_weekday_end"`
	ClosingSundayEnd    *string `json:"closing_sunday_end"`
	ClosedWeekdays      []int64 `json:"closed_weekdays"`
}

func (r updateClinicSettingsRequest) toServiceInput() service.UpdateClinicSettingsInput {
	return service.UpdateClinicSettingsInput{
		ClosingAmPmBoundary: r.ClosingAmPmBoundary,
		ClosingWeekdayEnd:   r.ClosingWeekdayEnd,
		ClosingSundayEnd:    r.ClosingSundayEnd,
		ClosedWeekdays:      r.ClosedWeekdays,
	}
}

type createSpecialPeriodRequest struct {
	StartDate    string `json:"start_date"    binding:"required"` // YYYY-MM-DD
	EndDate      string `json:"end_date"      binding:"required"` // YYYY-MM-DD
	AmPmBoundary string `json:"am_pm_boundary" binding:"required"`
	PmEnd        string `json:"pm_end"         binding:"required"`
	Note         string `json:"note"`
}

func (r createSpecialPeriodRequest) toServiceInput() *service.CreateSpecialPeriodInput {
	return &service.CreateSpecialPeriodInput{
		StartDate:    r.StartDate,
		EndDate:      r.EndDate,
		AmPmBoundary: r.AmPmBoundary,
		PmEnd:        r.PmEnd,
		Note:         r.Note,
	}
}

type updateSpecialPeriodRequest struct {
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	AmPmBoundary *string `json:"am_pm_boundary"`
	PmEnd        *string `json:"pm_end"`
	Note         *string `json:"note"`
}

func (r updateSpecialPeriodRequest) toServiceInput() service.UpdateSpecialPeriodInput {
	return service.UpdateSpecialPeriodInput{
		StartDate:    r.StartDate,
		EndDate:      r.EndDate,
		AmPmBoundary: r.AmPmBoundary,
		PmEnd:        r.PmEnd,
		Note:         r.Note,
	}
}

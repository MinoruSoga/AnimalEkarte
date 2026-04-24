package handler

type updateClinicSettingsRequest struct {
	ClosingAmPmBoundary *string `json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   *string `json:"closing_weekday_end"`
	ClosingSundayEnd    *string `json:"closing_sunday_end"`
	ClosedWeekdays      []int64 `json:"closed_weekdays"`
}

type createSpecialPeriodRequest struct {
	StartDate    string `json:"start_date"    binding:"required"` // YYYY-MM-DD
	EndDate      string `json:"end_date"      binding:"required"` // YYYY-MM-DD
	AmPmBoundary string `json:"am_pm_boundary" binding:"required"`
	PmEnd        string `json:"pm_end"         binding:"required"`
	Note         string `json:"note"`
}

type updateSpecialPeriodRequest struct {
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	AmPmBoundary *string `json:"am_pm_boundary"`
	PmEnd        *string `json:"pm_end"`
	Note         *string `json:"note"`
}

package clinic

type UpdateClinicSettingsRequest struct {
	ClosingAmPmBoundary *string `json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   *string `json:"closing_weekday_end"`
	ClosingSundayEnd    *string `json:"closing_sunday_end"`
	ClosedWeekdays      []int64 `json:"closed_weekdays"`
}

func (r UpdateClinicSettingsRequest) ToServiceInput() UpdateClinicSettingsInput {
	return UpdateClinicSettingsInput(r)
}

type CreateSpecialPeriodRequest struct {
	StartDate    string `json:"start_date"    binding:"required"` // YYYY-MM-DD
	EndDate      string `json:"end_date"      binding:"required"` // YYYY-MM-DD
	AmPmBoundary string `json:"am_pm_boundary" binding:"required"`
	PmEnd        string `json:"pm_end"         binding:"required"`
	Note         string `json:"note"`
}

func (r *CreateSpecialPeriodRequest) ToServiceInput() *CreateSpecialPeriodInput {
	return &CreateSpecialPeriodInput{
		StartDate:    r.StartDate,
		EndDate:      r.EndDate,
		AmPmBoundary: r.AmPmBoundary,
		PmEnd:        r.PmEnd,
		Note:         r.Note,
	}
}

type UpdateSpecialPeriodRequest struct {
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	AmPmBoundary *string `json:"am_pm_boundary"`
	PmEnd        *string `json:"pm_end"`
	Note         *string `json:"note"`
}

func (r UpdateSpecialPeriodRequest) ToServiceInput() UpdateSpecialPeriodInput {
	return UpdateSpecialPeriodInput(r)
}

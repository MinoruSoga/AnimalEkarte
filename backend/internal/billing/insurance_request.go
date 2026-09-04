package billing

type createInsuranceRequest struct {
	Name         string `json:"name"          binding:"required,max=255"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
	CoverageRate *int   `json:"coverage_rate" binding:"omitempty,min=0,max=100"`
	ContactPhone string `json:"contact_phone"`
	SortOrder    int    `json:"sort_order"`
}

func (r createInsuranceRequest) toServiceInput() *CreateInsuranceInput {
	return &CreateInsuranceInput{
		Name:         r.Name,
		IsActive:     r.IsActive,
		Description:  r.Description,
		CoverageRate: r.CoverageRate,
		ContactPhone: r.ContactPhone,
		SortOrder:    r.SortOrder,
	}
}

type updateInsuranceRequest struct {
	Name         *string `json:"name" binding:"omitempty,max=255"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
	CoverageRate *int    `json:"coverage_rate" binding:"omitempty,min=0,max=100"`
	ContactPhone *string `json:"contact_phone"`
	SortOrder    *int    `json:"sort_order"`
}

func (r updateInsuranceRequest) toServiceInput() *UpdateInsuranceInput {
	return &UpdateInsuranceInput{
		Name:         r.Name,
		IsActive:     r.IsActive,
		Description:  r.Description,
		CoverageRate: r.CoverageRate,
		ContactPhone: r.ContactPhone,
		SortOrder:    r.SortOrder,
	}
}

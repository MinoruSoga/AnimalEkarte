package handler

type createInsuranceRequest struct {
	Name         string `json:"name"          binding:"required"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
	CoverageRate *float64 `json:"coverage_rate"`
	ContactPhone string   `json:"contact_phone"`
	SortOrder    int      `json:"sort_order"`
}

type updateInsuranceRequest struct {
	Name         string   `json:"name"`
	IsActive     *bool    `json:"is_active"`
	Description  string   `json:"description"`
	CoverageRate *float64 `json:"coverage_rate"`
	ContactPhone string `json:"contact_phone"`
	SortOrder    int    `json:"sort_order"`
}

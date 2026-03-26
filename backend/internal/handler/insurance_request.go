package handler

type createInsuranceRequest struct {
	Name         string `json:"name"          binding:"required"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
	CoverageRate *int   `json:"coverage_rate"`
	ContactPhone string `json:"contact_phone"`
	SortOrder    int    `json:"sort_order"`
}

type updateInsuranceRequest struct {
	Name         *string `json:"name"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
	CoverageRate *int    `json:"coverage_rate"`
	ContactPhone *string `json:"contact_phone"`
	SortOrder    *int    `json:"sort_order"`
}

// reorderInsuranceRequest は保険並び替えリクエスト。
type reorderInsuranceRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

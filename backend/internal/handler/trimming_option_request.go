package handler

type createTrimmingOptionRequest struct {
	Name         string `json:"name"        binding:"required"`
	Price        *int64 `json:"price"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
	Duration     *int   `json:"duration"`
	IsCombinable bool   `json:"is_combinable"`
	SortOrder    int    `json:"sort_order"`
}

type updateTrimmingOptionRequest struct {
	Name         *string `json:"name"`
	Price        *int64  `json:"price"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
	Duration     *int    `json:"duration"`
	IsCombinable *bool   `json:"is_combinable"`
	SortOrder    *int    `json:"sort_order"`
}

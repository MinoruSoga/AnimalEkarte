package handler

type createServiceTypeRequest struct {
	Name        string `json:"name"        binding:"required"`
	Color       string `json:"color"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type updateServiceTypeRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

type reorderServiceTypeRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

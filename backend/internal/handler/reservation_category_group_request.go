package handler

type createReservationCategoryGroupRequest struct {
	Name      string `json:"name"       binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type updateReservationCategoryGroupRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

type reorderReservationCategoryGroupRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

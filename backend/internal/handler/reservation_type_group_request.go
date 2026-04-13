package handler

type createReservationTypeGroupRequest struct {
	Name      string `json:"name"       binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type updateReservationTypeGroupRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

type reorderReservationTypeGroupRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

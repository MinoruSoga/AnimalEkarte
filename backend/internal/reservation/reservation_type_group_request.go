package reservation

type createReservationTypeGroupRequest struct {
	Name      string `json:"name"       binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	// IsActive is *bool so JSON binding can distinguish omitted / false / true.
	// Omitted (nil) resolves to true in toServiceInput.
	IsActive  *bool  `json:"is_active"`
}

func (r createReservationTypeGroupRequest) toServiceInput() *CreateReservationTypeGroupInput {
	return &CreateReservationTypeGroupInput{
		Name:      r.Name,
		Color:     r.Color,
		SortOrder: r.SortOrder,
		IsActive:  resolveBoolDefaultTrue(r.IsActive),
	}
}

type updateReservationTypeGroupRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

func (r updateReservationTypeGroupRequest) toServiceInput() *UpdateReservationTypeGroupInput {
	return &UpdateReservationTypeGroupInput{
		Name:      r.Name,
		Color:     r.Color,
		SortOrder: r.SortOrder,
		IsActive:  r.IsActive,
	}
}

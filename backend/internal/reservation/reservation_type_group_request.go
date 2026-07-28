package reservation

type createReservationTypeGroupRequest struct {
	Name      string `json:"name"       binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func (r createReservationTypeGroupRequest) toServiceInput() *CreateReservationTypeGroupInput {
	return &CreateReservationTypeGroupInput{
		Name:      r.Name,
		Color:     r.Color,
		SortOrder: r.SortOrder,
		IsActive:  r.IsActive,
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

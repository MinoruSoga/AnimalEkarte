package handler

import "github.com/animal-ekarte/backend/internal/service"

type createReservationTypeGroupRequest struct {
	Name      string `json:"name"       binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func (r createReservationTypeGroupRequest) toServiceInput() *service.CreateReservationTypeGroupInput {
	return &service.CreateReservationTypeGroupInput{
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

func (r updateReservationTypeGroupRequest) toServiceInput() *service.UpdateReservationTypeGroupInput {
	return &service.UpdateReservationTypeGroupInput{
		Name:      r.Name,
		Color:     r.Color,
		SortOrder: r.SortOrder,
		IsActive:  r.IsActive,
	}
}

package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationCategoryGroupResponse struct {
	ID        uint64    `json:"id"`
	ClinicID  uint64    `json:"clinic_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toReservationCategoryGroupResponse(g *model.ReservationCategoryGroup) reservationCategoryGroupResponse {
	return reservationCategoryGroupResponse{
		ID:        g.ID,
		ClinicID:  g.ClinicID,
		Name:      g.Name,
		Color:     g.Color,
		SortOrder: g.SortOrder,
		IsActive:  g.IsActive,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func toReservationCategoryGroupResponseList(items []model.ReservationCategoryGroup) []reservationCategoryGroupResponse {
	list := make([]reservationCategoryGroupResponse, 0, len(items))
	for i := range items {
		list = append(list, toReservationCategoryGroupResponse(&items[i]))
	}
	return list
}

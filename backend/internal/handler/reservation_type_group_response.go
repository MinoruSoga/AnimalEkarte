package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationTypeGroupResponse struct {
	ID        uint64    `json:"id"`
	ClinicID  uint64    `json:"clinic_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toReservationTypeGroupResponse(g *model.ReservationTypeGroup) reservationTypeGroupResponse {
	return reservationTypeGroupResponse{
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

func toReservationTypeGroupResponseList(items []model.ReservationTypeGroup) []reservationTypeGroupResponse {
	list := make([]reservationTypeGroupResponse, 0, len(items))
	for i := range items {
		list = append(list, toReservationTypeGroupResponse(&items[i]))
	}
	return list
}

package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type occupationResponse struct {
	ID          string    `json:"id"`
	ClinicID    string    `json:"clinic_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toOccupationResponse(occ *model.Occupation) occupationResponse {
	return occupationResponse{
		ID:          strconv.FormatUint(occ.ID, 10),
		ClinicID:    strconv.FormatUint(occ.ClinicID, 10),
		Name:        occ.Name,
		Description: occ.Description,
		IsActive:    occ.IsActive,
		SortOrder:   occ.SortOrder,
		CreatedAt:   occ.CreatedAt,
		UpdatedAt:   occ.UpdatedAt,
	}
}

func toOccupationResponseList(items []model.Occupation) []occupationResponse {
	list := make([]occupationResponse, 0, len(items))
	for i := range items {
		list = append(list, toOccupationResponse(&items[i]))
	}
	return list
}

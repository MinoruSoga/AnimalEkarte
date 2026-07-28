package staff

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type occupationResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toOccupationResponse(occ *model.Occupation) occupationResponse {
	return occupationResponse{
		ID:          occ.ID,
		ClinicID:    occ.ClinicID,
		Name:        occ.Name,
		Description: occ.Description,
		IsActive:    occ.IsActive,
		SortOrder:   occ.SortOrder,
		CreatedAt:   localTime(occ.CreatedAt),
		UpdatedAt:   localTime(occ.UpdatedAt),
	}
}

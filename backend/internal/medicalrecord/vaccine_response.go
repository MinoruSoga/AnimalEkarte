package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type vaccineResponse struct {
	ID          uint64                `json:"id"`
	ClinicID    uint64                `json:"clinic_id"`
	Name        string                `json:"name"`
	Price       *int64                `json:"price,omitempty"`
	IsActive    bool                  `json:"is_active"`
	Description string                `json:"description"`
	Species     *model.VaccineSpecies `json:"species,omitempty"`
	Interval    string                `json:"interval"`
	InventoryID *uint64               `json:"inventory_id,omitempty"`
	ParentID    *uint64               `json:"parent_id,omitempty"`
	SortOrder   int                   `json:"sort_order"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

func toVaccineResponse(v *model.Vaccine) vaccineResponse {
	return vaccineResponse{
		ID:          v.ID,
		ClinicID:    v.ClinicID,
		Name:        v.Name,
		Price:       v.Price,
		IsActive:    v.IsActive,
		Description: v.Description,
		Species:     v.Species,
		Interval:    v.Interval,
		InventoryID: v.InventoryID,
		ParentID:    v.ParentID,
		SortOrder:   v.SortOrder,
		CreatedAt:   httpapi.LocalTime(v.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(v.UpdatedAt),
	}
}

package trimming

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingOptionResponse struct {
	ID           uint64    `json:"id"`
	ClinicID     uint64    `json:"clinic_id"`
	Name         string    `json:"name"`
	Price        *int64    `json:"price,omitempty"`
	IsActive     bool      `json:"is_active"`
	Description  string    `json:"description"`
	Duration     *int      `json:"duration,omitempty"`
	IsCombinable bool      `json:"is_combinable"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toTrimmingOptionResponse(o *model.TrimmingOption) trimmingOptionResponse {
	return trimmingOptionResponse{
		ID:           o.ID,
		ClinicID:     o.ClinicID,
		Name:         o.Name,
		Price:        o.Price,
		IsActive:     o.IsActive,
		Description:  o.Description,
		Duration:     o.Duration,
		IsCombinable: o.IsCombinable,
		SortOrder:    o.SortOrder,
		CreatedAt:    localTime(o.CreatedAt),
		UpdatedAt:    localTime(o.UpdatedAt),
	}
}

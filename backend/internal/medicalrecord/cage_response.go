package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/model"
)

type cageResponse struct {
	ID          uint64         `json:"id"`
	ClinicID    uint64         `json:"clinic_id"`
	Name        string         `json:"name"`
	CageType    model.CageType `json:"cage_type"`
	CageSize    model.CageSize `json:"cage_size"`
	Price       *int64         `json:"price,omitempty"`
	IsActive    bool           `json:"is_active"`
	Description string         `json:"description"`
	SortOrder   int            `json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func toCageResponse(c *model.Cage) cageResponse {
	return cageResponse{
		ID:          c.ID,
		ClinicID:    c.ClinicID,
		Name:        c.Name,
		CageType:    c.CageType,
		CageSize:    c.CageSize,
		Price:       c.Price,
		IsActive:    c.IsActive,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		CreatedAt:   httpapi.LocalTime(c.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(c.UpdatedAt),
	}
}

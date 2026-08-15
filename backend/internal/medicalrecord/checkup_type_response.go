package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type checkupTypeResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *int64    `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	Interval    string    `json:"interval"`
	TargetAge   string    `json:"target_age"`
	ParentID    *uint64   `json:"parent_id,omitempty"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toCheckupTypeResponse(ct *model.CheckupType) checkupTypeResponse {
	return checkupTypeResponse{
		ID:          ct.ID,
		ClinicID:    ct.ClinicID,
		Name:        ct.Name,
		Price:       ct.Price,
		IsActive:    ct.IsActive,
		Description: ct.Description,
		Interval:    ct.Interval,
		TargetAge:   ct.TargetAge,
		ParentID:    ct.ParentID,
		SortOrder:   ct.SortOrder,
		CreatedAt:   httpapi.LocalTime(ct.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(ct.UpdatedAt),
	}
}

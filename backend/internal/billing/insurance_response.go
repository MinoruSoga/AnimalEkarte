package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type insuranceResponse struct {
	ID           uint64    `json:"id"`
	ClinicID     uint64    `json:"clinic_id"`
	Name         string    `json:"name"`
	IsActive     bool      `json:"is_active"`
	Description  string    `json:"description"`
	CoverageRate int       `json:"coverage_rate"`
	ContactPhone string    `json:"contact_phone"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toInsuranceResponse(ins *model.Insurance) insuranceResponse {
	return insuranceResponse{
		ID:           ins.ID,
		ClinicID:     ins.ClinicID,
		Name:         ins.Name,
		IsActive:     ins.IsActive,
		Description:  ins.Description,
		CoverageRate: ins.CoverageRate,
		ContactPhone: ins.ContactPhone,
		SortOrder:    ins.SortOrder,
		CreatedAt:    httpapi.LocalTime(ins.CreatedAt),
		UpdatedAt:    httpapi.LocalTime(ins.UpdatedAt),
	}
}

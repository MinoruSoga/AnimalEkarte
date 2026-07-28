package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type hospitalizationPlanResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *int64    `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	BodySize    *string   `json:"body_size,omitempty"`
	BillingUnit *string   `json:"billing_unit,omitempty"`
	TaxType     string    `json:"tax_type"`
	TaxRate     float64   `json:"tax_rate"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toHospitalizationPlanResponse(p *model.HospitalizationPlan) hospitalizationPlanResponse {
	var bodySize *string
	if p.BodySize != nil {
		s := string(*p.BodySize)
		bodySize = &s
	}
	var billingUnit *string
	if p.BillingUnit != nil {
		s := string(*p.BillingUnit)
		billingUnit = &s
	}
	return hospitalizationPlanResponse{
		ID:          p.ID,
		ClinicID:    p.ClinicID,
		Name:        p.Name,
		Price:       p.Price,
		IsActive:    p.IsActive,
		Description: p.Description,
		BodySize:    bodySize,
		BillingUnit: billingUnit,
		TaxType:     string(p.TaxType),
		TaxRate:     p.TaxRate,
		SortOrder:   p.SortOrder,
		CreatedAt:   httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(p.UpdatedAt),
	}
}

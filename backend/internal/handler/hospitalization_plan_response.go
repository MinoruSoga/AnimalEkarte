package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type hospitalizationPlanResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *float64  `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	BodySize    *string   `json:"body_size,omitempty"`
	BillingUnit *string   `json:"billing_unit,omitempty"`
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
		SortOrder:   p.SortOrder,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toHospitalizationPlanResponseList(plans []model.HospitalizationPlan) []hospitalizationPlanResponse {
	result := make([]hospitalizationPlanResponse, 0, len(plans))
	for i := range plans {
		result = append(result, toHospitalizationPlanResponse(&plans[i]))
	}
	return result
}

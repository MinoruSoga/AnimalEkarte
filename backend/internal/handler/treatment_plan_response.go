package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type treatmentPlanResponse struct {
	ID                string    `json:"id"`
	MedicalRecordID   *string   `json:"medical_record_id,omitempty"`
	HospitalizationID *string   `json:"hospitalization_id,omitempty"`
	TreatmentContent  string    `json:"treatment_content"`
	Memo              string    `json:"memo"`
	Insurance         bool      `json:"insurance"`
	UnitPrice         float64   `json:"unit_price"`
	Quantity          int       `json:"quantity"`
	DiscountRate      float64   `json:"discount_rate"`
	DiscountAmount    float64   `json:"discount_amount"`
	Subtotal          float64   `json:"subtotal"`
	SortOrder         int       `json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toTreatmentPlanResponse(p *model.TreatmentPlan) treatmentPlanResponse {
	r := treatmentPlanResponse{
		ID:               strconv.FormatUint(p.ID, 10),
		TreatmentContent: p.TreatmentContent,
		Memo:             p.Memo,
		Insurance:        p.Insurance,
		UnitPrice:        p.UnitPrice,
		Quantity:         p.Quantity,
		DiscountRate:     p.DiscountRate,
		DiscountAmount:   p.DiscountAmount,
		Subtotal:         p.Subtotal,
		SortOrder:        p.SortOrder,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
	if p.MedicalRecordID != nil {
		s := strconv.FormatUint(*p.MedicalRecordID, 10)
		r.MedicalRecordID = &s
	}
	if p.HospitalizationID != nil {
		s := strconv.FormatUint(*p.HospitalizationID, 10)
		r.HospitalizationID = &s
	}
	return r
}

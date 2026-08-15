package medicalrecord

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// TreatmentPlanResponse is the treatment-plan list/create/update HTTP wire DTO
// (domain-owned). TASK-444-S2: tygo source for frontend hospitalization-responses.ts.
// ID / FK fields are string on the wire (uint64 formatted) — not models.TreatmentPlan numbers.
// Nested under GET /hospitalizations/:id/treatment-plans (or medical-records/...); not embedded
// on HospitalizationResponse / MedicalRecordResponse detail payloads.
type TreatmentPlanResponse struct {
	ID                string    `json:"id"`
	MedicalRecordID   *string   `json:"medical_record_id,omitempty"`
	HospitalizationID *string   `json:"hospitalization_id,omitempty"`
	TreatmentContent  string    `json:"treatment_content"`
	Memo              string    `json:"memo"`
	IsInsurance       bool      `json:"is_insurance"`
	UnitPrice         int64     `json:"unit_price"`
	Quantity          float64   `json:"quantity"`
	DiscountRate      float64   `json:"discount_rate"`
	DiscountAmount    int64     `json:"discount_amount"`
	Subtotal          int64     `json:"subtotal"`
	SortOrder         int       `json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toTreatmentPlanResponse(p *model.TreatmentPlan) TreatmentPlanResponse {
	r := TreatmentPlanResponse{
		ID:               strconv.FormatUint(p.ID, 10),
		TreatmentContent: p.TreatmentContent,
		Memo:             p.Memo,
		IsInsurance:      p.IsInsurance,
		UnitPrice:        p.UnitPrice,
		Quantity:         p.Quantity,
		DiscountRate:     p.DiscountRate,
		DiscountAmount:   p.DiscountAmount,
		Subtotal:         p.Subtotal,
		SortOrder:        p.SortOrder,
		CreatedAt:        httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(p.UpdatedAt),
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

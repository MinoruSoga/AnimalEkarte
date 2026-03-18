package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type estimateItemResponse struct {
	ID                    uint64    `json:"id"`
	EstimateID            uint64    `json:"estimate_id"`
	Name                  string    `json:"name"`
	Category              string    `json:"category"`
	UnitPrice             float64   `json:"unit_price"`
	Quantity              float64   `json:"quantity"`
	TaxRate               float64   `json:"tax_rate"`
	DiscountRate          float64   `json:"discount_rate"`
	DiscountAmount        float64   `json:"discount_amount"`
	IsInsuranceApplicable bool      `json:"is_insurance_applicable"`
	SortOrder             int       `json:"sort_order"`
	CreatedAt             time.Time `json:"created_at"`
}

type estimateResponse struct {
	ID              uint64                 `json:"id"`
	ClinicID        uint64                 `json:"clinic_id"`
	EstimateNo      string                 `json:"estimate_no"`
	MedicalRecordID *uint64                `json:"medical_record_id,omitempty"`
	Title           string                 `json:"title"`
	OwnerID         *uint64                `json:"owner_id,omitempty"`
	Owner           *ownerSummaryResponse  `json:"owner,omitempty"`
	Status          string                 `json:"status"`
	Subtotal        float64                `json:"subtotal"`
	TaxTotal        float64                `json:"tax_total"`
	TotalAmount     float64                `json:"total_amount"`
	InsuranceAmount float64                `json:"insurance_amount"`
	DiscountAmount  float64                `json:"discount_amount"`
	ValidUntil      *time.Time             `json:"valid_until,omitempty"`
	Comment         string                 `json:"comment"`
	Notes           string                 `json:"notes"`
	CreatedBy       *uint64                `json:"created_by,omitempty"`
	Items           []estimateItemResponse `json:"items,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func toEstimateItemResponse(item *model.EstimateItem) estimateItemResponse {
	return estimateItemResponse{
		ID:                    item.ID,
		EstimateID:            item.EstimateID,
		Name:                  item.Name,
		Category:              string(item.Category),
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		TaxRate:               item.TaxRate,
		DiscountRate:          item.DiscountRate,
		DiscountAmount:        item.DiscountAmount,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		SortOrder:             item.SortOrder,
		CreatedAt:             item.CreatedAt,
	}
}

func toEstimateResponse(e *model.Estimate) estimateResponse {
	items := make([]estimateItemResponse, 0, len(e.Items))
	for i := range e.Items {
		items = append(items, toEstimateItemResponse(&e.Items[i]))
	}
	return estimateResponse{
		ID:              e.ID,
		ClinicID:        e.ClinicID,
		EstimateNo:      e.EstimateNo,
		MedicalRecordID: e.MedicalRecordID,
		Title:           e.Title,
		OwnerID:         e.OwnerID,
		Owner:           toOwnerSummary(e.Owner),
		Status:          string(e.Status),
		Subtotal:        e.Subtotal,
		TaxTotal:        e.TaxTotal,
		TotalAmount:     e.TotalAmount,
		InsuranceAmount: e.InsuranceAmount,
		DiscountAmount:  e.DiscountAmount,
		ValidUntil:      e.ValidUntil,
		Comment:         e.Comment,
		Notes:           e.Notes,
		CreatedBy:       e.CreatedBy,
		Items:           items,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func toEstimateResponseList(estimates []model.Estimate) []estimateResponse {
	result := make([]estimateResponse, 0, len(estimates))
	for i := range estimates {
		result = append(result, toEstimateResponse(&estimates[i]))
	}
	return result
}

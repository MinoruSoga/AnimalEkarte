package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type estimateItemResponse struct {
	ID                    uint64    `json:"id"`
	EstimateID            uint64    `json:"estimate_id"`
	Name                  string    `json:"name"`
	Category              string    `json:"category"`
	UnitPrice             int64     `json:"unit_price"`
	Quantity              float64   `json:"quantity"`
	Subtotal              int64     `json:"subtotal"`
	TaxType               string    `json:"tax_type"`
	TaxRate               float64   `json:"tax_rate"`
	TaxAmount             int64     `json:"tax_amount"`
	DiscountRate          float64   `json:"discount_rate"`
	DiscountAmount        int64     `json:"discount_amount"`
	IsInsuranceApplicable bool      `json:"is_insurance_applicable"`
	SortOrder             int       `json:"sort_order"`
	CreatedAt             time.Time `json:"created_at"`
}

type estimateResponse struct {
	ID                   uint64                 `json:"id"`
	ClinicID             uint64                 `json:"clinic_id"`
	EstimateNo           string                 `json:"estimate_no"`
	MedicalRecordID      *uint64                `json:"medical_record_id,omitempty"`
	Title                string                 `json:"title"`
	OwnerID              *uint64                `json:"owner_id,omitempty"`
	Owner                *ownerSummaryResponse  `json:"owner,omitempty"`
	PetID                *uint64                `json:"pet_id,omitempty"`
	Pet                  *petSummaryResponse    `json:"pet,omitempty"`
	Status               string                 `json:"status"`
	Subtotal             int64                  `json:"subtotal"`
	TaxTotal             int64                  `json:"tax_total"`
	TotalAmount          int64                  `json:"total_amount"`
	InsuranceAmount      int64                  `json:"insurance_amount"`
	DiscountAmount       int64                  `json:"discount_amount"`
	ValidUntil           *time.Time             `json:"valid_until,omitempty"`
	Comment              string                 `json:"comment"`
	Notes                string                 `json:"notes"`
	CreatedBy            *uint64                `json:"created_by,omitempty"`
	SupersedesEstimateID *uint64                `json:"supersedes_estimate_id,omitempty"`
	Items                []estimateItemResponse `json:"items,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

func toEstimateItemResponse(item *model.EstimateItem) estimateItemResponse {
	// MDL-01: item subtotal is post-discount, matching BillingItem / CalculateTaxAmount base.
	subtotal := max(int64(float64(item.UnitPrice)*item.Quantity)-item.DiscountAmount, 0)
	taxAmount := item.CalculateTaxAmount()
	return estimateItemResponse{
		ID:                    item.ID,
		EstimateID:            item.EstimateID,
		Name:                  item.Name,
		Category:              string(item.Category),
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		Subtotal:              subtotal,
		TaxType:               string(item.TaxType),
		TaxRate:               item.TaxRate,
		TaxAmount:             taxAmount,
		DiscountRate:          item.DiscountRate,
		DiscountAmount:        item.DiscountAmount,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		SortOrder:             item.SortOrder,
		CreatedAt:             httpapi.LocalTime(item.CreatedAt),
	}
}

func toEstimateResponse(e *model.Estimate) estimateResponse {
	items := make([]estimateItemResponse, 0, len(e.Items))
	for i := range e.Items {
		items = append(items, toEstimateItemResponse(&e.Items[i]))
	}
	return estimateResponse{
		ID:                   e.ID,
		ClinicID:             e.ClinicID,
		EstimateNo:           e.EstimateNo,
		MedicalRecordID:      e.MedicalRecordID,
		Title:                e.Title,
		OwnerID:              e.OwnerID,
		Owner:                toOwnerSummary(e.Owner),
		PetID:                e.PetID,
		Pet:                  toPetSummary(e.Pet),
		Status:               string(e.Status),
		Subtotal:             e.Subtotal,
		TaxTotal:             e.TaxTotal,
		TotalAmount:          e.TotalAmount,
		InsuranceAmount:      e.InsuranceAmount,
		DiscountAmount:       e.DiscountAmount,
		ValidUntil:           e.ValidUntil,
		Comment:              e.Comment,
		Notes:                e.Notes,
		CreatedBy:            e.CreatedBy,
		SupersedesEstimateID: e.SupersedesEstimateID,
		Items:                items,
		CreatedAt:            httpapi.LocalTime(e.CreatedAt),
		UpdatedAt:            httpapi.LocalTime(e.UpdatedAt),
	}
}

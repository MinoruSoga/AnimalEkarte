package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type billingItemResponse struct {
	ID                    uint64    `json:"id"`
	BillingID             uint64    `json:"billing_id"`
	Category              string    `json:"category"`
	Name                  string    `json:"name"`
	UnitPrice             float64   `json:"unit_price"`
	Quantity              float64   `json:"quantity"`
	TaxRate               float64   `json:"tax_rate"`
	IsInsuranceApplicable bool      `json:"is_insurance_applicable"`
	Source                string    `json:"source"`
	SortOrder             int       `json:"sort_order"`
	CreatedAt             time.Time `json:"created_at"`
}

type paymentResponse struct {
	ID              uint64    `json:"id"`
	BillingID       uint64    `json:"billing_id"`
	Subtotal        float64   `json:"subtotal"`
	TaxTotal        float64   `json:"tax_total"`
	TotalAmount     float64   `json:"total_amount"`
	InsuranceName   string    `json:"insurance_name"`
	InsuranceRatio  float64   `json:"insurance_ratio"`
	InsuranceAmount float64   `json:"insurance_amount"`
	DiscountAmount  float64   `json:"discount_amount"`
	BillingAmount   float64   `json:"billing_amount"`
	ReceivedAmount  float64   `json:"received_amount"`
	ChangeAmount    float64   `json:"change_amount"`
	Method          string    `json:"method"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type accountingResponse struct {
	ID                uint64                `json:"id"`
	ClinicID          uint64                `json:"clinic_id"`
	MedicalRecordID   *uint64               `json:"medical_record_id,omitempty"`
	HospitalizationID *uint64               `json:"hospitalization_id,omitempty"`
	OwnerID           *uint64               `json:"owner_id,omitempty"`
	PetID             *uint64               `json:"pet_id,omitempty"`
	Subtotal          int                   `json:"subtotal"`
	TaxTotal          int                   `json:"tax_total"`
	TotalAmount       int                   `json:"total_amount"`
	HasInsurance      bool                  `json:"has_insurance"`
	Status            string                `json:"status"`
	ScheduledDate     time.Time             `json:"scheduled_date"`
	CompletedAt       *time.Time            `json:"completed_at,omitempty"`
	Memo              string                `json:"memo"`
	Items             []billingItemResponse `json:"items,omitempty"`
	Payments          []paymentResponse     `json:"payments,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

func toBillingItemResponse(item *model.BillingItem) billingItemResponse {
	return billingItemResponse{
		ID:                    item.ID,
		BillingID:             item.BillingID,
		Category:              string(item.Category),
		Name:                  item.Name,
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		TaxRate:               item.TaxRate,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		Source:                string(item.Source),
		SortOrder:             item.SortOrder,
		CreatedAt:             item.CreatedAt,
	}
}

func toPaymentResponse(p *model.Payment) paymentResponse {
	return paymentResponse{
		ID:              p.ID,
		BillingID:       p.BillingID,
		Subtotal:        p.Subtotal,
		TaxTotal:        p.TaxTotal,
		TotalAmount:     p.TotalAmount,
		InsuranceName:   p.InsuranceName,
		InsuranceRatio:  p.InsuranceRatio,
		InsuranceAmount: p.InsuranceAmount,
		DiscountAmount:  p.DiscountAmount,
		BillingAmount:   p.BillingAmount,
		ReceivedAmount:  p.ReceivedAmount,
		ChangeAmount:    p.ChangeAmount,
		Method:          string(p.Method),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func toAccountingResponse(b *model.Billing) accountingResponse {
	items := make([]billingItemResponse, 0, len(b.Items))
	for i := range b.Items {
		items = append(items, toBillingItemResponse(&b.Items[i]))
	}
	payments := make([]paymentResponse, 0, len(b.Payments))
	for i := range b.Payments {
		payments = append(payments, toPaymentResponse(&b.Payments[i]))
	}
	return accountingResponse{
		ID:                b.ID,
		ClinicID:          b.ClinicID,
		MedicalRecordID:   b.MedicalRecordID,
		HospitalizationID: b.HospitalizationID,
		OwnerID:           b.OwnerID,
		PetID:             b.PetID,
		Subtotal:          b.Subtotal,
		TaxTotal:          b.TaxTotal,
		TotalAmount:       b.TotalAmount,
		HasInsurance:      b.HasInsurance,
		Status:            string(b.Status),
		ScheduledDate:     b.ScheduledDate,
		CompletedAt:       b.CompletedAt,
		Memo:              b.Memo,
		Items:             items,
		Payments:          payments,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
}

func toAccountingResponseList(billings []model.Billing) []accountingResponse {
	result := make([]accountingResponse, 0, len(billings))
	for i := range billings {
		result = append(result, toAccountingResponse(&billings[i]))
	}
	return result
}

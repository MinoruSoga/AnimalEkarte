package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type billingItemResponse struct {
	ID                    uint64    `json:"id"`
	BillingID             uint64    `json:"billing_id"`
	Category              string    `json:"category"`
	Name                  string    `json:"name"`
	UnitPrice             int64     `json:"unit_price"`
	Quantity              float64   `json:"quantity"`
	Subtotal              int64     `json:"subtotal"`
	TaxType               string    `json:"tax_type"`
	TaxRate               float64   `json:"tax_rate"`
	TaxAmount             int64     `json:"tax_amount"`
	IsInsuranceApplicable bool      `json:"is_insurance_applicable"`
	Source                string    `json:"source"`
	SortOrder             int       `json:"sort_order"`
	CreatedAt             time.Time `json:"created_at"`
}

type paymentResponse struct {
	ID              uint64    `json:"id"`
	BillingID       uint64    `json:"billing_id"`
	Subtotal        int64     `json:"subtotal"`
	TaxTotal        int64     `json:"tax_total"`
	TotalAmount     int64     `json:"total_amount"`
	InsuranceName   string    `json:"insurance_name"`
	InsuranceRatio  float64   `json:"insurance_ratio"`
	InsuranceAmount int64     `json:"insurance_amount"`
	DiscountAmount  int64     `json:"discount_amount"`
	BillingAmount   int64     `json:"billing_amount"`
	ReceivedAmount  int64     `json:"received_amount"`
	ChangeAmount    int64     `json:"change_amount"`
	Method          string    `json:"method"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type accountingOwnerSummary struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"owner_name"`
}

type accountingPetSummary struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type accountingResponse struct {
	ID                uint64                  `json:"id"`
	ClinicID          uint64                  `json:"clinic_id"`
	MedicalRecordID   *uint64                 `json:"medical_record_id,omitempty"`
	HospitalizationID *uint64                 `json:"hospitalization_id,omitempty"`
	OwnerID           *uint64                 `json:"owner_id,omitempty"`
	PetID             *uint64                 `json:"pet_id,omitempty"`
	Owner             *accountingOwnerSummary `json:"owner,omitempty"`
	Pet               *accountingPetSummary   `json:"pet,omitempty"`
	Subtotal          int                     `json:"subtotal"`
	TaxTotal          int                     `json:"tax_total"`
	TotalAmount       int                     `json:"total_amount"`
	HasInsurance      bool                    `json:"has_insurance"`
	Status            string                  `json:"status"`
	ScheduledDate     time.Time               `json:"scheduled_date"`
	CompletedAt       *time.Time              `json:"completed_at,omitempty"`
	Memo              string                  `json:"memo"`
	Items             []billingItemResponse   `json:"items,omitempty"`
	Payments          []paymentResponse       `json:"payments,omitempty"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
}

func toBillingItemResponse(item *model.BillingItem) billingItemResponse {
	subtotal := int64(float64(item.UnitPrice) * item.Quantity)
	taxAmount := service.CalculateTaxAmount(item.UnitPrice, item.Quantity, item.TaxType, item.TaxRate)
	return billingItemResponse{
		ID:                    item.ID,
		BillingID:             item.BillingID,
		Category:              string(item.Category),
		Name:                  item.Name,
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		Subtotal:              subtotal,
		TaxType:               string(item.TaxType),
		TaxRate:               item.TaxRate,
		TaxAmount:             taxAmount,
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

	var owner *accountingOwnerSummary
	if b.Owner != nil {
		owner = &accountingOwnerSummary{
			ID:        b.Owner.ID,
			OwnerName: b.Owner.OwnerName,
		}
	}

	var pet *accountingPetSummary
	if b.Pet != nil {
		pet = &accountingPetSummary{
			ID:   b.Pet.ID,
			Name: b.Pet.Name,
		}
	}

	return accountingResponse{
		ID:                b.ID,
		ClinicID:          b.ClinicID,
		MedicalRecordID:   b.MedicalRecordID,
		HospitalizationID: b.HospitalizationID,
		OwnerID:           b.OwnerID,
		PetID:             b.PetID,
		Owner:             owner,
		Pet:               pet,
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

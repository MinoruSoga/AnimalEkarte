package handler

import "time"

// createAccountingRequest は会計作成リクエスト。
type createAccountingRequest struct {
	MedicalRecordID   *uint64    `json:"medical_record_id"`
	HospitalizationID *uint64    `json:"hospitalization_id"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	Subtotal          int        `json:"subtotal"`
	TaxTotal          int        `json:"tax_total"`
	TotalAmount       int        `json:"total_amount"`
	HasInsurance      bool       `json:"has_insurance"`
	Status            string     `json:"status"`
	ScheduledDate     time.Time  `json:"scheduled_date" binding:"required"`
	CompletedAt       *time.Time `json:"completed_at"`
	Memo              string     `json:"memo"`
}

// updateAccountingRequest は会計更新リクエスト。
type updateAccountingRequest struct {
	MedicalRecordID   *uint64    `json:"medical_record_id"`
	HospitalizationID *uint64    `json:"hospitalization_id"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	Subtotal          int        `json:"subtotal"`
	TaxTotal          int        `json:"tax_total"`
	TotalAmount       int        `json:"total_amount"`
	HasInsurance      bool       `json:"has_insurance"`
	Status            string     `json:"status"`
	ScheduledDate     *time.Time `json:"scheduled_date"`
	CompletedAt       *time.Time `json:"completed_at"`
	Memo              string     `json:"memo"`
}

// createBillingItemRequest は明細作成リクエスト。
type createBillingItemRequest struct {
	BillingID             uint64  `json:"billing_id" binding:"required"`
	Category              string  `json:"category"`
	Name                  string  `json:"name" binding:"required"`
	UnitPrice             int64   `json:"unit_price"`
	Quantity              float64 `json:"quantity"`
	TaxType               string  `json:"tax_type"`
	TaxRate               float64 `json:"tax_rate"`
	IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
	Source                string  `json:"source"`
	SortOrder             int     `json:"sort_order"`
}

// updateBillingItemRequest は明細更新リクエスト（nil = 未指定）。
type updateBillingItemRequest struct {
	UnitPrice             *int64   `json:"unit_price"`
	Quantity              *float64 `json:"quantity"`
	TaxType               *string  `json:"tax_type"`
	TaxRate               *float64 `json:"tax_rate"`
	IsInsuranceApplicable *bool    `json:"is_insurance_applicable"`
}

// createRefundRequest は返金登録リクエスト。
type createRefundRequest struct {
	Amount int64  `json:"amount" binding:"required,min=1"`
	Reason string `json:"reason"`
}

package handler

import "time"

// createAccountingRequest は会計作成リクエスト。
type createAccountingRequest struct {
	MedicalRecordID   *uint64    `json:"medical_record_id"`
	HospitalizationID *uint64    `json:"hospitalization_id"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	Subtotal          int64      `json:"subtotal"      binding:"min=0"`
	TaxTotal          int64      `json:"tax_total"     binding:"min=0"`
	TotalAmount       int64      `json:"total_amount"  binding:"min=0"`
	HasInsurance      bool       `json:"has_insurance"`
	Status            string     `json:"status"        binding:"omitempty,oneof=waiting pending completed cancelled"`
	ScheduledDate     time.Time  `json:"scheduled_date" binding:"required"`
	CompletedAt       *time.Time `json:"completed_at"`
	Memo              string     `json:"memo"`
}

// updateAccountingRequest は会計更新リクエスト。
// nil フィールドは更新しない（GORM ゼロ値スキップ問題を回避するためポインタ型を使用）。
// Payment フィールドが含まれている場合、会計完了時に Payment を同時 upsert する。
type updateAccountingRequest struct {
	MedicalRecordID   *uint64    `json:"medical_record_id"`
	HospitalizationID *uint64    `json:"hospitalization_id"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	Subtotal          *int64     `json:"subtotal"`
	TaxTotal          *int64     `json:"tax_total"`
	TotalAmount       *int64     `json:"total_amount"`
	HasInsurance      *bool      `json:"has_insurance"`
	Status            *string    `json:"status"        binding:"omitempty,oneof=waiting pending completed cancelled"`
	ScheduledDate     *time.Time `json:"scheduled_date"`
	CompletedAt       *time.Time `json:"completed_at"`
	Memo              *string    `json:"memo"`
	// Payment フィールド（会計完了時に同時送信される）
	PaymentMethod   *string  `json:"payment_method"  binding:"omitempty,oneof=cash credit_card electronic_money"`
	InsuranceRatio  *float64 `json:"insurance_ratio"`
	InsuranceName   *string  `json:"insurance_name"`
	InsuranceAmount *int64   `json:"insurance_amount"`
	DiscountAmount  *int64   `json:"discount_amount"`
	BillingAmount   *int64   `json:"billing_amount"`
	ReceivedAmount  *int64   `json:"received_amount"`
	ChangeAmount    *int64   `json:"change_amount"`
}

// createBillingItemRequest は明細作成リクエスト。
type createBillingItemRequest struct {
	BillingID             uint64  `json:"billing_id" binding:"required"`
	Category              string  `json:"category"  binding:"omitempty,oneof=examination test procedure surgery medicine food goods other"`
	Name                  string  `json:"name"      binding:"required"`
	UnitPrice             int64   `json:"unit_price" binding:"min=0"`
	Quantity              float64 `json:"quantity"   binding:"min=0"`
	TaxType               string  `json:"tax_type"  binding:"omitempty,oneof=included excluded exempt"`
	TaxRate               float64 `json:"tax_rate"`
	IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
	Source                string  `json:"source"    binding:"omitempty,oneof=medical_record manual hospitalization"`
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

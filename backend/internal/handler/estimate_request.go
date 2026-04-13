package handler

import "time"

// createEstimateRequest は見積書作成リクエスト
type createEstimateRequest struct {
	MedicalRecordID *uint64    `json:"medical_record_id"`
	Title           string     `json:"title" binding:"required,min=1,max=255"`
	OwnerID         *uint64    `json:"owner_id"`
	Status          string     `json:"status"`
	Subtotal        int64      `json:"subtotal"      binding:"min=0"`
	TaxTotal        int64      `json:"tax_total"     binding:"min=0"`
	TotalAmount     int64      `json:"total_amount"  binding:"min=0"`
	InsuranceAmount int64      `json:"insurance_amount"`
	DiscountAmount  int64      `json:"discount_amount"`
	ValidUntil      *time.Time `json:"valid_until"`
	Comment         string     `json:"comment"`
	Notes           string     `json:"notes"`
	CreatedBy       *uint64    `json:"created_by"`
}

// updateEstimateRequest は見積書更新リクエスト（PATCH: nil = 未送信）
type updateEstimateRequest struct {
	Title           *string    `json:"title"`
	Status          *string    `json:"status"`
	Subtotal        *int64     `json:"subtotal"`
	TaxTotal        *int64     `json:"tax_total"`
	TotalAmount     *int64     `json:"total_amount"`
	InsuranceAmount *int64     `json:"insurance_amount"`
	DiscountAmount  *int64     `json:"discount_amount"`
	ValidUntil      *time.Time `json:"valid_until"`
	ClearValidUntil bool       `json:"clear_valid_until"`
	Comment         *string    `json:"comment"`
	Notes           *string    `json:"notes"`
}

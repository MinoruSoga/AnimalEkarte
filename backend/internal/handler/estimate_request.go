package handler

import "time"

// createEstimateRequest は見積書作成リクエスト
type createEstimateRequest struct {
	MedicalRecordID *uint64    `json:"medical_record_id"`
	Title           string     `json:"title" binding:"required"`
	OwnerID         *uint64    `json:"owner_id"`
	Status          string     `json:"status"`
	Subtotal        float64    `json:"subtotal"`
	TaxTotal        float64    `json:"tax_total"`
	TotalAmount     float64    `json:"total_amount"`
	InsuranceAmount float64    `json:"insurance_amount"`
	DiscountAmount  float64    `json:"discount_amount"`
	ValidUntil      *time.Time `json:"valid_until"`
	Comment         string     `json:"comment"`
	Notes           string     `json:"notes"`
	CreatedBy       *uint64    `json:"created_by"`
}

// updateEstimateRequest は見積書更新リクエスト（PATCH: nil = 未送信）
type updateEstimateRequest struct {
	Title           *string    `json:"title"`
	Status          *string    `json:"status"`
	Subtotal        *float64   `json:"subtotal"`
	TaxTotal        *float64   `json:"tax_total"`
	TotalAmount     *float64   `json:"total_amount"`
	InsuranceAmount *float64   `json:"insurance_amount"`
	DiscountAmount  *float64   `json:"discount_amount"`
	ValidUntil      *time.Time `json:"valid_until"`
	ClearValidUntil bool       `json:"clear_valid_until"`
	Comment         *string    `json:"comment"`
	Notes           *string    `json:"notes"`
}

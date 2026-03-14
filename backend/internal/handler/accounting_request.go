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

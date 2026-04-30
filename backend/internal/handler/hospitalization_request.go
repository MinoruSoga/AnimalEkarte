package handler

import "time"

// dischargeWithBillingRequest は退院+会計作成のバインド struct
type dischargeWithBillingRequest struct {
	DischargeDate    time.Time `json:"discharge_date" binding:"required"`
	CreateAccounting bool      `json:"create_accounting"`
}

// createHospitalizationRequest は入院作成のバインド struct
type createHospitalizationRequest struct {
	OwnerID              uint64    `json:"owner_id"               binding:"required"`
	PetID                uint64    `json:"pet_id"                 binding:"required"`
	HospitalizationType  string    `json:"hospitalization_type"   binding:"required,oneof=hospitalization hotel"`
	StartDate            time.Time `json:"start_date"             binding:"required"`
	EndDate              time.Time `json:"end_date"               binding:"required"`
	Status               string    `json:"status"                 binding:"omitempty,oneof=admitted discharged reserved"`
	CageID               *uint64   `json:"cage_id"`
	DoctorID             *uint64   `json:"doctor_id"`
	Memo                 string    `json:"memo"`
	OwnerRequest         string    `json:"owner_request"`
	StaffNotes           string    `json:"staff_notes"`
	IsInsurance          bool      `json:"is_insurance"`
	InsuranceCompanyName *string   `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string   `json:"insurance_number,omitempty"`
}

// updateHospitalizationRequest は入院更新のバインド struct
type updateHospitalizationRequest struct {
	OwnerID              *uint64    `json:"owner_id"`
	PetID                *uint64    `json:"pet_id"`
	HospitalizationType  *string    `json:"hospitalization_type"   binding:"omitempty,oneof=hospitalization hotel"`
	StartDate            *time.Time `json:"start_date"`
	EndDate              *time.Time `json:"end_date"`
	Status               *string    `json:"status"                 binding:"omitempty,oneof=admitted discharged reserved"`
	CageID               *uint64    `json:"cage_id"`
	DoctorID             *uint64    `json:"doctor_id"`
	Memo                 *string    `json:"memo"`
	OwnerRequest         *string    `json:"owner_request"`
	StaffNotes           *string    `json:"staff_notes"`
	IsInsurance          *bool      `json:"is_insurance,omitempty"`
	InsuranceCompanyName *string    `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string    `json:"insurance_number,omitempty"`
}

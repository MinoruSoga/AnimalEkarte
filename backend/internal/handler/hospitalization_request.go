package handler

import "time"

// createHospitalizationRequest は入院作成のバインド struct
type createHospitalizationRequest struct {
	OwnerID             uint64    `json:"owner_id"              binding:"required"`
	PetID               uint64    `json:"pet_id"                binding:"required"`
	HospitalizationType string    `json:"hospitalization_type"  binding:"required"`
	StartDate           time.Time `json:"start_date"            binding:"required"`
	EndDate             time.Time `json:"end_date"              binding:"required"`
	Status              string    `json:"status"`
	CageID              *uint64   `json:"cage_id"`
	DoctorID            *uint64   `json:"doctor_id"`
	Memo                string    `json:"memo"`
	OwnerRequest        string    `json:"owner_request"`
	StaffNotes          string    `json:"staff_notes"`
}

// updateHospitalizationRequest は入院更新のバインド struct
type updateHospitalizationRequest struct {
	OwnerID             *uint64    `json:"owner_id"`
	PetID               *uint64    `json:"pet_id"`
	HospitalizationType *string    `json:"hospitalization_type"`
	StartDate           *time.Time `json:"start_date"`
	EndDate             *time.Time `json:"end_date"`
	Status              *string    `json:"status"`
	CageID              *uint64    `json:"cage_id"`
	DoctorID            *uint64    `json:"doctor_id"`
	Memo                *string    `json:"memo"`
	OwnerRequest        *string    `json:"owner_request"`
	StaffNotes          *string    `json:"staff_notes"`
}

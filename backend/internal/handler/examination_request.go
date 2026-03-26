package handler

import "time"

// createExaminationRequest は検査作成のバインド struct
type createExaminationRequest struct {
	MedicalRecordID *uint64   `json:"medical_record_id"`
	PetID           *uint64   `json:"pet_id"`
	ExamTypeID      uint64    `json:"exam_type_id"      binding:"required"`
	DoctorID        *uint64   `json:"doctor_id"`
	Date            time.Time `json:"date"              binding:"required"`
	ResultSummary   string    `json:"result_summary"`
	Machine         string    `json:"machine"`
	Status          string    `json:"status"`
}

// updateExaminationRequest は検査更新のバインド struct
type updateExaminationRequest struct {
	MedicalRecordID *uint64    `json:"medical_record_id"`
	PetID           *uint64    `json:"pet_id"`
	ExamTypeID      uint64     `json:"exam_type_id"`
	DoctorID        *uint64    `json:"doctor_id"`
	Date            *time.Time `json:"date"`
	ResultSummary   *string    `json:"result_summary"`
	Machine         *string    `json:"machine"`
	Status          *string    `json:"status"`
}

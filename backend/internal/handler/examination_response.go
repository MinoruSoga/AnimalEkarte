package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type examinationResponse struct {
	ID              uint64    `json:"id"`
	ClinicID        uint64    `json:"clinic_id"`
	MedicalRecordID *uint64   `json:"medical_record_id,omitempty"`
	PetID           *uint64   `json:"pet_id,omitempty"`
	ExamTypeID      uint64    `json:"exam_type_id"`
	DoctorID        *uint64   `json:"doctor_id,omitempty"`
	Date            time.Time `json:"date"`
	ResultSummary   string    `json:"result_summary"`
	Machine         string    `json:"machine"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toExaminationResponse(exam *model.Examination) examinationResponse {
	return examinationResponse{
		ID:              exam.ID,
		ClinicID:        exam.ClinicID,
		MedicalRecordID: exam.MedicalRecordID,
		PetID:           exam.PetID,
		ExamTypeID:      exam.ExamTypeID,
		DoctorID:        exam.DoctorID,
		Date:            exam.Date,
		ResultSummary:   exam.ResultSummary,
		Machine:         exam.Machine,
		Status:          string(exam.Status),
		CreatedAt:       exam.CreatedAt,
		UpdatedAt:       exam.UpdatedAt,
	}
}

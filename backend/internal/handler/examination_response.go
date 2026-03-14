package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type examItemResponse struct {
	ID              uint64    `json:"id"`
	ExamID          uint64    `json:"exam_id"`
	Name            string    `json:"name"`
	InspectionValue string    `json:"inspection_value"`
	NormalValue     string    `json:"normal_value"`
	Result          string    `json:"result"`
	Unit            string    `json:"unit"`
	Ref             string    `json:"ref"`
	Status          string    `json:"status"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
}

type examinationResponse struct {
	ID              uint64                   `json:"id"`
	MedicalRecordID uint64                   `json:"medical_record_id"`
	PetID           *uint64                  `json:"pet_id,omitempty"`
	ExamTypeID      uint64                   `json:"exam_type_id"`
	DoctorID        *uint64                  `json:"doctor_id,omitempty"`
	Date            time.Time                `json:"date"`
	ResultSummary   string                   `json:"result_summary"`
	Machine         string                   `json:"machine"`
	Status          string                   `json:"status"`
	Items           []examItemResponse       `json:"items,omitempty"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
	Pet             *petSummaryResponse      `json:"pet,omitempty"`
	ExaminationType        *examTypeSummaryResponse `json:"exam_type,omitempty"`
	Doctor          *staffSummaryResponse    `json:"doctor,omitempty"`
}

func toExamItemResponse(i *model.ExaminationItem) examItemResponse {
	return examItemResponse{
		ID:              i.ID,
		ExamID:          i.ExamID,
		Name:            i.Name,
		InspectionValue: i.InspectionValue,
		NormalValue:     i.NormalValue,
		Result:          i.Result,
		Unit:            i.Unit,
		Ref:             i.Ref,
		Status:          string(i.Status),
		SortOrder:       i.SortOrder,
		CreatedAt:       i.CreatedAt,
	}
}

func toExaminationResponse(e *model.Examination) examinationResponse {
	items := make([]examItemResponse, 0, len(e.Items))
	for i := range e.Items {
		items = append(items, toExamItemResponse(&e.Items[i]))
	}
	return examinationResponse{
		ID:              e.ID,
		MedicalRecordID: e.MedicalRecordID,
		PetID:           e.PetID,
		ExamTypeID:      e.ExamTypeID,
		DoctorID:        e.DoctorID,
		Date:            e.Date,
		ResultSummary:   e.ResultSummary,
		Machine:         e.Machine,
		Status:          string(e.Status),
		Items:           items,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		Pet:             toPetSummary(e.Pet),
		ExaminationType:        toExamTypeSummary(e.ExaminationType),
		Doctor:          toStaffSummary(e.Doctor),
	}
}

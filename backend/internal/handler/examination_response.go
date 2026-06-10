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
		Date:            localTime(exam.Date),
		ResultSummary:   exam.ResultSummary,
		Machine:         exam.Machine,
		Status:          string(exam.Status),
		CreatedAt:       localTime(exam.CreatedAt),
		UpdatedAt:       localTime(exam.UpdatedAt),
	}
}

// examResultResponse は exam_results 1 行分のレスポンス。
type examResultResponse struct {
	ID              uint64    `json:"id"`
	ExamID          uint64    `json:"exam_id"`
	ExamTypeFieldID *uint64   `json:"exam_type_field_id,omitempty"`
	Name            string    `json:"name"`
	InspectionValue string    `json:"inspection_value"`
	NormalValue     string    `json:"normal_value"`
	Unit            string    `json:"unit"`
	ReferenceValue  string    `json:"reference_value"`
	RefMin          *float64  `json:"ref_min,omitempty"`
	RefMax          *float64  `json:"ref_max,omitempty"`
	IsAbnormal      bool      `json:"is_abnormal"`
	Status          string    `json:"status"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toExamResultResponse(item *model.ExamResult) examResultResponse {
	return examResultResponse{
		ID:              item.ID,
		ExamID:          item.ExamID,
		ExamTypeFieldID: item.ExamTypeItemID,
		Name:            item.Name,
		InspectionValue: item.InspectionValue,
		NormalValue:     item.NormalValue,
		Unit:            item.Unit,
		ReferenceValue:  item.ReferenceValue,
		RefMin:          item.RefMin,
		RefMax:          item.RefMax,
		IsAbnormal:      item.IsAbnormal,
		Status:          string(item.Status),
		SortOrder:       item.SortOrder,
		CreatedAt:       localTime(item.CreatedAt),
		UpdatedAt:       localTime(item.UpdatedAt),
	}
}

// examItemsResponse は items 一括 GET / PUT のレスポンスエンベロープ。
type examItemsResponse struct {
	Items []examResultResponse `json:"items"`
}

func toExamItemsResponse(items []model.ExamResult) examItemsResponse {
	out := make([]examResultResponse, 0, len(items))
	for i := range items {
		out = append(out, toExamResultResponse(&items[i]))
	}
	return examItemsResponse{Items: out}
}

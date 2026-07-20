package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/httpapi"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// examTypeSummaryResponse は検査要約内で使用する検査種別の要約型
type examTypeSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

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
	// リレーション: 一覧で飼主名/ペット名/検査種別/担当医を表示するため。Preload 時のみ埋まる。
	// items は検査項目明細用の別 API (GET /examinations/:id/items) で取得するため、ここには含めない。
	Pet      *petSummaryResponse      `json:"pet,omitempty"`
	Doctor   *staffSummaryResponse    `json:"doctor,omitempty"`
	ExamType *examTypeSummaryResponse `json:"exam_type,omitempty"`
}

func toExaminationResponse(exam *model.Examination) examinationResponse {
	resp := examinationResponse{
		ID:              exam.ID,
		ClinicID:        exam.ClinicID,
		MedicalRecordID: exam.MedicalRecordID,
		PetID:           exam.PetID,
		ExamTypeID:      exam.ExamTypeID,
		DoctorID:        exam.DoctorID,
		Date:            httpapi.LocalTime(exam.Date),
		ResultSummary:   exam.ResultSummary,
		Machine:         exam.Machine,
		Status:          string(exam.Status),
		CreatedAt:       httpapi.LocalTime(exam.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(exam.UpdatedAt),
		Pet:             toPetSummary(exam.Pet),
		Doctor:          toStaffSummary(exam.Doctor),
	}
	if exam.ExaminationType != nil {
		resp.ExamType = &examTypeSummaryResponse{
			ID:   exam.ExaminationType.ID,
			Name: exam.ExaminationType.Name,
		}
	}
	return resp
}

// examResultResponse は exam_results 1 行分のレスポンス。
type examResultResponse struct {
	ID              uint64    `json:"id"`
	ExamID          uint64    `json:"exam_id"`
	ExamTypeFieldID *uint64   `json:"exam_type_field_id,omitempty"`
	Name            string    `json:"name"`
	InspectionValue string    `json:"inspection_value"`
	NormalValue     string    `json:"normal_value"`
	Result          string    `json:"result"`
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
		Result:          item.Result,
		Unit:            item.Unit,
		ReferenceValue:  item.ReferenceValue,
		RefMin:          item.RefMin,
		RefMax:          item.RefMax,
		IsAbnormal:      item.IsAbnormal,
		Status:          string(item.Status),
		SortOrder:       item.SortOrder,
		CreatedAt:       httpapi.LocalTime(item.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(item.UpdatedAt),
	}
}

// examItemsResponse は items 一括 GET / PUT のレスポンスエンベロープ。
type examItemsResponse struct {
	Items []examResultResponse `json:"items"`
}

func toExamItemsResponse(items []model.ExamResult) examItemsResponse {
	out := httpapi.MapSlice(items, toExamResultResponse)
	return examItemsResponse{Items: out}
}

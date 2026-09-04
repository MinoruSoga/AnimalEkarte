package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/model"
)

// examTypeSummaryResponse は検査要約内で使用する検査種別の要約型
type examTypeSummaryResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Price *int64 `json:"price,omitempty"`
}

type examinationResponse struct {
	ID                     uint64    `json:"id"`
	ClinicID               uint64    `json:"clinic_id"`
	MedicalRecordID        *uint64   `json:"medical_record_id,omitempty"`
	PetID                  *uint64   `json:"pet_id,omitempty"`
	ExamTypeID             uint64    `json:"exam_type_id"`
	DoctorID               *uint64   `json:"doctor_id,omitempty"`
	Date                   time.Time `json:"date"`
	ResultSummary          string    `json:"result_summary"`
	Machine                string    `json:"machine"`
	Status                 string    `json:"status"`
	CurrentRevisionVersion *uint64   `json:"current_revision_version,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	// リレーション: 一覧で飼主名/ペット名/検査種別/担当医を表示するため。Preload 時のみ埋まる。
	Pet      *PetSummaryResponse      `json:"pet,omitempty"`
	Doctor   *StaffSummaryResponse    `json:"doctor,omitempty"`
	ExamType *examTypeSummaryResponse `json:"exam_type,omitempty"`
	Items    *[]examResultResponse    `json:"items,omitempty"`
}

func toExaminationResponse(exam *model.Examination) examinationResponse {
	resp := examinationResponse{
		ID:                     exam.ID,
		ClinicID:               exam.ClinicID,
		MedicalRecordID:        exam.MedicalRecordID,
		PetID:                  exam.PetID,
		ExamTypeID:             exam.ExamTypeID,
		DoctorID:               exam.DoctorID,
		Date:                   httpapi.LocalTime(exam.Date),
		ResultSummary:          exam.ResultSummary,
		Machine:                exam.Machine,
		Status:                 string(exam.Status),
		CurrentRevisionVersion: exam.CurrentRevisionVersion,
		CreatedAt:              httpapi.LocalTime(exam.CreatedAt),
		UpdatedAt:              httpapi.LocalTime(exam.UpdatedAt),
		Pet:                    toPetSummary(exam.Pet),
		Doctor:                 toStaffSummary(exam.Doctor),
	}
	if exam.ExaminationType != nil {
		resp.ExamType = &examTypeSummaryResponse{
			ID:    exam.ExaminationType.ID,
			Name:  exam.ExaminationType.Name,
			Price: exam.ExaminationType.Price,
		}
	}
	return resp
}

func toExaminationResponseWithItems(exam *model.Examination) examinationResponse {
	resp := toExaminationResponse(exam)
	items := httpapi.MapSlice(exam.Items, toExamResultResponse)
	resp.Items = &items
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
	QualitativeMin  *string   `json:"qualitative_min,omitempty"`
	QualitativeMax  *string   `json:"qualitative_max,omitempty"`
	IsAssessed      bool      `json:"is_assessed"`
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
		QualitativeMin:  item.QualitativeMin,
		QualitativeMax:  item.QualitativeMax,
		IsAssessed: isExamResultAssessed(
			item.InspectionValue,
			item.RefMin,
			item.RefMax,
			item.QualitativeMin,
			item.QualitativeMax,
		),
		IsAbnormal: item.IsAbnormal,
		Status:     string(item.Status),
		SortOrder:  item.SortOrder,
		CreatedAt:  httpapi.LocalTime(item.CreatedAt),
		UpdatedAt:  httpapi.LocalTime(item.UpdatedAt),
	}
}

func isExamResultAssessed(
	inspectionValue string,
	refMin, refMax *float64,
	qualitativeMin, qualitativeMax *string,
) bool {
	return assessExamResult(
		inspectionValue,
		refMin,
		refMax,
		qualitativeMin,
		qualitativeMax,
	).isAssessed
}

// examItemsResponse は items 一括 GET / PUT のレスポンスエンベロープ。
type examItemsResponse struct {
	Items []examResultResponse `json:"items"`
}

func toExamItemsResponse(items []model.ExamResult) examItemsResponse {
	out := httpapi.MapSlice(items, toExamResultResponse)
	return examItemsResponse{Items: out}
}

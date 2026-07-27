package medicalrecord

import (
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// labExamReportSummaryResponse は GetLabJobReportSummaries の wire 形状。
// json タグは model.LabExamReportSummary から移設したもの（P7）。
type labExamReportSummaryResponse struct {
	ExamID        uint64     `json:"exam_id"`
	ClinicID      string     `json:"clinic_id"`
	JobID         *uuid.UUID `json:"job_id,omitempty"`
	Date          string     `json:"date"`
	ExamTypeName  string     `json:"exam_type_name"`
	Status        string     `json:"status"`
	ResultCount   int        `json:"result_count"`
	AbnormalCount int        `json:"abnormal_count"`
	Machine       string     `json:"machine"`
	CreatedAt     string     `json:"created_at"`
}

// labExamReportDetailResponse は GetLabExamReport の wire 形状。
// json タグは model.LabExamReportDetail から移設したもの（P7）。
type labExamReportDetailResponse struct {
	ExamID          uint64                      `json:"exam_id"`
	ClinicID        string                      `json:"clinic_id"`
	JobID           *uuid.UUID                  `json:"job_id,omitempty"`
	PetID           *uint64                     `json:"pet_id,omitempty"`
	MedicalRecordID *uint64                     `json:"medical_record_id,omitempty"`
	DoctorID        *uint64                     `json:"doctor_id,omitempty"`
	Date            string                      `json:"date"`
	ExamTypeName    string                      `json:"exam_type_name"`
	Status          string                      `json:"status"`
	Machine         string                      `json:"machine"`
	Items           []labExamResultItemResponse `json:"items"`
	CreatedAt       string                      `json:"created_at"`
	UpdatedAt       string                      `json:"updated_at"`
}

// labExamResultItemResponse は labExamReportDetailResponse.Items の wire 形状。
// json タグは model.LabExamResultItem から移設したもの（P7）。
type labExamResultItemResponse struct {
	Name            string   `json:"name"`
	InspectionValue string   `json:"inspection_value"`
	NormalValue     string   `json:"normal_value"`
	Unit            string   `json:"unit"`
	ReferenceValue  string   `json:"reference_value"`
	RefMin          *float64 `json:"ref_min,omitempty"`
	RefMax          *float64 `json:"ref_max,omitempty"`
	QualitativeMin  *string  `json:"qualitative_min,omitempty"`
	QualitativeMax  *string  `json:"qualitative_max,omitempty"`
	IsAssessed      bool     `json:"is_assessed"`
	IsAbnormal      bool     `json:"is_abnormal"`
	Status          string   `json:"status"`
	SortOrder       int      `json:"sort_order"`
}

// toLabExamReportSummaryResponse は model.LabExamReportSummary を 1:1 でコピーする（変換ロジックなし）。
func toLabExamReportSummaryResponse(s *model.LabExamReportSummary) labExamReportSummaryResponse {
	return labExamReportSummaryResponse{
		ExamID:        s.ExamID,
		ClinicID:      s.ClinicID,
		JobID:         s.JobID,
		Date:          s.Date,
		ExamTypeName:  s.ExamTypeName,
		Status:        s.Status,
		ResultCount:   s.ResultCount,
		AbnormalCount: s.AbnormalCount,
		Machine:       s.Machine,
		CreatedAt:     s.CreatedAt,
	}
}

// toLabExamReportDetailResponse は model.LabExamReportDetail を 1:1 でコピーする（変換ロジックなし）。
func toLabExamReportDetailResponse(d *model.LabExamReportDetail) labExamReportDetailResponse {
	return labExamReportDetailResponse{
		ExamID:          d.ExamID,
		ClinicID:        d.ClinicID,
		JobID:           d.JobID,
		PetID:           d.PetID,
		MedicalRecordID: d.MedicalRecordID,
		DoctorID:        d.DoctorID,
		Date:            d.Date,
		ExamTypeName:    d.ExamTypeName,
		Status:          d.Status,
		Machine:         d.Machine,
		Items:           httpapi.MapSlice(d.Items, toLabExamResultItemResponse),
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

// toLabExamResultItemResponse は保存済み基準値から評価状態を算出して wire 形状へ変換する。
func toLabExamResultItemResponse(i *model.LabExamResultItem) labExamResultItemResponse {
	return labExamResultItemResponse{
		Name:            i.Name,
		InspectionValue: i.InspectionValue,
		NormalValue:     i.NormalValue,
		Unit:            i.Unit,
		ReferenceValue:  i.ReferenceValue,
		RefMin:          i.RefMin,
		RefMax:          i.RefMax,
		QualitativeMin:  i.QualitativeMin,
		QualitativeMax:  i.QualitativeMax,
		IsAssessed: isExamResultAssessed(
			i.InspectionValue,
			i.RefMin,
			i.RefMax,
			i.QualitativeMin,
			i.QualitativeMax,
		),
		IsAbnormal: i.IsAbnormal,
		Status:     i.Status,
		SortOrder:  i.SortOrder,
	}
}

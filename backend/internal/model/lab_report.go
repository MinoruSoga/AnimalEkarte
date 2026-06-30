package model

import "github.com/google/uuid"

// LabReportFilter は LabReportQueryService のリスト/サマリクエリのフィルター入力。
// nil フィールドはフィルターなしを意味する。
// PII-safe: owner 情報・pet 名は含めない（ID のみ）。
type LabReportFilter struct {
	PetID      *uint64
	ExamTypeID *uint64
	Status     *string
	StartDate  *string // YYYY-MM-DD
	EndDate    *string // YYYY-MM-DD
	IsAbnormal *bool   // true: is_abnormal=true の exam_result を持つ exam のみ
}

// LabExamReportSummary は ListJobReportSummaries / ListExamReports の 1 件。
// result_summary は PHI 分類未確認のため除外する（Phase 4B.1 決定2）。
// owner 名・pet 名・raw 検査値は含めない。
type LabExamReportSummary struct {
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

// LabExamReportDetail は GetExamReport の詳細 DTO。
// result_summary は PHI 分類未確認のため omitempty NULL（Phase 4B.1 決定2）。
// owner 名・pet 名・raw デバイスペイロードは含めない。
type LabExamReportDetail struct {
	ExamID          uint64              `json:"exam_id"`
	ClinicID        string              `json:"clinic_id"`
	JobID           *uuid.UUID          `json:"job_id,omitempty"`
	PetID           *uint64             `json:"pet_id,omitempty"`
	MedicalRecordID *uint64             `json:"medical_record_id,omitempty"`
	DoctorID        *uint64             `json:"doctor_id,omitempty"`
	Date            string              `json:"date"`
	ExamTypeName    string              `json:"exam_type_name"`
	Status          string              `json:"status"`
	Machine         string              `json:"machine"`
	Items           []LabExamResultItem `json:"items"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

// LabExamResultItem は LabExamReportDetail の検査結果 1 件。
// 定量値・定性値・参照値を含む。raw デバイスペイロードは含めない。
type LabExamResultItem struct {
	Name            string   `json:"name"`
	InspectionValue string   `json:"inspection_value"`
	NormalValue     string   `json:"normal_value"`
	Unit            string   `json:"unit"`
	ReferenceValue  string   `json:"reference_value"`
	RefMin          *float64 `json:"ref_min,omitempty"`
	RefMax          *float64 `json:"ref_max,omitempty"`
	IsAbnormal      bool     `json:"is_abnormal"`
	Status          string   `json:"status"`
	SortOrder       int      `json:"sort_order"`
}

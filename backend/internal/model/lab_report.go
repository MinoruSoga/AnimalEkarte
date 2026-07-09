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
	ExamID        uint64
	ClinicID      string
	JobID         *uuid.UUID
	Date          string
	ExamTypeName  string
	Status        string
	ResultCount   int
	AbnormalCount int
	Machine       string
	CreatedAt     string
}

// LabExamReportDetail は GetExamReport の詳細 DTO。
// result_summary は PHI 分類未確認のため omitempty NULL（Phase 4B.1 決定2）。
// owner 名・pet 名・raw デバイスペイロードは含めない。
type LabExamReportDetail struct {
	ExamID          uint64
	ClinicID        string
	JobID           *uuid.UUID
	PetID           *uint64
	MedicalRecordID *uint64
	DoctorID        *uint64
	Date            string
	ExamTypeName    string
	Status          string
	Machine         string
	Items           []LabExamResultItem
	CreatedAt       string
	UpdatedAt       string
}

// LabExamResultItem は LabExamReportDetail の検査結果 1 件。
// 定量値・定性値・参照値を含む。raw デバイスペイロードは含めない。
type LabExamResultItem struct {
	Name            string
	InspectionValue string
	NormalValue     string
	Unit            string
	ReferenceValue  string
	RefMin          *float64
	RefMax          *float64
	IsAbnormal      bool
	Status          string
	SortOrder       int
}

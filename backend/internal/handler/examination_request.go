package handler

import "time"

// createExaminationRequest は検査作成のバインド struct
type createExaminationRequest struct {
	MedicalRecordID *uint64   `json:"medical_record_id"  binding:"omitempty"`
	PetID           *uint64   `json:"pet_id"`
	ExamTypeID      uint64    `json:"exam_type_id"       binding:"required"`
	DoctorID        *uint64   `json:"doctor_id"`
	Date            time.Time `json:"date"              binding:"required"`
	ResultSummary   string    `json:"result_summary"`
	Machine         string    `json:"machine"`
	Status          string    `json:"status"             binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
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
	Status          *string    `json:"status"             binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
}

// upsertExamItemRequest は検査項目 1 行分のバインド struct。
// status / is_abnormal は受け付けない（サーバ側で ref_min/ref_max から計算する）。
type upsertExamItemRequest struct {
	ExamTypeFieldID *uint64  `json:"exam_type_field_id"`
	Name            string   `json:"name"             binding:"required"`
	InspectionValue string   `json:"inspection_value"`
	NormalValue     string   `json:"normal_value"`
	Unit            string   `json:"unit"`
	ReferenceValue  string   `json:"reference_value"`
	RefMin          *float64 `json:"ref_min"`
	RefMax          *float64 `json:"ref_max"`
	SortOrder       int      `json:"sort_order"`
}

// replaceExamItemsRequest は検査項目の一括置換 PUT リクエスト。
// items が nil の場合は空配列として扱う（全削除と等価）。
type replaceExamItemsRequest struct {
	Items []upsertExamItemRequest `json:"items" binding:"dive"`
}

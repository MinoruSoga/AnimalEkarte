package handler

import "time"

// createMedicalRecordRequest はカルテ作成のバインド struct
// FE互換: FEから送信されるフィールドと BE期待値を統一
type createMedicalRecordRequest struct {
	// 基本フィールド
	RecordNo      string     `json:"record_no"`                   // optional; 自動生成される
	Date          *time.Time `json:"date"`                        // optional
	VisitDate     *string    `json:"visit_date"`                  // FE送信フィールド（"YYYY-MM-DD"形式）
	VisitType     string     `json:"visit_type"`                  // FE送信フィールド（無視してよい）
	OwnerID       *string    `json:"owner_id" binding:"required"` // FE送信（string）→ uint64に変換
	PetID         *string    `json:"pet_id" binding:"required"`   // FE送信（string）→ uint64に変換
	DoctorID      *string    `json:"doctor_id"`                   // FE送信（string）→ uint64に変換
	AppointmentID *string    `json:"appointment_id"`              // FE送信（string）→ uint64に変換
	Status        string     `json:"status"                       binding:"omitempty,oneof=draft finalized"`

	// BE-006: 次回来院推奨日
	NextVisitRecommendedDate *string `json:"next_visit_recommended_date"` // "YYYY-MM-DD" or null

	// ClinicalPlan関連フィールド（原子的作成用）
	ChiefComplaint       *string `json:"chief_complaint"`
	ChiefComplaintTypeID *uint64 `json:"chief_complaint_type_id"`
	Plan                 *string `json:"plan"`       // ClinicalPlan.TreatmentPolicyへ
	Assessment           *string `json:"assessment"` // ClinicalPlan.DiagnosisDetailsへ
	Notes                *string `json:"notes"`
	Diagnosis1CategoryID *uint64 `json:"diagnosis_1_category_id"`
	Diagnosis1NameID     *uint64 `json:"diagnosis_1_name_id"`
	Diagnosis2CategoryID *uint64 `json:"diagnosis_2_category_id"`
	Diagnosis2NameID     *uint64 `json:"diagnosis_2_name_id"`
}

// patchMedicalRecordRecommendationReasonRequest は受診推奨理由更新リクエスト（FEAT-381-2）。
// Reason は revisit / checkup / prevention / exam のいずれか、または "" (未設定)。
type patchMedicalRecordRecommendationReasonRequest struct {
	Reason string `json:"reason" binding:"max=100"`
}

// updateMedicalRecordRequest はカルテ更新のバインド struct
type updateMedicalRecordRequest struct {
	Date                     *time.Time `json:"date"`
	OwnerID                  *uint64    `json:"owner_id"`
	PetID                    *uint64    `json:"pet_id"`
	DoctorID                 *uint64    `json:"doctor_id"`
	AppointmentID            *uint64    `json:"appointment_id"`
	Status                   *string    `json:"status"      binding:"omitempty,oneof=draft finalized"`
	Version                  *int       `json:"version"`                     // 楽観的ロック用
	NextVisitRecommendedDate *string    `json:"next_visit_recommended_date"` // "YYYY-MM-DD" or null
}

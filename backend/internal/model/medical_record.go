package model

import (
	"time"

	"github.com/google/uuid"
)

// MedicalRecord 電子カルテモデル
type MedicalRecord struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	RecordNo       string     `json:"record_no" gorm:"type:varchar(20);uniqueIndex:idx_mr_record_no"`
	PetID          uuid.UUID  `json:"pet_id" gorm:"type:uuid;not null;index:idx_mr_pet_id"`
	OwnerID        uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	DoctorID       *uuid.UUID `json:"doctor_id" gorm:"type:uuid"`
	VisitDate      time.Time  `json:"visit_date" gorm:"index:idx_mr_visit_date"`
	VisitType      string     `json:"visit_type" gorm:"type:varchar(10)"` // 初診, 再診
	ChiefComplaint string     `json:"chief_complaint" gorm:"type:text"`
	Subjective     string     `json:"subjective" gorm:"type:text"`    // S: 主観的情報
	Objective      string     `json:"objective" gorm:"type:text"`     // O: 客観的情報
	Assessment     string     `json:"assessment" gorm:"type:text"`    // A: 評価
	Plan           string     `json:"plan" gorm:"type:text"`          // P: 計画
	SurgeryNotes   string     `json:"surgery_notes" gorm:"type:text"` // S: 手術・特記事項
	Diagnosis      string     `json:"diagnosis" gorm:"type:text"`
	Treatment      string     `json:"treatment" gorm:"type:text"`
	Prescription   string     `json:"prescription" gorm:"type:text"`
	Notes          string     `json:"notes" gorm:"type:text"`
	Status         string     `json:"status" gorm:"type:varchar(10);default:'作成中'"` // 作成中, 確定済
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Pet   *Pet   `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner *Owner `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
}

// PaginatedMedicalRecords ページングされたカルテ一覧レスポンス
type PaginatedMedicalRecords struct {
	Records     []MedicalRecord `json:"records"`
	CurrentPage int             `json:"current_page"`
	PerPage     int             `json:"per_page"`
	Total       int64           `json:"total"`
	TotalPages  int             `json:"total_pages"`
	HasNext     bool            `json:"has_next"`
	HasPrev     bool            `json:"has_prev"`
}

// TableName テーブル名を指定
func (MedicalRecord) TableName() string {
	return "medical_records"
}

// CreateMedicalRecordRequest カルテ作成リクエスト
type CreateMedicalRecordRequest struct {
	PetID          string `json:"pet_id" binding:"required"`
	OwnerID        string `json:"owner_id" binding:"required"`
	DoctorID       string `json:"doctor_id"`
	VisitDate      string `json:"visit_date" binding:"required"`
	VisitType      string `json:"visit_type"`
	ChiefComplaint string `json:"chief_complaint"`
	Subjective     string `json:"subjective"`
	Objective      string `json:"objective"`
	Assessment     string `json:"assessment"`
	Plan           string `json:"plan"`
	SurgeryNotes   string `json:"surgery_notes"`
	Diagnosis      string `json:"diagnosis"`
	Treatment      string `json:"treatment"`
	Prescription   string `json:"prescription"`
	Notes          string `json:"notes"`
	Status         string `json:"status"`
}

// UpdateMedicalRecordRequest カルテ更新リクエスト
type UpdateMedicalRecordRequest struct {
	PetID          *string `json:"pet_id"`
	OwnerID        *string `json:"owner_id"`
	DoctorID       *string `json:"doctor_id"`
	VisitDate      *string `json:"visit_date"`
	VisitType      *string `json:"visit_type"`
	ChiefComplaint *string `json:"chief_complaint"`
	Subjective     *string `json:"subjective"`
	Objective      *string `json:"objective"`
	Assessment     *string `json:"assessment"`
	Plan           *string `json:"plan"`
	SurgeryNotes   *string `json:"surgery_notes"`
	Diagnosis      *string `json:"diagnosis"`
	Treatment      *string `json:"treatment"`
	Prescription   *string `json:"prescription"`
	Notes          *string `json:"notes"`
	Status         *string `json:"status"`
}

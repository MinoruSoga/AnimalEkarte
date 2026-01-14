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
	Subjective     string     `json:"subjective" gorm:"type:text"`  // S: 主観的情報
	Objective      string     `json:"objective" gorm:"type:text"`   // O: 客観的情報
	Assessment     string     `json:"assessment" gorm:"type:text"`  // A: 評価
	Plan           string     `json:"plan" gorm:"type:text"`        // P: 計画
	SurgeryNotes   string     `json:"surgery_notes" gorm:"type:text"` // S: 手術・特記事項
	Diagnosis      string     `json:"diagnosis" gorm:"type:text"`
	Treatment      string     `json:"treatment" gorm:"type:text"`
	Prescription   string     `json:"prescription" gorm:"type:text"`
	Notes          string     `json:"notes" gorm:"type:text"`
	Status         string     `json:"status" gorm:"type:varchar(20);default:'作成中'"` // 作成中, 確定済
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Pet   *Pet   `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner *Owner `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
}

// TableName テーブル名を指定
func (MedicalRecord) TableName() string {
	return "medical_records"
}

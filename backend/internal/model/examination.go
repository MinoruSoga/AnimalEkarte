package model

import (
	"time"

	"github.com/google/uuid"
)

// Examination 検査記録モデル
type Examination struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	PetID           uuid.UUID  `json:"pet_id" gorm:"type:uuid;not null"`
	OwnerID         uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	DoctorID        *uuid.UUID `json:"doctor_id" gorm:"type:uuid"`
	MedicalRecordID *uuid.UUID `json:"medical_record_id" gorm:"type:uuid"`
	ExaminationDate time.Time  `json:"examination_date"`
	TestType        string     `json:"test_type" gorm:"type:varchar(100)"`
	Machine         string     `json:"machine" gorm:"type:varchar(100)"`
	Status          string     `json:"status" gorm:"type:varchar(20);default:'依頼中'"` // 依頼中, 検査中, 完了
	ResultSummary   string     `json:"result_summary" gorm:"type:text"`
	Items           string     `json:"items" gorm:"type:json"` // JSON array
	Notes           string     `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Pet           *Pet           `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner         *Owner         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	MedicalRecord *MedicalRecord `json:"medical_record,omitempty" gorm:"foreignKey:MedicalRecordID"`
}

// TableName テーブル名を指定
func (Examination) TableName() string {
	return "examinations"
}

// CreateExaminationRequest 検査作成リクエスト
type CreateExaminationRequest struct {
	PetID           uuid.UUID  `json:"pet_id" binding:"required"`
	OwnerID         uuid.UUID  `json:"owner_id" binding:"required"`
	DoctorID        *uuid.UUID `json:"doctor_id"`
	MedicalRecordID *uuid.UUID `json:"medical_record_id"`
	ExaminationDate time.Time  `json:"examination_date" binding:"required"`
	TestType        string     `json:"test_type" binding:"required"`
	Machine         string     `json:"machine"`
	ResultSummary   string     `json:"result_summary"`
	Items           string     `json:"items"`
	Notes           string     `json:"notes"`
}

// UpdateExaminationRequest 検査更新リクエスト
type UpdateExaminationRequest struct {
	Status          string     `json:"status"`
	ResultSummary   string     `json:"result_summary"`
	Items           string     `json:"items"`
	Notes           string     `json:"notes"`
	ExaminationDate *time.Time `json:"examination_date"`
	TestType        string     `json:"test_type"`
	Machine         string     `json:"machine"`
}

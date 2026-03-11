package model

import (
	"time"

	"github.com/google/uuid"
)

type ExaminationStatus string

const (
	ExaminationStatusPending    ExaminationStatus = "依頼中"
	ExaminationStatusInProgress ExaminationStatus = "検査中"
	ExaminationStatusCompleted  ExaminationStatus = "完了"
)

type ExaminationResultStatus string

const (
	ExaminationResultStatusNormal ExaminationResultStatus = "normal"
	ExaminationResultStatusHigh   ExaminationResultStatus = "high"
	ExaminationResultStatusLow    ExaminationResultStatus = "low"
)

type ExaminationRecord struct {
	ID                uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID   *uuid.UUID        `gorm:"type:uuid"                                      json:"medical_record_id,omitempty"`
	PetID             *uuid.UUID        `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	Date              time.Time         `gorm:"type:date;not null"                             json:"date"`
	OwnerName         string            `gorm:"not null;default:''"                            json:"owner_name"`
	PetName           string            `gorm:"not null;default:''"                            json:"pet_name"`
	ExaminationTypeID uuid.UUID         `gorm:"type:uuid;not null"                             json:"examination_type_id"`
	DoctorID          *uuid.UUID        `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	Status            ExaminationStatus `gorm:"type:examination_status;default:'依頼中'"          json:"status"`
	ResultSummary     string            `gorm:"default:''"                                     json:"result_summary"`
	Machine           string            `gorm:"default:''"                                     json:"machine"`
	CreatedAt         time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord   *MedicalRecord          `gorm:"foreignKey:MedicalRecordID"     json:"medical_record,omitempty"`
	Pet             *Pet                    `gorm:"foreignKey:PetID"               json:"pet,omitempty"`
	ExaminationType ExaminationType         `gorm:"foreignKey:ExaminationTypeID"   json:"examination_type,omitempty"`
	Doctor          *Staff                  `gorm:"foreignKey:DoctorID"            json:"doctor,omitempty"`
	Items           []ExaminationRecordItem `gorm:"foreignKey:ExaminationRecordID" json:"items,omitempty"`
}

func (ExaminationRecord) TableName() string { return "examination_records" }

type ExaminationRecordItem struct {
	ID                  uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExaminationRecordID uuid.UUID               `gorm:"type:uuid;not null"                             json:"examination_record_id"`
	Name                string                  `gorm:"not null;default:''"                            json:"name"`
	InspectionValue     string                  `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue         string                  `gorm:"default:''"                                     json:"normal_value"`
	Result              string                  `gorm:"default:''"                                     json:"result"`
	Unit                string                  `gorm:"default:''"                                     json:"unit"`
	Ref                 string                  `gorm:"default:''"                                     json:"ref"`
	Status              ExaminationResultStatus `gorm:"type:examination_result_status;default:'normal'" json:"status"`
	SortOrder           int                     `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt           time.Time               `gorm:"autoCreateTime"                                 json:"created_at"`
}

func (ExaminationRecordItem) TableName() string { return "examination_record_items" }

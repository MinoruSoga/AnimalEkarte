package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ExaminationRevisionKind is the sole discriminator between an editable working
// snapshot and an immutable official snapshot.
type ExaminationRevisionKind string

const (
	ExaminationRevisionKindWorking  ExaminationRevisionKind = "working"
	ExaminationRevisionKindOfficial ExaminationRevisionKind = "official"
)

// ExaminationDisplaySnapshot contains the human-readable identity/master labels that
// must remain stable even after mutable source rows are renamed or transferred.
type ExaminationDisplaySnapshot struct {
	MedicalRecordNo        string `json:"medical_record_no"`
	PetName                string `json:"pet_name"`
	MedicalRecordOwnerName string `json:"medical_record_owner_name"`
	PetOwnerName           string `json:"pet_owner_name"`
	SpeciesName            string `json:"species_name"`
	ExamTypeName           string `json:"exam_type_name"`
	DoctorName             string `json:"doctor_name"`
}

// ExaminationRevision is an append-only snapshot of one examination parent version.
// Mutable identity and master rows are intentionally not modeled as GORM relations.
type ExaminationRevision struct {
	ID                   uint64                  `gorm:"primaryKey;autoIncrement"             json:"id"`
	ClinicID             uint64                  `gorm:"not null"                             json:"clinic_id"`
	ExaminationID        uint64                  `gorm:"not null"                             json:"examination_id"`
	Version              uint64                  `gorm:"not null"                             json:"version"`
	Kind                 ExaminationRevisionKind `gorm:"type:text;not null"                   json:"kind"`
	Status               ExaminationStatus       `gorm:"type:exam_status;not null"            json:"status"`
	MedicalRecordID      *uint64                 `                                              json:"medical_record_id,omitempty"`
	PetID                *uint64                 `                                              json:"pet_id,omitempty"`
	MedicalRecordOwnerID *uint64                 `                                              json:"medical_record_owner_id,omitempty"`
	PetOwnerID           *uint64                 `                                              json:"pet_owner_id,omitempty"`
	AnimalSpeciesID      *uint64                 `                                              json:"animal_species_id,omitempty"`
	ExamTypeID           uint64                  `gorm:"not null"                             json:"exam_type_id"`
	DoctorID             *uint64                 `                                              json:"doctor_id,omitempty"`
	JobID                *uuid.UUID              `gorm:"type:uuid"                            json:"job_id,omitempty"`
	ActorID              uint64                  `gorm:"not null"                             json:"actor_id"`
	Date                 time.Time               `gorm:"type:date;not null"                   json:"date"`
	ResultSummary        string                  `gorm:"not null;default:''"                  json:"result_summary"`
	Machine              string                  `gorm:"not null;default:''"                  json:"machine"`
	DisplaySnapshot      json.RawMessage         `gorm:"type:jsonb;not null"                  json:"display_snapshot"`
	SchemaVersion        int16                   `gorm:"type:smallint;not null;default:1"     json:"schema_version"`
	ChangeReason         *string                 `gorm:"type:text"                            json:"change_reason"`
	CreatedAt            time.Time               `gorm:"not null;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
}

func (ExaminationRevision) TableName() string { return "examination_revisions" }

// ExaminationRevisionItem is an immutable item snapshot owned by one revision triple.
type ExaminationRevisionItem struct {
	ID              uint64                  `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64                  `gorm:"not null"                                       json:"clinic_id"`
	ExaminationID   uint64                  `gorm:"not null"                                       json:"examination_id"`
	Version         uint64                  `gorm:"not null"                                       json:"version"`
	ExamTypeFieldID *uint64                 `gorm:"column:exam_type_field_id"                      json:"exam_type_field_id,omitempty"`
	Name            string                  `gorm:"not null;default:''"                            json:"name"`
	InspectionValue string                  `gorm:"not null;default:''"                            json:"inspection_value"`
	NormalValue     string                  `gorm:"not null;default:''"                            json:"normal_value"`
	Result          string                  `gorm:"not null;default:''"                            json:"result"`
	Unit            string                  `gorm:"not null;default:''"                            json:"unit"`
	ReferenceValue  string                  `gorm:"not null;default:''"                            json:"reference_value"`
	RefMin          *float64                `gorm:"column:ref_min;type:decimal(10,4)"              json:"ref_min,omitempty"`
	RefMax          *float64                `gorm:"column:ref_max;type:decimal(10,4)"              json:"ref_max,omitempty"`
	QualitativeMin  *string                 `gorm:"column:qualitative_min"                         json:"qualitative_min,omitempty"`
	QualitativeMax  *string                 `gorm:"column:qualitative_max"                         json:"qualitative_max,omitempty"`
	IsAssessed      bool                    `gorm:"column:is_assessed;not null"                     json:"is_assessed"`
	IsAbnormal      bool                    `gorm:"column:is_abnormal;not null"                     json:"is_abnormal"`
	Status          ExaminationResultStatus `gorm:"type:exam_result_status;not null"                json:"status"`
	SortOrder       int                     `gorm:"type:integer;not null;default:0"                  json:"sort_order"`
	CreatedAt       time.Time               `gorm:"not null;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
}

func (ExaminationRevisionItem) TableName() string { return "examination_revision_items" }

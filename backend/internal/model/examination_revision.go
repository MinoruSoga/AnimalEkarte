package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
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
	ID                         uint64            `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID                   uint64            `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	ExaminationID              uint64            `gorm:"column:examination_id;type:bigint;not null" json:"examination_id"`
	Version                    uint64            `gorm:"column:version;type:bigint;not null" json:"version"`
	Kind                       string            `gorm:"column:kind;type:text;not null" json:"kind"`
	Status                     ExaminationStatus `gorm:"column:status;type:exam_status;not null" json:"status"`
	MedicalRecordID            *uint64           `gorm:"column:medical_record_id;type:bigint" json:"medical_record_id,omitempty"`
	PetID                      *uint64           `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	MedicalRecordOwnerID       *uint64           `gorm:"column:medical_record_owner_id;type:bigint" json:"medical_record_owner_id,omitempty"`
	PetOwnerID                 *uint64           `gorm:"column:pet_owner_id;type:bigint" json:"pet_owner_id,omitempty"`
	AnimalSpeciesID            *uint64           `gorm:"column:animal_species_id;type:bigint" json:"animal_species_id,omitempty"`
	ExamTypeID                 uint64            `gorm:"column:exam_type_id;type:bigint;not null" json:"exam_type_id"`
	DoctorID                   *uint64           `gorm:"column:doctor_id;type:bigint" json:"doctor_id,omitempty"`
	JobID                      *uuid.UUID        `gorm:"column:job_id;type:uuid" json:"job_id,omitempty"`
	ActorID                    uint64            `gorm:"column:actor_id;type:bigint;not null" json:"actor_id"`
	Date                       *time.Time        `gorm:"column:date;type:date" json:"date"`
	ResultSummary              string            `gorm:"column:result_summary;type:text;not null;default:''" json:"result_summary"`
	Machine                    string            `gorm:"column:machine;type:text;not null;default:''" json:"machine"`
	DisplaySnapshot            json.RawMessage   `gorm:"column:display_snapshot;type:jsonb;not null" json:"display_snapshot"`
	SchemaVersion              int16             `gorm:"column:schema_version;type:smallint;not null;default:1" json:"schema_version"`
	ChangeReason               *string           `gorm:"column:change_reason;type:text" json:"change_reason"`
	CreatedAt                  time.Time         `gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
	PerformedDatePrecision     *string           `gorm:"column:performed_date_precision;type:text" json:"-"`
	MeasuredAt                 *time.Time        `gorm:"column:measured_at;type:timestamptz" json:"-"`
	ResultSourceKind           *string           `gorm:"column:result_source_kind;type:text" json:"-"`
	ResultSourceDetails        *string           `gorm:"column:result_source_details;type:text" json:"-"`
	PreviousOfficialRevisionID *uint64           `gorm:"column:previous_official_revision_id;type:bigint" json:"-"`
}

func (ExaminationRevision) TableName() string { return "examination_revisions" }

// ExaminationRevisionItem is an immutable item snapshot owned by one revision triple.
type ExaminationRevisionItem struct {
	ID              uint64                  `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64                  `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	ExaminationID   uint64                  `gorm:"column:examination_id;type:bigint;not null" json:"examination_id"`
	Version         uint64                  `gorm:"column:version;type:bigint;not null" json:"version"`
	ExamTypeFieldID *uint64                 `gorm:"column:exam_type_field_id;type:bigint" json:"exam_type_field_id,omitempty"`
	Name            string                  `gorm:"column:name;type:text;not null;default:''" json:"name"`
	InspectionValue string                  `gorm:"column:inspection_value;type:text;not null;default:''" json:"inspection_value"`
	NormalValue     string                  `gorm:"column:normal_value;type:text;not null;default:''" json:"normal_value"`
	Result          string                  `gorm:"column:result;type:text;not null;default:''" json:"result"`
	Unit            string                  `gorm:"column:unit;type:text;not null;default:''" json:"unit"`
	ReferenceValue  string                  `gorm:"column:reference_value;type:text;not null;default:''" json:"reference_value"`
	RefMin          *ExactNumeric           `gorm:"column:ref_min;type:decimal(10,4)" json:"ref_min,omitempty"`
	RefMax          *ExactNumeric           `gorm:"column:ref_max;type:decimal(10,4)" json:"ref_max,omitempty"`
	QualitativeMin  *string                 `gorm:"column:qualitative_min;type:text" json:"qualitative_min,omitempty"`
	QualitativeMax  *string                 `gorm:"column:qualitative_max;type:text" json:"qualitative_max,omitempty"`
	IsAssessed      bool                    `gorm:"column:is_assessed;type:boolean;not null" json:"is_assessed"`
	IsAbnormal      bool                    `gorm:"column:is_abnormal;type:boolean;not null" json:"is_abnormal"`
	Status          ExaminationResultStatus `gorm:"column:status;type:exam_result_status;not null" json:"status"`
	SortOrder       int                     `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	CreatedAt       time.Time               `gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
}

func (ExaminationRevisionItem) TableName() string { return "examination_revision_items" }

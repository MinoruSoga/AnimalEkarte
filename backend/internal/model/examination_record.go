package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExaminationStatus string

const (
	ExaminationStatusPending       ExaminationStatus = "pending"
	ExaminationStatusInProgress    ExaminationStatus = "in_progress"
	ExaminationStatusResultEntered ExaminationStatus = "result_entered"
	ExaminationStatusCompleted     ExaminationStatus = "completed"
	ExaminationStatusConfirmed     ExaminationStatus = "confirmed"
)

type ExaminationResultStatus string

const (
	ExaminationResultStatusNormal ExaminationResultStatus = "normal"
	ExaminationResultStatusHigh   ExaminationResultStatus = "high"
	ExaminationResultStatusLow    ExaminationResultStatus = "low"
)

type Examination struct {
	ID              uint64  `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64  `gorm:"not null"                                       json:"clinic_id"`
	MedicalRecordID *uint64 `                                                      json:"medical_record_id,omitempty"`
	PetID           *uint64 `                                                      json:"pet_id,omitempty"`
	ExamTypeID      uint64  `gorm:"not null"                                       json:"exam_type_id"`
	DoctorID        *uint64 `                                                      json:"doctor_id,omitempty"`
	// JobID は lab_import_jobs.id への nullable FK。手動作成の exam は NULL。
	// TASK-032: composite (clinic_id, job_id) ON DELETE RESTRICT（SET NULL から置換）。
	JobID         *uuid.UUID        `gorm:"type:uuid"                                      json:"job_id,omitempty"`
	Date          time.Time         `gorm:"type:date;not null"                             json:"date"`
	ResultSummary string            `gorm:"default:''"                                     json:"result_summary"`
	Machine       string            `gorm:"default:''"                                     json:"machine"`
	Status        ExaminationStatus `gorm:"type:exam_status;default:'pending'"                json:"status"`
	// CurrentRevisionVersion points at the append-only working/official revision selected by
	// the examination workflow. A nil pointer denotes a legacy or not-yet-revisioned row.
	CurrentRevisionVersion *uint64        `gorm:"column:current_revision_version"                    json:"current_revision_version,omitempty"`
	DeletedAt              gorm.DeletedAt `                                                      json:"-"`
	CreatedAt              time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt              time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord   *MedicalRecord   `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet             *Pet             `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	ExaminationType *ExaminationType `gorm:"foreignKey:ExamTypeID"      json:"exam_type,omitempty"`
	Doctor          *Staff           `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
	Items           []ExamResult     `gorm:"foreignKey:ExamID"          json:"items,omitempty"`
}

func (Examination) TableName() string { return "exams" }

type ExamResult struct {
	ID              uint64                  `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamID          uint64                  `gorm:"not null"                                       json:"exam_id"`
	ExamTypeItemID  *uint64                 `gorm:"column:exam_type_field_id"                      json:"exam_type_field_id,omitempty"`
	Name            string                  `gorm:"not null;default:''"                            json:"name"`
	InspectionValue string                  `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string                  `gorm:"default:''"                                     json:"normal_value"`
	Result          string                  `gorm:"default:''"                                     json:"result"`
	Unit            string                  `gorm:"default:''"                                     json:"unit"`
	ReferenceValue  string                  `gorm:"column:reference_value;default:''"              json:"reference_value"`
	RefMin          *float64                `gorm:"column:ref_min;type:decimal(10,4)"              json:"ref_min,omitempty"`
	RefMax          *float64                `gorm:"column:ref_max;type:decimal(10,4)"              json:"ref_max,omitempty"`
	QualitativeMin  *string                 `gorm:"column:qualitative_min"                         json:"qualitative_min,omitempty"`
	QualitativeMax  *string                 `gorm:"column:qualitative_max"                         json:"qualitative_max,omitempty"`
	IsAbnormal      bool                    `gorm:"column:is_abnormal;default:false"               json:"is_abnormal"`
	Status          ExaminationResultStatus `gorm:"type:exam_result_status;default:'normal'"        json:"status"`
	SortOrder       int                     `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt       time.Time               `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time               `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	ExamTypeField *ExamTypeField `gorm:"foreignKey:ExamTypeItemID" json:"exam_type_field,omitempty"`
}

func (ExamResult) TableName() string { return "exam_results" }

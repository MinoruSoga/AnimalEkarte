package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
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
	ID                     uint64             `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	MedicalRecordID        *uint64            `gorm:"column:medical_record_id;type:bigint" json:"medical_record_id,omitempty"`
	ClinicID               uint64             `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	PetID                  *uint64            `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	Date                   *time.Time         `gorm:"column:date;type:date" json:"date"`
	ExamTypeID             uint64             `gorm:"column:exam_type_id;type:bigint;not null" json:"exam_type_id"`
	DoctorID               *uint64            `gorm:"column:doctor_id;type:bigint" json:"doctor_id,omitempty"`
	Status                 *ExaminationStatus `gorm:"column:status;type:exam_status;default:'pending'" json:"status"`
	ResultSummary          string             `gorm:"column:result_summary;type:text;not null;default:''" json:"result_summary"`
	Machine                string             `gorm:"column:machine;type:text;not null;default:''" json:"machine"`
	CreatedAt              time.Time          `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time          `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt     `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	JobID                  *uuid.UUID         `gorm:"column:job_id;type:uuid" json:"job_id,omitempty"`
	CurrentRevisionVersion *uint64            `gorm:"column:current_revision_version;type:bigint" json:"current_revision_version,omitempty"`
	PerformedDatePrecision *string            `gorm:"column:performed_date_precision;type:text" json:"-"`
	MedicalRecord          *MedicalRecord     `gorm:"foreignKey:MedicalRecordID;references:ID" json:"medical_record,omitempty"`
	Pet                    *Pet               `gorm:"foreignKey:PetID;references:ID"           json:"pet,omitempty"`
	ExaminationType        *ExaminationType   `gorm:"foreignKey:ExamTypeID;references:ID"      json:"exam_type,omitempty"`
	Doctor                 *Staff             `gorm:"foreignKey:DoctorID;references:ID"        json:"doctor,omitempty"`
	Items                  []ExamResult       `gorm:"foreignKey:ExamID;references:ID"          json:"items,omitempty"`
}

func (Examination) TableName() string { return "exams" }

type ExamResult struct {
	ID              uint64                   `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ExamID          uint64                   `gorm:"column:exam_id;type:bigint;not null" json:"exam_id"`
	ExamTypeItemID  *uint64                  `gorm:"column:exam_type_field_id;type:bigint" json:"exam_type_field_id,omitempty"`
	Name            string                   `gorm:"column:name;type:text;not null;default:''" json:"name"`
	InspectionValue string                   `gorm:"column:inspection_value;type:text;not null;default:''" json:"inspection_value"`
	NormalValue     string                   `gorm:"column:normal_value;type:text;not null;default:''" json:"normal_value"`
	Result          string                   `gorm:"column:result;type:text;not null;default:''" json:"result"`
	Unit            string                   `gorm:"column:unit;type:text;not null;default:''" json:"unit"`
	ReferenceValue  string                   `gorm:"column:reference_value;type:text;not null;default:''" json:"reference_value"`
	RefMin          *ExactNumeric            `gorm:"column:ref_min;type:decimal(10,4)" json:"ref_min,omitempty"`
	RefMax          *ExactNumeric            `gorm:"column:ref_max;type:decimal(10,4)" json:"ref_max,omitempty"`
	IsAbnormal      *bool                    `gorm:"column:is_abnormal;type:boolean;default:false" json:"is_abnormal"`
	Status          *ExaminationResultStatus `gorm:"column:status;type:exam_result_status;default:'normal'" json:"status"`
	SortOrder       *int                     `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt       time.Time                `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time                `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	QualitativeMin  *string                  `gorm:"column:qualitative_min;type:text" json:"qualitative_min,omitempty"`
	QualitativeMax  *string                  `gorm:"column:qualitative_max;type:text" json:"qualitative_max,omitempty"`
	ExamTypeField   *ExamTypeField           `gorm:"foreignKey:ExamTypeItemID;references:ID" json:"exam_type_field,omitempty"`
}

func (ExamResult) TableName() string { return "exam_results" }

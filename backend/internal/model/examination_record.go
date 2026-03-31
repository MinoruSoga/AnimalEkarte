package model

import (
	"time"

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
	ID              uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64            `gorm:"not null"                                       json:"clinic_id"`
	MedicalRecordID *uint64           `                                                      json:"medical_record_id,omitempty"`
	PetID           *uint64           `                                                      json:"pet_id,omitempty"`
	ExamTypeID      uint64            `gorm:"not null"                                       json:"exam_type_id"`
	DoctorID        *uint64           `                                                      json:"doctor_id,omitempty"`
	Date            time.Time         `gorm:"type:date;not null"                             json:"date"`
	ResultSummary   string            `gorm:"default:''"                                     json:"result_summary"`
	Machine         string            `gorm:"default:''"                                     json:"machine"`
	Status          ExaminationStatus `gorm:"type:examination_status;default:'pending'"        json:"status"`
	DeletedAt       gorm.DeletedAt    `                                                      json:"-" swaggerignore:"true"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord   *MedicalRecord    `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet             *Pet              `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	ExaminationType *ExaminationType  `gorm:"foreignKey:ExamTypeID"      json:"exam_type,omitempty"`
	Doctor          *Staff            `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
	Items           []ExaminationItem `gorm:"foreignKey:ExamID"          json:"items,omitempty"`
}

func (Examination) TableName() string { return "exams" }

type ExaminationItem struct {
	ID              uint64                  `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamID          uint64                  `gorm:"not null"                                       json:"exam_id"`
	ExamTypeItemID  *uint64                 `                                                      json:"exam_type_item_id,omitempty"`
	Name            string                  `gorm:"not null;default:''"                            json:"name"`
	InspectionValue string                  `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string                  `gorm:"default:''"                                     json:"normal_value"`
	Result          string                  `gorm:"default:''"                                     json:"result"`
	Unit            string                  `gorm:"default:''"                                     json:"unit"`
	Ref             string                  `gorm:"default:''"                                     json:"ref"`
	Status          ExaminationResultStatus `gorm:"type:examination_result_status;default:'normal'" json:"status"`
	SortOrder       int                     `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt       time.Time               `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time               `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	ExaminationTypeItem *ExaminationTypeItem `gorm:"foreignKey:ExamTypeItemID" json:"exam_type_item,omitempty"`
}

func (ExaminationItem) TableName() string { return "exam_items" }

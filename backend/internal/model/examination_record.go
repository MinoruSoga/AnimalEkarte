package model

import (
	"time"

	"gorm.io/gorm"
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
type Exam struct {
	ID              uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64            `gorm:"not null"                                       json:"medical_record_id"`
	PetID           *uint64           `                                                      json:"pet_id,omitempty"`
	ExamTypeID      uint64            `gorm:"not null"                                       json:"exam_type_id"`
	DoctorID        *uint64           `                                                      json:"doctor_id,omitempty"`
	Date            time.Time         `gorm:"type:date;not null"                             json:"date"`
	ResultSummary   string            `gorm:"default:''"                                     json:"result_summary"`
	Machine         string            `gorm:"default:''"                                     json:"machine"`
	Status          ExaminationStatus `gorm:"type:examination_status;default:'依頼中'"          json:"status"`
	DeletedAt       gorm.DeletedAt    `                                                      json:"deleted_at" swaggerignore:"true"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	ExamType      *ExamType      `gorm:"foreignKey:ExamTypeID"      json:"exam_type,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
	Items         []ExamItem     `gorm:"foreignKey:ExamID"          json:"items,omitempty"`
}

func (Exam) TableName() string { return "exams" }
type ExamItem struct {
	ID              uint64                  `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamID          uint64                  `gorm:"not null"                                       json:"exam_id"`
	Name            string                  `gorm:"not null;default:''"                            json:"name"`
	InspectionValue string                  `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string                  `gorm:"default:''"                                     json:"normal_value"`
	Result          string                  `gorm:"default:''"                                     json:"result"`
	Unit            string                  `gorm:"default:''"                                     json:"unit"`
	Ref             string                  `gorm:"default:''"                                     json:"ref"`
	Status          ExaminationResultStatus `gorm:"type:examination_result_status;default:'normal'" json:"status"`
	SortOrder       int                     `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time               `gorm:"autoCreateTime"                                 json:"created_at"`
}

func (ExamItem) TableName() string { return "exam_items" }

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MedicalRecordStatus string

const (
	MedicalRecordStatusDraft     MedicalRecordStatus = "作成中"
	MedicalRecordStatusFinalized MedicalRecordStatus = "確定済"
)

type MedicalRecord struct {
	ID                       uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID                 uuid.UUID           `gorm:"type:uuid;not null"                             json:"clinic_id"`
	RecordNo                 string              `gorm:"not null"                                       json:"record_no"`
	Date                     time.Time           `gorm:"type:date;not null"                             json:"date"`
	OwnerID                  *uuid.UUID          `gorm:"type:uuid"                                      json:"owner_id,omitempty"`
	PetID                    *uuid.UUID          `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	DoctorID                 *uuid.UUID          `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	ReservationAppointmentID *uuid.UUID          `gorm:"type:uuid"                                      json:"reservation_appointment_id,omitempty"`
	Status                   MedicalRecordStatus `gorm:"type:medical_record_status;default:'作成中'"       json:"status"`
	CreatedAt                time.Time           `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt                time.Time           `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt                gorm.DeletedAt      `                                                      json:"deleted_at"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"         json:"doctor,omitempty"`
	ClinicalPlan  *ClinicalPlan  `gorm:"foreignKey:MedicalRecordID"  json:"clinical_plan,omitempty"`
	Inquiry       *Inquiry       `gorm:"foreignKey:MedicalRecordID"  json:"inquiry,omitempty"`
	Treatments    []Treatment    `gorm:"foreignKey:MedicalRecordID"  json:"treatments,omitempty"`
	Vitals        []Vital        `gorm:"foreignKey:MedicalRecordID"  json:"vitals,omitempty"`
	Exams         []Exam         `gorm:"foreignKey:MedicalRecordID"  json:"exams,omitempty"`
	Vaccinations  []Vaccination  `gorm:"foreignKey:MedicalRecordID"  json:"vaccinations,omitempty"`
	Checkups      []Checkup      `gorm:"foreignKey:MedicalRecordID"  json:"checkups,omitempty"`
	Estimates     []Estimate     `gorm:"foreignKey:MedicalRecordID"  json:"estimates,omitempty"`
	BillingReview *BillingReview `gorm:"foreignKey:MedicalRecordID"  json:"billing_review,omitempty"`
}

func (MedicalRecord) TableName() string { return "medical_records" }

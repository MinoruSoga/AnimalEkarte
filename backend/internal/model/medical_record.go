package model

import (
	"time"

	"gorm.io/gorm"
)

// @name MedicalRecordStatus
type MedicalRecordStatus string

const (
	MedicalRecordStatusDraft     MedicalRecordStatus = "draft"
	MedicalRecordStatusFinalized MedicalRecordStatus = "finalized"
)

// @name MedicalRecord
type MedicalRecord struct {
	ID                       uint64              `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID                 uint64              `gorm:"not null"                                       json:"clinic_id"`
	RecordNo                 string              `gorm:"not null"                                       json:"record_no"`
	Date                     time.Time           `gorm:"type:date;not null"                             json:"date"`
	OwnerID                  *uint64             `                                                      json:"owner_id,omitempty"`
	PetID                    *uint64             `                                                      json:"pet_id,omitempty"`
	DoctorID                 *uint64             `                                                      json:"doctor_id,omitempty"`
	ReservationAppointmentID *uint64             `                                                      json:"reservation_appointment_id,omitempty"`
	Status                   MedicalRecordStatus `gorm:"type:medical_record_status;default:'draft'"      json:"status"`
	CreatedAt                time.Time           `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt                time.Time           `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt                gorm.DeletedAt      `                                                      json:"deleted_at" swaggerignore:"true"`

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

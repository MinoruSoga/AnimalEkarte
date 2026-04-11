package model

import (
	"time"

	"gorm.io/gorm"
)

type MedicalRecordStatus string

const (
	MedicalRecordStatusDraft     MedicalRecordStatus = "draft"
	MedicalRecordStatusFinalized MedicalRecordStatus = "finalized"
)

type MedicalRecord struct {
	ID                       uint64              `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID                 uint64              `gorm:"not null"                                       json:"clinic_id"`
	RecordNo                 string              `gorm:"not null"                                       json:"record_no"`
	Date                     time.Time           `gorm:"type:date;not null"                             json:"date"`
	OwnerID                  *uint64             `                                                      json:"owner_id,omitempty"`
	PetID                    *uint64             `                                                      json:"pet_id,omitempty"`
	DoctorID                 *uint64             `                                                      json:"doctor_id,omitempty"`
	ReservationAppointmentID *uint64             `gorm:"column:appointment_id"                          json:"appointment_id,omitempty"`
	Status                   MedicalRecordStatus `gorm:"type:medical_record_status;default:'draft'"      json:"status"`
	Version                  int                 `gorm:"default:1"                                       json:"version"`
	CreatedAt                time.Time           `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt                time.Time           `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt                gorm.DeletedAt      `                                                      json:"-"`
	VisitCount               int64               `gorm:"-"                                              json:"visit_count"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"         json:"doctor,omitempty"`
	ClinicalPlan  *ClinicalPlan  `gorm:"foreignKey:MedicalRecordID"  json:"clinical_plan,omitempty"`
	Inquiry       *Inquiry       `gorm:"foreignKey:MedicalRecordID"  json:"inquiry,omitempty"`
	Treatments    []Treatment    `gorm:"foreignKey:MedicalRecordID"  json:"treatments,omitempty"`
	Vitals        []VitalRecord  `gorm:"foreignKey:MedicalRecordID"  json:"vitals,omitempty"`
	Exams         []Examination  `gorm:"foreignKey:MedicalRecordID"  json:"exams,omitempty"`
	Vaccinations  []Vaccination  `gorm:"foreignKey:MedicalRecordID"  json:"vaccinations,omitempty"`
	Checkups      []Checkup      `gorm:"foreignKey:MedicalRecordID"  json:"checkups,omitempty"`
	Estimates     []Estimate     `gorm:"foreignKey:MedicalRecordID"  json:"estimates,omitempty"`
	BillingReview *BillingReview `gorm:"foreignKey:MedicalRecordID"  json:"billing_review,omitempty"`
	Billing       *Billing       `gorm:"foreignKey:MedicalRecordID"  json:"billing,omitempty"`
}

func (MedicalRecord) TableName() string { return "medical_records" }

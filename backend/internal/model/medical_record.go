package model

import (
	"gorm.io/gorm"
	"time"
)

type MedicalRecordStatus string

const (
	MedicalRecordStatusDraft     MedicalRecordStatus = "draft"
	MedicalRecordStatusFinalized MedicalRecordStatus = "finalized"
)

type MedicalRecord struct {
	ID                          uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID                    uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	RecordNo                    string         `gorm:"column:record_no;type:text;not null" json:"record_no"`
	Date                        time.Time      `gorm:"column:date;type:date;not null" json:"date"`
	OwnerID                     *uint64        `gorm:"column:owner_id;type:bigint" json:"owner_id,omitempty"`
	PetID                       *uint64        `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	DoctorID                    *uint64        `gorm:"column:doctor_id;type:bigint" json:"doctor_id,omitempty"`
	AppointmentID               *uint64        `gorm:"column:appointment_id;type:bigint" json:"appointment_id,omitempty"`
	Version                     int            `gorm:"column:version;type:integer;not null;default:1" json:"version"`
	EnteredBy                   *uint64        `gorm:"column:entered_by;type:bigint" json:"entered_by,omitempty"`
	RecommendationReason        *string        `gorm:"column:recommendation_reason;type:varchar(100)" json:"recommendation_reason,omitempty"`
	VisitType                   *VisitType     `gorm:"column:visit_type;type:visit_type" json:"visit_type,omitempty"`
	CreatedAt                   time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt                   time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt                   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	RevisitIntakeTargetID       *uint64        `gorm:"column:revisit_intake_target_id;type:bigint" json:"-"`
	ConversationSummary         *string        `gorm:"column:conversation_summary;type:text" json:"-"`
	ConversationSummaryRevision *int64         `gorm:"column:conversation_summary_revision;type:bigint" json:"-"`
	ConversationSummarySource   *string        `gorm:"column:conversation_summary_source;type:text" json:"-"`
	ClinicalSetGeneration       *int64         `gorm:"column:clinical_set_generation;type:bigint" json:"-"`
	ClinicalSetDigest           *[]byte        `gorm:"column:clinical_set_digest;type:bytea" json:"-"`
	VisitCount                  int64          `gorm:"-"                                              json:"visit_count"`
	Owner                       *Owner         `gorm:"foreignKey:ClinicID,OwnerID;references:ClinicID,ID"          json:"owner,omitempty"`
	Pet                         *Pet           `gorm:"foreignKey:ClinicID,PetID;references:ClinicID,ID"            json:"pet,omitempty"`
	Doctor                      *Staff         `gorm:"foreignKey:DoctorID;references:ID"         json:"doctor,omitempty"`
	EnteredByStaff              *Staff         `gorm:"foreignKey:EnteredBy;references:ID"        json:"entered_by_staff,omitempty" tygo:"-"`
	ClinicalPlan                *ClinicalPlan  `gorm:"foreignKey:MedicalRecordID,ClinicID;references:ID,ClinicID"  json:"clinical_plan,omitempty"`
	Inquiry                     *Inquiry       `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"inquiry,omitempty"`
	Treatments                  []Treatment    `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"treatments,omitempty"`
	Vitals                      []VitalRecord  `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"vitals,omitempty"`
	Exams                       []Examination  `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"exams,omitempty"`
	Vaccinations                []Vaccination  `gorm:"foreignKey:MedicalRecordID,ClinicID;references:ID,ClinicID"  json:"vaccinations,omitempty"`
	Checkups                    []Checkup      `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"checkups,omitempty"`
	Estimates                   []Estimate     `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"estimates,omitempty"`
	Billing                     *Billing       `gorm:"foreignKey:MedicalRecordID;references:ID"  json:"billing,omitempty"`
}

func (MedicalRecord) TableName() string { return "medical_records" }

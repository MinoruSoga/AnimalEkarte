package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReservationStatus string

const (
	ReservationStatusConfirmed      ReservationStatus = "confirmed"
	ReservationStatusPending        ReservationStatus = "pending"
	ReservationStatusCancelled      ReservationStatus = "cancelled"
	ReservationStatusCheckedIn      ReservationStatus = "checked_in"
	ReservationStatusInConsultation ReservationStatus = "in_consultation"
	ReservationStatusAccounting     ReservationStatus = "accounting"
	ReservationStatusCompleted      ReservationStatus = "completed"
)

type VisitType string

const (
	VisitTypeFirst   VisitType = "first"
	VisitTypeRevisit VisitType = "revisit"
)

type ReservationAppointment struct {
	ID            uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID      uuid.UUID         `gorm:"type:uuid;not null"                             json:"clinic_id"`
	StartTime     time.Time         `gorm:"not null"                                       json:"start_time"`
	EndTime       time.Time         `gorm:"not null"                                       json:"end_time"`
	OwnerID       *uuid.UUID        `gorm:"type:uuid"                                      json:"owner_id,omitempty"`
	PetID         *uuid.UUID        `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	VisitType     VisitType         `gorm:"type:visit_type;not null;default:'revisit'"     json:"visit_type"`
	ServiceTypeID uuid.UUID         `gorm:"type:uuid;not null"                             json:"service_type_id"`
	DoctorID      *uuid.UUID        `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	IsDesignated  bool              `gorm:"default:false"                                  json:"is_designated"`
	Status        ReservationStatus `gorm:"type:reservation_status;default:'pending'"      json:"status"`
	Notes         string            `gorm:"default:''"                                     json:"notes"`
	DeletedAt     gorm.DeletedAt    `                                                      json:"deleted_at"`
	CreatedAt     time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner       *Owner       `gorm:"foreignKey:OwnerID"       json:"owner,omitempty"`
	Pet         *Pet         `gorm:"foreignKey:PetID"         json:"pet,omitempty"`
	ServiceType *ServiceType `gorm:"foreignKey:ServiceTypeID" json:"service_type,omitempty"`
	Doctor      *Staff       `gorm:"foreignKey:DoctorID"      json:"doctor,omitempty"`
}

func (ReservationAppointment) TableName() string { return "reservation_appointments" }

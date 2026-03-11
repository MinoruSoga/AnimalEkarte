package model

import (
	"time"

	"github.com/google/uuid"
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
	StartTime     time.Time         `gorm:"not null"                                       json:"start_time"`
	EndTime       time.Time         `gorm:"not null"                                       json:"end_time"`
	OwnerName     string            `gorm:"not null;default:''"                            json:"owner_name"`
	PetName       string            `gorm:"not null;default:''"                            json:"pet_name"`
	PetID         *uuid.UUID        `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	VisitType     VisitType         `gorm:"type:visit_type;not null;default:'revisit'"     json:"visit_type"`
	ServiceTypeID *uuid.UUID        `gorm:"type:uuid"                                      json:"service_type_id,omitempty"`
	DoctorID      uuid.UUID         `gorm:"type:uuid;not null"                             json:"doctor_id"`
	IsDesignated  bool              `gorm:"default:false"                                  json:"is_designated"`
	Status        ReservationStatus `gorm:"type:reservation_status;default:'pending'"      json:"status"`
	Notes         string            `gorm:"default:''"                                     json:"notes"`
	CreatedAt     time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Pet         *Pet         `gorm:"foreignKey:PetID"         json:"pet,omitempty"`
	ServiceType *ServiceType `gorm:"foreignKey:ServiceTypeID" json:"service_type,omitempty"`
	Doctor      Staff        `gorm:"foreignKey:DoctorID"      json:"doctor,omitempty"`
}

func (ReservationAppointment) TableName() string { return "reservation_appointments" }

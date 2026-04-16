package model

import (
	"encoding/json"
	"time"

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

// ReservationSource indicates whether a reservation was created via LINE or manually.
type ReservationSource string

const (
	ReservationSourceManual ReservationSource = "manual"
	ReservationSourceLine   ReservationSource = "line"
)

type Appointment struct {
	ID                uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID          uint64            `gorm:"not null"                                       json:"clinic_id"`
	StartTime         time.Time         `gorm:"not null"                                       json:"start_time"`
	EndTime           time.Time         `gorm:"not null"                                       json:"end_time"`
	OwnerID           *uint64           `                                                      json:"owner_id,omitempty"`
	PetID             *uint64           `                                                      json:"pet_id,omitempty"`
	VisitType         VisitType         `gorm:"type:visit_type;not null;default:'revisit'"     json:"visit_type"`
	ReservationTypeID uint64            `gorm:"not null"                                       json:"reservation_type_id"`
	DoctorID          *uint64           `                                                      json:"doctor_id,omitempty"`
	IsDesignated      bool              `gorm:"default:false"                                  json:"is_designated"`
	Status            ReservationStatus `gorm:"type:reservation_status;default:'pending'"      json:"status"`
	Notes             string            `gorm:"default:''"                                     json:"notes"`
	DeletedAt         gorm.DeletedAt    `                                                      json:"-"`
	CreatedAt         time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// LINE予約用フィールド
	Source           ReservationSource `gorm:"type:reservation_source;not null;default:'manual'" json:"source"`
	CreatedBy        *uint64           `gorm:""                                                  json:"created_by,omitempty"`
	LineCustomerID   *uint64           `                                                         json:"line_customer_id,omitempty"`
	IsStaffDelegated bool              `gorm:"not null;default:false"                            json:"is_staff_delegated"`
	CustomerFields   json.RawMessage   `gorm:"type:jsonb;not null;default:'{}'"                  json:"customer_fields"`

	// Relations
	Owner           *Owner           `gorm:"foreignKey:OwnerID"              json:"owner,omitempty"`
	Pet             *Pet             `gorm:"foreignKey:PetID"                json:"pet,omitempty"`
	ReservationType *ReservationType `gorm:"foreignKey:ReservationTypeID"    json:"reservation_type,omitempty"`
	Doctor          *Staff           `gorm:"foreignKey:DoctorID"             json:"doctor,omitempty"`
	CreatedByStaff  *Staff           `gorm:"foreignKey:CreatedBy"            json:"created_by_staff,omitempty" tygo:"-"`
	LineCustomer    *LineCustomer    `gorm:"foreignKey:LineCustomerID"       json:"line_customer,omitempty"`
}

func (Appointment) TableName() string { return "appointments" }

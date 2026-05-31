package model

import "time"

// StaffReservationCapability records which reservation types a staff member can handle.
type StaffReservationCapability struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID          uint64    `gorm:"not null"                 json:"clinic_id"`
	StaffID           uint64    `gorm:"not null"                 json:"staff_id"`
	ReservationTypeID uint64    `gorm:"not null"                 json:"reservation_type_id"`
	CreatedAt         time.Time `gorm:"autoCreateTime"           json:"created_at"`

	// Relations
	Clinic          *Clinic          `gorm:"foreignKey:ClinicID"          json:"clinic,omitempty"`
	Staff           *Staff           `gorm:"foreignKey:StaffID"           json:"staff,omitempty"`
	ReservationType *ReservationType `gorm:"foreignKey:ReservationTypeID" json:"reservation_type,omitempty"`
}

func (StaffReservationCapability) TableName() string { return "staff_reservation_capabilities" }
